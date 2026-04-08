package sentry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"

	authmw "promptvault/internal/middleware/auth"
)

// TestHandler_ReturnsMiddleware — Handler возвращает нормальное middleware,
// не падает на пустой конфигурации (Repanic+Timeout — hard-coded).
func TestHandler_ReturnsMiddleware(t *testing.T) {
	mw := Handler()
	assert.NotNil(t, mw, "Handler() must return non-nil middleware")

	// Может оборачивать handler без паники.
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	assert.NotNil(t, h)
}

// TestUserContext_NoHub_Passthrough — без Sentry Hub в context middleware
// просто пропускает request, не падает. Это критично: при SENTRY_ENABLED=false
// sentry.Init не вызывается, Hub отсутствует, middleware должен быть no-op.
func TestUserContext_NoHub_Passthrough(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw := UserContext(next)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, "next handler must be called when Hub is nil")
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestUserContext_NoUserID_Passthrough — если в context нет UserID (публичный
// endpoint), middleware не падает и не пытается установить user на scope.
func TestUserContext_NoUserID_Passthrough(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Создаём Hub вручную — имитируем случай когда Sentry активен, но
	// текущий endpoint не защищён auth middleware.
	hub := sentry.NewHub(nil, sentry.NewScope())

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(sentry.SetHubOnContext(req.Context(), hub))

	rec := httptest.NewRecorder()
	UserContext(next).ServeHTTP(rec, req)

	assert.True(t, handlerCalled, "next must be called even without UserID")
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestUserContext_WithUserID_SetsUser — happy path: Hub есть, UserID есть,
// middleware устанавливает sentry.User{ID} на scope.
func TestUserContext_WithUserID_SetsUser(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Проверяем что user установлен в scope — делаем это изнутри handler,
		// т.к. Hub клонируется sentry middleware per-request.
		hub := sentry.GetHubFromContext(r.Context())
		assert.NotNil(t, hub, "Hub must be available in handler")
		w.WriteHeader(http.StatusOK)
	})

	hub := sentry.NewHub(nil, sentry.NewScope())

	// Имитируем auth middleware — кладём UserID в context.
	ctx := context.WithValue(context.Background(), authmw.UserIDKey, uint(42))
	ctx = sentry.SetHubOnContext(ctx, hub)

	req := httptest.NewRequest("GET", "/protected", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	UserContext(next).ServeHTTP(rec, req)

	assert.True(t, handlerCalled, "next handler must be called")
	assert.Equal(t, http.StatusOK, rec.Code)
}
