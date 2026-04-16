package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"promptvault/internal/infrastructure/config"
	repo "promptvault/internal/interface/repository"
	iservice "promptvault/internal/interface/service"
	"promptvault/internal/middleware/ratelimit"
	"promptvault/internal/models"
)

type Service struct {
	users           repo.UserRepository
	linkedAccounts  repo.LinkedAccountRepository
	verifications   repo.VerificationRepository
	email           iservice.EmailSender
	secret          []byte
	frontendURL     string
	accessDuration  time.Duration
	refreshDuration time.Duration
	backgroundWg    sync.WaitGroup
	forgotLimiter   *ratelimit.Limiter[string]
}

func NewService(cfg *config.Config, users repo.UserRepository, linkedAccounts repo.LinkedAccountRepository, verifications repo.VerificationRepository, emailSvc iservice.EmailSender) *Service {
	accessDur, err := time.ParseDuration(cfg.JWT.AccessDuration)
	if err != nil || accessDur == 0 {
		if err != nil {
			slog.Warn("invalid JWT access duration, using default", "configured", cfg.JWT.AccessDuration, "default", DefaultAccessDuration, "error", err)
		}
		accessDur = DefaultAccessDuration
	}

	refreshDur, err := time.ParseDuration(cfg.JWT.RefreshDuration)
	if err != nil || refreshDur == 0 {
		if err != nil {
			slog.Warn("invalid JWT refresh duration, using default", "configured", cfg.JWT.RefreshDuration, "default", DefaultRefreshDuration, "error", err)
		}
		refreshDur = DefaultRefreshDuration
	}

	return &Service{
		users:           users,
		linkedAccounts:  linkedAccounts,
		verifications:   verifications,
		email:           emailSvc,
		secret:          []byte(cfg.JWT.Secret),
		frontendURL:     cfg.Server.FrontendURL,
		accessDuration:  accessDur,
		refreshDuration: refreshDur,
		forgotLimiter:   ratelimit.NewLimiterWithWindow[string](3, 15*time.Minute, ratelimit.StringHash),
	}
}

// runBackground запускает функцию в горутине с отслеживанием через WaitGroup.
func (s *Service) runBackground(fn func()) {
	s.backgroundWg.Add(1)
	go func() {
		defer s.backgroundWg.Done()
		fn()
	}()
}

// WaitBackground ожидает завершения фоновых задач с таймаутом.
func (s *Service) WaitBackground(timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		s.backgroundWg.Wait()
		close(done)
	}()
	select {
	case <-done:
		slog.Info("background tasks completed")
	case <-time.After(timeout):
		slog.Warn("background tasks timeout", "timeout", timeout)
	}
}

