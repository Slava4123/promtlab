package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"promptvault/internal/infrastructure/config"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

var yandexEndpoint = oauth2.Endpoint{
	AuthURL:  "https://oauth.yandex.ru/authorize",
	TokenURL: "https://oauth.yandex.ru/token",
}

type OAuthService struct {
	users          repo.UserRepository
	linkedAccounts repo.LinkedAccountRepository
	jwt            *Service
	githubCfg      *oauth2.Config
	googleCfg      *oauth2.Config
	yandexCfg      *oauth2.Config
}

func NewOAuthService(cfg *config.Config, users repo.UserRepository, linkedAccounts repo.LinkedAccountRepository, jwtSvc *Service) *OAuthService {
	svc := &OAuthService{
		users:          users,
		linkedAccounts: linkedAccounts,
		jwt:            jwtSvc,
	}

	if cfg.OAuth.GitHub.ClientID != "" {
		svc.githubCfg = &oauth2.Config{
			ClientID:     cfg.OAuth.GitHub.ClientID,
			ClientSecret: cfg.OAuth.GitHub.ClientSecret,
			Endpoint:     github.Endpoint,
			RedirectURL:  cfg.OAuth.CallbackBase + "/api/auth/oauth/github/callback",
			Scopes:       []string{"user:email"},
		}
	}

	if cfg.OAuth.Google.ClientID != "" {
		svc.googleCfg = &oauth2.Config{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  cfg.OAuth.CallbackBase + "/api/auth/oauth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
		}
	}

	if cfg.OAuth.Yandex.ClientID != "" {
		svc.yandexCfg = &oauth2.Config{
			ClientID:     cfg.OAuth.Yandex.ClientID,
			ClientSecret: cfg.OAuth.Yandex.ClientSecret,
			Endpoint:     yandexEndpoint,
			RedirectURL:  cfg.OAuth.CallbackBase + "/api/auth/oauth/yandex/callback",
		}
	}

	return svc
}

// --- Auth URLs (with PKCE) ---

// AuthURL возвращает (url, state, pkceVerifier, error).
// Verifier нужно сохранить в cookie и передать при Exchange.
func (s *OAuthService) GitHubAuthURL() (string, string, string, error) {
	return s.authURL(s.githubCfg)
}

func (s *OAuthService) GoogleAuthURL() (string, string, string, error) {
	return s.authURL(s.googleCfg)
}

func (s *OAuthService) YandexAuthURL() (string, string, string, error) {
	return s.authURL(s.yandexCfg)
}

func (s *OAuthService) authURL(cfg *oauth2.Config) (string, string, string, error) {
	if cfg == nil {
		return "", "", "", ErrOAuthNotConfigured
	}
	state, err := generateState()
	if err != nil {
		return "", "", "", err
	}
	verifier := oauth2.GenerateVerifier()
	url := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	return url, state, verifier, nil
}

// --- Exchange (with PKCE) ---

func (s *OAuthService) ExchangeGitHub(ctx context.Context, code, verifier string) (*models.User, *TokenPair, error) {
	if s.githubCfg == nil {
		return nil, nil, ErrOAuthNotConfigured
	}

	token, err := s.githubCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrOAuthExchangeFailed, err)
	}

	profile, err := s.fetchGitHubProfile(ctx, token)
	if err != nil {
		return nil, nil, err
	}

	return s.upsertOAuthUser(ctx, ProviderGitHub, profile.id, profile.email, profile.name, profile.avatarURL)
}

func (s *OAuthService) ExchangeGoogle(ctx context.Context, code, verifier string) (*models.User, *TokenPair, error) {
	if s.googleCfg == nil {
		return nil, nil, ErrOAuthNotConfigured
	}

	token, err := s.googleCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrOAuthExchangeFailed, err)
	}

	profile, err := s.fetchGoogleProfile(ctx, token)
	if err != nil {
		return nil, nil, err
	}

	return s.upsertOAuthUser(ctx, ProviderGoogle, profile.id, profile.email, profile.name, profile.avatarURL)
}

func (s *OAuthService) ExchangeYandex(ctx context.Context, code, verifier string) (*models.User, *TokenPair, error) {
	if s.yandexCfg == nil {
		return nil, nil, ErrOAuthNotConfigured
	}

	token, err := s.yandexCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrOAuthExchangeFailed, err)
	}

	profile, err := s.fetchYandexProfile(ctx, token)
	if err != nil {
		return nil, nil, err
	}

	return s.upsertOAuthUser(ctx, ProviderYandex, profile.id, profile.email, profile.name, profile.avatarURL)
}

