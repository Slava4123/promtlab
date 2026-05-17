package models

import "time"

// ReferralPendingReward — отложенный grant пригласившего, ожидающий refund-окна.
// На webhook payment.succeeded subscription создаёт row с eligible_at = now + 14d.
// ReferralRewardLoop через час делает SELECT WHERE eligible_at < now → grant + DELETE.
// UNIQUE на referee_id (одна награда на одного реферри).
type ReferralPendingReward struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ReferrerID uint      `gorm:"not null" json:"referrer_id"`
	RefereeID  uint      `gorm:"not null;uniqueIndex" json:"referee_id"`
	PaymentID  uint      `gorm:"not null" json:"payment_id"`
	EligibleAt time.Time `gorm:"not null;index" json:"eligible_at"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ReferralPendingReward) TableName() string { return "referral_pending_rewards" }
