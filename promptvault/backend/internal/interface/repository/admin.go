package repository

import (
	"context"
	"time"

	"promptvault/internal/models"
)

// UserListFilter — параметры для админ-страницы /admin/users.
// Nil-указатели и пустые строки означают «не фильтровать по этому полю».
type UserListFilter struct {
	// Query — ILIKE-поиск по email и username (case-insensitive).
	Query string
	// Role — "user"/"admin"/"" (пусто = все).
	Role string
	// Status — "active"/"frozen"/"" (пусто = все).
	Status string
	// SortBy — "created_at" (default) или "email".
	SortBy string
	// SortDesc — true для DESC (default для created_at — тоже true: свежие сверху).
	SortDesc bool
	Page     int
	PageSize int
}

// UserSummary — компактное представление юзера для списка /admin/users.
// Без counts (тяжёлые агрегации берутся только в GetUserDetail).
type UserSummary struct {
	ID            uint
	Email         string
	Name          string
	Username      string
	Role          string
	Status        string
	EmailVerified bool
	CreatedAt     time.Time
}

// UserDetail — детальное представление юзера для /admin/users/{id}.
// Aggregation-поля заполняются через JOINs/subqueries в admin_repo.GetUserDetail.
type UserDetail struct {
	User             *models.User
	PromptCount      int64
	CollectionCount  int64
	BadgeCount       int64
	TotalUsage       int64
	LinkedProviders  []string
	UnlockedBadgeIDs []string // список badge_id'шек разблокированных у юзера — для admin UI
}

// AdminRepository агрегирует read-методы для админ-панели + UpdateStatus
// для freeze/unfreeze. Не дублирует существующие методы UserRepository —
// отвечает только за специфичные для admin запросы.
type AdminRepository interface {
	// ListUsers возвращает страницу юзеров по фильтру + общий count.
	// Упорядочено по SortBy/SortDesc, по умолчанию — created_at DESC.
	ListUsers(ctx context.Context, filter UserListFilter) ([]UserSummary, int64, error)

	// GetUserDetail возвращает расширенное представление одного юзера
	// с агрегациями (prompt_count, collection_count, badge_count, total_usage).
	// Если юзера не существует — возвращает ErrNotFound.
	GetUserDetail(ctx context.Context, userID uint) (*UserDetail, error)

	// UpdateStatus обновляет users.status для freeze/unfreeze.
	// Валидация значения выполняется в usecase-слое.
	UpdateStatus(ctx context.Context, userID uint, status models.UserStatus) error

	// CountUsers возвращает агрегации для /admin/health dashboard.
	// Одним методом вместо 4-х отдельных, чтобы уменьшить round-trips в DB.
	CountUsers(ctx context.Context) (total, admins, active, frozen int64, err error)
}
