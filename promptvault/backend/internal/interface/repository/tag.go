package repository

import (
	"context"

	"promptvault/internal/models"
)

type TagRepository interface {
	GetOrCreate(ctx context.Context, name, color string, userID uint, teamID *uint) (*models.Tag, error)
	List(ctx context.Context, userID uint, teamID *uint) ([]models.Tag, error)
	GetByID(ctx context.Context, id uint) (*models.Tag, error)
	GetByIDs(ctx context.Context, ids []uint) ([]models.Tag, error)
	Delete(ctx context.Context, id uint) error
	SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Tag, error)
}
