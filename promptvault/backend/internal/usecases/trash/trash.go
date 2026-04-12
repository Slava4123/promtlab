package trash

import (
	"context"
	"errors"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	"promptvault/internal/usecases/teamcheck"
)

type Service struct {
	trash repo.TrashRepository
	teams repo.TeamRepository
}

func NewService(trash repo.TrashRepository, teams repo.TeamRepository) *Service {
	return &Service{trash: trash, teams: teams}
}

// ---------- List ----------

func (s *Service) ListDeletedPrompts(ctx context.Context, userID uint, teamIDs []uint, page, pageSize int) ([]models.Prompt, int64, error) {
	if err := s.checkMembership(ctx, teamIDs, userID); err != nil {
		return nil, 0, err
	}
	return s.trash.ListDeletedPrompts(ctx, userID, teamIDs, page, pageSize)
}

func (s *Service) ListDeletedCollections(ctx context.Context, userID uint, teamIDs []uint) ([]models.Collection, error) {
	if err := s.checkMembership(ctx, teamIDs, userID); err != nil {
		return nil, err
	}
	return s.trash.ListDeletedCollections(ctx, userID, teamIDs)
}

// ---------- Count ----------

func (s *Service) Count(ctx context.Context, userID uint, teamIDs []uint) (repo.TrashCounts, error) {
	if err := s.checkMembership(ctx, teamIDs, userID); err != nil {
		return repo.TrashCounts{}, err
	}
	return s.trash.CountDeleted(ctx, userID, teamIDs)
}

// ---------- Restore ----------

func (s *Service) Restore(ctx context.Context, itemType ItemType, id, userID uint) error {
	switch itemType {
	case TypePrompt:
		return s.restorePrompt(ctx, id, userID)
	case TypeCollection:
		return s.restoreCollection(ctx, id, userID)
	default:
		return ErrInvalidType
	}
}

func (s *Service) restorePrompt(ctx context.Context, id, userID uint) error {
	p, err := s.trash.GetDeletedPrompt(ctx, id)
	if err != nil {
		return s.mapRepoErr(err)
	}
	if err := s.checkOwnerOrEditor(ctx, p.UserID, p.TeamID, userID); err != nil {
		return err
	}
	return s.trash.RestorePrompt(ctx, id)
}

func (s *Service) restoreCollection(ctx context.Context, id, userID uint) error {
	c, err := s.trash.GetDeletedCollection(ctx, id)
	if err != nil {
		return s.mapRepoErr(err)
	}
	if err := s.checkOwnerOrEditor(ctx, c.UserID, c.TeamID, userID); err != nil {
		return err
	}
	return s.trash.RestoreCollection(ctx, id)
}

// ---------- Permanent delete ----------

func (s *Service) PermanentDelete(ctx context.Context, itemType ItemType, id, userID uint) error {
	switch itemType {
	case TypePrompt:
		return s.permanentDeletePrompt(ctx, id, userID)
	case TypeCollection:
		return s.permanentDeleteCollection(ctx, id, userID)
	default:
		return ErrInvalidType
	}
}

func (s *Service) permanentDeletePrompt(ctx context.Context, id, userID uint) error {
	p, err := s.trash.GetDeletedPrompt(ctx, id)
	if err != nil {
		return s.mapRepoErr(err)
	}
	if err := s.checkOwnerOrEditor(ctx, p.UserID, p.TeamID, userID); err != nil {
		return err
	}
	return s.trash.HardDeletePrompt(ctx, id)
}

func (s *Service) permanentDeleteCollection(ctx context.Context, id, userID uint) error {
	c, err := s.trash.GetDeletedCollection(ctx, id)
	if err != nil {
		return s.mapRepoErr(err)
	}
	if err := s.checkOwnerOrEditor(ctx, c.UserID, c.TeamID, userID); err != nil {
		return err
	}
	return s.trash.HardDeleteCollection(ctx, id)
}

// ---------- Empty trash ----------

func (s *Service) EmptyTrash(ctx context.Context, userID uint, teamIDs []uint) (int64, error) {
	if err := s.checkEditorForTeams(ctx, teamIDs, userID); err != nil {
		return 0, err
	}
	return s.trash.EmptyTrash(ctx, userID, teamIDs)
}

// ---------- helpers ----------

func (s *Service) checkMembership(ctx context.Context, teamIDs []uint, userID uint) error {
	if len(teamIDs) == 0 {
		return nil
	}
	return mapTeamError(teamcheck.RequireMembership(ctx, s.teams, teamIDs, userID))
}

func (s *Service) checkEditorForTeams(ctx context.Context, teamIDs []uint, userID uint) error {
	for _, tid := range teamIDs {
		if err := teamcheck.RequireEditor(ctx, s.teams, &tid, userID); err != nil {
			return mapTeamError(err)
		}
	}
	return nil
}

func (s *Service) checkOwnerOrEditor(ctx context.Context, ownerID uint, teamID *uint, userID uint) error {
	if teamID == nil {
		if ownerID != userID {
			return ErrForbidden
		}
		return nil
	}
	return mapTeamError(teamcheck.RequireEditor(ctx, s.teams, teamID, userID))
}

func (s *Service) mapRepoErr(err error) error {
	if errors.Is(err, repo.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func mapTeamError(err error) error {
	return teamcheck.MapError(err, ErrForbidden, ErrViewerReadOnly)
}
