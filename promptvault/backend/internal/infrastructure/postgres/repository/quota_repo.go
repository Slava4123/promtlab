package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"promptvault/internal/models"
)

type quotaRepo struct {
	db *gorm.DB
}

func NewQuotaRepository(db *gorm.DB) *quotaRepo {
	return &quotaRepo{db: db}
}

func (r *quotaRepo) CountPrompts(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Prompt{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

func (r *quotaRepo) CountCollections(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Collection{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

func (r *quotaRepo) CountTeamsOwned(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Team{}).
		Where("created_by = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *quotaRepo) CountActiveShareLinks(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ShareLink{}).
		Where("user_id = ? AND is_active = true", userID).
		Count(&count).Error
	return count, err
}

func (r *quotaRepo) CountTeamMembers(ctx context.Context, teamID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.TeamMember{}).
		Where("team_id = ?", teamID).
		Count(&count).Error
	return int(count), err
}

func (r *quotaRepo) GetDailyUsage(ctx context.Context, userID uint, date time.Time, featureType string) (int, error) {
	var usage models.DailyFeatureUsage
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND usage_date = ? AND feature_type = ?", userID, date.Format("2006-01-02"), featureType).
		First(&usage).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return usage.Count, nil
}

func (r *quotaRepo) GetTotalUsage(ctx context.Context, userID uint, featureType string) (int, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.DailyFeatureUsage{}).
		Where("user_id = ? AND feature_type = ?", userID, featureType).
		Select("COALESCE(SUM(count), 0)").
		Scan(&total).Error
	return int(total), err
}

func (r *quotaRepo) IncrementDailyUsage(ctx context.Context, userID uint, date time.Time, featureType string) error {
	return r.db.WithContext(ctx).Exec(
		`INSERT INTO daily_feature_usage (user_id, usage_date, feature_type, count)
		 VALUES (?, ?, ?, 1)
		 ON CONFLICT (user_id, usage_date, feature_type)
		 DO UPDATE SET count = daily_feature_usage.count + 1`,
		userID, date.Format("2006-01-02"), featureType,
	).Error
}
