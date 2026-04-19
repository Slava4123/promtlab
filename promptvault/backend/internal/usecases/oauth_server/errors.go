package oauth_server

import "errors"

// Доменные ошибки OAuth-сервера. Маппинг на OAuth 2.1 error codes (RFC 6749 §5.2)
// делает HTTP-слой в delivery/http/oauth_server/errors.go.
var (
	ErrClientNotFound       = errors.New("oauth: client not found")
	ErrInvalidRedirectURI   = errors.New("oauth: redirect_uri is not registered for this client")
	ErrInvalidGrant         = errors.New("oauth: code/refresh token invalid, expired or already used")
	ErrInvalidRequest       = errors.New("oauth: request is missing required parameters")
	ErrUnsupportedGrantType = errors.New("oauth: grant_type not supported")
	ErrUnsupportedResponseType = errors.New("oauth: response_type not supported")
	ErrInvalidScope         = errors.New("oauth: requested scope is not allowed")
	ErrInvalidToken         = errors.New("oauth: access token is invalid, expired or revoked")
	ErrPKCERequired         = errors.New("oauth: PKCE code_challenge is required")
	ErrResourceMismatch     = errors.New("oauth: resource parameter does not match authorized resource")
)
