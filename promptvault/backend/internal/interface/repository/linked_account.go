package repository

import (
	"context"

	"promptvault/internal/models"
)

type LinkedAccountRepository interface {
	Create(ctx context.Context, la *models.LinkedAccount) error
	GetByUserID(ctx context.Context, userID uint) ([]models.LinkedAccount, error)
	GetByProviderID(ctx context.Context, provider, providerID string) (*models.LinkedAccount, error)
	Delete(ctx context.Context, userID uint, provider string) error
	CountByUserID(ctx context.Context, userID uint) (int64, error)
}