// --- Link (привязка провайдера к существующему аккаунту) ---

func (s *OAuthService) LinkGitHub(ctx context.Context, userID uint, code, verifier string) error {
	if s.githubCfg == nil {
		return ErrOAuthNotConfigured
	}
	token, err := s.githubCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOAuthExchangeFailed, err)
	}
	profile, err := s.fetchGitHubProfile(ctx, token)
	if err != nil {
		return err
	}
	return s.linkProvider(ctx, userID, ProviderGitHub, profile)
}

func (s *OAuthService) LinkGoogle(ctx context.Context, userID uint, code, verifier string) error {
	if s.googleCfg == nil {
		return ErrOAuthNotConfigured
	}
	token, err := s.googleCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOAuthExchangeFailed, err)
	}
	profile, err := s.fetchGoogleProfile(ctx, token)
	if err != nil {
		return err
	}
	return s.linkProvider(ctx, userID, ProviderGoogle, profile)
}

func (s *OAuthService) LinkYandex(ctx context.Context, userID uint, code, verifier string) error {
	if s.yandexCfg == nil {
		return ErrOAuthNotConfigured
	}
	token, err := s.yandexCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOAuthExchangeFailed, err)
	}
	profile, err := s.fetchYandexProfile(ctx, token)
	if err != nil {
		return err
	}
	return s.linkProvider(ctx, userID, ProviderYandex, profile)
}

func (s *OAuthService) linkProvider(ctx context.Context, userID uint, provider string, profile *oauthProfile) error {
	existing, err := s.linkedAccounts.GetByProviderID(ctx, provider, profile.id)
	if err == nil {
		if existing.UserID == userID {
			return ErrProviderAlreadyLinked
		}
		return ErrProviderLinkedToOther
	}
	return s.linkedAccounts.Create(ctx, &models.LinkedAccount{
		UserID:     userID,
		Provider:   provider,
		ProviderID: profile.id,
	})
}

// --- Upsert ---

func (s *OAuthService) upsertOAuthUser(ctx context.Context, provider, providerID, oauthEmail, name, avatarURL string) (*models.User, *TokenPair, error) {
	if oauthEmail == "" {
		return nil, nil, fmt.Errorf("%w: провайдер не вернул email", ErrOAuthProfileFailed)
	}

	// 1. Ищем по provider+providerID в linked_accounts
	la, err := s.linkedAccounts.GetByProviderID(ctx, provider, providerID)
	if err == nil {
		// Уже привязан — обновляем профиль и возвращаем
		user, err := s.users.GetByID(ctx, la.UserID)
		if err != nil {
			return nil, nil, err
		}
		user.Name = name
		user.AvatarURL = avatarURL
		if err := s.users.Update(ctx, user); err != nil {
			return nil, nil, err
		}

		tokens, err := s.jwt.issueTokens(ctx, user)
		if err != nil {
			return nil, nil, err
		}
		s.touchLastLogin(user.ID)
		return user, tokens, nil
	}

	// 2. Ищем по email — привязываем провайдер к существующему аккаунту
	user, err := s.users.GetByEmail(ctx, oauthEmail)
	if err == nil {
		// Аккаунт с таким email уже есть — привязываем нового провайдера
		if err := s.linkedAccounts.Create(ctx, &models.LinkedAccount{
			UserID:     user.ID,
			Provider:   provider,
			ProviderID: providerID,
		}); err != nil {
			return nil, nil, err
		}
		user.Name = name
		user.AvatarURL = avatarURL
		user.EmailVerified = true
		if err := s.users.Update(ctx, user); err != nil {
			return nil, nil, err
		}

		tokens, err := s.jwt.issueTokens(ctx, user)
		if err != nil {
			return nil, nil, err
		}
		s.touchLastLogin(user.ID)
		return user, tokens, nil
	}

	// 3. Новый пользователь. M-7: referredBy приходит из oauth_ref cookie через
	// ctx (OAuthHandler кладёт туда значение перед вызовом, если оно есть).
	user = &models.User{
		Email:         oauthEmail,
		Name:          name,
		AvatarURL:     avatarURL,
		EmailVerified: true,
		ReferredBy:    referredByFromCtx(ctx),
	}
	if err := createUserWithReferralCode(ctx, s.users, user); err != nil {
		return nil, nil, err
	}

	if err := s.linkedAccounts.Create(ctx, &models.LinkedAccount{
		UserID:     user.ID,
		Provider:   provider,
		ProviderID: providerID,
	}); err != nil {
		return nil, nil, err
	}

	tokens, err := s.jwt.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	s.touchLastLogin(user.ID)

	return user, tokens, nil
}

