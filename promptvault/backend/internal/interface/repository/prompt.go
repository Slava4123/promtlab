package repository

import (
	"context"

	"promptvault/internal/models"
)

type PromptListFilter struct {
	UserID       uint
	TeamIDs      []uint
	CollectionID *uint
	TagIDs       []uint
	FavoriteOnly bool
	Query        string
	Page         int
	PageSize     int
}

type PromptRepository interface {
	Create(ctx context.Context, prompt *models.Prompt) error
	GetByID(ctx context.Context, id uint) (*models.Prompt, error)
	Update(ctx context.Context, prompt *models.Prompt) error
	SoftDelete(ctx context.Context, id uint) error
	List(ctx context.Context, filter PromptListFilter) ([]models.Prompt, int64, error)
	SetFavorite(ctx context.Context, id uint, favorite bool) error
	IncrementUsage(ctx context.Context, id uint) error
	SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Prompt, error)
}
