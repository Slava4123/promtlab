package repository

import (
	"context"
	"time"

	"promptvault/internal/models"
)

type ShareLinkRepository interface {
	Create(ctx context.Context, link *models.ShareLink) error
	GetByToken(ctx context.Context, token string) (*models.ShareLink, error)
	GetActiveByPromptID(ctx context.Context, promptID uint) (*models.ShareLink, error)
	Deactivate(ctx context.Context, promptID uint) error
	IncrementViewCount(ctx context.Context, id uint) error
	// CleanupExpired — Phase 16-Y. Удаляет ссылки с expires_at < now()-grace.
	// Grace period (30d) даёт шанс показать страницу «истекла» вместо 404
	// сразу после expires_at; через grace удаляем окончательно.
	// Возвращает количество удалённых строк.
	CleanupExpired(ctx context.Context, grace time.Duration) (int64, error)
}
