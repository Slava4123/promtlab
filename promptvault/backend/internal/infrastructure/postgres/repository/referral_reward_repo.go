package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type referralRewardRepo struct {
	db *gorm.DB
}

func NewReferralRewardRepository(db *gorm.DB) repo.ReferralRewardRepository {
	return &referralRewardRepo{db: db}
}

// Create — INSERT pending. Возвращает error на UNIQUE violation (referee_id).
// Idempotency обеспечивается на уровне БД через uniqueIndex по referee_id —
// double-fire webhook'а на тот же payment не создаст дубль.
func (r *referralRewardRepo) Create(ctx context.Context, pending *models.ReferralPendingReward) error {
	return r.db.WithContext(ctx).Create(pending).Error
}

// ListEligible — кандидаты на grant: eligible_at < ts (refund-окно истекло).
// ORDER BY eligible_at ASC — старые сначала; LIMIT N — порционная обработка
// в ReferralRewardLoop. Индекс idx_referral_pending_eligible_at делает запрос
// O(log N + K).
func (r *referralRewardRepo) ListEligible(ctx context.Context, ts time.Time, limit int) ([]models.ReferralPendingReward, error) {
	var rows []models.ReferralPendingReward
	err := r.db.WithContext(ctx).
		Where("eligible_at < ?", ts).
		Order("eligible_at ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

// FindByReferee — idempotency-check для webhook handler'а: уже создавали
// pending по этому referee? Возвращает (nil, nil) если не найдено — caller
// не должен путать «нет записи» и «ошибка БД».
func (r *referralRewardRepo) FindByReferee(ctx context.Context, refereeID uint) (*models.ReferralPendingReward, error) {
	var row models.ReferralPendingReward
	err := r.db.WithContext(ctx).Where("referee_id = ?", refereeID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Delete — финализация после успешного grant'а реферальной награды.
// Hard delete: после grant'а pending больше не нужна (история — в audit/users.referral_rewarded_at).
func (r *referralRewardRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ReferralPendingReward{}, id).Error
}
