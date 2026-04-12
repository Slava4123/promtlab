package repository

import (
	"context"

	"promptvault/internal/models"
)

type TrashCounts struct {
	Prompts     int64 `json:"prompts"`
	Collections int64 `json:"collections"`
}

type TrashRepository interface {
	// List
	ListDeletedPrompts(ctx context.Context, userID uint, teamIDs []uint, page, pageSize int) ([]models.Prompt, int64, error)
	ListDeletedCollections(ctx context.Context, userID uint, teamIDs []uint) ([]models.Collection, error)

	// Count
	CountDeleted(ctx context.Context, userID uint, teamIDs []uint) (TrashCounts, error)

	// Get deleted by ID (returns ErrNotFound if not in trash)
	GetDeletedPrompt(ctx context.Context, id uint) (*models.Prompt, error)
	GetDeletedCollection(ctx context.Context, id uint) (*models.Collection, error)

	// Restore (set deleted_at = NULL)
	RestorePrompt(ctx context.Context, id uint) error
	RestoreCollection(ctx context.Context, id uint) error

	// Hard delete (permanent)
	HardDeletePrompt(ctx context.Context, id uint) error
	HardDeleteCollection(ctx context.Context, id uint) error

	// Bulk operations
	PurgeExpired(ctx context.Context, retentionDays int) (int64, error)
	EmptyTrash(ctx context.Context, userID uint, teamIDs []uint) (int64, error)
}
