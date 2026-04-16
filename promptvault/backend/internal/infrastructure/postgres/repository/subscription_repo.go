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
	// past_due с истёкшим периодом тоже экспайрим — если retry-попытки не помогли,
	// подписка не должна зависнуть в past_due навечно.
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("status IN ? AND current_period_end < ?",
			[]models.SubscriptionStatus{models.SubStatusActive, models.SubStatusPastDue}, before).
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

func (r *subscriptionRepo) ListReadyForRenewal(ctx context.Context, before time.Time, retryAfter time.Time, maxAttempts int) ([]models.Subscription, error) {
	var subs []models.Subscription
	// Два случая:
	//  1. active подписки, чей период заканчивается в ближайшие `before` (первая попытка).
	//  2. past_due подписки, у которых была попытка раньше retryAfter и attempts < max.
	// Обе ветки требуют auto_renew=true и непустой rebill_id.
	err := r.db.WithContext(ctx).
		Where("auto_renew = ? AND rebill_id <> '' AND (("+
			"status = ? AND current_period_end <= ?) OR ("+
			"status = ? AND renewal_attempts < ? AND "+
			"(last_renewal_attempt_at IS NULL OR last_renewal_attempt_at <= ?)))",
			true,
			models.SubStatusActive, before,
			models.SubStatusPastDue, maxAttempts, retryAfter).
		Find(&subs).Error
	return subs, err
}

func (r *subscriptionRepo) ExtendPeriod(ctx context.Context, subID uint, newPeriodEnd time.Time) error {
	// При успешном продлении сбрасываем retry-счётчики, pre_expire_stage и
	// возвращаем в active (восстановление после retry-успеха или ручного продления).
	// Без сброса pre_expire_stage юзер на продлённую подписку не получит
	// reminder на следующем цикле.
	return r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("id = ?", subID).
		Updates(map[string]any{
			"current_period_start":    time.Now(),
			"current_period_end":      newPeriodEnd,
			"status":                  models.SubStatusActive,
			"renewal_attempts":        0,
			"last_renewal_attempt_at": nil,
			"pre_expire_stage":        0,
			"updated_at":              time.Now(),
		}).Error
}

func (r *subscriptionRepo) RecordRenewalFailure(ctx context.Context, subID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("id = ?", subID).
		Updates(map[string]any{
			"status":                  models.SubStatusPastDue,
			"renewal_attempts":        gorm.Expr("renewal_attempts + 1"),
			"last_renewal_attempt_at": now,
			"updated_at":              now,
		}).Error
}

// ListPreExpiring возвращает active подписки с auto_renew=false,
// у которых period_end попадает в окно (now, upTo], и pre_expire_stage < minStage.
// Используется ReminderLoop — auto_renew=true юзеры получают retry-уведомления
// из RenewalLoop, им pre-expire напоминания не нужны.
// Preload плана — чтобы ReminderLoop мог показать читаемое имя в письме.
func (r *subscriptionRepo) ListPreExpiring(ctx context.Context, now, upTo time.Time, minStage int16) ([]models.Subscription, error) {
	var subs []models.Subscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("status = ?", models.SubStatusActive).
		Where("auto_renew = ?", false).
		Where("current_period_end > ?", now).
		Where("current_period_end <= ?", upTo).
		Where("pre_expire_stage < ?", minStage).
		Order("current_period_end ASC").
		Limit(200).
		Find(&subs).Error
	return subs, err
}

func (r *subscriptionRepo) SetPreExpireStage(ctx context.Context, subID uint, stage int16) error {
	return r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("id = ?", subID).
		Updates(map[string]any{
			"pre_expire_stage": stage,
			"updated_at":       time.Now(),
		}).Error
}
