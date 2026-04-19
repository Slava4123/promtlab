package oauth_server

import "time"

// TTL'ы согласно OAuth 2.1 Security BCP.
const (
	// AuthorizationCodeTTL — короткий, один раз использованный (one-time).
	// RFC 6749 §4.1.2 рекомендует ≤ 10 мин, мы ставим 60 сек как в spec.
	AuthorizationCodeTTL = 60 * time.Second

	// AccessTokenTTL — умеренный, чтобы минимизировать impact утечки.
	AccessTokenTTL = 1 * time.Hour

	// RefreshTokenTTL — долгий, но не вечный. Rotation на каждом use.
	RefreshTokenTTL = 30 * 24 * time.Hour

	// Поддерживаемые response_type/grant_type.
	ResponseTypeCode   = "code"
	GrantTypeAuthCode  = "authorization_code"
	GrantTypeRefresh   = "refresh_token"
	CodeChallengeS256  = "S256"
	TokenAuthMethodNone = "none" // PKCE-only public clients
)

// DefaultScope выдаётся в RFC 7591 registration'е если клиент не запросил явный scope.
const DefaultScope = "mcp:read mcp:write"
