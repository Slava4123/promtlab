package team

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// ===================== Mocks =====================

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

// ===================== Helpers =====================

func newTestService() (*Service, *mockTeamRepo, *mockUserRepo) {
	tr := new(mockTeamRepo)
	ur := new(mockUserRepo)
	svc := NewService(tr, ur)
	return svc, tr, ur
}

// ===================== Create =====================

func TestCreate_Success(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	tr.On("CreateWithOwner", ctx, mock.AnythingOfType("*models.Team"), uint(1)).Return(nil)

	team, err := svc.Create(ctx, 1, CreateInput{Name: "My Team", Description: "desc"})

	assert.NoError(t, err)
	assert.Equal(t, "My Team", team.Name)
	assert.Equal(t, "desc", team.Description)
	assert.Equal(t, uint(1), team.CreatedBy)
	assert.NotEmpty(t, team.Slug)
	tr.AssertNumberOfCalls(t, "CreateWithOwner", 1)
}

func TestCreate_SlugCollisionRetry(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	// First two calls fail (slug collision), third succeeds
	tr.On("CreateWithOwner", ctx, mock.AnythingOfType("*models.Team"), uint(1)).
		Return(errors.New("unique constraint violation")).Once()
	tr.On("CreateWithOwner", ctx, mock.AnythingOfType("*models.Team"), uint(1)).
		Return(errors.New("unique constraint violation")).Once()
	tr.On("CreateWithOwner", ctx, mock.AnythingOfType("*models.Team"), uint(1)).
		Return(nil).Once()

	team, err := svc.Create(ctx, 1, CreateInput{Name: "My Team"})

	assert.NoError(t, err)
	assert.NotNil(t, team)
	tr.AssertNumberOfCalls(t, "CreateWithOwner", 3)
}

func TestCreate_AllRetriesFail(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	dbErr := errors.New("unique constraint violation")
	tr.On("CreateWithOwner", ctx, mock.AnythingOfType("*models.Team"), uint(1)).
		Return(dbErr)

	team, err := svc.Create(ctx, 1, CreateInput{Name: "My Team"})

	assert.Nil(t, team)
	assert.Error(t, err)
	assert.Equal(t, dbErr, err)
	tr.AssertNumberOfCalls(t, "CreateWithOwner", 3)
}

// ===================== GetBySlug (checkAccess — viewer) =====================

func TestGetBySlug_OwnerAllowed(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team", Name: "My Team"}
	member := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}
	members := []models.TeamMember{*member}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(member, nil)
	tr.On("ListMembers", ctx, uint(10)).Return(members, nil)

	result, resultMembers, err := svc.GetBySlug(ctx, "my-team", 1)

	assert.NoError(t, err)
	assert.Equal(t, "My Team", result.Name)
	assert.Len(t, resultMembers, 1)
}

func TestGetBySlug_ViewerAllowed(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team", Name: "My Team"}
	member := &models.TeamMember{TeamID: 10, UserID: 2, Role: models.RoleViewer}
	members := []models.TeamMember{*member}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(2)).Return(member, nil)
	tr.On("ListMembers", ctx, uint(10)).Return(members, nil)

	result, _, err := svc.GetBySlug(ctx, "my-team", 2)

	assert.NoError(t, err)
	assert.Equal(t, "My Team", result.Name)
}

func TestGetBySlug_NonMember_ReturnsForbidden(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(99)).Return(nil, repo.ErrNotFound)

	_, _, err := svc.GetBySlug(ctx, "my-team", 99)

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestGetBySlug_TeamNotFound(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	tr.On("GetBySlug", ctx, "nope").Return(nil, repo.ErrNotFound)

	_, _, err := svc.GetBySlug(ctx, "nope", 1)

	assert.ErrorIs(t, err, ErrNotFound)
}

// ===================== Update (checkAccess — owner) =====================

func TestUpdate_OwnerAllowed(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team", Name: "Old Name", Description: "Old Desc"}
	member := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(member, nil)
	tr.On("Update", ctx, mock.AnythingOfType("*models.Team")).Return(nil)

	newName := "New Name"
	result, err := svc.Update(ctx, "my-team", 1, UpdateInput{Name: &newName})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
	assert.Equal(t, "Old Desc", result.Description) // unchanged
}

