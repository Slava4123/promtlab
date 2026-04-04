package tag

import (
	"context"
	"errors"
	"strings"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	"promptvault/internal/usecases/teamcheck"
)

type Service struct {
	tags  repo.TagRepository
	teams repo.TeamRepository
}

func NewService(tags repo.TagRepository, teams repo.TeamRepository) *Service {
	return &Service{tags: tags, teams: teams}
}

func (s *Service) List(ctx context.Context, userID uint, teamID *uint) ([]models.Tag, error) {
	return s.tags.List(ctx, userID, teamID)
}

func (s *Service) Create(ctx context.Context, name, color string, userID uint, teamID *uint) (*models.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNameEmpty
	}
	if color == "" {
		color = "#6366f1"
	}
	// Проверка роли для командного тега
	if err := teamcheck.RequireEditor(ctx, s.teams, teamID, userID); err != nil {
		return nil, mapTeamError(err)
	}
	return s.tags.GetOrCreate(ctx, name, color, userID, teamID)
}

func (s *Service) Delete(ctx context.Context, id, userID uint) error {
	tag, err := s.tags.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	// Командный тег — проверить membership + editor+
	if tag.TeamID != nil {
		if err := teamcheck.RequireEditor(ctx, s.teams, tag.TeamID, userID); err != nil {
			return mapTeamError(err)
		}
	} else if tag.UserID != userID {
		// Личный тег — проверить владельца
		return ErrForbidden
	}

	return s.tags.Delete(ctx, id)
}

func mapTeamError(err error) error {
	return teamcheck.MapError(err, ErrForbidden, ErrViewerReadOnly)
}
