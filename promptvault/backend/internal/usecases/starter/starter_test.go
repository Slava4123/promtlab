package starter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- mocks ---

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) SearchUsers(ctx context.Context, query string, limit int) ([]models.User, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *mockUserRepo) Update(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) SetQuotaWarningSentOn(ctx context.Context, userID uint, date time.Time) error {
	return m.Called(ctx, userID, date).Error(0)
}
func (m *mockUserRepo) TouchLastLogin(ctx context.Context, userID uint) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockUserRepo) ListInactiveForReengagement(ctx context.Context, inactiveBefore, sentBefore time.Time, limit int) ([]models.User, error) {
	args := m.Called(ctx, inactiveBefore, sentBefore, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *mockUserRepo) MarkReengagementSent(ctx context.Context, userID uint) error {
	return m.Called(ctx, userID).Error(0)
}

type mockStarterRepo struct{ mock.Mock }

func (m *mockStarterRepo) InstallTemplates(ctx context.Context, userID uint, prompts []*models.Prompt) (time.Time, error) {
	args := m.Called(ctx, userID, prompts)
	return args.Get(0).(time.Time), args.Error(1)
}

// --- helpers ---

func newSvcWithMocks(t *testing.T) (*Service, *mockUserRepo, *mockStarterRepo) {
	t.Helper()
	users := new(mockUserRepo)
	starter := new(mockStarterRepo)
	svc, err := NewService(starter, users)
	require.NoError(t, err)
	return svc, users, starter
}

// --- tests ---

func TestService_NewService_LoadsEmbeddedCatalog(t *testing.T) {
	users := new(mockUserRepo)
	starter := new(mockStarterRepo)
	svc, err := NewService(starter, users)

	require.NoError(t, err)
	require.NotNil(t, svc.catalog)
	assert.Equal(t, 1, svc.catalog.Version)
	assert.Equal(t, "ru", svc.catalog.Lang)
	assert.Len(t, svc.catalog.Categories, 4, "должно быть ровно 4 категории")
	assert.Len(t, svc.catalog.Templates, 30, "должно быть ровно 30 промптов в каталоге")
	// Indexes built
	assert.Equal(t, len(svc.catalog.Templates), len(svc.templatesByID))
}

func TestService_Catalog_AllTemplateCategoriesValid(t *testing.T) {
	svc, _, _ := newSvcWithMocks(t)
	catIDs := make(map[string]bool, len(svc.catalog.Categories))
	for _, c := range svc.catalog.Categories {
		catIDs[c.ID] = true
	}
	for _, tpl := range svc.catalog.Templates {
		assert.True(t, catIDs[tpl.Category],
			"template %q ссылается на несуществующую категорию %q", tpl.ID, tpl.Category)
	}
}

func TestService_ListCatalog_ReturnsEmbedded(t *testing.T) {
	svc, _, _ := newSvcWithMocks(t)
	c := svc.ListCatalog()
	assert.NotNil(t, c)
	assert.Equal(t, svc.catalog, c)
}

func TestService_Install_HappyPath(t *testing.T) {
	svc, users, starter := newSvcWithMocks(t)

	users.On("GetByID", mock.Anything, uint(42)).
		Return(&models.User{ID: 42, OnboardingCompletedAt: nil}, nil)

	expectedTime := time.Now().UTC()
	starter.On("InstallTemplates", mock.Anything, uint(42), mock.MatchedBy(func(prompts []*models.Prompt) bool {
		return len(prompts) == 1 &&
			prompts[0].UserID == 42 &&
			prompts[0].TeamID == nil &&
			prompts[0].Title == "Code Review (PR на русском)"
	})).Return(expectedTime, nil)

	result, err := svc.Install(context.Background(), 42, []string{"dev-code-review"})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Prompts, 1)
	assert.Equal(t, expectedTime, result.CompletedAt)
	users.AssertExpectations(t)
	starter.AssertExpectations(t)
}

func TestService_Install_EmptyArray_StillMarksCompleted(t *testing.T) {
	svc, users, starter := newSvcWithMocks(t)

	users.On("GetByID", mock.Anything, uint(7)).
		Return(&models.User{ID: 7, OnboardingCompletedAt: nil}, nil)
	expectedTime := time.Now().UTC()
	starter.On("InstallTemplates", mock.Anything, uint(7), []*models.Prompt{}).
		Return(expectedTime, nil)

	result, err := svc.Install(context.Background(), 7, []string{})

	require.NoError(t, err)
	assert.Empty(t, result.Prompts)
	assert.Equal(t, expectedTime, result.CompletedAt)
}