// touchLastLogin обновляет users.last_login_at в background — триггер для
// re-engagement (M-5d). Ошибку игнорируем: lifecycle-метрика, не критична.
func (s *OAuthService) touchLastLogin(userID uint) {
	go func() {
		if err := s.users.TouchLastLogin(context.Background(), userID); err != nil {
			slog.Warn("oauth.touch_last_login.failed", "user_id", userID, "error", err)
		}
	}()
}

// --- Profile fetchers ---

type oauthProfile struct {
	id        string
	email     string
	name      string
	avatarURL string
}

func (s *OAuthService) fetchGitHubProfile(ctx context.Context, token *oauth2.Token) (*oauthProfile, error) {
	client := s.githubCfg.Client(ctx, token)

	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOAuthProfileFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: github /user returned %d", ErrOAuthProfileFailed, resp.StatusCode)
	}

	var raw struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOAuthProfileFailed, err)
	}

	name := raw.Name
	if name == "" {
		name = raw.Login
	}

	// S-3 defence-in-depth: всегда идём в /user/emails как основной источник,
	// даже если /user вернул email. Причины:
	// (1) /user.email может отдать unverified primary email;
	// (2) attacker может подменить ответ /user при MITM middleware;
	// (3) email там может быть no-reply@users.noreply.github.com — не настоящий.
	// /user/emails даёт нам явный флаг Verified.
	verifiedEmail, err := s.fetchGitHubVerifiedEmail(ctx, client)
	if err != nil || verifiedEmail == "" {
		// fallback на /user.email только если не получили ничего — логируем нарушение.
		if raw.Email != "" {
			slog.Warn("oauth.github.emails_endpoint_failed_using_user_email", "error", err)
		}
		verifiedEmail = raw.Email
	}

	return &oauthProfile{
		id:        fmt.Sprintf("%d", raw.ID),
		email:     verifiedEmail,
		name:      name,
		avatarURL: raw.AvatarURL,
	}, nil
}

// fetchGitHubVerifiedEmail возвращает primary verified email (S-3 defence-in-depth).
// Переименование из fetchGitHubEmail — подчёркиваем, что метод гарантирует
// Verified=true, а не просто primary. Если нет ни одного verified — возвращает "".
func (s *OAuthService) fetchGitHubVerifiedEmail(ctx context.Context, client *http.Client) (string, error) {
	return s.fetchGitHubEmail(ctx, client)
}

func (s *OAuthService) fetchGitHubEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github /user/emails returned %d", resp.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", nil
}

func (s *OAuthService) fetchGoogleProfile(ctx context.Context, token *oauth2.Token) (*oauthProfile, error) {
	client := s.googleCfg.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOAuthProfileFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: google /userinfo returned %d", ErrOAuthProfileFailed, resp.StatusCode)
	}

	var raw struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOAuthProfileFailed, err)
	}

	// S-3 defence-in-depth: отклоняем unverified Google email'ы. Google помечает
	// email как verified если доменный провайдер это подтвердил (gmail автоматом,
	// G Suite — по SPF/DKIM). Unverified = возможно угнанный, не доверяем.
	if !raw.VerifiedEmail {
		return nil, fmt.Errorf("%w: google email not verified", ErrOAuthProfileFailed)
	}

	return &oauthProfile{
		id:        raw.ID,
		email:     raw.Email,
		name:      raw.Name,
		avatarURL: raw.Picture,
	}, nil
}

func (s *OAuthService) fetchYandexProfile(ctx context.Context, token *oauth2.Token) (*oauthProfile, error) {
	client := s.yandexCfg.Client(ctx, token)

	resp, err := client.Get("https://login.yandex.ru/info?format=json")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOAuthProfileFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: yandex /info returned %d", ErrOAuthProfileFailed, resp.StatusCode)
	}

	var raw struct {
		ID           string `json:"id"`
		DefaultEmail string `json:"default_email"`
		DisplayName  string `json:"display_name"`
		RealName     string `json:"real_name"`
		DefaultAvatarID string `json:"default_avatar_id"`
		IsAvatarEmpty   bool   `json:"is_avatar_empty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOAuthProfileFailed, err)
	}

	name := raw.DisplayName
	if name == "" {
		name = raw.RealName
	}

	var avatarURL string
	if !raw.IsAvatarEmpty {
		avatarURL = fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", raw.DefaultAvatarID)
	}

	return &oauthProfile{
		id:        raw.ID,
		email:     raw.DefaultEmail,
		name:      name,
		avatarURL: avatarURL,
	}, nil
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
