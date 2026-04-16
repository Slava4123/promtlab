package prompt

import (
	"context"

	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- PromptRepository mock ---

type mockPromptRepo struct{ mock.Mock }

func (m *mockPromptRepo) Create(ctx context.Context, p *models.Prompt) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockPromptRepo) GetByID(ctx context.Context, id uint) (*models.Prompt, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}
func (m *mockPromptRepo) Update(ctx context.Context, p *models.Prompt) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockPromptRepo) SoftDelete(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockPromptRepo) List(ctx context.Context, f repo.PromptListFilter) ([]models.Prompt, int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]models.Prompt), args.Get(1).(int64), args.Error(2)
}
func (m *mockPromptRepo) SetFavorite(ctx context.Context, id uint, fav bool) error {
	return m.Called(ctx, id, fav).Error(0)
}
func (m *mockPromptRepo) IncrementUsage(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockPromptRepo) SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Prompt, error) {
	args := m.Called(ctx, userID, teamID, query, limit)
	return args.Get(0).([]models.Prompt), args.Error(1)
}
func (m *mockPromptRepo) UpdateLastUsed(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockPromptRepo) ListRecent(ctx context.Context, userID uint, teamID *uint, limit int) ([]models.Prompt, error) {
	args := m.Called(ctx, userID, teamID, limit)
	return args.Get(0).([]models.Prompt), args.Error(1)
}
func (m *mockPromptRepo) LogUsage(ctx context.Context, userID, promptID uint) error {
	return m.Called(ctx, userID, promptID).Error(0)
}
func (m *mockPromptRepo) ListUsageHistory(ctx context.Context, userID uint, teamID *uint, page, pageSize int) ([]models.PromptUsageLog, int64, error) {
	args := m.Called(ctx, userID, teamID, page, pageSize)
	return args.Get(0).([]models.PromptUsageLog), args.Get(1).(int64), args.Error(2)
}
func (m *mockPromptRepo) SuggestByPrefix(ctx context.Context, userID uint, teamID *uint, prefix string, limit int) ([]string, error) {
	args := m.Called(ctx, userID, teamID, prefix, limit)
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockPromptRepo) GetPublicBySlug(ctx context.Context, slug string) (*models.Prompt, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}
func (m *mockPromptRepo) ListPublic(ctx context.Context, limit int) ([]models.Prompt, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Prompt), args.Error(1)
}

// --- VersionRepository mock ---

type mockVersionRepo struct{ mock.Mock }

func (m *mockVersionRepo) CreateWithNextVersion(ctx context.Context, v *models.PromptVersion) error {
	args := m.Called(ctx, v)
	// Имитируем атомарное присвоение номера версии
	if v.VersionNumber == 0 {
		v.VersionNumber = 1
	}
	return args.Error(0)
}
func (m *mockVersionRepo) ListByPromptID(ctx context.Context, promptID uint, page, pageSize int) ([]models.PromptVersion, int64, error) {
	args := m.Called(ctx, promptID, page, pageSize)
	return args.Get(0).([]models.PromptVersion), args.Get(1).(int64), args.Error(2)
}
func (m *mockVersionRepo) GetByIDForPrompt(ctx context.Context, versionID, promptID uint) (*models.PromptVersion, error) {
	args := m.Called(ctx, versionID, promptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PromptVersion), args.Error(1)
}

// --- TagRepository mock ---

type mockTagRepo struct{ mock.Mock }

func (m *mockTagRepo) GetOrCreate(ctx context.Context, name, color string, userID uint, teamID *uint) (*models.Tag, error) {
	args := m.Called(ctx, name, color, userID, teamID)
	return args.Get(0).(*models.Tag), args.Error(1)
}
func (m *mockTagRepo) List(ctx context.Context, userID uint, teamID *uint) ([]models.Tag, error) {
	args := m.Called(ctx, userID, teamID)
	return args.Get(0).([]models.Tag), args.Error(1)
}
func (m *mockTagRepo) GetByIDs(ctx context.Context, ids []uint) ([]models.Tag, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.Tag), args.Error(1)
}
func (m *mockTagRepo) Delete(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockTagRepo) DeleteOrphans(ctx context.Context, userID uint, teamID *uint) error {
	return m.Called(ctx, userID, teamID).Error(0)
}
func (m *mockTagRepo) SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Tag, error) {
	args := m.Called(ctx, userID, teamID, query, limit)
	return args.Get(0).([]models.Tag), args.Error(1)
}
func (m *mockTagRepo) GetByID(ctx context.Context, id uint) (*models.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tag), args.Error(1)
}
func (m *mockTagRepo) SuggestByPrefix(ctx context.Context, userID uint, teamID *uint, prefix string, limit int) ([]string, error) {
	args := m.Called(ctx, userID, teamID, prefix, limit)
	return args.Get(0).([]string), args.Error(1)
}

// --- CollectionRepository mock ---

type mockCollectionRepo struct{ mock.Mock }

func (m *mockCollectionRepo) Create(ctx context.Context, c *models.Collection) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockCollectionRepo) GetByID(ctx context.Context, id uint) (*models.Collection, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Collection), args.Error(1)
}
func (m *mockCollectionRepo) Update(ctx context.Context, c *models.Collection) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockCollectionRepo) Delete(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockCollectionRepo) CountPrompts(ctx context.Context, collectionID uint) (int64, error) {
	args := m.Called(ctx, collectionID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockCollectionRepo) GetByIDs(ctx context.Context, ids []uint) ([]models.Collection, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.Collection), args.Error(1)
}
func (m *mockCollectionRepo) ListWithCounts(ctx context.Context, userID uint, teamIDs []uint) ([]models.CollectionWithCount, error) {
	args := m.Called(ctx, userID, teamIDs)
	return args.Get(0).([]models.CollectionWithCount), args.Error(1)
}
func (m *mockCollectionRepo) SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Collection, error) {
	args := m.Called(ctx, userID, teamID, query, limit)
	return args.Get(0).([]models.Collection), args.Error(1)
}
func (m *mockCollectionRepo) SuggestByPrefix(ctx context.Context, userID uint, teamID *uint, prefix string, limit int) ([]string, error) {
	args := m.Called(ctx, userID, teamID, prefix, limit)
	return args.Get(0).([]string), args.Error(1)
}