func TestService_Install_UnknownTemplateID_RejectsBeforeRepoCall(t *testing.T) {
	svc, users, starter := newSvcWithMocks(t)

	users.On("GetByID", mock.Anything, uint(1)).
		Return(&models.User{ID: 1, OnboardingCompletedAt: nil}, nil)

	_, err := svc.Install(context.Background(), 1, []string{"dev-code-review", "id-which-does-not-exist"})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownTemplate)
	starter.AssertNotCalled(t, "InstallTemplates")
}

func TestService_Install_AlreadyCompleted_Returns409(t *testing.T) {
	svc, users, starter := newSvcWithMocks(t)

	now := time.Now().UTC()
	users.On("GetByID", mock.Anything, uint(5)).
		Return(&models.User{ID: 5, OnboardingCompletedAt: &now}, nil)

	_, err := svc.Install(context.Background(), 5, []string{"dev-code-review"})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAlreadyCompleted)
	starter.AssertNotCalled(t, "InstallTemplates")
}

func TestService_Install_UserNotFound(t *testing.T) {
	svc, users, _ := newSvcWithMocks(t)

	users.On("GetByID", mock.Anything, uint(999)).Return(nil, repo.ErrNotFound)

	_, err := svc.Install(context.Background(), 999, []string{"dev-code-review"})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestService_Install_RepoError_Propagates(t *testing.T) {
	svc, users, starter := newSvcWithMocks(t)

	users.On("GetByID", mock.Anything, uint(3)).
		Return(&models.User{ID: 3, OnboardingCompletedAt: nil}, nil)
	dbErr := errors.New("constraint violation")
	starter.On("InstallTemplates", mock.Anything, uint(3), mock.Anything).
		Return(time.Time{}, dbErr)

	_, err := svc.Install(context.Background(), 3, []string{"dev-code-review"})

	require.Error(t, err)
	assert.Equal(t, dbErr, err)
}

func TestService_Install_DeduplicatesTemplateIDs(t *testing.T) {
	// API контракт не запрещает дубли в install array. Service должен схлопнуть
	// их перед batch insert, иначе юзер получит несколько идентичных промптов.
	svc, users, starter := newSvcWithMocks(t)

	users.On("GetByID", mock.Anything, uint(11)).
		Return(&models.User{ID: 11, OnboardingCompletedAt: nil}, nil)
	starter.On("InstallTemplates", mock.Anything, uint(11),
		mock.MatchedBy(func(p []*models.Prompt) bool {
			return len(p) == 1 && p[0].Title == "Code Review (PR на русском)"
		}),
	).Return(time.Now().UTC(), nil)

	_, err := svc.Install(context.Background(), 11,
		[]string{"dev-code-review", "dev-code-review", "dev-code-review"})

	require.NoError(t, err)
	starter.AssertExpectations(t)
}

func TestService_Install_TxLevelGuard_MapsConflictToAlreadyCompleted(t *testing.T) {
	// Регрессия на TOCTOU race: pre-check Service пропускает юзера
	// (OnboardingCompletedAt == nil), но conditional UPDATE в репозитории
	// не находит строки и возвращает repo.ErrConflict (concurrent install
	// или удалённый юзер). Service должен замапить это в ErrAlreadyCompleted,
	// чтобы клиент получил 409 и засинкал состояние.
	svc, users, starter := newSvcWithMocks(t)

	users.On("GetByID", mock.Anything, uint(20)).
		Return(&models.User{ID: 20, OnboardingCompletedAt: nil}, nil)
	starter.On("InstallTemplates", mock.Anything, uint(20), mock.Anything).
		Return(time.Time{}, repo.ErrConflict)

	_, err := svc.Install(context.Background(), 20, []string{"dev-code-review"})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAlreadyCompleted)
}

func TestService_Install_PromptFieldsCopiedFromCatalog(t *testing.T) {
	svc, users, starter := newSvcWithMocks(t)

	users.On("GetByID", mock.Anything, uint(10)).
		Return(&models.User{ID: 10, OnboardingCompletedAt: nil}, nil)

	var capturedPrompts []*models.Prompt
	starter.On("InstallTemplates", mock.Anything, uint(10), mock.Anything).
		Run(func(args mock.Arguments) {
			capturedPrompts = args.Get(2).([]*models.Prompt)
		}).
		Return(time.Now().UTC(), nil)

	_, err := svc.Install(context.Background(), 10, []string{"dev-code-review"})

	require.NoError(t, err)
	require.Len(t, capturedPrompts, 1)
	p := capturedPrompts[0]
	tpl := svc.templatesByID["dev-code-review"]
	assert.Equal(t, tpl.Title, p.Title)
	assert.Equal(t, tpl.Content, p.Content)
	assert.Equal(t, tpl.Model, p.Model)
	assert.Equal(t, uint(10), p.UserID)
	assert.Nil(t, p.TeamID)
	// Sanity: реальные значения из catalog.json
	assert.Contains(t, p.Title, "Code Review")
	assert.Contains(t, p.Content, "{{язык}}")
}
