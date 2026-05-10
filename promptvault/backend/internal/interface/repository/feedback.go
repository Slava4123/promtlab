package repository

import (
	"context"

	"promptvault/internal/models"
)

// FeedbackListFilter — параметры выборки для admin /admin/feedbacks.
// Все поля опциональны: пустая строка/значение = не фильтруем.
type FeedbackListFilter struct {
	Type     models.FeedbackType   // "" = любой тип
	Status   models.FeedbackStatus // "" = любой статус
	Query    string                // ILIKE по message (case-insensitive substring)
	Page     int                   // 1-based; ≤0 → 1
	PageSize int                   // ≤0 или >100 → 20
}

// FeedbackListItem — denormalized строка списка с user-полями для UI таблицы.
// Не наследуем модель Feedback напрямую, чтобы JSON shape не зависел от GORM
// и был стабилен для admin frontend (имена в snake_case через JSON tags).
type FeedbackListItem struct {
	ID        uint                  `json:"id"`
	UserID    uint                  `json:"user_id"`
	UserEmail string                `json:"user_email"`
	UserName  string                `json:"user_name"`
	Type      models.FeedbackType   `json:"type"`
	Status    models.FeedbackStatus `json:"status"`
	Message   string                `json:"message"`
	PageURL   string                `json:"page_url"`
	CreatedAt string                `json:"created_at"` // RFC3339, форматируется в repo
}

// FeedbackDetail — полный detail для GET /admin/feedbacks/:id.
// Идентичен FeedbackListItem; вынесен отдельным типом, чтобы было место для
// будущих полей (например, attachments, reply, и т.п.) без поломки list shape.
type FeedbackDetail = FeedbackListItem

type FeedbackRepository interface {
	// Create — used by user submission (POST /api/feedback). NOT mutating.
	Create(ctx context.Context, feedback *models.Feedback) error

	// List возвращает страницу feedback'ов и общее число под фильтром.
	// Сортировка: created_at DESC (новые сверху).
	List(ctx context.Context, filter FeedbackListFilter) ([]FeedbackListItem, int64, error)

	// GetByID возвращает один feedback с user-полями.
	// nil + ErrNotFound (или похожая обёртка) если не найден — caller обрабатывает.
	GetByID(ctx context.Context, id uint) (*FeedbackDetail, error)

	// UpdateStatus меняет status. Идемпотентно: тот же status не считается ошибкой.
	UpdateStatus(ctx context.Context, id uint, status models.FeedbackStatus) error

	// Delete удаляет feedback навсегда. Hard delete (нет soft).
	// Используется только админом для спама/PII.
	Delete(ctx context.Context, id uint) error
}
