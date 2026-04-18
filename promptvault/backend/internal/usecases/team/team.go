package team

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	repo "promptvault/internal/interface/repository"
	iservice "promptvault/internal/interface/service"
	"promptvault/internal/models"
	quotauc "promptvault/internal/usecases/quota"
)

type Service struct {
	teams  repo.TeamRepository
	users  repo.UserRepository
	email  iservice.EmailSender
	quotas *quotauc.Service
}

func NewService(teams repo.TeamRepository, users repo.UserRepository, quotas *quotauc.Service) *Service {
	return &Service{teams: teams, users: users, quotas: quotas}
}

// SetEmail sets the email service for team notifications.
func (s *Service) SetEmail(email iservice.EmailSender) {
	s.email = email
}

func (s *Service) Create(ctx context.Context, userID uint, input CreateInput) (*models.Team, error) {
	// Проверка квоты команд
	if s.quotas != nil {
		if err := s.quotas.CheckTeamQuota(ctx, userID); err != nil {
			return nil, err
		}
	}

	// Retry на случай slug collision (unique index)
	var lastErr error
	for range 3 {
		team := &models.Team{
			Slug:        generateSlug(input.Name),
			Name:        input.Name,
			Description: input.Description,
			CreatedBy:   userID,
		}
		if err := s.teams.CreateWithOwner(ctx, team, userID); err != nil {
			lastErr = err
			continue
		}
		return team, nil
	}
	return nil, lastErr
}

func (s *Service) GetBySlug(ctx context.Context, slug string, userID uint) (*models.Team, []models.TeamMember, error) {
	team, _, err := s.checkAccess(ctx, slug, userID, models.RoleViewer)
	if err != nil {
		return nil, nil, err
	}

	members, err := s.teams.ListMembers(ctx, team.ID)
	if err != nil {
		return nil, nil, err
	}

	return team, members, nil
}

func (s *Service) List(ctx context.Context, userID uint) ([]TeamListItem, error) {
	rows, err := s.teams.ListByUserIDWithRolesAndCounts(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]TeamListItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, TeamListItem{
			Team:        r.Team,
			Role:        r.Role,
			MemberCount: r.MemberCount,
		})
	}

	return items, nil
}

func (s *Service) Update(ctx context.Context, slug string, userID uint, input UpdateInput) (*models.Team, error) {
	team, _, err := s.checkAccess(ctx, slug, userID, models.RoleOwner)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		team.Name = *input.Name
	}
	if input.Description != nil {
		team.Description = *input.Description
	}

	if err := s.teams.Update(ctx, team); err != nil {
		return nil, err
	}
	return team, nil
}

func (s *Service) Delete(ctx context.Context, slug string, userID uint) error {
	team, _, err := s.checkAccess(ctx, slug, userID, models.RoleOwner)
	if err != nil {
		return err
	}
	return s.teams.Delete(ctx, team.ID)
}

func (s *Service) InviteMember(ctx context.Context, slug string, userID uint, input AddMemberInput) (*models.TeamInvitation, error) {
	team, _, err := s.checkAccess(ctx, slug, userID, models.RoleOwner)
	if err != nil {
		return nil, err
	}

	// Проверка квоты участников команды (по плану владельца)
	if s.quotas != nil {
		if err := s.quotas.CheckTeamMemberQuota(ctx, team.ID, team.CreatedBy); err != nil {
			return nil, err
		}
	}

	if input.Role == models.RoleOwner {
		return nil, ErrInvalidRole
	}

	var targetUser *models.User
	if strings.HasPrefix(input.Query, "@") {
		uname := strings.TrimPrefix(input.Query, "@")
		if uname == "" {
			return nil, ErrUserNotFound
		}
		targetUser, err = s.users.GetByUsername(ctx, uname)
	} else {
		targetUser, err = s.users.GetByEmail(ctx, input.Query)
	}
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if targetUser.ID == userID {
		return nil, ErrCannotInviteSelf
	}

	// Уже участник?
	_, err = s.teams.GetMember(ctx, team.ID, targetUser.ID)
	if err == nil {
		return nil, ErrAlreadyMember
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return nil, err
	}

	// Уже есть pending приглашение?
	_, err = s.teams.GetPendingInvitation(ctx, team.ID, targetUser.ID)
	if err == nil {
		return nil, ErrAlreadyInvited
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return nil, err
	}

	inv := &models.TeamInvitation{
		TeamID:    team.ID,
		UserID:    targetUser.ID,
		InviterID: userID,
		Role:      input.Role,
		Status:    models.InvitationPending,
	}
	if err := s.teams.CreateInvitation(ctx, inv); err != nil {
		return nil, err
	}

	inv.Team = *team
	inviter, err := s.users.GetByID(ctx, userID)
	if err != nil {
		slog.Error("failed to load inviter", "user_id", userID, "error", err)
	} else {
		inv.Inviter = *inviter
	}

	// Отправляем email-уведомление о приглашении (fire-and-forget, допустимо для некритичного уведомления)
	if s.email != nil && s.email.Configured() {
		inviterName := "Пользователь"
		if inviter != nil {
			inviterName = inviter.Name
		}
		targetEmail := targetUser.Email
		teamName := team.Name
		go func() {
			if err := s.email.SendTeamInvitation(targetEmail, teamName, inviterName); err != nil {
				slog.Error("team invitation email failed", "error", err)
			}
		}()
	}

	return inv, nil
}

