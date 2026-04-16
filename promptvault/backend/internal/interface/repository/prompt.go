package repository

import (
	"context"

	"promptvault/internal/models"
)

type PromptListFilter struct {
	UserID       uint
	TeamIDs      []uint
	CollectionID *uint
	TagIDs       []uint
	FavoriteOnly bool
	Query        string
	Page         int
	PageSize     int
}

type PromptRepository interface {
	Create(ctx context.Context, prompt *models.Prompt) error
	GetByID(ctx context.Context, id uint) (*models.Prompt, error)
	Update(ctx context.Context, prompt *models.Prompt) error
	SoftDelete(ctx context.Context, id uint) error
	List(ctx context.Context, filter PromptListFilter) ([]models.Prompt, int64, error)
	SetFavorite(ctx context.Context, id uint, favorite bool) error
	IncrementUsage(ctx context.Context, id uint) error
	SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Prompt, error)
	UpdateLastUsed(ctx context.Context, id uint) error
	ListRecent(ctx context.Context, userID uint, teamID *uint, limit int) ([]models.Prompt, error)
	LogUsage(ctx context.Context, userID, promptID uint) error
	ListUsageHistory(ctx context.Context, userID uint, teamID *uint, page, pageSize int) ([]models.PromptUsageLog, int64, error)
	SuggestByPrefix(ctx context.Context, userID uint, teamID *uint, prefix string, limit int) ([]string, error)

	// GetPublicBySlug возвращает публичный промпт по slug. is_public=true,
	// deleted_at IS NULL. Используется в GET /api/public/prompts/:slug (без auth).
	GetPublicBySlug(ctx context.Context, slug string) (*models.Prompt, error)

	// ListPublic — для sitemap.xml. Возвращает (id, slug, updated_at) с LIMIT.
	ListPublic(ctx context.Context, limit int) ([]models.Prompt, error)
}
