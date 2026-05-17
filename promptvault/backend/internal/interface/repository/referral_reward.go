package repository

import (
	"context"
	"time"

	"promptvault/internal/models"
)

// ReferralRewardRepository — pending'и для отложенного grant'а реферальной награды.
// Записи живут от webhook'а payment.succeeded до grant'а через 14 дней.
type ReferralRewardRepository interface {
	// Create — INSERT pending. Возвращает error на UNIQUE violation (referee_id).
	Create(ctx context.Context, pending *models.ReferralPendingReward) error
	// ListEligible — SELECT WHERE eligible_at < ts ORDER BY eligible_at LIMIT N.
	ListEligible(ctx context.Context, ts time.Time, limit int) ([]models.ReferralPendingReward, error)
	// FindByReferee — для idempotency check. nil + nil если не найдено.
	FindByReferee(ctx context.Context, refereeID uint) (*models.ReferralPendingReward, error)
	// Delete — после успешного grant'а.
	Delete(ctx context.Context, id uint) error
}
