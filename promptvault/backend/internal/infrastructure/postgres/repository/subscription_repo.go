package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type subscriptionRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *subscriptionRepo {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *subscriptionRepo) GetActiveByUserID(ctx context.Context, userID uint) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("user_id = ? AND status IN ?", userID, []string{"active", "past_due"}).
		First(&sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepo) Update(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *subscriptionRepo) ListExpiring(ctx context.Context, before time.Time) ([]models.Subscription, error) {
	var subs []models.Subscription
	err := r.db.WithContext(ctx).
		Where("status = ? AND current_period_end < ?", models.SubStatusActive, before).
		Find(&subs).Error
	return subs, err
}

func (r *subscriptionRepo) ActivateWithPlanUpdate(ctx context.Context, sub *models.Subscription, userID uint, planID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(sub).Error; err != nil {
			return err
		}
		return tx.Model(&models.User{}).
			Where("id = ?", userID).
			Update("plan_id", planID).Error
	})
}

func (r *subscriptionRepo) CancelAtPeriodEnd(ctx context.Context, subID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("id = ?", subID).
		Updates(map[string]any{
			"cancel_at_period_end": true,
			"cancelled_at":        now,
			"updated_at":          now,
		}).Error
}

func (r *subscriptionRepo) ExpireAndDowngrade(ctx context.Context, subID uint, userID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&models.Subscription{}).
			Where("id = ?", subID).
			Updates(map[string]any{
				"status":     models.SubStatusExpired,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		return tx.Model(&models.User{}).
			Where("id = ?", userID).
			Update("plan_id", "free").Error
	})
}

func (r *subscriptionRepo) SetRebillId(ctx context.Context, subID uint, rebillID string) error {
	return r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("id = ?", subID).
		Update("rebill_id", rebillID).Error
}

func (r *subscriptionRepo) SetAutoRenew(ctx context.Context, subID uint, autoRenew bool) error {
	return r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("id = ?", subID).
		Update("auto_renew", autoRenew).Error
}

func (r *subscriptionRepo) ListReadyForRenewal(ctx context.Context, before time.Time) ([]models.Subscription, error) {
	var subs []models.Subscription
	err := r.db.WithContext(ctx).
		Where("status = ? AND auto_renew = ? AND rebill_id <> '' AND current_period_end <= ?",
			models.SubStatusActive, true, before).
		Find(&subs).Error
	return subs, err
}

func (r *subscriptionRepo) ExtendPeriod(ctx context.Context, subID uint, newPeriodEnd time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("id = ?", subID).
		Updates(map[string]any{
			"current_period_start": time.Now(),
			"current_period_end":   newPeriodEnd,
			"updated_at":           time.Now(),
		}).Error
}
