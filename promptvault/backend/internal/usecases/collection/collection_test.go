package collection

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// ===================== Mocks =====================

type mockCollectionRepo struct{ mock.Mock }

func (m *mockCollectionRepo) Create(ctx context.Context, c *models.Collection) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockCollectionRepo) GetByID(ctx context.Context, id uint) (*models.Collection, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Collection), args.Error(1)
}
func (m *mockCollectionRepo) ListWithCounts(ctx context.Context, userID uint, teamIDs []uint) ([]models.CollectionWithCount, error) {
	args := m.Called(ctx, userID, teamIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CollectionWithCount), args.Error(1)
}
func (m *mockCollectionRepo) SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Collection, error) {
	args := m.Called(ctx, userID, teamID, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Collection), args.Error(1)
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

func newTestService() (*Service, *mockCollectionRepo, *mockTeamRepo) {
	cr := new(mockCollectionRepo)
	tr := new(mockTeamRepo)
	svc := NewService(cr, tr)
	return svc, cr, tr
}

func uint_ptr(v uint) *uint { return &v }

// ===================== Create =====================

func TestCreate_PersonalSuccess(t *testing.T) {
	svc, cr, _ := newTestService()
	ctx := context.Background()

	cr.On("Create", ctx, mock.AnythingOfType("*models.Collection")).Return(nil)

	c, err := svc.Create(ctx, 10, "Моя коллекция", "Описание", "#ff0000", "📝", nil)

	assert.NoError(t, err)
	assert.Equal(t, "Моя коллекция", c.Name)
	assert.Equal(t, "#ff0000", c.Color)
	assert.Equal(t, uint(10), c.UserID)
	assert.Nil(t, c.TeamID)
	cr.AssertCalled(t, "Create", ctx, mock.Anything)
}

func TestCreate_DefaultColor(t *testing.T) {
	svc, cr, _ := newTestService()
	ctx := context.Background()

	cr.On("Create", ctx, mock.AnythingOfType("*models.Collection")).Return(nil)

	c, err := svc.Create(ctx, 10, "Тест", "", "", "", nil)

	assert.NoError(t, err)
	assert.Equal(t, "#8b5cf6", c.Color)
}

func TestCreate_TeamEditorSuccess(t *testing.T) {
	svc, cr, tr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	tr.On("GetMember", ctx, uint(5), uint(10)).
		Return(&models.TeamMember{Role: models.RoleEditor}, nil)
	cr.On("Create", ctx, mock.AnythingOfType("*models.Collection")).Return(nil)

	c, err := svc.Create(ctx, 10, "Командная", "Описание", "", "", teamID)

	assert.NoError(t, err)
	assert.Equal(t, teamID, c.TeamID)
	assert.Equal(t, "#8b5cf6", c.Color) // default color
}

func TestCreate_TeamViewerForbidden(t *testing.T) {
	svc, _, tr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	tr.On("GetMember", ctx, uint(5), uint(10)).
		Return(&models.TeamMember{Role: models.RoleViewer}, nil)

	_, err := svc.Create(ctx, 10, "Командная", "", "", "", teamID)

	assert.ErrorIs(t, err, ErrViewerReadOnly)
}

func TestCreate_TeamNotMember(t *testing.T) {
	svc, _, tr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	tr.On("GetMember", ctx, uint(5), uint(10)).
		Return(nil, repo.ErrNotFound)

	_, err := svc.Create(ctx, 10, "Командная", "", "", "", teamID)

	assert.ErrorIs(t, err, ErrForbidden)
}

// ===================== GetByID =====================

func TestGetByID_PersonalOwner(t *testing.T) {
	svc, cr, _ := newTestService()
	ctx := context.Background()

	cr.On("GetByID", ctx, uint(1)).Return(&models.Collection{
		ID: 1, UserID: 10, Name: "Личная",
	}, nil)

	c, err := svc.GetByID(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, "Личная", c.Name)
}

func TestGetByID_PersonalOtherUserForbidden(t *testing.T) {
	svc, cr, _ := newTestService()
	ctx := context.Background()

	cr.On("GetByID", ctx, uint(1)).Return(&models.Collection{
		ID: 1, UserID: 10, Name: "Личная",
	}, nil)

	_, err := svc.GetByID(ctx, 1, 999)

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestGetByID_NotFound(t *testing.T) {
	svc, cr, _ := newTestService()
	ctx := context.Background()

	cr.On("GetByID", ctx, uint(99)).Return(nil, repo.ErrNotFound)

	_, err := svc.GetByID(ctx, 99, 10)

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestGetByID_TeamMember(t *testing.T) {
	svc, cr, tr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	cr.On("GetByID", ctx, uint(1)).Return(&models.Collection{
		ID: 1, UserID: 10, TeamID: teamID, Name: "Командная",
	}, nil)
	tr.On("GetMember", ctx, uint(5), uint(20)).
		Return(&models.TeamMember{Role: models.RoleViewer}, nil)

	c, err := svc.GetByID(ctx, 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, "Командная", c.Name)
}

func TestGetByID_TeamNonMember(t *testing.T) {
	svc, cr, tr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	cr.On("GetByID", ctx, uint(1)).Return(&models.Collection{
		ID: 1, UserID: 10, TeamID: teamID, Name: "Командная",
	}, nil)
	tr.On("GetMember", ctx, uint(5), uint(999)).
		Return(nil, repo.ErrNotFound)

	_, err := svc.GetByID(ctx, 1, 999)

	assert.ErrorIs(t, err, ErrForbidden)
}

// ===================== Update =====================

func TestUpdate_Success(t *testing.T) {
	svc, cr, _ := newTestService()
	ctx := context.Background()

	cr.On("GetByID", ctx, uint(1)).Return(&models.Collection{
		ID: 1, UserID: 10, Name: "Старое", Color: "#000000",
	}, nil)
	cr.On("Update", ctx, mock.AnythingOfType("*models.Collection")).Return(nil)

	c, err := svc.Update(ctx, 1, 10, "Новое", "Описание", "#ff0000", "📝")

	assert.NoError(t, err)
	assert.Equal(t, "Новое", c.Name)
	assert.Equal(t, "#ff0000", c.Color)
	assert.Equal(t, "📝", c.Icon)
}

func TestUpdate_TeamViewerForbidden(t *testing.T) {
	svc, cr, tr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	// GetByID succeeds — user is a team member (viewer can read)
	cr.On("GetByID", ctx, uint(1)).Return(&models.Collection{
		ID: 1, UserID: 10, TeamID: teamID, Name: "Командная",
	}, nil)
	// GetMember called twice: once for GetByID check, once for checkTeamEditAccess
	tr.On("GetMember", ctx, uint(5), uint(20)).
		Return(&models.TeamMember{Role: models.RoleViewer}, nil)

	_, err := svc.Update(ctx, 1, 20, "Новое", "", "", "")

	assert.ErrorIs(t, err, ErrViewerReadOnly)
}

func TestUpdate_NotFound(t *testing.T) {
	svc, cr, _ := newTestService()
	ctx := context.Background()

	cr.On("GetByID", ctx, uint(99)).Return(nil, repo.ErrNotFound)

	_, err := svc.Update(ctx, 99, 10, "Новое", "", "", "")

	assert.ErrorIs(t, err, ErrNotFound)
}

// ===================== Delete =====================

func TestDelete_PersonalSuccess(t *testing.T) {
	svc, cr, _ := newTestService()
	ctx := context.Background()

	cr.On("GetByID", ctx, uint(1)).Return(&models.Collection{
		ID: 1, UserID: 10, Name: "Личная",
	}, nil)
	cr.On("Delete", ctx, uint(1)).Return(nil)

	err := svc.Delete(ctx, 1, 10)

	assert.NoError(t, err)
	cr.AssertCalled(t, "Delete", ctx, uint(1))
}

func TestDelete_TeamViewerForbidden(t *testing.T) {
	svc, cr, tr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	cr.On("GetByID", ctx, uint(1)).Return(&models.Collection{
		ID: 1, UserID: 10, TeamID: teamID, Name: "Командная",
	}, nil)
	// GetMember called twice: once in GetByID, once in checkTeamEditAccess
	tr.On("GetMember", ctx, uint(5), uint(20)).
		Return(&models.TeamMember{Role: models.RoleViewer}, nil)

	err := svc.Delete(ctx, 1, 20)

	assert.ErrorIs(t, err, ErrViewerReadOnly)
}

func TestDelete_TeamEditorSuccess(t *testing.T) {
	svc, cr, tr := newTestService()
	ctx := context.Background()
	teamID := uint_ptr(5)

	cr.On("GetByID", ctx, uint(1)).Return(&models.Collection{
		ID: 1, UserID: 10, TeamID: teamID, Name: "Командная",
	}, nil)
	tr.On("GetMember", ctx, uint(5), uint(20)).
		Return(&models.TeamMember{Role: models.RoleEditor}, nil)
	cr.On("Delete", ctx, uint(1)).Return(nil)

	err := svc.Delete(ctx, 1, 20)

	assert.NoError(t, err)
}

func TestDelete_NotFound(t *testing.T) {
	svc, cr, _ := newTestService()
	ctx := context.Background()

	cr.On("GetByID", ctx, uint(99)).Return(nil, repo.ErrNotFound)

	err := svc.Delete(ctx, 99, 10)

	assert.ErrorIs(t, err, ErrNotFound)
}

// ===================== List =====================

func TestList_Personal(t *testing.T) {
	svc, cr, _ := newTestService()
	ctx := context.Background()

	expected := []models.CollectionWithCount{
		{Collection: models.Collection{ID: 1, UserID: 10, Name: "A"}, PromptCount: 3},
		{Collection: models.Collection{ID: 2, UserID: 10, Name: "B"}, PromptCount: 1},
	}
	cr.On("ListWithCounts", ctx, uint(10), []uint(nil)).Return(expected, nil)

	result, err := svc.List(ctx, 10, nil)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "A", result[0].Name)
}

func TestList_TeamWithMembership(t *testing.T) {
	svc, cr, tr := newTestService()
	ctx := context.Background()
	teamIDs := []uint{5}

	tr.On("GetMember", ctx, uint(5), uint(10)).
		Return(&models.TeamMember{Role: models.RoleEditor}, nil)

	expected := []models.CollectionWithCount{
		{Collection: models.Collection{ID: 3, UserID: 10, TeamID: uint_ptr(5), Name: "Командная"}, PromptCount: 5},
	}
	cr.On("ListWithCounts", ctx, uint(10), teamIDs).Return(expected, nil)

	result, err := svc.List(ctx, 10, teamIDs)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestList_TeamNotMember(t *testing.T) {
	svc, _, tr := newTestService()
	ctx := context.Background()
	teamIDs := []uint{5}

	tr.On("GetMember", ctx, uint(5), uint(999)).
		Return(nil, repo.ErrNotFound)

	_, err := svc.List(ctx, 999, teamIDs)

	assert.ErrorIs(t, err, ErrForbidden)
}
