package prompt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	promptuc "promptvault/internal/usecases/prompt"
)

// --- mocks (нужны здесь т.к. _test.go не импортируются) ---

type mPromptRepo struct{ mock.Mock }

func (m *mPromptRepo) Create(ctx context.Context, p *models.Prompt) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mPromptRepo) GetByID(ctx context.Context, id uint) (*models.Prompt, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}
func (m *mPromptRepo) Update(ctx context.Context, p *models.Prompt) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mPromptRepo) SoftDelete(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mPromptRepo) List(ctx context.Context, f repo.PromptListFilter) ([]models.Prompt, int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]models.Prompt), args.Get(1).(int64), args.Error(2)
}
func (m *mPromptRepo) SetFavorite(ctx context.Context, id uint, fav bool) error {
	return m.Called(ctx, id, fav).Error(0)
}
func (m *mPromptRepo) IncrementUsage(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mPromptRepo) SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Prompt, error) {
	args := m.Called(ctx, userID, teamID, query, limit)
	return args.Get(0).([]models.Prompt), args.Error(1)
}

type mVersionRepo struct{ mock.Mock }

func (m *mVersionRepo) CreateWithNextVersion(ctx context.Context, v *models.PromptVersion) error {
	return m.Called(ctx, v).Error(0)
}
func (m *mVersionRepo) ListByPromptID(ctx context.Context, promptID uint, page, pageSize int) ([]models.PromptVersion, int64, error) {
	args := m.Called(ctx, promptID, page, pageSize)
	return args.Get(0).([]models.PromptVersion), args.Get(1).(int64), args.Error(2)
}
func (m *mVersionRepo) GetByIDForPrompt(ctx context.Context, versionID, promptID uint) (*models.PromptVersion, error) {
	args := m.Called(ctx, versionID, promptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PromptVersion), args.Error(1)
}

type mTagRepo struct{ mock.Mock }

func (m *mTagRepo) GetOrCreate(ctx context.Context, name, color string, userID uint, teamID *uint) (*models.Tag, error) {
	args := m.Called(ctx, name, color, userID, teamID)
	return args.Get(0).(*models.Tag), args.Error(1)
}
func (m *mTagRepo) List(ctx context.Context, userID uint, teamID *uint) ([]models.Tag, error) {
	args := m.Called(ctx, userID, teamID)
	return args.Get(0).([]models.Tag), args.Error(1)
}
func (m *mTagRepo) GetByIDs(ctx context.Context, ids []uint) ([]models.Tag, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.Tag), args.Error(1)
}
func (m *mTagRepo) Delete(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mTagRepo) SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Tag, error) {
	args := m.Called(ctx, userID, teamID, query, limit)
	return args.Get(0).([]models.Tag), args.Error(1)
}
func (m *mTagRepo) GetByID(ctx context.Context, id uint) (*models.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tag), args.Error(1)
}

type mCollRepo struct{ mock.Mock }

func (m *mCollRepo) Create(ctx context.Context, c *models.Collection) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mCollRepo) GetByID(ctx context.Context, id uint) (*models.Collection, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Collection), args.Error(1)
}
func (m *mCollRepo) Update(ctx context.Context, c *models.Collection) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mCollRepo) Delete(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mCollRepo) CountPrompts(ctx context.Context, collectionID uint) (int64, error) {
	args := m.Called(ctx, collectionID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mCollRepo) GetByIDs(ctx context.Context, ids []uint) ([]models.Collection, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.Collection), args.Error(1)
}
func (m *mCollRepo) ListWithCounts(ctx context.Context, userID uint, teamIDs []uint) ([]models.CollectionWithCount, error) {
	args := m.Called(ctx, userID, teamIDs)
	return args.Get(0).([]models.CollectionWithCount), args.Error(1)
}
func (m *mCollRepo) SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Collection, error) {
	args := m.Called(ctx, userID, teamID, query, limit)
	return args.Get(0).([]models.Collection), args.Error(1)
}

// --- helpers ---

func setupHandler() (*Handler, *mPromptRepo, *mVersionRepo) {
	pr := new(mPromptRepo)
	vr := new(mVersionRepo)
	tr := new(mTagRepo)
	cr := new(mCollRepo)
	svc := promptuc.NewService(pr, tr, cr, vr, nil)
	return NewHandler(svc), pr, vr
}

func makeReq(method, url string, userID uint, params map[string]string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, url, nil)
	ctx := context.WithValue(req.Context(), authmw.UserIDKey, userID)
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	return req.WithContext(ctx), httptest.NewRecorder()
}

// ===== ListVersions handler =====

func TestHandler_ListVersions_Success(t *testing.T) {
	h, pr, vr := setupHandler()

	prompt := &models.Prompt{ID: 1, UserID: 10, Title: "Test"}
	pr.On("GetByID", mock.Anything, uint(1)).Return(prompt, nil)

	versions := []models.PromptVersion{
		{ID: 2, PromptID: 1, VersionNumber: 2, Title: "v2", Content: "c2", CreatedAt: time.Now()},
		{ID: 1, PromptID: 1, VersionNumber: 1, Title: "v1", Content: "c1", CreatedAt: time.Now()},
	}
	vr.On("ListByPromptID", mock.Anything, uint(1), 1, 20).Return(versions, int64(2), nil)

	req, w := makeReq("GET", "/api/prompts/1/versions?page=1&page_size=20", 10, map[string]string{"id": "1"})
	h.ListVersions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(2), resp["total"])
	assert.True(t, resp["has_more"] == false)
	items := resp["items"].([]any)
	assert.Len(t, items, 2)
	assert.Equal(t, float64(2), items[0].(map[string]any)["version_number"])
}

