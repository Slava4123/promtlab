package repository

import (
	"context"
	"time"

	"promptvault/internal/models"
)

type PinStatus struct {
	PinnedPersonal bool
	PinnedTeam     bool
	PinnedAt       *time.Time
}

type PinRepository interface {
	Create(ctx context.Context, pin *models.PromptPin) error
	Delete(ctx context.Context, promptID, userID uint, teamWide bool) error
	Get(ctx context.Context, promptID, userID uint, teamWide bool) (*models.PromptPin, error)
	GetStatuses(ctx context.Context, promptIDs []uint, userID uint) (map[uint]PinStatus, error)
	ListPinned(ctx context.Context, userID uint, teamID *uint, limit int) ([]models.Prompt, error)
}
