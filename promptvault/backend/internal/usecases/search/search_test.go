package search

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
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

// --- tests ---

func TestSearch_EmptyQuery(t *testing.T) {
	svc := NewService(new(mPromptRepo), new(mCollRepo), new(mTagRepo))

	out, err := svc.Search(context.Background(), 1, nil, "")
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Empty(t, out.Prompts)
	assert.Empty(t, out.Collections)
	assert.Empty(t, out.Tags)
}

func TestSearch_WhitespaceQuery(t *testing.T) {
	svc := NewService(new(mPromptRepo), new(mCollRepo), new(mTagRepo))

	out, err := svc.Search(context.Background(), 1, nil, "   ")
	assert.NoError(t, err)
	assert.Empty(t, out.Prompts)
}

func TestSearch_MixedResults(t *testing.T) {
	pr := new(mPromptRepo)
	cr := new(mCollRepo)
	tr := new(mTagRepo)

	pr.On("SearchByQuery", mock.Anything, uint(10), (*uint)(nil), "test", 5).Return([]models.Prompt{
		{ID: 1, Title: "Test prompt", Content: "Short content"},
	}, nil)
	cr.On("SearchByQuery", mock.Anything, uint(10), (*uint)(nil), "test", 3).Return([]models.Collection{
		{ID: 2, Name: "Test coll", Color: "#ff0000", Icon: "folder"},
	}, nil)
	tr.On("SearchByQuery", mock.Anything, uint(10), (*uint)(nil), "test", 3).Return([]models.Tag{
		{ID: 3, Name: "testing", Color: "#00ff00"},
	}, nil)

	svc := NewService(pr, cr, tr)
	out, err := svc.Search(context.Background(), 10, nil, "test")

	assert.NoError(t, err)
	assert.Len(t, out.Prompts, 1)
	assert.Equal(t, "prompt", out.Prompts[0].Type)
	assert.Equal(t, "Test prompt", out.Prompts[0].Title)
	assert.Equal(t, "Short content", out.Prompts[0].Description)

	assert.Len(t, out.Collections, 1)
	assert.Equal(t, "collection", out.Collections[0].Type)
	assert.Equal(t, "#ff0000", out.Collections[0].Color)

	assert.Len(t, out.Tags, 1)
	assert.Equal(t, "tag", out.Tags[0].Type)
	assert.Equal(t, "#00ff00", out.Tags[0].Color)
}

func TestSearch_ContentTruncation(t *testing.T) {
	pr := new(mPromptRepo)
	cr := new(mCollRepo)
	tr := new(mTagRepo)

	longContent := strings.Repeat("A", 200)
	pr.On("SearchByQuery", mock.Anything, uint(1), (*uint)(nil), "long", 5).Return([]models.Prompt{
		{ID: 1, Title: "Long", Content: longContent},
	}, nil)
	cr.On("SearchByQuery", mock.Anything, uint(1), (*uint)(nil), "long", 3).Return([]models.Collection{}, nil)
	tr.On("SearchByQuery", mock.Anything, uint(1), (*uint)(nil), "long", 3).Return([]models.Tag{}, nil)

	svc := NewService(pr, cr, tr)
	out, err := svc.Search(context.Background(), 1, nil, "long")

	assert.NoError(t, err)
	assert.Len(t, out.Prompts, 1)
	// 120 символов + "..."
	assert.Equal(t, 123, len(out.Prompts[0].Description))
	assert.True(t, strings.HasSuffix(out.Prompts[0].Description, "..."))
}

func TestSearch_NoTruncationForShortContent(t *testing.T) {
	pr := new(mPromptRepo)
	cr := new(mCollRepo)
	tr := new(mTagRepo)

	pr.On("SearchByQuery", mock.Anything, uint(1), (*uint)(nil), "short", 5).Return([]models.Prompt{
		{ID: 1, Title: "Short", Content: "Brief"},
	}, nil)
	cr.On("SearchByQuery", mock.Anything, uint(1), (*uint)(nil), "short", 3).Return([]models.Collection{}, nil)
	tr.On("SearchByQuery", mock.Anything, uint(1), (*uint)(nil), "short", 3).Return([]models.Tag{}, nil)

	svc := NewService(pr, cr, tr)
	out, err := svc.Search(context.Background(), 1, nil, "short")

	assert.NoError(t, err)
	assert.Equal(t, "Brief", out.Prompts[0].Description)
}