func (s *Service) ListMyInvitations(ctx context.Context, userID uint) ([]models.TeamInvitation, error) {
	return s.teams.ListPendingByUserID(ctx, userID)
}

func (s *Service) ListTeamInvitations(ctx context.Context, slug string, userID uint) ([]models.TeamInvitation, error) {
	team, _, err := s.checkAccess(ctx, slug, userID, models.RoleOwner)
	if err != nil {
		return nil, err
	}
	return s.teams.ListPendingByTeamID(ctx, team.ID)
}

func (s *Service) AcceptInvitation(ctx context.Context, invitationID, userID uint) error {
	inv, err := s.teams.GetInvitationByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrInvitationNotFound
		}
		return err
	}

	if inv.UserID != userID {
		return ErrForbidden
	}
	if inv.Status != models.InvitationPending {
		return ErrInvitationNotFound
	}

	// Проверка что не уже участник
	_, err = s.teams.GetMember(ctx, inv.TeamID, inv.UserID)
	if err == nil {
		return ErrAlreadyMember
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return err
	}

	member := &models.TeamMember{
		TeamID: inv.TeamID,
		UserID: inv.UserID,
		Role:   inv.Role,
	}
	return s.teams.AcceptInvitationTx(ctx, inv.ID, member)
}

func (s *Service) DeclineInvitation(ctx context.Context, invitationID, userID uint) error {
	inv, err := s.teams.GetInvitationByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrInvitationNotFound
		}
		return err
	}

	if inv.UserID != userID {
		return ErrForbidden
	}
	if inv.Status != models.InvitationPending {
		return ErrInvitationNotFound
	}

	return s.teams.UpdateInvitationStatus(ctx, inv.ID, models.InvitationDeclined)
}

func (s *Service) CancelInvitation(ctx context.Context, slug string, userID, invitationID uint) error {
	team, _, err := s.checkAccess(ctx, slug, userID, models.RoleOwner)
	if err != nil {
		return err
	}

	inv, err := s.teams.GetInvitationByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrInvitationNotFound
		}
		return err
	}

	// Проверка что приглашение принадлежит этой команде
	if inv.TeamID != team.ID {
		return ErrInvitationNotFound
	}

	if inv.Status != models.InvitationPending {
		return ErrInvitationNotFound
	}

	return s.teams.DeleteInvitation(ctx, inv.ID)
}

func (s *Service) UpdateMemberRole(ctx context.Context, slug string, userID, targetUserID uint, role models.TeamRole) error {
	team, _, err := s.checkAccess(ctx, slug, userID, models.RoleOwner)
	if err != nil {
		return err
	}

	if role == models.RoleOwner {
		return ErrInvalidRole
	}

	targetMember, err := s.teams.GetMember(ctx, team.ID, targetUserID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if targetMember.Role == models.RoleOwner {
		return ErrCannotChangeOwnerRole
	}

	return s.teams.UpdateMemberRole(ctx, team.ID, targetUserID, role)
}

func (s *Service) RemoveMember(ctx context.Context, slug string, userID, targetUserID uint) error {
	team, callerMember, err := s.checkAccess(ctx, slug, userID, models.RoleViewer)
	if err != nil {
		return err
	}

	// Не-owner может покинуть только сам
	if userID != targetUserID && callerMember.Role != models.RoleOwner {
		return ErrNotOwner
	}

	targetMember, err := s.teams.GetMember(ctx, team.ID, targetUserID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if targetMember.Role == models.RoleOwner {
		return ErrCannotRemoveOwner
	}

	return s.teams.RemoveMember(ctx, team.ID, targetUserID)
}

// IsMember проверяет, что userID является участником команды teamID (в любой роли).
// Используется при валидации scoped API-keys с team_id.
func (s *Service) IsMember(ctx context.Context, teamID, userID uint) (bool, error) {
	_, err := s.teams.GetMember(ctx, teamID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// checkAccess проверяет, что пользователь является участником команды с минимальной ролью
func (s *Service) checkAccess(ctx context.Context, slug string, userID uint, minRole models.TeamRole) (*models.Team, *models.TeamMember, error) {
	team, err := s.teams.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}

	member, err := s.teams.GetMember(ctx, team.ID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, nil, ErrForbidden
		}
		return nil, nil, err
	}

	if roleLevel(member.Role) < roleLevel(minRole) {
		return nil, nil, ErrNotOwner
	}

	return team, member, nil
}

func roleLevel(r models.TeamRole) int {
	switch r {
	case models.RoleOwner:
		return 3
	case models.RoleEditor:
		return 2
	case models.RoleViewer:
		return 1
	default:
		return 0
	}
}
