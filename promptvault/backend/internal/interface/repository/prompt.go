package repository

import (
	"context"
	"time"

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

	// Keyset cursor pagination для MCP list_prompts (C-3).
	// Если AfterID != nil и AfterUpdatedAt != nil — используется keyset-ветка
	// `WHERE (updated_at, id) < ($ts, $id)` вместо OFFSET.
	// Когда заданы, Page игнорируется.
	AfterID        *uint
	AfterUpdatedAt *time.Time
}

type PromptRepository interface {
	Create(ctx context.Context, prompt *models.Prompt) error
	GetByID(ctx context.Context, id uint) (*models.Prompt, error)
	// GetMeta — облегчённая GetByID без Preload Tags/Collections.
	// MN-37: GetByID делает 3 SELECT'а (prompt + tags + collections) даже когда
	// caller'у нужна только проверка существования или owner_id для access-check.
	// Используется в chain.AddStep, share access-checks и т.п. — экономит ~2/3
	// query на каждом обращении.
	GetMeta(ctx context.Context, id uint) (*models.Prompt, error)
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

	// MergeWith soft-deletes mergeID после проверки что оба промпта принадлежат
	// userID. Возвращает gorm.ErrRecordNotFound если любой prompt не найден или
	// не принадлежит юзеру. Возвращает errors.New("cannot merge prompt with itself")
	// если keepID == mergeID.
	MergeWith(ctx context.Context, keepID, mergeID, userID uint) error
}