func TestUpdate_ViewerBlocked(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	member := &models.TeamMember{TeamID: 10, UserID: 2, Role: models.RoleViewer}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(2)).Return(member, nil)

	newName := "Hacked"
	_, err := svc.Update(ctx, "my-team", 2, UpdateInput{Name: &newName})

	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestUpdate_EditorBlocked(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	member := &models.TeamMember{TeamID: 10, UserID: 3, Role: models.RoleEditor}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(3)).Return(member, nil)

	newName := "Hacked"
	_, err := svc.Update(ctx, "my-team", 3, UpdateInput{Name: &newName})

	assert.ErrorIs(t, err, ErrNotOwner)
}

// ===================== Delete (checkAccess — owner) =====================

func TestDelete_OwnerSuccess(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	member := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(member, nil)
	tr.On("Delete", ctx, uint(10)).Return(nil)

	err := svc.Delete(ctx, "my-team", 1)

	assert.NoError(t, err)
	tr.AssertCalled(t, "Delete", ctx, uint(10))
}

func TestDelete_ViewerBlocked(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	member := &models.TeamMember{TeamID: 10, UserID: 2, Role: models.RoleViewer}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(2)).Return(member, nil)

	err := svc.Delete(ctx, "my-team", 2)

	assert.ErrorIs(t, err, ErrNotOwner)
}

// ===================== InviteMember =====================

func TestInviteMember_Success(t *testing.T) {
	svc, tr, ur := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team", Name: "My Team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}
	targetUser := &models.User{ID: 5, Email: "target@example.com", Name: "Target"}
	inviter := &models.User{ID: 1, Email: "owner@example.com", Name: "Owner"}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)
	ur.On("GetByEmail", ctx, "target@example.com").Return(targetUser, nil)
	tr.On("GetMember", ctx, uint(10), uint(5)).Return(nil, repo.ErrNotFound)
	tr.On("GetPendingInvitation", ctx, uint(10), uint(5)).Return(nil, repo.ErrNotFound)
	tr.On("CreateInvitation", ctx, mock.AnythingOfType("*models.TeamInvitation")).Return(nil)
	ur.On("GetByID", ctx, uint(1)).Return(inviter, nil)

	inv, err := svc.InviteMember(ctx, "my-team", 1, AddMemberInput{
		Query: "target@example.com",
		Role:  models.RoleEditor,
	})

	assert.NoError(t, err)
	assert.NotNil(t, inv)
	assert.Equal(t, uint(10), inv.TeamID)
	assert.Equal(t, uint(5), inv.UserID)
	assert.Equal(t, models.RoleEditor, inv.Role)
	assert.Equal(t, models.InvitationPending, inv.Status)
	assert.Equal(t, "My Team", inv.Team.Name)
	assert.Equal(t, "Owner", inv.Inviter.Name)
}

func TestInviteMember_CannotInviteSelf(t *testing.T) {
	svc, tr, ur := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}
	selfUser := &models.User{ID: 1, Email: "owner@example.com"}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)
	ur.On("GetByEmail", ctx, "owner@example.com").Return(selfUser, nil)

	_, err := svc.InviteMember(ctx, "my-team", 1, AddMemberInput{
		Query: "owner@example.com",
		Role:  models.RoleEditor,
	})

	assert.ErrorIs(t, err, ErrCannotInviteSelf)
}

func TestInviteMember_AlreadyMember(t *testing.T) {
	svc, tr, ur := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}
	targetUser := &models.User{ID: 5, Email: "target@example.com"}
	existingMember := &models.TeamMember{TeamID: 10, UserID: 5, Role: models.RoleViewer}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)
	ur.On("GetByEmail", ctx, "target@example.com").Return(targetUser, nil)
	tr.On("GetMember", ctx, uint(10), uint(5)).Return(existingMember, nil)

	_, err := svc.InviteMember(ctx, "my-team", 1, AddMemberInput{
		Query: "target@example.com",
		Role:  models.RoleEditor,
	})

	assert.ErrorIs(t, err, ErrAlreadyMember)
}

