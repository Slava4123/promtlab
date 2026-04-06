package collection

import (
	"context"
	"errors"
	"regexp"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	"promptvault/internal/usecases/teamcheck"
)

var validHexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type Service struct {
	collections repo.CollectionRepository
	teams       repo.TeamRepository
}

func NewService(collections repo.CollectionRepository, teams repo.TeamRepository) *Service {
	return &Service{collections: collections, teams: teams}
}

func (s *Service) Create(ctx context.Context, userID uint, name, description, color, icon string, teamID *uint) (*models.Collection, error) {
	// Проверка роли для командной коллекции
	if err := teamcheck.RequireEditor(ctx, s.teams, teamID, userID); err != nil {
		return nil, mapTeamError(err)
	}

	if color == "" {
		color = "#8b5cf6"
	} else if !validHexColor.MatchString(color) {
		return nil, ErrInvalidColor
	}
	if len(icon) > 30 {
		return nil, ErrInvalidIcon
	}
	c := &models.Collection{
		UserID:      userID,
		TeamID:      teamID,
		Name:        name,
		Description: description,
		Color:       color,
		Icon:        icon,
	}
	if err := s.collections.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) GetByID(ctx context.Context, id, userID uint) (*models.Collection, error) {
	c, err := s.collections.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Командная коллекция — проверять membership
	if c.TeamID != nil {
		_, err := s.teams.GetMember(ctx, *c.TeamID, userID)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return nil, ErrForbidden
			}
			return nil, err
		}
		return c, nil
	}

	// Личная коллекция — проверять user_id
	if c.UserID != userID {
		return nil, ErrForbidden
	}
	return c, nil
}

func (s *Service) List(ctx context.Context, userID uint, teamIDs []uint) ([]models.CollectionWithCount, error) {
	// Проверка membership для командных коллекций
	if len(teamIDs) > 0 {
		if err := teamcheck.RequireMembership(ctx, s.teams, teamIDs, userID); err != nil {
			return nil, mapTeamError(err)
		}
	}
	return s.collections.ListWithCounts(ctx, userID, teamIDs)
}

func (s *Service) Update(ctx context.Context, id, userID uint, name, description, color, icon string) (*models.Collection, error) {
	c, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Командная коллекция — viewer не может редактировать
	if err := s.checkTeamEditAccess(ctx, c, userID); err != nil {
		return nil, err
	}

	if name != "" {
		c.Name = name
	}
	c.Description = description
	if color != "" {
		if !validHexColor.MatchString(color) {
			return nil, ErrInvalidColor
		}
		c.Color = color
	}
	if icon != "" {
		if len(icon) > 30 {
			return nil, ErrInvalidIcon
		}
		c.Icon = icon
	}

	if err := s.collections.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) CountPrompts(ctx context.Context, collectionID uint) (int64, error) {
	return s.collections.CountPrompts(ctx, collectionID)
}

func (s *Service) Delete(ctx context.Context, id, userID uint) error {
	c, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if err := s.checkTeamEditAccess(ctx, c, userID); err != nil {
		return err
	}
	return s.collections.Delete(ctx, id)
}

// checkTeamEditAccess проверяет, что пользователь имеет роль editor+ для командной коллекции
func (s *Service) checkTeamEditAccess(ctx context.Context, c *models.Collection, userID uint) error {
	return mapTeamError(teamcheck.RequireEditor(ctx, s.teams, c.TeamID, userID))
}

func mapTeamError(err error) error {
	return teamcheck.MapError(err, ErrForbidden, ErrViewerReadOnly)
}
