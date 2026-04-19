package oauth_server

import (
	"encoding/json"
	"time"
)

// RegisterClientInput — RFC 7591 Dynamic Client Registration request.
type RegisterClientInput struct {
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	Scope                   string   `json:"scope,omitempty"`
}

// RegisterClientOutput — RFC 7591 response.
type RegisterClientOutput struct {
	ClientID                string    `json:"client_id"`
	ClientSecret            string    `json:"client_secret,omitempty"` // пусто для public clients
	ClientIDIssuedAt        int64     `json:"client_id_issued_at"`
	ClientName              string    `json:"client_name"`
	RedirectURIs            []string  `json:"redirect_uris"`
	GrantTypes              []string  `json:"grant_types"`
	ResponseTypes           []string  `json:"response_types"`
	TokenEndpointAuthMethod string    `json:"token_endpoint_auth_method"`
	Scope                   string    `json:"scope"`
	CreatedAt               time.Time `json:"-"`
}

// AuthorizeInput — параметры GET /oauth/authorize после логина пользователя.
// UserID устанавливает HTTP-handler из JWT-сессии пользователя.
type AuthorizeInput struct {
	UserID              uint
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	Resource            string // RFC 8707 — canonical URI целевого MCP
}

// AuthorizeOutput содержит raw-код и state для передачи клиенту через redirect.
type AuthorizeOutput struct {
	Code  string
	State string
}

// ExchangeCodeInput — POST /oauth/token с grant_type=authorization_code.
type ExchangeCodeInput struct {
	ClientID     string
	Code         string
	RedirectURI  string
	CodeVerifier string
	Resource     string
}

// RefreshTokenInput — POST /oauth/token с grant_type=refresh_token.
type RefreshTokenInput struct {
	ClientID     string
	RefreshToken string
	Scope        string // optional, должен быть подмножеством оригинального
	Resource     string
}

// TokenResponse — RFC 6749 §5.1.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"` // Bearer
	ExpiresIn    int64  `json:"expires_in"` // секунды
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
}

// ValidatedAccessToken — результат валидации для MCP middleware.
type ValidatedAccessToken struct {
	UserID   uint
	ClientID string
	Scope    string
	Resource string
	Policy   json.RawMessage // сериализованный models.Policy, для rehydrate в вызывающем слое
}
