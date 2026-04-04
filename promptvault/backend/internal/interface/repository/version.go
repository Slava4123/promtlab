package repository

import (
	"context"

	"promptvault/internal/models"
)

type VersionRepository interface {
	// CreateWithNextVersion атомарно вычисляет следующий номер версии и создаёт запись в одной транзакции.
	CreateWithNextVersion(ctx context.Context, v *models.PromptVersion) error
	ListByPromptID(ctx context.Context, promptID uint, page, pageSize int) ([]models.PromptVersion, int64, error)
	// GetByIDForPrompt возвращает версию только если она принадлежит указанному промпту.
	GetByIDForPrompt(ctx context.Context, versionID, promptID uint) (*models.PromptVersion, error)
}