func TestInviteMember_AlreadyInvited(t *testing.T) {
	svc, tr, ur := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}
	targetUser := &models.User{ID: 5, Email: "target@example.com"}
	existingInv := &models.TeamInvitation{ID: 100, TeamID: 10, UserID: 5, Status: models.InvitationPending}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)
	ur.On("GetByEmail", ctx, "target@example.com").Return(targetUser, nil)
	tr.On("GetMember", ctx, uint(10), uint(5)).Return(nil, repo.ErrNotFound)
	tr.On("GetPendingInvitation", ctx, uint(10), uint(5)).Return(existingInv, nil)

	_, err := svc.InviteMember(ctx, "my-team", 1, AddMemberInput{
		Query: "target@example.com",
		Role:  models.RoleEditor,
	})

	assert.ErrorIs(t, err, ErrAlreadyInvited)
}

func TestInviteMember_CannotAssignOwnerRole(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)

	_, err := svc.InviteMember(ctx, "my-team", 1, AddMemberInput{
		Query: "target@example.com",
		Role:  models.RoleOwner,
	})

	assert.ErrorIs(t, err, ErrInvalidRole)
}

func TestInviteMember_UserNotFound(t *testing.T) {
	svc, tr, ur := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)
	ur.On("GetByEmail", ctx, "nobody@example.com").Return(nil, repo.ErrNotFound)

	_, err := svc.InviteMember(ctx, "my-team", 1, AddMemberInput{
		Query: "nobody@example.com",
		Role:  models.RoleEditor,
	})

	assert.ErrorIs(t, err, ErrUserNotFound)
}

// ===================== AcceptInvitation =====================

func TestAcceptInvitation_Success(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	inv := &models.TeamInvitation{
		ID:     100,
		TeamID: 10,
		UserID: 5,
		Role:   models.RoleEditor,
		Status: models.InvitationPending,
	}

	tr.On("GetInvitationByID", ctx, uint(100)).Return(inv, nil)
	tr.On("GetMember", ctx, uint(10), uint(5)).Return(nil, repo.ErrNotFound)
	tr.On("AcceptInvitationTx", ctx, uint(100), mock.MatchedBy(func(m *models.TeamMember) bool {
		return m.TeamID == 10 && m.UserID == 5 && m.Role == models.RoleEditor
	})).Return(nil)

	err := svc.AcceptInvitation(ctx, 100, 5)

	assert.NoError(t, err)
	tr.AssertCalled(t, "AcceptInvitationTx", ctx, uint(100), mock.Anything)
}

func TestAcceptInvitation_WrongUser_Forbidden(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	inv := &models.TeamInvitation{
		ID:     100,
		TeamID: 10,
		UserID: 5,
		Status: models.InvitationPending,
	}

	tr.On("GetInvitationByID", ctx, uint(100)).Return(inv, nil)

	err := svc.AcceptInvitation(ctx, 100, 99) // userID 99 != inv.UserID 5

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestAcceptInvitation_AlreadyAccepted_ReturnsNotFound(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	inv := &models.TeamInvitation{
		ID:     100,
		TeamID: 10,
		UserID: 5,
		Status: models.InvitationAccepted, // not pending
	}

	tr.On("GetInvitationByID", ctx, uint(100)).Return(inv, nil)

	err := svc.AcceptInvitation(ctx, 100, 5)

	assert.ErrorIs(t, err, ErrInvitationNotFound)
}

func TestAcceptInvitation_InvitationDoesNotExist(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	tr.On("GetInvitationByID", ctx, uint(999)).Return(nil, repo.ErrNotFound)

	err := svc.AcceptInvitation(ctx, 999, 5)

	assert.ErrorIs(t, err, ErrInvitationNotFound)
}

func TestAcceptInvitation_AlreadyMember(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	inv := &models.TeamInvitation{
		ID:     100,
		TeamID: 10,
		UserID: 5,
		Status: models.InvitationPending,
	}
	existingMember := &models.TeamMember{TeamID: 10, UserID: 5, Role: models.RoleViewer}

	tr.On("GetInvitationByID", ctx, uint(100)).Return(inv, nil)
	tr.On("GetMember", ctx, uint(10), uint(5)).Return(existingMember, nil)

	err := svc.AcceptInvitation(ctx, 100, 5)

	assert.ErrorIs(t, err, ErrAlreadyMember)
}

// ===================== RemoveMember =====================

func TestRemoveMember_OwnerRemovesEditor(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}
	editorMember := &models.TeamMember{TeamID: 10, UserID: 5, Role: models.RoleEditor}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)
	tr.On("GetMember", ctx, uint(10), uint(5)).Return(editorMember, nil)
	tr.On("RemoveMember", ctx, uint(10), uint(5)).Return(nil)

	err := svc.RemoveMember(ctx, "my-team", 1, 5)

	assert.NoError(t, err)
	tr.AssertCalled(t, "RemoveMember", ctx, uint(10), uint(5))
}

