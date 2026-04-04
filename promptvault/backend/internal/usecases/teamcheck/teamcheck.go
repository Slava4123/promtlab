package teamcheck

import (
	"context"
	"errors"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

var (
	ErrForbidden      = errors.New("Нет доступа")
	ErrViewerReadOnly = errors.New("Читатель не может выполнить это действие")
)

// RequireEditor проверяет, что пользователь имеет роль editor+ для командного ресурса.
// Для личных ресурсов (teamID == nil) сразу возвращает nil.
func RequireEditor(ctx context.Context, teams repo.TeamRepository, teamID *uint, userID uint) error {
	if teamID == nil {
		return nil
	}
	member, err := teams.GetMember(ctx, *teamID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrForbidden
		}
		return err
	}
	if member.Role == models.RoleViewer {
		return ErrViewerReadOnly
	}
	return nil
}

// RequireMembership проверяет, что пользователь является участником всех указанных команд.
func RequireMembership(ctx context.Context, teams repo.TeamRepository, teamIDs []uint, userID uint) error {
	for _, tid := range teamIDs {
		if _, err := teams.GetMember(ctx, tid, userID); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return ErrForbidden
			}
			return err
		}
	}
	return nil
}

// MapError translates teamcheck errors into caller-specific domain errors.
func MapError(err error, forbidden, viewerReadOnly error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrForbidden):
		return forbidden
	case errors.Is(err, ErrViewerReadOnly):
		return viewerReadOnly
	default:
		return err
	}
}
