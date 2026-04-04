package prompt

import (
	"context"
	"errors"
	"fmt"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	"promptvault/internal/usecases/teamcheck"
)

type Service struct {
	prompts     repo.PromptRepository
	tags        repo.TagRepository
	collections repo.CollectionRepository
	versions    repo.VersionRepository
	teams       repo.TeamRepository
}

func NewService(prompts repo.PromptRepository, tags repo.TagRepository, collections repo.CollectionRepository, versions repo.VersionRepository, teams repo.TeamRepository) *Service {
	return &Service{prompts: prompts, tags: tags, collections: collections, versions: versions, teams: teams}
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*models.Prompt, error) {
	// Проверка роли для командного промпта (viewer не может создавать)
	if err := teamcheck.RequireEditor(ctx, s.teams, in.TeamID, in.UserID); err != nil {
		return nil, mapTeamError(err)
	}

	p := &models.Prompt{
		UserID:  in.UserID,
		TeamID:  in.TeamID,
		Title:   in.Title,
		Content: in.Content,
		Model:   in.Model,
	}

	if len(in.TagIDs) > 0 {
		tags, err := s.tags.GetByIDs(ctx, in.TagIDs)
		if err != nil {
			return nil, err
		}
		// Проверка что теги принадлежат тому же workspace
		for _, t := range tags {
			if !sameWorkspace(in.TeamID, t.TeamID) {
				return nil, ErrForbidden
			}
		}
		p.Tags = tags
	}

	if len(in.CollectionIDs) > 0 {
		cols, err := s.collections.GetByIDs(ctx, in.CollectionIDs)
		if err != nil {
			return nil, err
		}
		// Проверка что коллекции принадлежат тому же workspace
		for _, c := range cols {
			if !sameWorkspace(in.TeamID, c.TeamID) {
				return nil, ErrForbidden
			}
		}
		p.Collections = cols
	}

	if err := s.prompts.Create(ctx, p); err != nil {
		return nil, err
	}

	return s.prompts.GetByID(ctx, p.ID)
}

func (s *Service) GetByID(ctx context.Context, id, userID uint) (*models.Prompt, error) {
	p, err := s.prompts.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Личный промпт — проверка по user_id
	if p.TeamID == nil {
		if p.UserID != userID {
			return nil, ErrForbidden
		}
		return p, nil
	}

	// Командный промпт — проверка membership (любая роль)
	_, err = s.teams.GetMember(ctx, *p.TeamID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	return p, nil
}

func (s *Service) List(ctx context.Context, filter repo.PromptListFilter) ([]models.Prompt, int64, error) {
	// Проверка membership для командных промптов
	if len(filter.TeamIDs) > 0 {
		if err := teamcheck.RequireMembership(ctx, s.teams, filter.TeamIDs, filter.UserID); err != nil {
			return nil, 0, mapTeamError(err)
		}
	}
	return s.prompts.List(ctx, filter)
}

func (s *Service) Update(ctx context.Context, id, userID uint, in UpdateInput) (*models.Prompt, error) {
	p, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Командный промпт — viewer не может редактировать
	if err := s.checkTeamEditAccess(ctx, p, userID); err != nil {
		return nil, err
	}

	// Атомарный снимок старого состояния перед мутацией (SELECT MAX + INSERT в одной транзакции)
	version := &models.PromptVersion{
		PromptID:   p.ID,
		Title:      p.Title,
		Content:    p.Content,
		Model:      p.Model,
		ChangeNote: in.ChangeNote,
	}
	if err := s.versions.CreateWithNextVersion(ctx, version); err != nil {
		return nil, err
	}

	if in.Title != nil {
		p.Title = *in.Title
	}
	if in.Content != nil {
		p.Content = *in.Content
	}
	if in.Model != nil {
		p.Model = *in.Model
	}
	if in.CollectionIDs != nil {
		cols, err := s.collections.GetByIDs(ctx, in.CollectionIDs)
		if err != nil {
			return nil, err
		}
		for _, c := range cols {
			if !sameWorkspace(p.TeamID, c.TeamID) {
				return nil, ErrForbidden
			}
		}
		p.Collections = cols
	}
	if in.TagIDs != nil {
		tags, err := s.tags.GetByIDs(ctx, in.TagIDs)
		if err != nil {
			return nil, err
		}
		for _, t := range tags {
			if !sameWorkspace(p.TeamID, t.TeamID) {
				return nil, ErrForbidden
			}
		}
		p.Tags = tags
	}

	if err := s.prompts.Update(ctx, p); err != nil {
		return nil, err
	}

	return s.prompts.GetByID(ctx, p.ID)
}

func (s *Service) Delete(ctx context.Context, id, userID uint) error {
	p, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if err := s.checkTeamEditAccess(ctx, p, userID); err != nil {
		return err
	}
	return s.prompts.SoftDelete(ctx, id)
}

func (s *Service) ToggleFavorite(ctx context.Context, id, userID uint) (*models.Prompt, error) {
	p, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if err := s.prompts.SetFavorite(ctx, id, !p.Favorite); err != nil {
		return nil, err
	}

	p.Favorite = !p.Favorite
	return p, nil
}

func (s *Service) IncrementUsage(ctx context.Context, id, userID uint) error {
	if _, err := s.GetByID(ctx, id, userID); err != nil {
		return err
	}
	return s.prompts.IncrementUsage(ctx, id)
}

func (s *Service) ListVersions(ctx context.Context, promptID, userID uint, page, pageSize int) ([]models.PromptVersion, int64, error) {
	if _, err := s.GetByID(ctx, promptID, userID); err != nil {
		return nil, 0, err
	}
	return s.versions.ListByPromptID(ctx, promptID, page, pageSize)
}

func (s *Service) RevertToVersion(ctx context.Context, promptID, userID, versionID uint) (*models.Prompt, error) {
	// Загрузить версию с проверкой принадлежности к промпту (один запрос)
	v, err := s.versions.GetByIDForPrompt(ctx, versionID, promptID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrVersionNotFound
		}
		return nil, err
	}

	// Откат через Update (проверит доступ userID + создаст снимок текущего состояния)
	changeNote := fmt.Sprintf("Откат к версии %d", v.VersionNumber)
	return s.Update(ctx, promptID, userID, UpdateInput{
		Title:      &v.Title,
		Content:    &v.Content,
		Model:      &v.Model,
		ChangeNote: changeNote,
	})
}

// checkTeamEditAccess проверяет, что пользователь имеет роль editor+ для командного промпта
func (s *Service) checkTeamEditAccess(ctx context.Context, p *models.Prompt, userID uint) error {
	return mapTeamError(teamcheck.RequireEditor(ctx, s.teams, p.TeamID, userID))
}

// mapTeamError транслирует ошибки teamcheck в доменные ошибки prompt
func mapTeamError(err error) error {
	return teamcheck.MapError(err, ErrForbidden, ErrViewerReadOnly)
}

// sameWorkspace проверяет что два TeamID указывают на один workspace
func sameWorkspace(a, b *uint) bool {
	if a == nil && b == nil {
		return true // оба личные
	}
	if a == nil || b == nil {
		return false // один личный, другой командный
	}
	return *a == *b
}
