package repository

import (
	"context"

	"promptvault/internal/models"
)

type ShareLinkRepository interface {
	Create(ctx context.Context, link *models.ShareLink) error
	GetByToken(ctx context.Context, token string) (*models.ShareLink, error)
	GetActiveByPromptID(ctx context.Context, promptID uint) (*models.ShareLink, error)
	Deactivate(ctx context.Context, promptID uint) error
	IncrementViewCount(ctx context.Context, id uint) error
}
