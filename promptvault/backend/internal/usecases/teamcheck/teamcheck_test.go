// MN-1 — pure-function tests for teamcheck package (RequireEditor / RequireMembership / MapError).
// Без mocking-инфраструктуры — только валидация семантики error-mapping и role gating.
package teamcheck

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// fakeTeamRepo — минимальный мок только для GetMember (единственное что использует teamcheck).
type fakeTeamRepo struct{ mock.Mock }

func (m *fakeTeamRepo) GetMember(ctx context.Context, teamID, userID uint) (*models.TeamMember, error) {
	args := m.Called(ctx, teamID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TeamMember), args.Error(1)
}

// Остальные методы — заглушки.
func (m *fakeTeamRepo) CreateWithOwner(context.Context, *models.Team, uint) error { panic("unused") }
func (m *fakeTeamRepo) GetBySlug(context.Context, string) (*models.Team, error)   { panic("unused") }
func (m *fakeTeamRepo) GetByID(context.Context, uint) (*models.Team, error)       { panic("unused") }
func (m *fakeTeamRepo) ListByUserID(context.Context, uint) ([]models.Team, error) { panic("unused") }
func (m *fakeTeamRepo) ListByUserIDWithRolesAndCounts(context.Context, uint) ([]models.TeamWithRoleAndCount, error) {
	panic("unused")
}
func (m *fakeTeamRepo) ListOwnedTeams(context.Context, uint) ([]models.Team, error) {
	panic("unused")
}
func (m *fakeTeamRepo) Update(context.Context, *models.Team) error           { panic("unused") }
func (m *fakeTeamRepo) Delete(context.Context, uint) error                   { panic("unused") }
func (m *fakeTeamRepo) UpdateMemberRole(context.Context, uint, uint, models.TeamRole) error {
	panic("unused")
}
func (m *fakeTeamRepo) RemoveMember(context.Context, uint, uint) error          { panic("unused") }
func (m *fakeTeamRepo) ListMembers(context.Context, uint) ([]models.TeamMember, error) {
	panic("unused")
}
func (m *fakeTeamRepo) CountMembers(context.Context, uint) (int, error)            { panic("unused") }
func (m *fakeTeamRepo) CreateInvitation(context.Context, *models.TeamInvitation) error {
	panic("unused")
}
func (m *fakeTeamRepo) GetInvitationByID(context.Context, uint) (*models.TeamInvitation, error) {
	panic("unused")
}
func (m *fakeTeamRepo) GetPendingInvitation(context.Context, uint, uint) (*models.TeamInvitation, error) {
	panic("unused")
}
func (m *fakeTeamRepo) ListPendingByUserID(context.Context, uint) ([]models.TeamInvitation, error) {
	panic("unused")
}
func (m *fakeTeamRepo) ListPendingByTeamID(context.Context, uint) ([]models.TeamInvitation, error) {
	panic("unused")
}
func (m *fakeTeamRepo) UpdateInvitationStatus(context.Context, uint, models.InvitationStatus) error {
	panic("unused")
}
func (m *fakeTeamRepo) DeleteInvitation(context.Context, uint) error { panic("unused") }
func (m *fakeTeamRepo) AcceptInvitationTx(context.Context, uint, *models.TeamMember) error {
	panic("unused")
}
func (m *fakeTeamRepo) UpdateBranding(context.Context, uint, string, string, string, string) error {
	panic("unused")
}
func (m *fakeTeamRepo) UpdateBrandLogoSource(context.Context, uint, string) error {
	panic("unused")
}

// --- RequireEditor ---

func TestRequireEditor_NilTeamID_ReturnsNil(t *testing.T) {
	if err := RequireEditor(context.Background(), nil, nil, 42); err != nil {
		t.Fatalf("expected nil for personal-mode (teamID=nil), got %v", err)
	}
}

func TestRequireEditor_OwnerOK(t *testing.T) {
	r := &fakeTeamRepo{}
	teamID := uint(7)
	r.On("GetMember", mock.Anything, teamID, uint(42)).
		Return(&models.TeamMember{TeamID: teamID, UserID: 42, Role: models.RoleOwner}, nil)

	if err := RequireEditor(context.Background(), r, &teamID, 42); err != nil {
		t.Fatalf("expected nil for owner, got %v", err)
	}
}