func (s *Service) getUser(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *Service) Register(ctx context.Context, userEmail, password, name, username, referredBy string) (*models.User, error) {
	existing, err := s.users.GetByEmail(ctx, userEmail)
	if err == nil {
		accounts, _ := s.linkedAccounts.GetByUserID(ctx, existing.ID)
		if len(accounts) > 0 {
			return nil, &EmailTakenError{Provider: providerName(accounts[0].Provider)}
		}
		if existing.EmailVerified {
			return nil, &EmailTakenError{}
		}
		// Email не подтверждён — обновим данные и переотправим код
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		existing.PasswordHash = string(hash)
		existing.Name = name
		if username != "" {
			if found, err := s.users.GetByUsername(ctx, username); err == nil && found.ID != existing.ID {
				return nil, ErrUsernameTaken
			} else if err != nil && !errors.Is(err, repo.ErrNotFound) {
				return nil, err
			}
			existing.Username = username
		}
		if err := s.users.Update(ctx, existing); err != nil {
			return nil, err
		}
		if s.email != nil && s.email.Configured() {
			s.runBackground(func() {
				if err := s.sendVerificationCode(context.Background(), existing); err != nil {
					slog.Error("verification code failed", "user_id", existing.ID, "error", err)
				}
			})
		}
		return existing, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if username != "" {
		if _, err := s.users.GetByUsername(ctx, username); err == nil {
			return nil, ErrUsernameTaken
		} else if !errors.Is(err, repo.ErrNotFound) {
			return nil, err
		}
	}

	user := &models.User{
		Email:         userEmail,
		PasswordHash:  string(hash),
		Name:          name,
		Username:      username,
		EmailVerified: false,
		ReferredBy:    referredBy, // M-7: может быть пустым, сохраняется как есть
	}

	if err := createUserWithReferralCode(ctx, s.users, user); err != nil {
		return nil, err
	}

	if s.email != nil && s.email.Configured() {
		s.runBackground(func() {
			if err := s.sendVerificationCode(context.Background(), user); err != nil {
				slog.Error("verification code failed", "user_id", user.ID, "error", err)
			}
		})
	}

	return user, nil
}

func providerName(p string) string {
	switch p {
	case ProviderGitHub:
		return "GitHub"
	case ProviderGoogle:
		return "Google"
	case ProviderYandex:
		return "Яндекс"
	default:
		return p
	}
}

func (s *Service) VerifyEmail(ctx context.Context, userEmail, code string) (*models.User, *TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, userEmail)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	v, err := s.verifications.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, nil, ErrInvalidCode
	}

	if err := s.verifyCode(ctx, v, code); err != nil {
		return nil, nil, err
	}

	user.EmailVerified = true
	if err := s.users.Update(ctx, user); err != nil {
		return nil, nil, err
	}

	if err := s.verifications.DeleteByUserID(ctx, user.ID); err != nil {
		slog.Error("failed to delete verification after email verify", "user_id", user.ID, "error", err)
	}

	// Welcome email async — не блокируем verify-flow, если SMTP медленный/упал.
	// VerifyEmail вызывается один раз на аккаунт (verifications-запись удаляется
	// выше), поэтому welcome тоже отправится один раз.
	s.sendWelcomeAsync(user)

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// sendWelcomeAsync шлёт приветственное письмо в background.
// No-op если email не настроен или у юзера пустой email (OAuth без email).
func (s *Service) sendWelcomeAsync(user *models.User) {
	if s.email == nil || !s.email.Configured() || user.Email == "" {
		return
	}
	s.runBackground(func() {
		if err := s.email.SendWelcome(user.Email, user.Name, s.frontendURL); err != nil {
			slog.Warn("welcome email failed", "user_id", user.ID, "error", err)
		}
	})
}

func (s *Service) ResendCode(ctx context.Context, userEmail string) error {
	user, err := s.users.GetByEmail(ctx, userEmail)
	if err != nil {
		return ErrUserNotFound
	}

	if user.EmailVerified {
		return nil
	}

	s.runBackground(func() {
		if err := s.sendVerificationCode(context.Background(), user); err != nil {
			slog.Error("resend verification code failed", "user_id", user.ID, "error", err)
		}
	})
	return nil
}

func (s *Service) sendVerificationCode(ctx context.Context, user *models.User) error {
	if err := s.verifications.DeleteByUserID(ctx, user.ID); err != nil {
		slog.Error("failed to delete old verification before sending new code", "user_id", user.ID, "error", err)
	}

	code, err := generateCode()
	if err != nil {
		slog.Error("failed to generate verification code", "user_id", user.ID, "error", err)
		return err
	}

	v := &models.EmailVerification{
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(VerificationCodeExpiry),
	}

	if err := s.verifications.Create(ctx, v); err != nil {
		return err
	}

	if err := s.email.SendVerificationCode(user.Email, code); err != nil {
		slog.Error("failed to send verification email", "email", user.Email, "error", err)
		return err
	}
	return nil
}

// verifyCode проверяет код с constant-time comparison и ограничением попыток.
func (s *Service) verifyCode(ctx context.Context, v *models.EmailVerification, code string) error {
	if v.Attempts >= models.MaxVerificationAttempts {
		if err := s.verifications.DeleteByUserID(ctx, v.UserID); err != nil {
			slog.Error("failed to delete verification after too many attempts", "user_id", v.UserID, "error", err)
		}
		return ErrTooManyAttempts
	}

	if time.Now().After(v.ExpiresAt) {
		return ErrExpiredCode
	}

	if subtle.ConstantTimeCompare([]byte(v.Code), []byte(code)) != 1 {
		if err := s.verifications.IncrementAttempts(ctx, v.ID); err != nil {
			return fmt.Errorf("increment verification attempts: %w", err)
		}
		return ErrInvalidCode
	}

	return nil
}

