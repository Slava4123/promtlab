package prompt

import (
	"context"
	"errors"
	"fmt"

	"log/slog"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	badgeuc "promptvault/internal/usecases/badge"
	quotauc "promptvault/internal/usecases/quota"
	streakuc "promptvault/internal/usecases/streak"
	"promptvault/internal/usecases/teamcheck"
)

type Service struct {
	prompts     repo.PromptRepository
	pins        repo.PinRepository
	tags        repo.TagRepository
	collections repo.CollectionRepository
	versions    repo.VersionRepository
	teams       repo.TeamRepository
	streaks     *streakuc.Service
	badges      *badgeuc.Service
	quotas      *quotauc.Service
}

func NewService(prompts repo.PromptRepository, tags repo.TagRepository, collections repo.CollectionRepository, versions repo.VersionRepository, teams repo.TeamRepository, pins repo.PinRepository, streaks *streakuc.Service, badges *badgeuc.Service, quotas *quotauc.Service) *Service {
	return &Service{prompts: prompts, tags: tags, collections: collections, versions: versions, teams: teams, pins: pins, streaks: streaks, badges: badges, quotas: quotas}
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*models.Prompt, []badgeuc.Badge, error) {
	// Проверка квоты промптов
	if s.quotas != nil {
		if err := s.quotas.CheckPromptQuota(ctx, in.UserID); err != nil {
			return nil, nil, err
		}
	}

	// Проверка роли для командного промпта (viewer не может создавать)
	if err := teamcheck.RequireEditor(ctx, s.teams, in.TeamID, in.UserID); err != nil {
		return nil, nil, mapTeamError(err)
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
			return nil, nil, err
		}
		// Проверка что теги принадлежат тому же workspace
		for _, t := range tags {
			if !sameWorkspace(in.TeamID, t.TeamID) {
				return nil, nil, ErrWorkspaceMismatch
			}
		}
		p.Tags = tags
	}

	if len(in.CollectionIDs) > 0 {
		cols, err := s.collections.GetByIDs(ctx, in.CollectionIDs)
		if err != nil {
			return nil, nil, err
		}
		// Проверка что коллекции принадлежат тому же workspace
		for _, c := range cols {
			if !sameWorkspace(in.TeamID, c.TeamID) {
				return nil, nil, ErrWorkspaceMismatch
			}
		}
		p.Collections = cols
	}

	if err := s.prompts.Create(ctx, p); err != nil {
		return nil, nil, err
	}

	if s.streaks != nil {
		s.streaks.RecordActivity(ctx, in.UserID, timezoneFromCtx(ctx))
	}

	// Badges evaluate — best-effort, возвращает newly unlocked (или nil).
	// Сломанная badge-эвалуация никогда не блокирует основной flow создания промпта.
	var newBadges []badgeuc.Badge
	if s.badges != nil {
		newBadges = s.badges.Evaluate(ctx, in.UserID, badgeuc.Event{
			Type:     badgeuc.EventPromptCreated,
			TeamID:   in.TeamID,
			PromptID: p.ID,
		})
	}

	result, err := s.prompts.GetByID(ctx, p.ID)
	if err != nil {
		return nil, nil, err
	}
	return result, newBadges, nil
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

// GetPublicBySlug — публичное получение промпта по slug (no auth).
func (s *Service) GetPublicBySlug(ctx context.Context, slug string) (*models.Prompt, error) {
	return s.prompts.GetPublicBySlug(ctx, slug)
}

// ListPublic — список публичных промптов (для sitemap).
func (s *Service) ListPublic(ctx context.Context, limit int) ([]models.Prompt, error) {
	return s.prompts.ListPublic(ctx, limit)
}

func (s *Service) Update(ctx context.Context, id, userID uint, in UpdateInput) (*models.Prompt, []badgeuc.Badge, error) {
	p, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, nil, err
	}

	// Командный промпт — viewer не может редактировать
	if err := s.checkTeamEditAccess(ctx, p, userID); err != nil {
		return nil, nil, err
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
		return nil, nil, err
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
			return nil, nil, err
		}
		for _, c := range cols {
			if !sameWorkspace(p.TeamID, c.TeamID) {
				return nil, nil, ErrWorkspaceMismatch
			}
		}
		p.Collections = cols
	}
	if in.TagIDs != nil {
		tags, err := s.tags.GetByIDs(ctx, in.TagIDs)
		if err != nil {
			return nil, nil, err
		}
		for _, t := range tags {
			if !sameWorkspace(p.TeamID, t.TeamID) {
				return nil, nil, ErrWorkspaceMismatch
			}
		}
		p.Tags = tags
	}

	// Public flag + slug generation.
	// При включении публичности лениво генерим slug если пусто. Slug не меняется
	// при re-publish, чтобы ссылки не протухали.
	if in.IsPublic != nil {
		p.IsPublic = *in.IsPublic
		if p.IsPublic && p.Slug == "" {
			title := p.Title
			if in.Title != nil {
				title = *in.Title
			}
			p.Slug = makeSlug(p.ID, title)
		}
	}

	if err := s.prompts.Update(ctx, p); err != nil {
		return nil, nil, err
	}

	// Удаляем теги-сироты (не привязанные ни к одному промпту)
	if in.TagIDs != nil {
		if err := s.tags.DeleteOrphans(ctx, p.UserID, p.TeamID); err != nil {
			slog.Warn("orphan tags cleanup failed", "error", err, "user_id", p.UserID)
		}
	}

	// Badges evaluate на событии prompt_updated (для бейджа Refactorer).
	// Best-effort: ошибки в Evaluate не блокируют основной flow.
	var newBadges []badgeuc.Badge
	if s.badges != nil {
		newBadges = s.badges.Evaluate(ctx, userID, badgeuc.Event{
			Type:     badgeuc.EventPromptUpdated,
			TeamID:   p.TeamID,
			PromptID: p.ID,
		})
	}

	result, err := s.prompts.GetByID(ctx, p.ID)
	if err != nil {
		return nil, nil, err
	}
	return result, newBadges, nil
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

