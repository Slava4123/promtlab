package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	authuc "promptvault/internal/usecases/auth"
)

type OAuthHandler struct {
	oauth         *authuc.OAuthService
	frontendURL   string
	jwtSecret     string
	secureCookies bool
}

func NewOAuthHandler(oauthSvc *authuc.OAuthService, frontendURL, jwtSecret string, secureCookies bool) *OAuthHandler {
	return &OAuthHandler{oauth: oauthSvc, frontendURL: frontendURL, jwtSecret: jwtSecret, secureCookies: secureCookies}
}

func (h *OAuthHandler) redirectWithTokens(w http.ResponseWriter, r *http.Request, tokens *authuc.TokenPair) {
	// Refresh token — HttpOnly cookie (не видна JS, не в URL).
	// Lax чтобы cookie отправлялась при возврате с OAuth-провайдера (top-level
	// cross-site navigation, которое Strict блокирует).
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 3600,
	})
	// Access token — URL fragment (# не отправляется серверу, не попадает в логи)
	fragment := url.Values{}
	fragment.Set("access_token", tokens.AccessToken)
	fragment.Set("expires_in", fmt.Sprintf("%d", tokens.ExpiresIn))
	http.Redirect(w, r, h.frontendURL+"/oauth/callback#"+fragment.Encode(), http.StatusTemporaryRedirect)
}

// --- PKCE + State cookies ---

func (h *OAuthHandler) setOAuthCookies(w http.ResponseWriter, state, verifier string) {
	http.SetCookie(w, &http.Cookie{
		Name: "oauth_state", Value: state,
		Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, MaxAge: 300,
	})
	http.SetCookie(w, &http.Cookie{
		Name: "oauth_verifier", Value: verifier,
		Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, MaxAge: 300,
	})
}

// maybeSetReferralCookie — если в query есть валидный ?ref=XXXXX, кладём его
// в HttpOnly-cookie на 5 минут (M-7). Callback прочитает и запишет в ctx.
// Валидируем — 8 символов alphanumeric, чтобы не таскать мусор.
func (h *OAuthHandler) maybeSetReferralCookie(w http.ResponseWriter, r *http.Request) {
	ref := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("ref")))
	if len(ref) != 8 {
		return
	}
	for _, ch := range ref {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name: "oauth_ref", Value: ref,
		Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, MaxAge: 300,
	})
}

// popReferralCookie — читает oauth_ref cookie и возвращает код (или "").
// Cookie сразу очищаем, чтобы не перезаписать реф у следующего OAuth-входа.
func (h *OAuthHandler) popReferralCookie(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie("oauth_ref")
	if err != nil || c.Value == "" {
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name: "oauth_ref", Value: "",
		Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, MaxAge: -1,
	})
	return c.Value
}

func (h *OAuthHandler) getVerifier(r *http.Request) string {
	c, err := r.Cookie("oauth_verifier")
	if err != nil {
		return ""
	}
	return c.Value
}

// --- Account Linking ---

// POST /api/auth/link/{provider} — инициация привязки (protected)
func (h *OAuthHandler) InitiateLink(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	userID := authmw.GetUserID(r.Context())

	var authURL, state, verifier string
	var err error

	switch provider {
	case "github":
		authURL, state, verifier, err = h.oauth.GitHubAuthURL()
	case "google":
		authURL, state, verifier, err = h.oauth.GoogleAuthURL()
	case "yandex":
		authURL, state, verifier, err = h.oauth.YandexAuthURL()
	default:
		httperr.Respond(w, httperr.BadRequest("Неизвестный провайдер: "+provider))
		return
	}
	if err != nil {
		respondError(w, err)
		return
	}

	h.setOAuthCookies(w, state, verifier)

	http.SetCookie(w, &http.Cookie{
		Name: "oauth_link", Value: h.signLinkCookie(userID),
		Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, MaxAge: 300,
	})

	utils.WriteOK(w, map[string]string{"redirect_url": authURL})
}

func (h *OAuthHandler) signLinkCookie(userID uint) string {
	data := fmt.Sprintf("%d:%d", userID, time.Now().Unix())
	mac := hmac.New(sha256.New, []byte(h.jwtSecret))
	mac.Write([]byte(data))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	return data + ":" + sig
}

func (h *OAuthHandler) verifyLinkCookie(value string) (uint, bool) {
	parts := strings.SplitN(value, ":", 3)
	if len(parts) != 3 {
		return 0, false
	}
	uid, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, false
	}
	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, false
	}
	if time.Now().Unix()-ts > 300 {
		return 0, false
	}
	data := parts[0] + ":" + parts[1]
	mac := hmac.New(sha256.New, []byte(h.jwtSecret))
	mac.Write([]byte(data))
	expected := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return 0, false
	}
	return uint(uid), true
}

func (h *OAuthHandler) clearOAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: "oauth_state", Value: "", Path: "/", HttpOnly: true, MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name: "oauth_verifier", Value: "", Path: "/", HttpOnly: true, MaxAge: -1,
	})
}

func (h *OAuthHandler) clearLinkCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: "oauth_link", Value: "", Path: "/", HttpOnly: true, MaxAge: -1,
	})
}

func (h *OAuthHandler) redirectLinkError(w http.ResponseWriter, r *http.Request, err error) {
	h.clearLinkCookie(w)
	code := "exchange_failed"
	switch {
	case errors.Is(err, authuc.ErrProviderLinkedToOther):
		code = "linked_to_other"
	case errors.Is(err, authuc.ErrProviderAlreadyLinked):
		code = "already_linked"
	case errors.Is(err, authuc.ErrOAuthNotConfigured):
		code = "not_configured"
	}
	slog.Error("oauth link failed", "error", err, "code", code)
	http.Redirect(w, r, h.frontendURL+"/settings/accounts?link_error="+code, http.StatusTemporaryRedirect)
}

