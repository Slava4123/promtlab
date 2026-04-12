package apikey

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	apikeyuc "promptvault/internal/usecases/apikey"
)

// --- mock repo ---

type mRepo struct{ mock.Mock }

func (m *mRepo) Create(ctx context.Context, key *models.APIKey) error {
	return m.Called(ctx, key).Error(0)
}
func (m *mRepo) ListByUserID(ctx context.Context, userID uint) ([]models.APIKey, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.APIKey), args.Error(1)
}
func (m *mRepo) GetByHash(ctx context.Context, hash string) (*models.APIKey, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.APIKey), args.Error(1)
}
func (m *mRepo) Delete(ctx context.Context, id, userID uint) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mRepo) UpdateLastUsed(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// --- helpers ---

func setupHandler() (*Handler, *mRepo) {
	r := new(mRepo)
	svc := apikeyuc.NewService(r, 5)
	return NewHandler(svc, 5), r
}

func makeReq(method, url string, userID uint, params map[string]string, body []byte) (*http.Request, *httptest.ResponseRecorder) {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	ctx := context.WithValue(req.Context(), authmw.UserIDKey, uint(userID))
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	return req.WithContext(ctx), httptest.NewRecorder()
}

// --- Create ---

func TestHandler_Create_Success(t *testing.T) {
	h, r := setupHandler()

	r.On("CountByUserID", mock.Anything, uint(1)).Return(int64(0), nil)
	r.On("Create", mock.Anything, mock.AnythingOfType("*models.APIKey")).Return(nil)

	body, _ := json.Marshal(CreateRequest{Name: "Test"})
	req, rr := makeReq(http.MethodPost, "/api/api-keys", 1, nil, body)

	h.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp CreatedAPIKeyResponse
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "Test", resp.Name)
	assert.Contains(t, resp.Key, "pvlt_")
	assert.True(t, len(resp.Key) >= 48)
	assert.Contains(t, resp.KeyPrefix, "pvlt_")
}

func TestHandler_Create_ValidationError(t *testing.T) {
	h, _ := setupHandler()

	body, _ := json.Marshal(map[string]string{"name": ""})
	req, rr := makeReq(http.MethodPost, "/api/api-keys", 1, nil, body)

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- List ---

func TestHandler_List_Success(t *testing.T) {
	h, r := setupHandler()

	r.On("ListByUserID", mock.Anything, uint(1)).Return([]models.APIKey{
		{ID: 1, Name: "Key1", KeyPrefix: "pvlt_aB3x"},
	}, nil)

	req, rr := makeReq(http.MethodGet, "/api/api-keys", 1, nil, nil)

	h.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp ListResponse
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(t, resp.Keys, 1)
	assert.Equal(t, "Key1", resp.Keys[0].Name)
}

func TestHandler_List_Empty(t *testing.T) {
	h, r := setupHandler()

	r.On("ListByUserID", mock.Anything, uint(1)).Return([]models.APIKey{}, nil)

	req, rr := makeReq(http.MethodGet, "/api/api-keys", 1, nil, nil)

	h.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp ListResponse
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.NotNil(t, resp.Keys)
	assert.Len(t, resp.Keys, 0)
}

// --- Revoke ---

func TestHandler_Revoke_Success(t *testing.T) {
	h, r := setupHandler()

	r.On("Delete", mock.Anything, uint(5), uint(1)).Return(nil)

	req, rr := makeReq(http.MethodDelete, "/api/api-keys/5", 1, map[string]string{"id": "5"}, nil)

	h.Revoke(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestHandler_Revoke_NotFound(t *testing.T) {
	h, r := setupHandler()

	r.On("Delete", mock.Anything, uint(99), uint(1)).Return(repo.ErrNotFound)

	req, rr := makeReq(http.MethodDelete, "/api/api-keys/99", 1, map[string]string{"id": "99"}, nil)

	h.Revoke(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandler_Revoke_InvalidID(t *testing.T) {
	h, _ := setupHandler()

	req, rr := makeReq(http.MethodDelete, "/api/api-keys/abc", 1, map[string]string{"id": "abc"}, nil)

	h.Revoke(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
