package repository

import (
	"context"

	"promptvault/internal/models"
)

type StreakRepository interface {
	RecordActivity(ctx context.Context, userID uint, today string) error
	GetByUserID(ctx context.Context, userID uint) (*models.UserStreak, error)
}
