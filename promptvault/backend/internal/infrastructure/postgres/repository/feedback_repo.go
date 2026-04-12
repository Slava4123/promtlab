package repository

import (
	"context"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type feedbackRepo struct {
	db *gorm.DB
}

func NewFeedbackRepository(db *gorm.DB) repo.FeedbackRepository {
	return &feedbackRepo{db: db}
}

func (r *feedbackRepo) Create(ctx context.Context, feedback *models.Feedback) error {
	return r.db.WithContext(ctx).Create(feedback).Error
}
