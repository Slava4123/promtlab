package repository

import (
	"context"

	"promptvault/internal/models"
)

type APIKeyRepository interface {
	Create(ctx context.Context, key *models.APIKey) error
	ListByUserID(ctx context.Context, userID uint) ([]models.APIKey, error)
	GetByHash(ctx context.Context, hash string) (*models.APIKey, error)
	Delete(ctx context.Context, id, userID uint) error
	UpdateLastUsed(ctx context.Context, id uint) error
	CountByUserID(ctx context.Context, userID uint) (int64, error)
}
