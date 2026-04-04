package repository

import (
	"context"

	"promptvault/internal/models"
)

type CollectionRepository interface {
	Create(ctx context.Context, c *models.Collection) error
	GetByID(ctx context.Context, id uint) (*models.Collection, error)
	Update(ctx context.Context, c *models.Collection) error
	Delete(ctx context.Context, id uint) error
	CountPrompts(ctx context.Context, collectionID uint) (int64, error)
	GetByIDs(ctx context.Context, ids []uint) ([]models.Collection, error)
	ListWithCounts(ctx context.Context, userID uint, teamIDs []uint) ([]models.CollectionWithCount, error)
	SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Collection, error)
}
