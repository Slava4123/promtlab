package repository

import (
	"context"

	"promptvault/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	SearchUsers(ctx context.Context, query string, limit int) ([]models.User, error)
	Update(ctx context.Context, user *models.User) error
}