func TestRemoveMember_EditorLeavesSelf(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	editorMember := &models.TeamMember{TeamID: 10, UserID: 5, Role: models.RoleEditor}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	// First GetMember call: checkAccess for userID=5
	tr.On("GetMember", ctx, uint(10), uint(5)).Return(editorMember, nil)
	tr.On("RemoveMember", ctx, uint(10), uint(5)).Return(nil)

	err := svc.RemoveMember(ctx, "my-team", 5, 5) // userID == targetUserID

	assert.NoError(t, err)
	tr.AssertCalled(t, "RemoveMember", ctx, uint(10), uint(5))
}

func TestRemoveMember_EditorCannotRemoveOther(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	editorMember := &models.TeamMember{TeamID: 10, UserID: 5, Role: models.RoleEditor}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(5)).Return(editorMember, nil)

	err := svc.RemoveMember(ctx, "my-team", 5, 6) // editor tries to remove userID=6

	assert.ErrorIs(t, err, ErrNotOwner)
}

func TestRemoveMember_CannotRemoveOwner(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)

	err := svc.RemoveMember(ctx, "my-team", 1, 1) // owner tries to remove self

	assert.ErrorIs(t, err, ErrCannotRemoveOwner)
}

// ===================== UpdateMemberRole =====================

func TestUpdateMemberRole_Success(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}
	viewerMember := &models.TeamMember{TeamID: 10, UserID: 5, Role: models.RoleViewer}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)
	tr.On("GetMember", ctx, uint(10), uint(5)).Return(viewerMember, nil)
	tr.On("UpdateMemberRole", ctx, uint(10), uint(5), models.RoleEditor).Return(nil)

	err := svc.UpdateMemberRole(ctx, "my-team", 1, 5, models.RoleEditor)

	assert.NoError(t, err)
	tr.AssertCalled(t, "UpdateMemberRole", ctx, uint(10), uint(5), models.RoleEditor)
}

func TestUpdateMemberRole_CannotChangeOwnerRole(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}
	otherOwner := &models.TeamMember{TeamID: 10, UserID: 2, Role: models.RoleOwner}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)
	tr.On("GetMember", ctx, uint(10), uint(2)).Return(otherOwner, nil)

	err := svc.UpdateMemberRole(ctx, "my-team", 1, 2, models.RoleEditor)

	assert.ErrorIs(t, err, ErrCannotChangeOwnerRole)
}

func TestUpdateMemberRole_CannotAssignOwnerRole(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)

	err := svc.UpdateMemberRole(ctx, "my-team", 1, 5, models.RoleOwner)

	assert.ErrorIs(t, err, ErrInvalidRole)
}

func TestUpdateMemberRole_TargetNotFound(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)
	tr.On("GetMember", ctx, uint(10), uint(99)).Return(nil, repo.ErrNotFound)

	err := svc.UpdateMemberRole(ctx, "my-team", 1, 99, models.RoleEditor)

	assert.ErrorIs(t, err, ErrUserNotFound)
}

