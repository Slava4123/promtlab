// Package metadata отдаёт OAuth 2.0 metadata endpoints для discovery:
//
//	GET /.well-known/oauth-protected-resource     — RFC 9728
//	GET /.well-known/oauth-authorization-server   — RFC 8414
//
// MCP клиенты (Claude.ai, Cursor) дискаверят OAuth-параметры через эти эндпоинты.
// MCP spec 2025-06-18 §Authorization Server Discovery требует их обязательно.
package metadata

import (
	"net/http"

	"promptvault/internal/delivery/http/utils"
)

// Config — данные для metadata-документов. Устанавливается из cfg при старте.
type Config struct {
	// Issuer — публичный base-URL (без trailing slash). Напр. "https://promtlabs.ru"
	Issuer string
	// ResourceServer — canonical URI MCP-сервера (RFC 8707 audience).
	ResourceServer string
}

type Handler struct {
	cfg Config
}

func NewHandler(cfg Config) *Handler {
	return &Handler{cfg: cfg}
}

// -----------------------------------------------------------------------------
// GET /.well-known/oauth-protected-resource  (RFC 9728)
// -----------------------------------------------------------------------------

// protectedResourceMetadata описывает наш MCP-сервер как OAuth resource server.
type protectedResourceMetadata struct {
	Resource               string   `json:"resource"`
	AuthorizationServers   []string `json:"authorization_servers"`
	ScopesSupported        []string `json:"scopes_supported"`
	BearerMethodsSupported []string `json:"bearer_methods_supported"`
	ResourceName           string   `json:"resource_name,omitempty"`
}

func (h *Handler) ProtectedResource(w http.ResponseWriter, _ *http.Request) {
	doc := protectedResourceMetadata{
		Resource:               h.cfg.ResourceServer,
		AuthorizationServers:   []string{h.cfg.Issuer},
		ScopesSupported:        []string{"mcp:read", "mcp:write"},
		BearerMethodsSupported: []string{"header"},
		ResourceName:           "PromptVault MCP",
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	utils.WriteJSON(w, http.StatusOK, doc)
}

// -----------------------------------------------------------------------------
// GET /.well-known/oauth-authorization-server  (RFC 8414)
// -----------------------------------------------------------------------------

type authorizationServerMetadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	RegistrationEndpoint              string   `json:"registration_endpoint"`
	RevocationEndpoint                string   `json:"revocation_endpoint"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
}

func (h *Handler) AuthorizationServer(w http.ResponseWriter, _ *http.Request) {
	doc := authorizationServerMetadata{
		Issuer:                            h.cfg.Issuer,
		AuthorizationEndpoint:             h.cfg.Issuer + "/oauth/authorize",
		TokenEndpoint:                     h.cfg.Issuer + "/oauth/token",
		RegistrationEndpoint:              h.cfg.Issuer + "/oauth/register",
		RevocationEndpoint:                h.cfg.Issuer + "/oauth/revoke",
		ResponseTypesSupported:            []string{"code"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token"},
		CodeChallengeMethodsSupported:     []string{"S256"},
		TokenEndpointAuthMethodsSupported: []string{"none", "client_secret_post"},
		ScopesSupported:                   []string{"mcp:read", "mcp:write"},
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	utils.WriteJSON(w, http.StatusOK, doc)
}
