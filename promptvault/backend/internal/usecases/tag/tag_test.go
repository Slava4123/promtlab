package tag

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// ===================== Mocks =====================

type mockTagRepo struct{ mock.Mock }

func (m *mockTagRepo) GetOrCreate(ctx context.Context, name, color string, userID uint, teamID *uint) (*models.Tag, error) {
	args := m.Called(ctx, name, color, userID, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tag), args.Error(1)
}
func (m *mockTagRepo) List(ctx context.Context, userID uint, teamID *uint) ([]models.Tag, error) {
	args := m.Called(ctx, userID, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Tag), args.Error(1)
}
func (m *mockTagRepo) GetByID(ctx context.Context, id uint) (*models.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tag), args.Error(1)
}
func (m *mockTagRepo) GetByIDs(ctx context.Context, ids []uint) ([]models.Tag, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Tag), args.Error(1)
}
func (m *mockTagRepo) SuggestByPrefix(ctx context.Context, userID uint, teamID *uint, prefix string, limit int) ([]string, error) {
	args := m.Called(ctx, userID, teamID, prefix, limit)
	return args.Get(0).([]string), args.Error(1)
}

// --- TeamRepository mock ---

type mockTeamRepo struct{ mock.Mock }

func (m *mockTeamRepo) CreateWithOwner(ctx context.Context, team *models.Team, ownerUserID uint) error {
	return m.Called(ctx, team, ownerUserID).Error(0)
}
func (m *mockTeamRepo) GetBySlug(ctx context.Context, slug string) (*models.Team, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}
func (m *mockTeamRepo) ListByUserID(ctx context.Context, userID uint) ([]models.Team, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Team), args.Error(1)
}
func (m *mockTeamRepo) ListByUserIDWithRolesAndCounts(ctx context.Context, userID uint) ([]models.TeamWithRoleAndCount, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamWithRoleAndCount), args.Error(1)
}
func (m *mockTeamRepo) Update(ctx context.Context, team *models.Team) error {
	return m.Called(ctx, team).Error(0)
}
func (m *mockTeamRepo) Delete(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockTeamRepo) GetMember(ctx context.Context, teamID, userID uint) (*models.TeamMember, error) {
	args := m.Called(ctx, teamID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TeamMember), args.Error(1)
}
func (m *mockTeamRepo) UpdateMemberRole(ctx context.Context, teamID, userID uint, role models.TeamRole) error {
	return m.Called(ctx, teamID, userID, role).Error(0)
}
func (m *mockTeamRepo) RemoveMember(ctx context.Context, teamID, userID uint) error {
	return m.Called(ctx, teamID, userID).Error(0)
}
func (m *mockTeamRepo) ListMembers(ctx context.Context, teamID uint) ([]models.TeamMember, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMember), args.Error(1)
}
func (m *mockTeamRepo) CountMembers(ctx context.Context, teamID uint) (int, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).(int), args.Error(1)
}
func (m *mockTeamRepo) CreateInvitation(ctx context.Context, inv *models.TeamInvitation) error {
	return m.Called(ctx, inv).Error(0)
}
func (m *mockTeamRepo) GetInvitationByID(ctx context.Context, id uint) (*models.TeamInvitation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TeamInvitation), args.Error(1)
}
func (m *mockTeamRepo) GetPendingInvitation(ctx context.Context, teamID, userID uint) (*models.TeamInvitation, error) {
	args := m.Called(ctx, teamID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TeamInvitation), args.Error(1)
}
func (m *mockTeamRepo) ListPendingByUserID(ctx context.Context, userID uint) ([]models.TeamInvitation, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamInvitation), args.Error(1)
}
func (m *mockTeamRepo) ListPendingByTeamID(ctx context.Context, teamID uint) ([]models.TeamInvitation, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamInvitation), args.Error(1)
}
func (m *mockTeamRepo) UpdateInvitationStatus(ctx context.Context, id uint, status models.InvitationStatus) error {
	return m.Called(ctx, id, status).Error(0)
}
func (m *mockTeamRepo) DeleteInvitation(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockTeamRepo) AcceptInvitationTx(ctx context.Context, invID uint, member *models.TeamMember) error {
	return m.Called(ctx, invID, member).Error(0)
}

// ===================== Helpers =====================

func newTestService() (*Service, *mockTagRepo, *mockTeamRepo) {
	tr := new(mockTagRepo)
	tmr := new(mockTeamRepo)
	svc := NewService(tr, tmr)
	return svc, tr, tmr
}

func uint_ptr(v uint) *uint { return &v }

// ===================== Create =====================

func TestCreate_PersonalSuccess(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	expected := &models.Tag{ID: 1, Name: "golang", Color: "#ff0000", UserID: 10}
	tr.On("GetOrCreate", ctx, "golang", "#ff0000", uint(10), (*uint)(nil)).Return(expected, nil)

	tag, err := svc.Create(ctx, "golang", "#ff0000", 10, nil)

	assert.NoError(t, err)
	assert.Equal(t, "golang", tag.Name)
	assert.Equal(t, "#ff0000", tag.Color)
}

func TestCreate_DefaultColor(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	expected := &models.Tag{ID: 1, Name: "test", Color: "#6366f1", UserID: 10}
	tr.On("GetOrCreate", ctx, "test", "#6366f1", uint(10), (*uint)(nil)).Return(expected, nil)

	tag, err := svc.Create(ctx, "test", "", 10, nil)

	assert.NoError(t, err)
	assert.Equal(t, "#6366f1", tag.Color)
}