func TestHandler_ListVersions_DefaultPagination(t *testing.T) {
	h, pr, vr := setupHandler()

	pr.On("GetByID", mock.Anything, uint(1)).Return(&models.Prompt{ID: 1, UserID: 10}, nil)
	vr.On("ListByPromptID", mock.Anything, uint(1), 1, 20).Return([]models.PromptVersion{}, int64(0), nil)

	req, w := makeReq("GET", "/api/prompts/1/versions", 10, map[string]string{"id": "1"})
	h.ListVersions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	vr.AssertCalled(t, "ListByPromptID", mock.Anything, uint(1), 1, 20)
}

func TestHandler_ListVersions_HasMore(t *testing.T) {
	h, pr, vr := setupHandler()

	pr.On("GetByID", mock.Anything, uint(1)).Return(&models.Prompt{ID: 1, UserID: 10}, nil)
	// 25 total, page_size=10 → has_more = true
	vr.On("ListByPromptID", mock.Anything, uint(1), 1, 10).Return(
		make([]models.PromptVersion, 10), int64(25), nil,
	)

	req, w := makeReq("GET", "/api/prompts/1/versions?page=1&page_size=10", 10, map[string]string{"id": "1"})
	h.ListVersions(w, req)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["has_more"])
	assert.Equal(t, float64(25), resp["total"])
}

func TestHandler_ListVersions_InvalidID(t *testing.T) {
	h, _, _ := setupHandler()

	req, w := makeReq("GET", "/api/prompts/abc/versions", 10, map[string]string{"id": "abc"})
	h.ListVersions(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_ListVersions_Forbidden(t *testing.T) {
	h, pr, _ := setupHandler()

	pr.On("GetByID", mock.Anything, uint(1)).Return(&models.Prompt{ID: 1, UserID: 10}, nil)

	req, w := makeReq("GET", "/api/prompts/1/versions", 999, map[string]string{"id": "1"})
	h.ListVersions(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ===== RevertToVersion handler =====

func TestHandler_RevertToVersion_Success(t *testing.T) {
	h, pr, vr := setupHandler()

	oldVer := &models.PromptVersion{
		ID: 5, PromptID: 1, VersionNumber: 1,
		Title: "Оригинал", Content: "Контент v1", Model: "gpt-4o",
	}
	vr.On("GetByIDForPrompt", mock.Anything, uint(5), uint(1)).Return(oldVer, nil)

	current := &models.Prompt{ID: 1, UserID: 10, Title: "Текущий", Content: "Текущий контент", Model: "gpt-4o"}
	pr.On("GetByID", mock.Anything, uint(1)).Return(current, nil)
	vr.On("CreateWithNextVersion", mock.Anything, mock.Anything).Return(nil)
	pr.On("Update", mock.Anything, mock.Anything).Return(nil)

	reverted := &models.Prompt{ID: 1, UserID: 10, Title: "Оригинал", Content: "Контент v1", Model: "gpt-4o"}
	pr.On("GetByID", mock.Anything, uint(1)).Return(reverted, nil)

	req, w := makeReq("POST", "/api/prompts/1/revert/5", 10, map[string]string{"id": "1", "versionId": "5"})
	h.RevertToVersion(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Оригинал", resp["title"])
}

func TestHandler_RevertToVersion_InvalidPromptID(t *testing.T) {
	h, _, _ := setupHandler()

	req, w := makeReq("POST", "/api/prompts/abc/revert/1", 10, map[string]string{"id": "abc", "versionId": "1"})
	h.RevertToVersion(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_RevertToVersion_InvalidVersionID(t *testing.T) {
	h, _, _ := setupHandler()

	req, w := makeReq("POST", "/api/prompts/1/revert/abc", 10, map[string]string{"id": "1", "versionId": "abc"})
	h.RevertToVersion(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_RevertToVersion_VersionNotFound(t *testing.T) {
	h, _, vr := setupHandler()

	vr.On("GetByIDForPrompt", mock.Anything, uint(99), uint(1)).Return(nil, repo.ErrNotFound)

	req, w := makeReq("POST", "/api/prompts/1/revert/99", 10, map[string]string{"id": "1", "versionId": "99"})
	h.RevertToVersion(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_RevertToVersion_Forbidden(t *testing.T) {
	h, pr, vr := setupHandler()

	oldVer := &models.PromptVersion{ID: 5, PromptID: 1, VersionNumber: 1, Title: "T", Content: "C", Model: "M"}
	vr.On("GetByIDForPrompt", mock.Anything, uint(5), uint(1)).Return(oldVer, nil)
	pr.On("GetByID", mock.Anything, uint(1)).Return(&models.Prompt{ID: 1, UserID: 10}, nil)

	req, w := makeReq("POST", "/api/prompts/1/revert/5", 999, map[string]string{"id": "1", "versionId": "5"})
	h.RevertToVersion(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
