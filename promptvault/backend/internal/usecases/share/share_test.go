package share

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- mocks ---

type mockShareRepo struct{ mock.Mock }

func (m *mockShareRepo) Create(ctx context.Context, link *models.ShareLink) error {
	return m.Called(ctx, link).Error(0)
}
func (m *mockShareRepo) GetByToken(ctx context.Context, token string) (*models.ShareLink, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShareLink), args.Error(1)
}
func (m *mockShareRepo) GetActiveByPromptID(ctx context.Context, promptID uint) (*models.ShareLink, error) {
	args := m.Called(ctx, promptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShareLink), args.Error(1)
}
func (m *mockShareRepo) Deactivate(ctx context.Context, promptID uint) error {
	return m.Called(ctx, promptID).Error(0)
}
func (m *mockShareRepo) IncrementViewCount(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}

type mockPromptRepo struct{ mock.Mock }

func (m *mockPromptRepo) GetByID(ctx context.Context, id uint) (*models.Prompt, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

// Stubs for the rest of PromptRepository interface.
func (m *mockPromptRepo) Create(ctx context.Context, p *models.Prompt) error {
	return m.Called(ctx, p).Error(0)
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
func (m *mockPromptRepo) SearchByQuery(ctx context.Context, userID uint, teamID *uint, q string, limit int) ([]models.Prompt, error) {
	args := m.Called(ctx, userID, teamID, q, limit)
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

type mockTeamRepo struct{ mock.Mock }

func (m *mockTeamRepo) GetMember(ctx context.Context, teamID, userID uint) (*models.TeamMember, error) {
	args := m.Called(ctx, teamID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TeamMember), args.Error(1)
}

// Stubs — share service uses only GetMember via teamcheck.
func (m *mockTeamRepo) CreateWithOwner(context.Context, *models.Team, uint) error { return nil }
func (m *mockTeamRepo) GetBySlug(context.Context, string) (*models.Team, error)   { return nil, nil }
func (m *mockTeamRepo) ListByUserID(context.Context, uint) ([]models.Team, error) { return nil, nil }
func (m *mockTeamRepo) ListByUserIDWithRolesAndCounts(context.Context, uint) ([]models.TeamWithRoleAndCount, error) {
	return nil, nil
}
func (m *mockTeamRepo) Update(context.Context, *models.Team) error      { return nil }
func (m *mockTeamRepo) Delete(context.Context, uint) error              { return nil }
func (m *mockTeamRepo) UpdateMemberRole(context.Context, uint, uint, models.TeamRole) error {
	return nil
}
func (m *mockTeamRepo) RemoveMember(context.Context, uint, uint) error { return nil }
func (m *mockTeamRepo) ListMembers(context.Context, uint) ([]models.TeamMember, error) {
	return nil, nil
}
func (m *mockTeamRepo) CountMembers(context.Context, uint) (int, error)             { return 0, nil }
func (m *mockTeamRepo) CreateInvitation(context.Context, *models.TeamInvitation) error { return nil }
func (m *mockTeamRepo) GetInvitationByID(context.Context, uint) (*models.TeamInvitation, error) {
	return nil, nil
}
func (m *mockTeamRepo) GetPendingInvitation(context.Context, uint, uint) (*models.TeamInvitation, error) {
	return nil, nil
}
func (m *mockTeamRepo) ListPendingByUserID(context.Context, uint) ([]models.TeamInvitation, error) {
	return nil, nil
}
func (m *mockTeamRepo) ListPendingByTeamID(context.Context, uint) ([]models.TeamInvitation, error) {
	return nil, nil
}
func (m *mockTeamRepo) UpdateInvitationStatus(context.Context, uint, models.InvitationStatus) error {
	return nil
}
func (m *mockTeamRepo) DeleteInvitation(context.Context, uint) error                      { return nil }
func (m *mockTeamRepo) AcceptInvitationTx(context.Context, uint, *models.TeamMember) error { return nil }

// --- helpers ---

func setupService() (*Service, *mockShareRepo, *mockPromptRepo, *mockTeamRepo) {
	sr := new(mockShareRepo)
	pr := new(mockPromptRepo)
	tr := new(mockTeamRepo)
	svc := NewService(sr, pr, tr, "https://app.test.ru", nil)
	return svc, sr, pr, tr
}

func personalPrompt(ownerID uint) *models.Prompt {
	return &models.Prompt{ID: 1, UserID: ownerID, Title: "Test", Content: "Hello"}
}

// --- tests ---

func TestGenerateToken(t *testing.T) {
	tok, err := generateToken()
	require.NoError(t, err)
	assert.True(t, len(tok) >= 25)
	assert.Equal(t, "ps_", tok[:3])
}

func TestCreateOrGet_New(t *testing.T) {
	svc, sr, pr, _ := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(personalPrompt(42), nil)
	sr.On("GetActiveByPromptID", ctx, uint(1)).Return(nil, repo.ErrNotFound)
	sr.On("Create", ctx, mock.Anything).Return(nil)

	info, created, err := svc.CreateOrGet(ctx, 1, 42)
	require.NoError(t, err)
	assert.True(t, created)
	assert.Contains(t, info.Token, "ps_")
	assert.Contains(t, info.URL, "https://app.test.ru/s/ps_")
}

func TestCreateOrGet_Existing(t *testing.T) {
	svc, sr, pr, _ := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(personalPrompt(42), nil)
	sr.On("GetActiveByPromptID", ctx, uint(1)).Return(&models.ShareLink{
		ID: 10, Token: "ps_existing", IsActive: true,
	}, nil)

	info, created, err := svc.CreateOrGet(ctx, 1, 42)
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, "ps_existing", info.Token)
	sr.AssertNotCalled(t, "Create")
}

func TestCreateOrGet_Forbidden(t *testing.T) {
	svc, _, pr, _ := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(personalPrompt(99), nil) // owned by 99, not 42

	_, _, err := svc.CreateOrGet(ctx, 1, 42)
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestCreateOrGet_PromptNotFound(t *testing.T) {
	svc, _, pr, _ := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(99)).Return(nil, repo.ErrNotFound)

	_, _, err := svc.CreateOrGet(ctx, 99, 42)
	assert.ErrorIs(t, err, ErrPromptNotFound)
}

func TestDeactivate(t *testing.T) {
	svc, sr, pr, _ := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(personalPrompt(42), nil)
	sr.On("Deactivate", ctx, uint(1)).Return(nil)

	err := svc.Deactivate(ctx, 1, 42)
	assert.NoError(t, err)
}

func TestDeactivate_NotFound(t *testing.T) {
	svc, sr, pr, _ := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(personalPrompt(42), nil)
	sr.On("Deactivate", ctx, uint(1)).Return(repo.ErrNotFound)

	err := svc.Deactivate(ctx, 1, 42)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestGetPublicPrompt(t *testing.T) {
	svc, sr, _, _ := setupService()
	ctx := context.Background()

	sr.On("GetByToken", ctx, "ps_abc").Return(&models.ShareLink{
		ID: 1, Token: "ps_abc", IsActive: true,
		Prompt: models.Prompt{
			ID: 1, Title: "Test", Content: "Hello", Model: "claude",
			Tags: []models.Tag{{Name: "go", Color: "#00ff00"}},
			User: models.User{Name: "Иван", AvatarURL: "https://avatar.test"},
		},
	}, nil)
	sr.On("IncrementViewCount", mock.Anything, uint(1)).Return(nil)

	info, err := svc.GetPublicPrompt(ctx, "ps_abc")
	require.NoError(t, err)
	assert.Equal(t, "Test", info.Title)
	assert.Equal(t, "Hello", info.Content)
	assert.Equal(t, "Иван", info.Author.Name)
	assert.Len(t, info.Tags, 1)
}

func TestGetPublicPrompt_NotFound(t *testing.T) {
	svc, sr, _, _ := setupService()
	ctx := context.Background()

	sr.On("GetByToken", ctx, "ps_invalid").Return(nil, repo.ErrNotFound)

	_, err := svc.GetPublicPrompt(ctx, "ps_invalid")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestGetByPromptID(t *testing.T) {
	svc, sr, pr, _ := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(personalPrompt(42), nil)
	sr.On("GetActiveByPromptID", ctx, uint(1)).Return(&models.ShareLink{
		ID: 10, Token: "ps_test", IsActive: true, ViewCount: 5,
	}, nil)

	info, err := svc.GetByPromptID(ctx, 1, 42)
	require.NoError(t, err)
	assert.Equal(t, 5, info.ViewCount)
	assert.Equal(t, "https://app.test.ru/s/ps_test", info.URL)
}

func TestGetByPromptID_NoLink(t *testing.T) {
	svc, sr, pr, _ := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(personalPrompt(42), nil)
	sr.On("GetActiveByPromptID", ctx, uint(1)).Return(nil, repo.ErrNotFound)

	_, err := svc.GetByPromptID(ctx, 1, 42)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestGetByPromptID_PromptNotFound(t *testing.T) {
	svc, _, pr, _ := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(99)).Return(nil, repo.ErrNotFound)

	_, err := svc.GetByPromptID(ctx, 99, 42)
	assert.ErrorIs(t, err, ErrPromptNotFound)
}

// --- team authorization tests ---

func teamPrompt(ownerID, teamID uint) *models.Prompt {
	tid := teamID
	return &models.Prompt{ID: 1, UserID: ownerID, TeamID: &tid, Title: "Team Test", Content: "Hello"}
}

func TestCreateOrGet_TeamEditor(t *testing.T) {
	svc, sr, pr, tr := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(teamPrompt(99, 5), nil)
	tr.On("GetMember", ctx, uint(5), uint(42)).Return(&models.TeamMember{
		UserID: 42, TeamID: 5, Role: models.RoleEditor,
	}, nil)
	sr.On("GetActiveByPromptID", ctx, uint(1)).Return(nil, repo.ErrNotFound)
	sr.On("Create", ctx, mock.Anything).Return(nil)

	info, created, err := svc.CreateOrGet(ctx, 1, 42)
	require.NoError(t, err)
	assert.True(t, created)
	assert.Contains(t, info.Token, "ps_")
}

func TestCreateOrGet_TeamViewer(t *testing.T) {
	svc, _, pr, tr := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(teamPrompt(99, 5), nil)
	tr.On("GetMember", ctx, uint(5), uint(42)).Return(&models.TeamMember{
		UserID: 42, TeamID: 5, Role: models.RoleViewer,
	}, nil)

	_, _, err := svc.CreateOrGet(ctx, 1, 42)
	assert.ErrorIs(t, err, ErrViewerReadOnly)
}

func TestCreateOrGet_TeamNonMember(t *testing.T) {
	svc, _, pr, tr := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(teamPrompt(99, 5), nil)
	tr.On("GetMember", ctx, uint(5), uint(42)).Return(nil, repo.ErrNotFound)

	_, _, err := svc.CreateOrGet(ctx, 1, 42)
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestDeactivate_TeamEditor(t *testing.T) {
	svc, sr, pr, tr := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(teamPrompt(99, 5), nil)
	tr.On("GetMember", ctx, uint(5), uint(42)).Return(&models.TeamMember{
		UserID: 42, TeamID: 5, Role: models.RoleEditor,
	}, nil)
	sr.On("Deactivate", ctx, uint(1)).Return(nil)

	err := svc.Deactivate(ctx, 1, 42)
	assert.NoError(t, err)
}

func TestDeactivate_TeamViewer(t *testing.T) {
	svc, _, pr, tr := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(teamPrompt(99, 5), nil)
	tr.On("GetMember", ctx, uint(5), uint(42)).Return(&models.TeamMember{
		UserID: 42, TeamID: 5, Role: models.RoleViewer,
	}, nil)

	err := svc.Deactivate(ctx, 1, 42)
	assert.ErrorIs(t, err, ErrViewerReadOnly)
}

func TestGetByPromptID_TeamMember(t *testing.T) {
	svc, sr, pr, tr := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(teamPrompt(99, 5), nil)
	tr.On("GetMember", ctx, uint(5), uint(42)).Return(&models.TeamMember{
		UserID: 42, TeamID: 5, Role: models.RoleViewer,
	}, nil)
	sr.On("GetActiveByPromptID", ctx, uint(1)).Return(&models.ShareLink{
		ID: 10, Token: "ps_team", IsActive: true,
	}, nil)

	info, err := svc.GetByPromptID(ctx, 1, 42)
	require.NoError(t, err)
	assert.Equal(t, "ps_team", info.Token)
}

func TestGetByPromptID_TeamNonMember(t *testing.T) {
	svc, _, pr, tr := setupService()
	ctx := context.Background()

	pr.On("GetByID", ctx, uint(1)).Return(teamPrompt(99, 5), nil)
	tr.On("GetMember", ctx, uint(5), uint(42)).Return(nil, repo.ErrNotFound)

	_, err := svc.GetByPromptID(ctx, 1, 42)
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestGetPublicPrompt_SoftDeletedPrompt(t *testing.T) {
	svc, sr, _, _ := setupService()
	ctx := context.Background()

	// Prompt.ID == 0 means GORM preload found no matching (non-deleted) prompt
	sr.On("GetByToken", ctx, "ps_deleted").Return(&models.ShareLink{
		ID: 1, Token: "ps_deleted", IsActive: true,
		Prompt: models.Prompt{}, // zero-value, ID == 0
	}, nil)

	_, err := svc.GetPublicPrompt(ctx, "ps_deleted")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestGenerateToken_Unique(t *testing.T) {
	tok1, err := generateToken()
	require.NoError(t, err)
	tok2, err := generateToken()
	require.NoError(t, err)
	assert.NotEqual(t, tok1, tok2)
}