func TestCreate_EmptyNameError(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.Create(ctx, "", "", 10, nil)

	assert.ErrorIs(t, err, ErrNameEmpty)
}

func TestCreate_WhitespaceOnlyNameError(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.Create(ctx, "   ", "", 10, nil)

	assert.ErrorIs(t, err, ErrNameEmpty)
}

func TestCreate_TeamEditorSuccess(t *testing.T) {
	svc, tr, tmr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	tmr.On("GetMember", ctx, uint(5), uint(10)).
		Return(&models.TeamMember{Role: models.RoleEditor}, nil)

	expected := &models.Tag{ID: 1, Name: "api", Color: "#6366f1", UserID: 10, TeamID: teamID}
	tr.On("GetOrCreate", ctx, "api", "#6366f1", uint(10), teamID).Return(expected, nil)

	tag, err := svc.Create(ctx, "api", "", 10, teamID)

	assert.NoError(t, err)
	assert.Equal(t, teamID, tag.TeamID)
}

func TestCreate_TeamViewerForbidden(t *testing.T) {
	svc, _, tmr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	tmr.On("GetMember", ctx, uint(5), uint(10)).
		Return(&models.TeamMember{Role: models.RoleViewer}, nil)

	_, err := svc.Create(ctx, "api", "", 10, teamID)

	assert.ErrorIs(t, err, ErrViewerReadOnly)
}

func TestCreate_TeamNotMember(t *testing.T) {
	svc, _, tmr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	tmr.On("GetMember", ctx, uint(5), uint(10)).
		Return(nil, repo.ErrNotFound)

	_, err := svc.Create(ctx, "api", "", 10, teamID)

	assert.ErrorIs(t, err, ErrForbidden)
}

// ===================== Delete =====================

func TestDelete_PersonalOwner(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	tr.On("GetByID", ctx, uint(1)).Return(&models.Tag{
		ID: 1, UserID: 10, Name: "golang",
	}, nil)
	tr.On("Delete", ctx, uint(1)).Return(nil)

	err := svc.Delete(ctx, 1, 10)

	assert.NoError(t, err)
	tr.AssertCalled(t, "Delete", ctx, uint(1))
}

func TestDelete_PersonalOtherUserForbidden(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	tr.On("GetByID", ctx, uint(1)).Return(&models.Tag{
		ID: 1, UserID: 10, Name: "golang",
	}, nil)

	err := svc.Delete(ctx, 1, 999)

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestDelete_NotFound(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	tr.On("GetByID", ctx, uint(99)).Return(nil, repo.ErrNotFound)

	err := svc.Delete(ctx, 99, 10)

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDelete_TeamEditorSuccess(t *testing.T) {
	svc, tr, tmr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	tr.On("GetByID", ctx, uint(1)).Return(&models.Tag{
		ID: 1, UserID: 10, TeamID: teamID, Name: "api",
	}, nil)
	tmr.On("GetMember", ctx, uint(5), uint(20)).
		Return(&models.TeamMember{Role: models.RoleEditor}, nil)
	tr.On("Delete", ctx, uint(1)).Return(nil)

	err := svc.Delete(ctx, 1, 20)

	assert.NoError(t, err)
}

func TestDelete_TeamViewerForbidden(t *testing.T) {
	svc, tr, tmr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	tr.On("GetByID", ctx, uint(1)).Return(&models.Tag{
		ID: 1, UserID: 10, TeamID: teamID, Name: "api",
	}, nil)
	tmr.On("GetMember", ctx, uint(5), uint(20)).
		Return(&models.TeamMember{Role: models.RoleViewer}, nil)

	err := svc.Delete(ctx, 1, 20)

	assert.ErrorIs(t, err, ErrViewerReadOnly)
}

func TestDelete_TeamNotMember(t *testing.T) {
	svc, tr, tmr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	tr.On("GetByID", ctx, uint(1)).Return(&models.Tag{
		ID: 1, UserID: 10, TeamID: teamID, Name: "api",
	}, nil)
	tmr.On("GetMember", ctx, uint(5), uint(999)).
		Return(nil, repo.ErrNotFound)

	err := svc.Delete(ctx, 1, 999)

	assert.ErrorIs(t, err, ErrForbidden)
}

// ===================== List =====================

func TestList_Success(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	expected := []models.Tag{
		{ID: 1, Name: "golang", UserID: 10},
		{ID: 2, Name: "python", UserID: 10},
	}
	tr.On("List", ctx, uint(10), (*uint)(nil)).Return(expected, nil)

	result, err := svc.List(ctx, 10, nil)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "golang", result[0].Name)
}

func TestList_WithTeamID(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	expected := []models.Tag{
		{ID: 3, Name: "api", UserID: 10, TeamID: teamID},
	}
	tr.On("List", ctx, uint(10), teamID).Return(expected, nil)

	result, err := svc.List(ctx, 10, teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestList_Empty(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	tr.On("List", ctx, uint(10), (*uint)(nil)).Return([]models.Tag{}, nil)

	result, err := svc.List(ctx, 10, nil)

	assert.NoError(t, err)
	assert.Empty(t, result)
}