// --- OAuth Redirects ---

// GET /api/auth/oauth/github
func (h *OAuthHandler) GitHubRedirect(w http.ResponseWriter, r *http.Request) {
	authURL, state, verifier, err := h.oauth.GitHubAuthURL()
	if err != nil {
		respondError(w, err)
		return
	}
	h.setOAuthCookies(w, state, verifier)
	h.maybeSetReferralCookie(w, r)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GET /api/auth/oauth/github/callback
func (h *OAuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	if err := h.validateState(r); err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}
	h.clearOAuthCookies(w)

	code := r.URL.Query().Get("code")
	if code == "" {
		httperr.Respond(w, httperr.BadRequest("missing code"))
		return
	}

	verifier := h.getVerifier(r)

	// Link flow
	if linkCookie, err := r.Cookie("oauth_link"); err == nil {
		if userID, ok := h.verifyLinkCookie(linkCookie.Value); ok {
			h.clearLinkCookie(w)
			if err := h.oauth.LinkGitHub(r.Context(), userID, code, verifier); err != nil {
				h.redirectLinkError(w, r, err)
				return
			}
			http.Redirect(w, r, h.frontendURL+"/settings/accounts?linked=github", http.StatusTemporaryRedirect)
			return
		}
		h.clearLinkCookie(w)
	}

	// Login flow
	ctx := authuc.WithReferredBy(r.Context(), h.popReferralCookie(w, r))
	_, tokens, err := h.oauth.ExchangeGitHub(ctx, code, verifier)
	if err != nil {
		respondError(w, err)
		return
	}
	h.redirectWithTokens(w, r, tokens)
}

// GET /api/auth/oauth/google
func (h *OAuthHandler) GoogleRedirect(w http.ResponseWriter, r *http.Request) {
	authURL, state, verifier, err := h.oauth.GoogleAuthURL()
	if err != nil {
		respondError(w, err)
		return
	}
	h.setOAuthCookies(w, state, verifier)
	h.maybeSetReferralCookie(w, r)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GET /api/auth/oauth/google/callback
func (h *OAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if err := h.validateState(r); err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}
	h.clearOAuthCookies(w)

	code := r.URL.Query().Get("code")
	if code == "" {
		httperr.Respond(w, httperr.BadRequest("missing code"))
		return
	}

	verifier := h.getVerifier(r)

	// Link flow
	if linkCookie, err := r.Cookie("oauth_link"); err == nil {
		if userID, ok := h.verifyLinkCookie(linkCookie.Value); ok {
			h.clearLinkCookie(w)
			if err := h.oauth.LinkGoogle(r.Context(), userID, code, verifier); err != nil {
				h.redirectLinkError(w, r, err)
				return
			}
			http.Redirect(w, r, h.frontendURL+"/settings/accounts?linked=google", http.StatusTemporaryRedirect)
			return
		}
		h.clearLinkCookie(w)
	}

	// Login flow
	ctx := authuc.WithReferredBy(r.Context(), h.popReferralCookie(w, r))
	_, tokens, err := h.oauth.ExchangeGoogle(ctx, code, verifier)
	if err != nil {
		respondError(w, err)
		return
	}
	h.redirectWithTokens(w, r, tokens)
}

// GET /api/auth/oauth/yandex
func (h *OAuthHandler) YandexRedirect(w http.ResponseWriter, r *http.Request) {
	authURL, state, verifier, err := h.oauth.YandexAuthURL()
	if err != nil {
		respondError(w, err)
		return
	}
	h.setOAuthCookies(w, state, verifier)
	h.maybeSetReferralCookie(w, r)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GET /api/auth/oauth/yandex/callback
func (h *OAuthHandler) YandexCallback(w http.ResponseWriter, r *http.Request) {
	if err := h.validateState(r); err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}
	h.clearOAuthCookies(w)

	code := r.URL.Query().Get("code")
	if code == "" {
		httperr.Respond(w, httperr.BadRequest("missing code"))
		return
	}

	verifier := h.getVerifier(r)

	// Link flow
	if linkCookie, err := r.Cookie("oauth_link"); err == nil {
		if userID, ok := h.verifyLinkCookie(linkCookie.Value); ok {
			h.clearLinkCookie(w)
			if err := h.oauth.LinkYandex(r.Context(), userID, code, verifier); err != nil {
				h.redirectLinkError(w, r, err)
				return
			}
			http.Redirect(w, r, h.frontendURL+"/settings/accounts?linked=yandex", http.StatusTemporaryRedirect)
			return
		}
		h.clearLinkCookie(w)
	}

	// Login flow
	ctx := authuc.WithReferredBy(r.Context(), h.popReferralCookie(w, r))
	_, tokens, err := h.oauth.ExchangeYandex(ctx, code, verifier)
	if err != nil {
		respondError(w, err)
		return
	}
	h.redirectWithTokens(w, r, tokens)
}

func (h *OAuthHandler) validateState(r *http.Request) error {
	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		return authuc.ErrOAuthStateMismatch
	}
	state := r.URL.Query().Get("state")
	// Constant-time compare — defence-in-depth; state 32 байта рандома,
	// timing attack тяжёлая, но != может leak'ать префикс при high-sample.
	if state == "" || subtle.ConstantTimeCompare([]byte(state), []byte(cookie.Value)) != 1 {
		return authuc.ErrOAuthStateMismatch
	}
	return nil
}
