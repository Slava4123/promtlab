package auth

import "time"

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
	// TokenTypePreAuth — short-lived токен, выдаваемый после успешной проверки
	// password для admin-юзера. Содержит только userID и короткий TTL (5 минут).
	// Обменивается на полный access/refresh pair через POST /api/auth/verify-totp
	// после проверки TOTP или backup code.
	TokenTypePreAuth TokenType = "pre_auth"

	DefaultAccessDuration         = 15 * time.Minute
	DefaultRefreshDuration        = 7 * 24 * time.Hour
	PreAuthTokenDuration          = 5 * time.Minute
	VerificationCodeExpiry        = 15 * time.Minute

	ProviderGitHub = "github"
	ProviderGoogle = "google"
	ProviderYandex = "yandex"
)
