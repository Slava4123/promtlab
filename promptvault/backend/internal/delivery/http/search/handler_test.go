package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	searchuc "promptvault/internal/usecases/search"
)

// --- mocks ---

type mPromptRepo struct{ mock.Mock }

func (m *mPromptRepo) Create(ctx context.Context, p *models.Prompt) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mPromptRepo) GetByID(ctx context.Context, id uint) (*models.Prompt, error) {
	args := m.Called(ctx, id)
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
func (m *mPromptRepo) UpdateLastUsed(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mPromptRepo) ListRecent(ctx context.Context, userID uint, teamID *uint, limit int) ([]models.Prompt, error) {
	args := m.Called(ctx, userID, teamID, limit)
	return args.Get(0).([]models.Prompt), args.Error(1)
}
func (m *mPromptRepo) LogUsage(ctx context.Context, userID, promptID uint) error {
	return m.Called(ctx, userID, promptID).Error(0)
}
func (m *mPromptRepo) ListUsageHistory(ctx context.Context, userID uint, teamID *uint, page, pageSize int) ([]models.PromptUsageLog, int64, error) {
	args := m.Called(ctx, userID, teamID, page, pageSize)
	return args.Get(0).([]models.PromptUsageLog), args.Get(1).(int64), args.Error(2)
}
func (m *mPromptRepo) SuggestByPrefix(ctx context.Context, userID uint, teamID *uint, prefix string, limit int) ([]string, error) {
	args := m.Called(ctx, userID, teamID, prefix, limit)
	return args.Get(0).([]string), args.Error(1)
}
func (m *mPromptRepo) GetPublicBySlug(ctx context.Context, slug string) (*models.Prompt, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}
func (m *mPromptRepo) ListPublic(ctx context.Context, limit int) ([]models.Prompt, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Prompt), args.Error(1)
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
func (m *mCollRepo) SuggestByPrefix(ctx context.Context, userID uint, teamID *uint, prefix string, limit int) ([]string, error) {
	args := m.Called(ctx, userID, teamID, prefix, limit)
	return args.Get(0).([]string), args.Error(1)
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
func (m *mTagRepo) DeleteOrphans(ctx context.Context, userID uint, teamID *uint) error {
	return m.Called(ctx, userID, teamID).Error(0)
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
func (m *mTagRepo) SuggestByPrefix(ctx context.Context, userID uint, teamID *uint, prefix string, limit int) ([]string, error) {
	args := m.Called(ctx, userID, teamID, prefix, limit)
	return args.Get(0).([]string), args.Error(1)
}

// --- helpers ---

func makeReq(query string, userID uint) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest("GET", "/api/search?q="+query, nil)
	ctx := context.WithValue(req.Context(), authmw.UserIDKey, userID)
	return req.WithContext(ctx), httptest.NewRecorder()
}

// --- tests ---

func TestHandler_Search_WithQuery(t *testing.T) {
	pr := new(mPromptRepo)
	cr := new(mCollRepo)
	tr := new(mTagRepo)

	pr.On("SearchByQuery", mock.Anything, uint(10), (*uint)(nil), "hello", 5).Return([]models.Prompt{
		{ID: 1, Title: "Hello world", Content: "Content"},
	}, nil)
	cr.On("SearchByQuery", mock.Anything, uint(10), (*uint)(nil), "hello", 3).Return([]models.Collection{}, nil)
	tr.On("SearchByQuery", mock.Anything, uint(10), (*uint)(nil), "hello", 3).Return([]models.Tag{}, nil)

	svc := searchuc.NewService(pr, cr, tr)
	h := NewHandler(svc)

	req, w := makeReq("hello", 10)
	h.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp searchuc.SearchOutput
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Prompts, 1)
	assert.Equal(t, "Hello world", resp.Prompts[0].Title)
	assert.Empty(t, resp.Collections)
	assert.Empty(t, resp.Tags)
}

func TestHandler_Search_EmptyQuery(t *testing.T) {
	svc := searchuc.NewService(new(mPromptRepo), new(mCollRepo), new(mTagRepo))
	h := NewHandler(svc)

	req, w := makeReq("", 10)
	h.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp searchuc.SearchOutput
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Empty(t, resp.Prompts)
	assert.Empty(t, resp.Collections)
	assert.Empty(t, resp.Tags)
}