func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("generate verification code: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func generateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token nonce: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// issueTokens генерирует новый nonce, сохраняет его в user и выдаёт token pair.
func (s *Service) issueTokens(ctx context.Context, user *models.User) (*TokenPair, error) {
	nonce, err := generateNonce()
	if err != nil {
		return nil, err
	}
	user.TokenNonce = nonce
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return s.generateTokenPair(user.ID, user.TokenNonce)
}

// IssueTokens — экспортированная обёртка над issueTokens. Используется из
// handler после успешной проверки TOTP для admin юзера (обменять pre_auth token
// на полный JWT pair). Также используется в /auth/verify-totp endpoint.
func (s *Service) IssueTokens(ctx context.Context, user *models.User) (*TokenPair, error) {
	return s.issueTokens(ctx, user)
}

// AuthenticatePassword проверяет credentials и возвращает user без issue tokens.
// Отличается от Login тем, что не создаёт новый nonce и не выдаёт refresh_token.
// Используется в login handler для admin flow: сначала проверить password,
// потом, если user=admin AND confirmed TOTP — вернуть pre_auth_token, и только
// после /verify-totp issue полные tokens.
func (s *Service) AuthenticatePassword(ctx context.Context, userEmail, password string) (*models.User, error) {
	user, err := s.users.GetByEmail(ctx, userEmail)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !user.HasPassword() {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.EmailVerified {
		return nil, ErrEmailNotVerified
	}
	return user, nil
}

// IssuePreAuthToken выдаёт short-lived (5 минут) токен с userID,
// используемый как «промежуточный билет» в admin login flow. НЕ является
// access token — не даёт доступа ни к одному protected endpoint, кроме
// POST /api/auth/verify-totp (где он обменивается на полный pair).
func (s *Service) IssuePreAuthToken(userID uint) (string, error) {
	return s.generateToken(userID, TokenTypePreAuth, "", time.Now(), PreAuthTokenDuration)
}

// ValidatePreAuthToken проверяет подпись и срок pre_auth токена, возвращает
// userID. Type в claims должен быть "pre_auth", иначе ErrInvalidToken.
func (s *Service) ValidatePreAuthToken(tokenStr string) (uint, error) {
	claims, err := s.ValidateToken(tokenStr, TokenTypePreAuth)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

func (s *Service) Login(ctx context.Context, userEmail, password string) (*models.User, *TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, userEmail)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if !user.HasPassword() {
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	if !user.EmailVerified {
		return nil, nil, ErrEmailNotVerified
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	// TouchLastLogin в background — триггер для re-engagement (M-5d). Ошибку
	// игнорируем: это не влияет на логин, только на email lifecycle.
	s.touchLastLoginAsync(user.ID)

	return user, tokens, nil
}

// touchLastLoginAsync обновляет users.last_login_at в background.
// Не блокирует response и не фэйлит login если UPDATE упадёт.
func (s *Service) touchLastLoginAsync(userID uint) {
	s.runBackground(func() {
		if err := s.users.TouchLastLogin(context.Background(), userID); err != nil {
			slog.Warn("auth.touch_last_login.failed", "user_id", userID, "error", err)
		}
	})
}

// InitiateSetPassword — шаг 1: отправляет код подтверждения на email (для OAuth-юзеров без пароля)
func (s *Service) InitiateSetPassword(ctx context.Context, userID uint) error {
	user, err := s.getUser(ctx, userID)
	if err != nil {
		return err
	}
	if user.HasPassword() {
		return ErrPasswordAlreadySet
	}
	if s.email == nil || !s.email.Configured() {
		return fmt.Errorf("email не настроен")
	}
	s.runBackground(func() {
		if err := s.sendSetPasswordCode(context.Background(), user); err != nil {
			slog.Error("set-password code failed", "user_id", user.ID, "error", err)
		}
	})
	return nil
}

// ConfirmSetPassword — шаг 2: проверяет код и устанавливает пароль
func (s *Service) ConfirmSetPassword(ctx context.Context, userID uint, code, password string) error {
	user, err := s.getUser(ctx, userID)
	if err != nil {
		return err
	}
	if user.HasPassword() {
		return ErrPasswordAlreadySet
	}

	v, err := s.verifications.GetByUserID(ctx, user.ID)
	if err != nil {
		return ErrInvalidCode
	}

	if err := s.verifyCode(ctx, v, code); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	user.EmailVerified = true
	if err := s.users.Update(ctx, user); err != nil {
		return err
	}

	if err := s.verifications.DeleteByUserID(ctx, user.ID); err != nil {
		slog.Error("failed to delete verification after set password", "user_id", user.ID, "error", err)
	}
	return nil
}

func (s *Service) sendSetPasswordCode(ctx context.Context, user *models.User) error {
	if err := s.verifications.DeleteByUserID(ctx, user.ID); err != nil {
		slog.Error("failed to delete old verification before sending set-password code", "user_id", user.ID, "error", err)
	}

	code, err := generateCode()
	if err != nil {
		return err
	}

	v := &models.EmailVerification{
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(VerificationCodeExpiry),
	}
	if err := s.verifications.Create(ctx, v); err != nil {
		return err
	}
	return s.email.SendSetPasswordCode(user.Email, code)
}

// ForgotPassword — отправляет код сброса на email (публичный, без авторизации)
func (s *Service) ForgotPassword(ctx context.Context, userEmail string) error {
	// Constant-time delay — не раскрываем существование аккаунта через timing
	defer func(start time.Time) {
		elapsed := time.Since(start)
		if elapsed < 200*time.Millisecond {
			time.Sleep(200*time.Millisecond - elapsed)
		}
	}(time.Now())

	// Per-email rate limiting — 3 запроса за 15 минут.
	// Наружу возвращаем nil (не раскрываем rate-limit), но логируем для observability.
	if !s.forgotLimiter.Allow(userEmail) {
		slog.Warn("auth.forgot.rate_limited", "email_suffix", emailSuffix(userEmail))
		return nil
	}

	user, err := s.users.GetByEmail(ctx, userEmail)
	if err != nil {
		// Отличаем "not found" от реальных БД-ошибок — прод-outage не должен быть тихим.
		if !errors.Is(err, ErrUserNotFound) {
			slog.Error("auth.forgot.db_error", "error", err)
		}
		return nil
	}
	if !user.HasPassword() {
		slog.Info("auth.forgot.no_password", "user_id", user.ID)
		return nil
	}
	if s.email == nil || !s.email.Configured() {
		slog.Warn("auth.forgot.email_not_configured")
		return nil
	}
	s.runBackground(func() {
		if err := s.sendResetCode(context.Background(), user); err != nil {
			slog.Error("reset code failed", "user_id", user.ID, "error", err)
		}
	})
	return nil
}

// emailSuffix — возвращает "@domain" для безопасного логирования без PII.
func emailSuffix(email string) string {
	if i := strings.Index(email, "@"); i >= 0 {
		return email[i:]
	}
	return ""
}

// ResetPassword — проверяет код и устанавливает новый пароль
func (s *Service) ResetPassword(ctx context.Context, userEmail, code, newPassword string) error {
	user, err := s.users.GetByEmail(ctx, userEmail)
	if err != nil {
		return ErrInvalidCode
	}

	v, err := s.verifications.GetByUserID(ctx, user.ID)
	if err != nil {
		return ErrInvalidCode
	}

	if err := s.verifyCode(ctx, v, code); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	if err := s.users.Update(ctx, user); err != nil {
		return err
	}

	// Инвалидируем все refresh tokens после сброса пароля
	if err := s.InvalidateTokens(ctx, user.ID); err != nil {
		return err
	}

	if err := s.verifications.DeleteByUserID(ctx, user.ID); err != nil {
		slog.Error("failed to delete verification after password reset", "user_id", user.ID, "error", err)
	}
	return nil
}

func (s *Service) sendResetCode(ctx context.Context, user *models.User) error {
	if err := s.verifications.DeleteByUserID(ctx, user.ID); err != nil {
		slog.Error("failed to delete old verification before sending reset code", "user_id", user.ID, "error", err)
	}

	code, err := generateCode()
	if err != nil {
		return err
	}

	v := &models.EmailVerification{
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(VerificationCodeExpiry),
	}
	if err := s.verifications.Create(ctx, v); err != nil {
		return err
	}
	return s.email.SendPasswordResetCode(user.Email, code)
}

// AdminResetUserPassword инициирует сброс пароля для юзера от имени админа.
// В отличие от ForgotPassword:
//   - не применяет per-email rate limit (админ сам решает когда сбрасывать),
//   - не скрывает несуществующих юзеров (caller уже проверил существование),
//   - не запускает фоновую горутину — отправка email синхронная, чтобы админ
//     видел реальный результат операции и мог повторить при ошибке SMTP,
//   - инвалидирует все refresh tokens юзера — после сброса старые сессии
//     перестают работать (юзер должен залогиниться заново с новым паролем).
//
// Сам password на этом этапе НЕ меняется. Юзер получает email с кодом и
// использует обычный flow /reset-password на frontend, чтобы установить новый
// пароль. Это безопаснее, чем temp-password: admin не видит новый пароль.
func (s *Service) AdminResetUserPassword(ctx context.Context, user *models.User) error {
	if s.email == nil || !s.email.Configured() {
		return errors.New("email service not configured")
	}
	if err := s.sendResetCode(ctx, user); err != nil {
		return err
	}
	return s.InvalidateTokens(ctx, user.ID)
}

func (s *Service) GetLinkedAccounts(ctx context.Context, userID uint) ([]models.LinkedAccount, error) {
	return s.linkedAccounts.GetByUserID(ctx, userID)
}

func (s *Service) UnlinkProvider(ctx context.Context, userID uint, provider string) error {
	count, err := s.linkedAccounts.CountByUserID(ctx, userID)
	if err != nil {
		return err
	}

	user, err := s.getUser(ctx, userID)
	if err != nil {
		return err
	}

	totalMethods := count
	if user.HasPassword() {
		totalMethods++
	}

	if totalMethods <= 1 {
		return ErrCannotUnlinkLast
	}

	return s.linkedAccounts.Delete(ctx, userID, provider)
}

func (s *Service) UpdateProfile(ctx context.Context, userID uint, name, avatarURL string, username *string) (*models.User, error) {
	user, err := s.getUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Name = name
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}
	if username != nil && *username != user.Username {
		if *username != "" {
			existing, err := s.users.GetByUsername(ctx, *username)
			if err == nil && existing.ID != userID {
				return nil, ErrUsernameTaken
			}
			if err != nil && !errors.Is(err, repo.ErrNotFound) {
				return nil, err
			}
		}
		user.Username = *username
	}
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	user, err := s.getUser(ctx, userID)
	if err != nil {
		return err
	}

	if !user.HasPassword() {
		return ErrNoPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrWrongPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	if err := s.users.Update(ctx, user); err != nil {
		return err
	}

	// Rotate nonce to invalidate all existing refresh tokens after password change
	if err := s.InvalidateTokens(ctx, userID); err != nil {
		return err
	}

	if s.email != nil && s.email.Configured() {
		s.runBackground(func() {
			if err := s.email.SendPasswordChangedNotification(user.Email); err != nil {
				slog.Error("password changed notification failed", "user_id", user.ID, "error", err)
			}
		})
	}
	return nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*models.User, *TokenPair, error) {
	claims, err := s.ValidateToken(refreshToken, TokenTypeRefresh)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, nil, ErrUserNotFound
		}
		return nil, nil, err
	}

	// Проверяем nonce — если не совпадает, токен был отозван (logout)
	if claims.Nonce != user.TokenNonce {
		return nil, nil, ErrInvalidToken
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// InvalidateTokens отзывает все refresh tokens пользователя (ротирует nonce).
func (s *Service) InvalidateTokens(ctx context.Context, userID uint) error {
	user, err := s.getUser(ctx, userID)
	if err != nil {
		return err
	}
	nonce, err := generateNonce()
	if err != nil {
		return err
	}
	user.TokenNonce = nonce
	return s.users.Update(ctx, user)
}

func (s *Service) Me(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *Service) ValidateAccessToken(token string) (*Claims, error) {
	return s.ValidateToken(token, TokenTypeAccess)
}

func (s *Service) ValidateToken(tokenStr string, expectedType TokenType) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid || claims.Type != expectedType {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *Service) generateTokenPair(userID uint, nonce string) (*TokenPair, error) {
	now := time.Now()

	accessToken, err := s.generateToken(userID, TokenTypeAccess, "", now, s.accessDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(userID, TokenTypeRefresh, nonce, now, s.refreshDuration)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessDuration.Seconds()),
	}, nil
}

func (s *Service) generateToken(userID uint, tokenType TokenType, nonce string, now time.Time, duration time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Type:   tokenType,
		Nonce:  nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}
