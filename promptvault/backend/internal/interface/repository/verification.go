package repository

import (
	"context"

	"promptvault/internal/models"
)

type VerificationRepository interface {
	Create(ctx context.Context, v *models.EmailVerification) error
	GetByUserID(ctx context.Context, userID uint) (*models.EmailVerification, error)
	IncrementAttempts(ctx context.Context, id uint) error
	DeleteByUserID(ctx context.Context, userID uint) error
}
