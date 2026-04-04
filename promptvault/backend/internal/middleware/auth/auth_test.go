package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	authuc "promptvault/internal/usecases/auth"
)

// --- Mock TokenValidator ---

type MockTokenValidator struct {
	mock.Mock
}

func (m *MockTokenValidator) ValidateAccessToken(token string) (*authuc.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authuc.Claims), args.Error(1)
}

// --- Tests ---

func TestMiddleware_NoAuthHeader(t *testing.T) {
	v := new(MockTokenValidator)
	handler := Middleware(v)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing authorization header")
	v.AssertNotCalled(t, "ValidateAccessToken")
}

func TestMiddleware_InvalidFormat(t *testing.T) {
	v := new(MockTokenValidator)
	handler := Middleware(v)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic xxx")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid authorization format")
	v.AssertNotCalled(t, "ValidateAccessToken")
}

func TestMiddleware_InvalidToken(t *testing.T) {
	v := new(MockTokenValidator)
	v.On("ValidateAccessToken", "bad-token").Return(nil, errors.New("token expired"))

	handler := Middleware(v)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid or expired token")
	v.AssertExpectations(t)
}

func TestMiddleware_ValidToken(t *testing.T) {
	v := new(MockTokenValidator)
	claims := &authuc.Claims{UserID: 42, Type: "access"}
	v.On("ValidateAccessToken", "good-token").Return(claims, nil)

	var capturedUserID uint
	handler := Middleware(v)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, uint(42), capturedUserID)
	v.AssertExpectations(t)
}

func TestGetUserID_EmptyContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	uid := GetUserID(req.Context())
	assert.Equal(t, uint(0), uid)
}
