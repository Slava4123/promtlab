// MN-3 — validateState unit tests (key защита OAuth callback).
// CSRF-protection через state cookie + query-param. Constant-time compare.
package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authuc "promptvault/internal/usecases/auth"
)

func newOAuthHandler() *OAuthHandler {
	return &OAuthHandler{frontendURL: "http://localhost", jwtSecret: "test-secret-32-bytes-minimum-len"}
}

func TestValidateState_NoStateCookie_Mismatch(t *testing.T) {
	h := newOAuthHandler()
	req := httptest.NewRequest("GET", "/callback?state=abc", nil)
	if err := h.validateState(req); !errors.Is(err, authuc.ErrOAuthStateMismatch) {
		t.Fatalf("expected ErrOAuthStateMismatch, got %v", err)
	}
}

func TestValidateState_EmptyQueryParam_Mismatch(t *testing.T) {
	h := newOAuthHandler()
	req := httptest.NewRequest("GET", "/callback", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "stored-state"})
	if err := h.validateState(req); !errors.Is(err, authuc.ErrOAuthStateMismatch) {
		t.Fatalf("expected ErrOAuthStateMismatch, got %v", err)
	}
}

func TestValidateState_StateMismatch_Refused(t *testing.T) {
	h := newOAuthHandler()
	req := httptest.NewRequest("GET", "/callback?state=different", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "stored-state"})
	if err := h.validateState(req); !errors.Is(err, authuc.ErrOAuthStateMismatch) {
		t.Fatalf("expected ErrOAuthStateMismatch, got %v", err)
	}
}

func TestValidateState_StateMatches_OK(t *testing.T) {
	h := newOAuthHandler()
	state := "32-bytes-of-random-state-value-here"
	req := httptest.NewRequest("GET", "/callback?state="+state, nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: state})
	if err := h.validateState(req); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

// MN-3: constant-time compare работает корректно для разной длины.
func TestValidateState_ShorterAttacker_Refused(t *testing.T) {
	h := newOAuthHandler()
	state := "long-state-value-from-cookie-stored"
	// Атакующий пытается подобрать prefix.
	req := httptest.NewRequest("GET", "/callback?state=long-st", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: state})
	if err := h.validateState(req); !errors.Is(err, authuc.ErrOAuthStateMismatch) {
		t.Fatalf("expected ErrOAuthStateMismatch на префикс-атаку, got %v", err)
	}
}
