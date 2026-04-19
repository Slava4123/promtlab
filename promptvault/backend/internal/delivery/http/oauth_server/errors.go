package oauth_server

import (
	"errors"
	"net/http"

	"promptvault/internal/delivery/http/utils"
	ucerr "promptvault/internal/usecases/oauth_server"
)

// OAuth 2.1 error codes (RFC 6749 §5.2).
const (
	errCodeInvalidRequest       = "invalid_request"
	errCodeInvalidClient        = "invalid_client"
	errCodeInvalidGrant         = "invalid_grant"
	errCodeUnauthorizedClient   = "unauthorized_client"
	errCodeUnsupportedGrantType = "unsupported_grant_type"
	errCodeInvalidScope         = "invalid_scope"
	errCodeServerError          = "server_error"
)

// oauthErrorResponse соответствует JSON-формату RFC 6749 §5.2.
type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// writeOAuthError пишет стандартный OAuth-JSON. status 400 для большинства
// client errors; 401 для invalid_client на /token.
func writeOAuthError(w http.ResponseWriter, status int, code, description string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	utils.WriteJSON(w, status, oauthErrorResponse{Error: code, ErrorDescription: description})
}

// mapDomainError — доменная ошибка → OAuth error code.
func mapDomainError(err error) (status int, code, description string) {
	switch {
	case errors.Is(err, ucerr.ErrClientNotFound):
		return http.StatusUnauthorized, errCodeInvalidClient, "unknown client"
	case errors.Is(err, ucerr.ErrInvalidRedirectURI):
		return http.StatusBadRequest, errCodeInvalidRequest, "redirect_uri mismatch"
	case errors.Is(err, ucerr.ErrInvalidGrant):
		return http.StatusBadRequest, errCodeInvalidGrant, "code or token invalid, expired or reused"
	case errors.Is(err, ucerr.ErrInvalidRequest):
		return http.StatusBadRequest, errCodeInvalidRequest, err.Error()
	case errors.Is(err, ucerr.ErrUnsupportedGrantType):
		return http.StatusBadRequest, errCodeUnsupportedGrantType, "grant_type not supported"
	case errors.Is(err, ucerr.ErrUnsupportedResponseType):
		return http.StatusBadRequest, errCodeInvalidRequest, "response_type not supported"
	case errors.Is(err, ucerr.ErrInvalidScope):
		return http.StatusBadRequest, errCodeInvalidScope, err.Error()
	case errors.Is(err, ucerr.ErrPKCERequired):
		return http.StatusBadRequest, errCodeInvalidRequest, "PKCE code_challenge is required"
	case errors.Is(err, ucerr.ErrResourceMismatch):
		return http.StatusBadRequest, errCodeInvalidRequest, "resource parameter mismatch"
	default:
		return http.StatusInternalServerError, errCodeServerError, "internal server error"
	}
}