func TestRequireEditor_EditorOK(t *testing.T) {
	r := &fakeTeamRepo{}
	teamID := uint(7)
	r.On("GetMember", mock.Anything, teamID, uint(42)).
		Return(&models.TeamMember{TeamID: teamID, UserID: 42, Role: models.RoleEditor}, nil)

	if err := RequireEditor(context.Background(), r, &teamID, 42); err != nil {
		t.Fatalf("expected nil for editor, got %v", err)
	}
}

func TestRequireEditor_Viewer_ReturnsViewerReadOnly(t *testing.T) {
	r := &fakeTeamRepo{}
	teamID := uint(7)
	r.On("GetMember", mock.Anything, teamID, uint(42)).
		Return(&models.TeamMember{TeamID: teamID, UserID: 42, Role: models.RoleViewer}, nil)

	err := RequireEditor(context.Background(), r, &teamID, 42)
	if !errors.Is(err, ErrViewerReadOnly) {
		t.Fatalf("expected ErrViewerReadOnly for viewer, got %v", err)
	}
}

func TestRequireEditor_NotMember_ReturnsForbidden(t *testing.T) {
	r := &fakeTeamRepo{}
	teamID := uint(7)
	r.On("GetMember", mock.Anything, teamID, uint(42)).
		Return(nil, repo.ErrNotFound)

	err := RequireEditor(context.Background(), r, &teamID, 42)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden for non-member, got %v", err)
	}
}

// --- RequireMembership ---

func TestRequireMembership_NoTeams_OK(t *testing.T) {
	if err := RequireMembership(context.Background(), nil, nil, 42); err != nil {
		t.Fatalf("expected nil for empty teamIDs, got %v", err)
	}
}

func TestRequireMembership_AllMembers_OK(t *testing.T) {
	r := &fakeTeamRepo{}
	r.On("GetMember", mock.Anything, uint(7), uint(42)).
		Return(&models.TeamMember{Role: models.RoleViewer}, nil)
	r.On("GetMember", mock.Anything, uint(8), uint(42)).
		Return(&models.TeamMember{Role: models.RoleEditor}, nil)

	if err := RequireMembership(context.Background(), r, []uint{7, 8}, 42); err != nil {
		t.Fatalf("expected nil for all members, got %v", err)
	}
}

func TestRequireMembership_NotMember_ReturnsForbidden(t *testing.T) {
	r := &fakeTeamRepo{}
	r.On("GetMember", mock.Anything, uint(7), uint(42)).
		Return(&models.TeamMember{Role: models.RoleViewer}, nil)
	r.On("GetMember", mock.Anything, uint(8), uint(42)).
		Return(nil, repo.ErrNotFound)

	err := RequireMembership(context.Background(), r, []uint{7, 8}, 42)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden when not a member of one team, got %v", err)
	}
}

// --- MapError ---

func TestMapError_Nil_ReturnsNil(t *testing.T) {
	if err := MapError(nil, errors.New("x"), errors.New("y")); err != nil {
		t.Fatalf("expected nil for nil input, got %v", err)
	}
}

func TestMapError_Forbidden_MapsToCustom(t *testing.T) {
	customForbidden := errors.New("custom forbidden")
	customViewer := errors.New("custom viewer")
	if err := MapError(ErrForbidden, customForbidden, customViewer); !errors.Is(err, customForbidden) {
		t.Fatalf("expected custom forbidden, got %v", err)
	}
}

func TestMapError_ViewerReadOnly_MapsToCustom(t *testing.T) {
	customForbidden := errors.New("custom forbidden")
	customViewer := errors.New("custom viewer")
	if err := MapError(ErrViewerReadOnly, customForbidden, customViewer); !errors.Is(err, customViewer) {
		t.Fatalf("expected custom viewer, got %v", err)
	}
}

func TestMapError_OtherError_PassedThrough(t *testing.T) {
	other := errors.New("db down")
	if err := MapError(other, errors.New("x"), errors.New("y")); !errors.Is(err, other) {
		t.Fatalf("expected pass-through of unknown error, got %v", err)
	}
}
