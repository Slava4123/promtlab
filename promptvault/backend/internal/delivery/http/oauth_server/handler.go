// Package oauth_server содержит HTTP-хэндлеры OAuth 2.1 Authorization Server endpoints:
//
//	POST /oauth/register       — RFC 7591 Dynamic Client Registration
//	GET  /oauth/authorize      — initiates PKCE authorization code flow
//	POST /oauth/token          — exchange code or refresh → (access, refresh)
//	POST /oauth/revoke         — RFC 7009 Token Revocation
//
// Metadata endpoints (.well-known/*) — в пакете delivery/http/metadata.
package oauth_server

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"promptvault/internal/delivery/http/utils"
	"promptvault/internal/middleware/auth"
	uc "promptvault/internal/usecases/oauth_server"
)

type Handler struct {
	svc *uc.Service
}

func NewHandler(svc *uc.Service) *Handler {
	return &Handler{svc: svc}
}

// -----------------------------------------------------------------------------
// POST /oauth/register — RFC 7591
// -----------------------------------------------------------------------------

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in uc.RegisterClientInput
	if err := utils.DecodeJSON(r, &in); err != nil {
		writeOAuthError(w, http.StatusBadRequest, errCodeInvalidRequest, "malformed request body")
		return
	}

	out, err := h.svc.RegisterClient(r.Context(), in)
	if err != nil {
		status, code, desc := mapDomainError(err)
		writeOAuthError(w, status, code, desc)
		return
	}

	// RFC 7591 §3.2.1 требует 201 + Cache-Control: no-store.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	utils.WriteJSON(w, http.StatusCreated, out)
}

// -----------------------------------------------------------------------------
// GET /oauth/authorize — требует залогиненного пользователя
// -----------------------------------------------------------------------------

// Authorize ожидает JWT-сессию пользователя в context (через auth.Middleware).
// Query-параметры: client_id, redirect_uri, response_type, scope, state,
// code_challenge, code_challenge_method, resource.
//
// При успехе отдаёт 302 на redirect_uri?code=...&state=...
// При ошибке валидации client_id / redirect_uri — JSON 400 (не redirect,
// чтобы не сливать код злоумышленнику). При остальных ошибках — redirect
// c ?error=… согласно RFC 6749 §4.1.2.1.
func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == 0 {
		// Не залогинен → редирект на login с return-URL. Здесь просто 401,
		// потому что full login flow завязан на frontend redirect.
		writeOAuthError(w, http.StatusUnauthorized, errCodeInvalidRequest, "user not authenticated")
		return
	}

	q := r.URL.Query()
	in := uc.AuthorizeInput{
		UserID:              userID,
		ClientID:            q.Get("client_id"),
		RedirectURI:         q.Get("redirect_uri"),
		Scope:               q.Get("scope"),
		State:               q.Get("state"),
		CodeChallenge:       q.Get("code_challenge"),
		CodeChallengeMethod: q.Get("code_challenge_method"),
		Resource:            q.Get("resource"),
	}

	if q.Get("response_type") != uc.ResponseTypeCode {
		writeOAuthError(w, http.StatusBadRequest, errCodeInvalidRequest, "response_type must be 'code'")
		return
	}

	out, err := h.svc.Authorize(r.Context(), in)
	if err != nil {
		// Для client/redirect ошибок — JSON (чтобы не ретранслировать атакующему).
		// Для остальных можно было бы делать redirect, но упрощённо — тоже JSON.
		status, code, desc := mapDomainError(err)
		writeOAuthError(w, status, code, desc)
		return
	}

	// Успех — 302 redirect с параметрами.
	target, _ := url.Parse(in.RedirectURI)
	qs := target.Query()
	qs.Set("code", out.Code)
	if out.State != "" {
		qs.Set("state", out.State)
	}
	target.RawQuery = qs.Encode()
	http.Redirect(w, r, target.String(), http.StatusFound)
}

// -----------------------------------------------------------------------------
// POST /oauth/token
// -----------------------------------------------------------------------------

func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, errCodeInvalidRequest, "malformed form")
		return
	}

	grantType := r.PostForm.Get("grant_type")
	var (
		resp *uc.TokenResponse
		err  error
	)

	switch grantType {
	case uc.GrantTypeAuthCode:
		resp, err = h.svc.ExchangeCode(r.Context(), uc.ExchangeCodeInput{
			ClientID:     r.PostForm.Get("client_id"),
			Code:         r.PostForm.Get("code"),
			RedirectURI:  r.PostForm.Get("redirect_uri"),
			CodeVerifier: r.PostForm.Get("code_verifier"),
			Resource:     r.PostForm.Get("resource"),
		})
	case uc.GrantTypeRefresh:
		resp, err = h.svc.RefreshToken(r.Context(), uc.RefreshTokenInput{
			ClientID:     r.PostForm.Get("client_id"),
			RefreshToken: r.PostForm.Get("refresh_token"),
			Scope:        r.PostForm.Get("scope"),
			Resource:     r.PostForm.Get("resource"),
		})
	default:
		writeOAuthError(w, http.StatusBadRequest, errCodeUnsupportedGrantType, "grant_type must be authorization_code or refresh_token")
		return
	}

	if err != nil {
		status, code, desc := mapDomainError(err)
		writeOAuthError(w, status, code, desc)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	utils.WriteJSON(w, http.StatusOK, resp)
}

// -----------------------------------------------------------------------------
// POST /oauth/revoke — RFC 7009
// -----------------------------------------------------------------------------

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, errCodeInvalidRequest, "malformed form")
		return
	}
	token := strings.TrimSpace(r.PostForm.Get("token"))
	if token == "" {
		writeOAuthError(w, http.StatusBadRequest, errCodeInvalidRequest, "token parameter required")
		return
	}
	if err := h.svc.Revoke(r.Context(), token); err != nil {
		slog.Error("oauth.revoke.error", "error", err)
		// RFC 7009 рекомендует 200 даже при ошибке.
	}
	w.WriteHeader(http.StatusOK)
}
