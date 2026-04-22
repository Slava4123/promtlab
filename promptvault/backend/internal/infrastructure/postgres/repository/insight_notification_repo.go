package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type insightNotificationRepo struct {
	db *gorm.DB
}

func NewInsightNotificationRepository(db *gorm.DB) repo.InsightNotificationRepository {
	return &insightNotificationRepo{db: db}
}

func (r *insightNotificationRepo) RecentlySent(ctx context.Context, userID uint, insightType string, within time.Duration) (bool, error) {
	threshold := time.Now().Add(-within)
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.InsightNotification{}).
		Where("user_id = ? AND insight_type = ? AND sent_at > ?", userID, insightType, threshold).
		Count(&count).Error
	return count > 0, err
}

func (r *insightNotificationRepo) Record(ctx context.Context, userID uint, insightType string) error {
	return r.db.WithContext(ctx).Create(&models.InsightNotification{
		UserID:      userID,
		InsightType: insightType,
		SentAt:      time.Now(),
	}).Error
}
