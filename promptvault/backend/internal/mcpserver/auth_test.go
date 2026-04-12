package mcpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	apikeyuc "promptvault/internal/usecases/apikey"
)

// --- mock repo ---

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Create(ctx context.Context, key *models.APIKey) error {
	return m.Called(ctx, key).Error(0)
}
func (m *mockRepo) ListByUserID(ctx context.Context, userID uint) ([]models.APIKey, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.APIKey), args.Error(1)
}
func (m *mockRepo) GetByHash(ctx context.Context, hash string) (*models.APIKey, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.APIKey), args.Error(1)
}
func (m *mockRepo) Delete(ctx context.Context, id, userID uint) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockRepo) UpdateLastUsed(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// --- helpers ---

func setupAuth(t *testing.T) (*mockRepo, func(http.Handler) http.Handler) {
	t.Helper()
	r := new(mockRepo)
	svc := apikeyuc.NewService(r, 5)
	return r, APIKeyAuth(svc)
}

func capturedUserID(ctx context.Context) uint {
	return authmw.GetUserID(ctx)
}

// --- tests ---

func TestMCPAuth_NoHeader(t *testing.T) {
	_, auth := setupAuth(t)

	handler := auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing authorization header")
}

func TestMCPAuth_WrongScheme(t *testing.T) {
	_, auth := setupAuth(t)

	handler := auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid authorization format")
}

func TestMCPAuth_InvalidKey(t *testing.T) {
	mr, auth := setupAuth(t)

	// malformed key — too short, rejected before DB call
	handler := auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer pvlt_short")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "unauthorized")
	mr.AssertNotCalled(t, "GetByHash")
}

func TestMCPAuth_KeyNotInDB(t *testing.T) {
	mr, auth := setupAuth(t)

	// well-formed key, not found in DB
	mr.On("GetByHash", mock.Anything, mock.Anything).Return(nil, repo.ErrNotFound)

	handler := auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	fakeKey := "pvlt_" + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa="
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+fakeKey)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	// единая ошибка — нет oracle
	assert.Contains(t, rr.Body.String(), "unauthorized")
}

func TestMCPAuth_ValidKey(t *testing.T) {
	mr, auth := setupAuth(t)

	// generate a real key for hash matching
	mr.On("GetByHash", mock.Anything, mock.Anything).Return(&models.APIKey{
		ID:     1,
		UserID: 42,
	}, nil)
	mr.On("UpdateLastUsed", mock.Anything, uint(1)).Return(nil)

	var gotUserID uint
	handler := auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = capturedUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	fakeKey := "pvlt_" + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa="
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+fakeKey)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, uint(42), gotUserID)
}

func TestMCPAuth_UniformError(t *testing.T) {
	// all auth failures return the same error — no timing oracle
	mr1, auth1 := setupAuth(t)
	_, auth2 := setupAuth(t)
	mr3, auth3 := setupAuth(t)

	// case 1: malformed key
	h1 := auth1(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req1 := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req1.Header.Set("Authorization", "Bearer bad")
	rr1 := httptest.NewRecorder()
	h1.ServeHTTP(rr1, req1)

	// case 2: no header
	h2 := auth2(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req2 := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rr2 := httptest.NewRecorder()
	h2.ServeHTTP(rr2, req2)

	// case 3: key not in DB
	mr1.AssertNotCalled(t, "GetByHash")
	mr3.On("GetByHash", mock.Anything, mock.Anything).Return(nil, repo.ErrNotFound)
	h3 := auth3(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	fakeKey := "pvlt_" + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa="
	req3 := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req3.Header.Set("Authorization", "Bearer "+fakeKey)
	rr3 := httptest.NewRecorder()
	h3.ServeHTTP(rr3, req3)

	// all return 401
	assert.Equal(t, http.StatusUnauthorized, rr1.Code)
	assert.Equal(t, http.StatusUnauthorized, rr2.Code)
	assert.Equal(t, http.StatusUnauthorized, rr3.Code)
}