func TestSearch_EmptyResults(t *testing.T) {
	pr := new(mPromptRepo)
	cr := new(mCollRepo)
	tr := new(mTagRepo)

	pr.On("SearchByQuery", mock.Anything, uint(1), (*uint)(nil), "xyz", 5).Return([]models.Prompt{}, nil)
	cr.On("SearchByQuery", mock.Anything, uint(1), (*uint)(nil), "xyz", 3).Return([]models.Collection{}, nil)
	tr.On("SearchByQuery", mock.Anything, uint(1), (*uint)(nil), "xyz", 3).Return([]models.Tag{}, nil)

	svc := NewService(pr, cr, tr)
	out, err := svc.Search(context.Background(), 1, nil, "xyz")

	assert.NoError(t, err)
	assert.Empty(t, out.Prompts)
	assert.Empty(t, out.Collections)
	assert.Empty(t, out.Tags)
}

// --- Suggest tests ---

func TestSuggest_EmptyPrefix(t *testing.T) {
	svc := NewService(new(mPromptRepo), new(mCollRepo), new(mTagRepo))

	out, err := svc.Suggest(context.Background(), 1, nil, "")
	assert.NoError(t, err)
	assert.Empty(t, out.Suggestions)
}

func TestSuggest_MixedResults(t *testing.T) {
	pr := new(mPromptRepo)
	cr := new(mCollRepo)
	tr := new(mTagRepo)

	pr.On("SuggestByPrefix", mock.Anything, uint(1), (*uint)(nil), "cod", suggestPrompts).
		Return([]string{"Code Review", "Coding Standards"}, nil)
	cr.On("SuggestByPrefix", mock.Anything, uint(1), (*uint)(nil), "cod", suggestCollections).
		Return([]string{"Code Templates"}, nil)
	tr.On("SuggestByPrefix", mock.Anything, uint(1), (*uint)(nil), "cod", suggestTags).
		Return([]string{"coding"}, nil)

	svc := NewService(pr, cr, tr)
	out, err := svc.Suggest(context.Background(), 1, nil, "cod")

	assert.NoError(t, err)
	assert.Len(t, out.Suggestions, 4)
	assert.Equal(t, "Code Review", out.Suggestions[0].Text)
	assert.Equal(t, "prompt", out.Suggestions[0].Type)
	assert.Equal(t, "Code Templates", out.Suggestions[2].Text)
	assert.Equal(t, "collection", out.Suggestions[2].Type)
	assert.Equal(t, "coding", out.Suggestions[3].Text)
	assert.Equal(t, "tag", out.Suggestions[3].Type)
}

func TestSuggest_Deduplication(t *testing.T) {
	pr := new(mPromptRepo)
	cr := new(mCollRepo)
	tr := new(mTagRepo)

	pr.On("SuggestByPrefix", mock.Anything, uint(1), (*uint)(nil), "test", suggestPrompts).
		Return([]string{"Test Prompt"}, nil)
	cr.On("SuggestByPrefix", mock.Anything, uint(1), (*uint)(nil), "test", suggestCollections).
		Return([]string{"test prompt"}, nil)
	tr.On("SuggestByPrefix", mock.Anything, uint(1), (*uint)(nil), "test", suggestTags).
		Return([]string{}, nil)

	svc := NewService(pr, cr, tr)
	out, err := svc.Suggest(context.Background(), 1, nil, "test")

	assert.NoError(t, err)
	assert.Len(t, out.Suggestions, 1)
	assert.Equal(t, "Test Prompt", out.Suggestions[0].Text)
}

func TestSuggest_LimitCapping(t *testing.T) {
	pr := new(mPromptRepo)
	cr := new(mCollRepo)
	tr := new(mTagRepo)

	pr.On("SuggestByPrefix", mock.Anything, uint(1), (*uint)(nil), "a", suggestPrompts).
		Return([]string{"A1", "A2", "A3", "A4"}, nil)
	cr.On("SuggestByPrefix", mock.Anything, uint(1), (*uint)(nil), "a", suggestCollections).
		Return([]string{"AC1", "AC2"}, nil)
	tr.On("SuggestByPrefix", mock.Anything, uint(1), (*uint)(nil), "a", suggestTags).
		Return([]string{"AT1"}, nil)

	svc := NewService(pr, cr, tr)
	out, err := svc.Suggest(context.Background(), 1, nil, "a")

	assert.NoError(t, err)
	assert.Len(t, out.Suggestions, 7)
}
