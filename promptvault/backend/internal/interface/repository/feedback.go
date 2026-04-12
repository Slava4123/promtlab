package repository

import (
	"context"

	"promptvault/internal/models"
)

type FeedbackRepository interface {
	Create(ctx context.Context, feedback *models.Feedback) error
}
