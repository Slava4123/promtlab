package repository

import (
	"context"

	"promptvault/internal/models"
)

type TeamRepository interface {
	CreateWithOwner(ctx context.Context, team *models.Team, ownerUserID uint) error
	GetBySlug(ctx context.Context, slug string) (*models.Team, error)
	ListByUserID(ctx context.Context, userID uint) ([]models.Team, error)
	Update(ctx context.Context, team *models.Team) error
	Delete(ctx context.Context, id uint) error
	GetMember(ctx context.Context, teamID, userID uint) (*models.TeamMember, error)
	UpdateMemberRole(ctx context.Context, teamID, userID uint, role models.TeamRole) error
	RemoveMember(ctx context.Context, teamID, userID uint) error
	ListMembers(ctx context.Context, teamID uint) ([]models.TeamMember, error)
	CountMembers(ctx context.Context, teamID uint) (int, error)

	// Invitations
	CreateInvitation(ctx context.Context, inv *models.TeamInvitation) error
	GetInvitationByID(ctx context.Context, id uint) (*models.TeamInvitation, error)
	GetPendingInvitation(ctx context.Context, teamID, userID uint) (*models.TeamInvitation, error)
	ListPendingByUserID(ctx context.Context, userID uint) ([]models.TeamInvitation, error)
	ListPendingByTeamID(ctx context.Context, teamID uint) ([]models.TeamInvitation, error)
	UpdateInvitationStatus(ctx context.Context, id uint, status models.InvitationStatus) error
	DeleteInvitation(ctx context.Context, id uint) error
	AcceptInvitationTx(ctx context.Context, invID uint, member *models.TeamMember) error
}
