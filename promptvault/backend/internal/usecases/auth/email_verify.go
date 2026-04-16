package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"promptvault/internal/models"
)

// Email verification flow. Q-13: выделено из auth.go — подтверждение email
// при регистрации и повторная отправка кода. Welcome-email после verify тоже
// здесь (он — логическое продолжение этого flow).

// VerifyEmail проверяет код подтверждения, помечает email_verified=true,
// удаляет verification-запись, асинхронно отправляет welcome email, выдаёт
// полный TokenPair для первого login'а.
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

// ResendCode переотправляет код верификации (fire-and-forget). Если юзер
// уже verified — silently no-op (чтобы не раскрывать существование аккаунта).
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

// sendVerificationCode — low-level создание verification-записи + отправка кода
// через SMTP. Вызывается из Register (первая отправка) и ResendCode (повторная).
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
// Удаляет запись при exceed MaxVerificationAttempts — защита от brute-force.
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

// generateCode — 6-значный numeric код для email-верификации.
func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("generate verification code: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
