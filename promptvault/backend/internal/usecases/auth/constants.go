package auth

import "time"

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"

	DefaultAccessDuration         = 15 * time.Minute
	DefaultRefreshDuration        = 7 * 24 * time.Hour
	VerificationCodeExpiry        = 15 * time.Minute

	ProviderGitHub = "github"
	ProviderGoogle = "google"
	ProviderYandex = "yandex"
)
