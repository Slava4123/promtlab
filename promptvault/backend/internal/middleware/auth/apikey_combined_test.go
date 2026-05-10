// MN-4 — APIKeyAuth + CombinedAuth tests.
package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	authuc "promptvault/internal/usecases/auth"
)

// --- mockAPIKeyValidator ---

type mockAPIKeyValidator struct{ mock.Mock }

func (m *mockAPIKeyValidator) ValidateKey(ctx context.Context, raw string) (uint, uint, error) {
	args := m.Called(ctx, raw)
	return args.Get(0).(uint), args.Get(1).(uint), args.Error(2)
}

// --- APIKeyAuth ---

func TestAPIKeyAuth_NoHeader_401(t *testing.T) {
	v := new(mockAPIKeyValidator)
	h := APIKeyAuth(v)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("next handler не должен быть вызван")
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing authorization header")
	v.AssertNotCalled(t, "ValidateKey")
}

func TestAPIKeyAuth_InvalidFormat_401(t *testing.T) {
	v := new(mockAPIKeyValidator)
	h := APIKeyAuth(v)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("not called")
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	req.Header.Set("Authorization", "Basic xxx")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid authorization format")
}

func TestAPIKeyAuth_BadKey_401(t *testing.T) {
	v := new(mockAPIKeyValidator)
	v.On("ValidateKey", mock.Anything, "pvlt_bad").Return(uint(0), uint(0), errors.New("not found"))
	h := APIKeyAuth(v)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("not called")
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	req.Header.Set("Authorization", "Bearer pvlt_bad")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "unauthorized")
}

func TestAPIKeyAuth_ValidKey_PutsUserIDInContext(t *testing.T) {
	v := new(mockAPIKeyValidator)
	v.On("ValidateKey", mock.Anything, "pvlt_good").Return(uint(42), uint(7), nil)
	var captured uint
	h := APIKeyAuth(v)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	req.Header.Set("Authorization", "Bearer pvlt_good")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, uint(42), captured)
}

// --- CombinedAuth ---

func TestCombinedAuth_NoHeader_401(t *testing.T) {
	jwt := new(MockTokenValidator)
	api := new(mockAPIKeyValidator)
	h := CombinedAuth(jwt, api)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("not called")
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCombinedAuth_APIKeyPrefix_RoutesToAPIKeyValidator(t *testing.T) {
	jwt := new(MockTokenValidator)
	api := new(mockAPIKeyValidator)
	api.On("ValidateKey", mock.Anything, "pvlt_xx").Return(uint(99), uint(1), nil)

	var captured uint
	h := CombinedAuth(jwt, api)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	req.Header.Set("Authorization", "Bearer pvlt_xx")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, uint(99), captured)
	jwt.AssertNotCalled(t, "ValidateAccessToken") // NOT routed to JWT
	api.AssertExpectations(t)
}

func TestCombinedAuth_JWT_RoutesToJWTValidator(t *testing.T) {
	jwt := new(MockTokenValidator)
	api := new(mockAPIKeyValidator)
	jwt.On("ValidateAccessToken", "eyJhbGc.fake.jwt").
		Return(&authuc.Claims{UserID: 42, Type: "access"}, nil)

	var captured uint
	h := CombinedAuth(jwt, api)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGc.fake.jwt")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, uint(42), captured)
	api.AssertNotCalled(t, "ValidateKey") // NOT routed to API-key
}

func TestCombinedAuth_BadAPIKey_401_NoFallthroughToJWT(t *testing.T) {
	// Bad API-key не должен retry'иться через JWT — это бы leaked timing-info
	// о том, какой формат токена был угадан правильно.
	jwt := new(MockTokenValidator)
	api := new(mockAPIKeyValidator)
	api.On("ValidateKey", mock.Anything, "pvlt_bad").Return(uint(0), uint(0), errors.New("nope"))

	h := CombinedAuth(jwt, api)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("not called")
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	req.Header.Set("Authorization", "Bearer pvlt_bad")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	jwt.AssertNotCalled(t, "ValidateAccessToken")
}