// ===================== DeclineInvitation =====================

func TestDeclineInvitation_Success(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	inv := &models.TeamInvitation{
		ID:     100,
		TeamID: 10,
		UserID: 5,
		Status: models.InvitationPending,
	}

	tr.On("GetInvitationByID", ctx, uint(100)).Return(inv, nil)
	tr.On("UpdateInvitationStatus", ctx, uint(100), models.InvitationDeclined).Return(nil)

	err := svc.DeclineInvitation(ctx, 100, 5)

	assert.NoError(t, err)
	tr.AssertCalled(t, "UpdateInvitationStatus", ctx, uint(100), models.InvitationDeclined)
}

func TestDeclineInvitation_WrongUser(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	inv := &models.TeamInvitation{
		ID:     100,
		TeamID: 10,
		UserID: 5,
		Status: models.InvitationPending,
	}

	tr.On("GetInvitationByID", ctx, uint(100)).Return(inv, nil)

	err := svc.DeclineInvitation(ctx, 100, 99)

	assert.ErrorIs(t, err, ErrForbidden)
}

// ===================== CancelInvitation =====================

func TestCancelInvitation_Success(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}
	inv := &models.TeamInvitation{ID: 100, TeamID: 10, UserID: 5, Status: models.InvitationPending}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)
	tr.On("GetInvitationByID", ctx, uint(100)).Return(inv, nil)
	tr.On("DeleteInvitation", ctx, uint(100)).Return(nil)

	err := svc.CancelInvitation(ctx, "my-team", 1, 100)

	assert.NoError(t, err)
	tr.AssertCalled(t, "DeleteInvitation", ctx, uint(100))
}

func TestCancelInvitation_WrongTeam(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}
	inv := &models.TeamInvitation{ID: 100, TeamID: 99, UserID: 5, Status: models.InvitationPending} // different team

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)
	tr.On("GetInvitationByID", ctx, uint(100)).Return(inv, nil)

	err := svc.CancelInvitation(ctx, "my-team", 1, 100)

	assert.ErrorIs(t, err, ErrInvitationNotFound)
}

// ===================== List =====================

func TestList_Success(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	rows := []models.TeamWithRoleAndCount{
		{Team: models.Team{ID: 10, Name: "Team A"}, Role: models.RoleOwner, MemberCount: 3},
		{Team: models.Team{ID: 20, Name: "Team B"}, Role: models.RoleViewer, MemberCount: 5},
	}

	tr.On("ListByUserIDWithRolesAndCounts", ctx, uint(1)).Return(rows, nil)

	items, err := svc.List(ctx, 1)

	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, models.RoleOwner, items[0].Role)
	assert.Equal(t, 3, items[0].MemberCount)
	assert.Equal(t, models.RoleViewer, items[1].Role)
	assert.Equal(t, 5, items[1].MemberCount)
}

// ===================== ListMyInvitations =====================

func TestListMyInvitations_Success(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	invitations := []models.TeamInvitation{
		{ID: 1, TeamID: 10, UserID: 5},
		{ID: 2, TeamID: 20, UserID: 5},
	}
	tr.On("ListPendingByUserID", ctx, uint(5)).Return(invitations, nil)

	result, err := svc.ListMyInvitations(ctx, 5)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// ===================== ListTeamInvitations =====================

func TestListTeamInvitations_Success(t *testing.T) {
	svc, tr, _ := newTestService()
	ctx := context.Background()

	team := &models.Team{ID: 10, Slug: "my-team"}
	ownerMember := &models.TeamMember{TeamID: 10, UserID: 1, Role: models.RoleOwner}
	invitations := []models.TeamInvitation{
		{ID: 1, TeamID: 10, UserID: 5},
	}

	tr.On("GetBySlug", ctx, "my-team").Return(team, nil)
	tr.On("GetMember", ctx, uint(10), uint(1)).Return(ownerMember, nil)
	tr.On("ListPendingByTeamID", ctx, uint(10)).Return(invitations, nil)

	result, err := svc.ListTeamInvitations(ctx, "my-team", 1)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}
