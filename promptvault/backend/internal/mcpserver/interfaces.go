package mcpserver

import (
	"context"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	analyticsuc "promptvault/internal/usecases/analytics"
	promptuc "promptvault/internal/usecases/prompt"
	searchuc "promptvault/internal/usecases/search"
	shareuc "promptvault/internal/usecases/share"
	teamuc "promptvault/internal/usecases/team"
	trashuc "promptvault/internal/usecases/trash"
)

type PromptService interface {
	Create(ctx context.Context, in promptuc.CreateInput) (*models.Prompt, error)
	GetByID(ctx context.Context, id, userID uint) (*models.Prompt, error)
	List(ctx context.Context, filter repo.PromptListFilter) ([]models.Prompt, int64, error)
	Update(ctx context.Context, id, userID uint, in promptuc.UpdateInput) (*models.Prompt, error)
	Delete(ctx context.Context, id, userID uint) error
	ListVersions(ctx context.Context, promptID, userID uint, page, pageSize int) ([]models.PromptVersion, int64, error)
	ToggleFavorite(ctx context.Context, id, userID uint) (*models.Prompt, error)
	TogglePin(ctx context.Context, in promptuc.PinInput) (*promptuc.PinResult, error)
	ListPinned(ctx context.Context, userID uint, teamID *uint, limit int) ([]models.Prompt, error)
	ListRecent(ctx context.Context, userID uint, teamID *uint, limit int) ([]models.Prompt, error)
	RevertToVersion(ctx context.Context, promptID, userID, versionID uint) (*models.Prompt, error)
	IncrementUsage(ctx context.Context, id, userID uint) error
}

type CollectionService interface {
	List(ctx context.Context, userID uint, teamIDs []uint) ([]models.CollectionWithCount, error)
	Create(ctx context.Context, userID uint, name, description, color, icon string, teamID *uint) (*models.Collection, error)
	Delete(ctx context.Context, id, userID uint) error
	GetByID(ctx context.Context, id, userID uint) (*models.Collection, error)
	Update(ctx context.Context, id, userID uint, name, description, color, icon string) (*models.Collection, error)
}

type TagService interface {
	List(ctx context.Context, userID uint, teamID *uint) ([]models.Tag, error)
	Create(ctx context.Context, name, color string, userID uint, teamID *uint) (*models.Tag, error)
	Delete(ctx context.Context, id, userID uint) error
}

type SearchService interface {
	Search(ctx context.Context, userID uint, teamID *uint, query string) (*searchuc.SearchOutput, error)
	Suggest(ctx context.Context, userID uint, teamID *uint, prefix string) (*searchuc.SuggestOutput, error)
}

type ShareService interface {
	CreateOrGet(ctx context.Context, promptID, userID uint) (*shareuc.ShareLinkInfo, bool, error)
	Deactivate(ctx context.Context, promptID, userID uint) error
}

type TeamService interface {
	List(ctx context.Context, userID uint) ([]teamuc.TeamListItem, error)
}

type TrashService interface {
	ListDeletedPrompts(ctx context.Context, userID uint, teamIDs []uint, page, pageSize int) ([]models.Prompt, int64, error)
	Restore(ctx context.Context, itemType trashuc.ItemType, id, userID uint) error
	PermanentDelete(ctx context.Context, itemType trashuc.ItemType, id, userID uint) error
}

type UserService interface {
	GetByID(ctx context.Context, id uint) (*models.User, error)
}

// ActivityService — MCP-интерфейс для team_activity_feed (Phase 14, B.3).
// Только read-path; запись через hooks в других usecases.
type ActivityService interface {
	ListByTeam(ctx context.Context, filter repo.TeamActivityFilter) ([]models.TeamActivityLog, *time.Time, error)
}

// AnalyticsService — MCP-интерфейс для analytics_summary / analytics_team_summary.
// Возвращает готовые dashboard-структуры из usecases/analytics.
type AnalyticsService interface {
	GetPersonalDashboard(ctx context.Context, userID uint, requestedRange analyticsuc.RangeID) (*analyticsuc.PersonalDashboard, error)
	GetTeamDashboard(ctx context.Context, userID, teamID uint, requestedRange analyticsuc.RangeID) (*analyticsuc.TeamDashboard, error)
}