func (s *Service) IncrementUsage(ctx context.Context, id, userID uint) ([]badgeuc.Badge, error) {
	p, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	// Логируем в историю (best-effort — не блокирует основной flow)
	if err := s.prompts.LogUsage(ctx, userID, id); err != nil {
		slog.Warn("usage log failed", "error", err, "prompt_id", id, "user_id", userID)
	}
	if s.streaks != nil {
		s.streaks.RecordActivity(ctx, userID, timezoneFromCtx(ctx))
	}
	if err := s.prompts.IncrementUsage(ctx, id); err != nil {
		return nil, err
	}

	// Badges evaluate — best-effort, после инкремента счётчика, чтобы
	// SumUsage вернул актуальное значение (важно для бейджа Advanced).
	var newBadges []badgeuc.Badge
	if s.badges != nil {
		newBadges = s.badges.Evaluate(ctx, userID, badgeuc.Event{
			Type:     badgeuc.EventPromptUsed,
			TeamID:   p.TeamID,
			PromptID: id,
		})
	}
	return newBadges, nil
}

func (s *Service) ListHistory(ctx context.Context, userID uint, teamID *uint, page, pageSize int) ([]models.PromptUsageLog, int64, error) {
	if teamID != nil {
		if err := teamcheck.RequireMembership(ctx, s.teams, []uint{*teamID}, userID); err != nil {
			return nil, 0, mapTeamError(err)
		}
	}
	return s.prompts.ListUsageHistory(ctx, userID, teamID, page, pageSize)
}

func (s *Service) ListVersions(ctx context.Context, promptID, userID uint, page, pageSize int) ([]models.PromptVersion, int64, error) {
	if _, err := s.GetByID(ctx, promptID, userID); err != nil {
		return nil, 0, err
	}
	return s.versions.ListByPromptID(ctx, promptID, page, pageSize)
}

func (s *Service) RevertToVersion(ctx context.Context, promptID, userID, versionID uint) (*models.Prompt, []badgeuc.Badge, error) {
	// Загрузить версию с проверкой принадлежности к промпту (один запрос)
	v, err := s.versions.GetByIDForPrompt(ctx, versionID, promptID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, nil, ErrVersionNotFound
		}
		return nil, nil, err
	}

	// Откат через Update (проверит доступ userID + создаст снимок текущего состояния).
	// Update триггерит EventPromptUpdated → возможен unlock Refactorer.
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

func (s *Service) TogglePin(ctx context.Context, in PinInput) (*PinResult, error) {
	p, err := s.GetByID(ctx, in.PromptID, in.UserID)
	if err != nil {
		return nil, err
	}

	// Командный пин: запрещён на личных промптах, требует роли editor+
	if in.TeamWide {
		if p.TeamID == nil {
			return nil, ErrPinForbidden
		}
		if err := s.checkTeamEditAccess(ctx, p, in.UserID); err != nil {
			return nil, ErrPinForbidden
		}
	}

	// Проверяем существование пина
	existing, err := s.pins.Get(ctx, in.PromptID, in.UserID, in.TeamWide)
	if err != nil && !errors.Is(err, repo.ErrNotFound) {
		return nil, err
	}

	if existing != nil {
		// Пин существует → удалить
		if err := s.pins.Delete(ctx, in.PromptID, in.UserID, in.TeamWide); err != nil {
			return nil, err
		}
		return &PinResult{Pinned: false, TeamWide: in.TeamWide}, nil
	}

	// Пин не существует → создать
	pin := &models.PromptPin{
		PromptID:   in.PromptID,
		UserID:     in.UserID,
		IsTeamWide: in.TeamWide,
	}
	if err := s.pins.Create(ctx, pin); err != nil {
		return nil, err
	}
	return &PinResult{Pinned: true, TeamWide: in.TeamWide}, nil
}

func (s *Service) ListPinned(ctx context.Context, userID uint, teamID *uint, limit int) ([]models.Prompt, error) {
	if teamID != nil {
		if err := teamcheck.RequireMembership(ctx, s.teams, []uint{*teamID}, userID); err != nil {
			return nil, mapTeamError(err)
		}
	}
	return s.pins.ListPinned(ctx, userID, teamID, limit)
}

func (s *Service) ListRecent(ctx context.Context, userID uint, teamID *uint, limit int) ([]models.Prompt, error) {
	if teamID != nil {
		if err := teamcheck.RequireMembership(ctx, s.teams, []uint{*teamID}, userID); err != nil {
			return nil, mapTeamError(err)
		}
	}
	return s.prompts.ListRecent(ctx, userID, teamID, limit)
}

func (s *Service) GetPinStatuses(ctx context.Context, promptIDs []uint, userID uint) (map[uint]repo.PinStatus, error) {
	return s.pins.GetStatuses(ctx, promptIDs, userID)
}

// sameWorkspace проверяет что два TeamID указывают на один workspace
type timezoneKey struct{}

func ContextWithTimezone(ctx context.Context, tz string) context.Context {
	return context.WithValue(ctx, timezoneKey{}, tz)
}

func timezoneFromCtx(ctx context.Context) string {
	if tz, ok := ctx.Value(timezoneKey{}).(string); ok {
		return tz
	}
	return ""
}

func sameWorkspace(a, b *uint) bool {
	if a == nil && b == nil {
		return true // оба личные
	}
	if a == nil || b == nil {
		return false // один личный, другой командный
	}
	return *a == *b
}
