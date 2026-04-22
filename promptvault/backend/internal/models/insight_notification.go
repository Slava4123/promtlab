package models

import "time"

// InsightNotification — лог факта отправки email-уведомления по конкретному
// типу инсайта (миграция 000049). Используется для rate-limit 1 письмо/неделю
// через `SELECT ... WHERE sent_at > NOW() - INTERVAL '7 days'`.
// Append-only, без UPDATE/DELETE.
type InsightNotification struct {
	UserID      uint      `gorm:"primaryKey;not null" json:"user_id"`
	InsightType string    `gorm:"primaryKey;size:50;not null" json:"insight_type"`
	SentAt      time.Time `gorm:"primaryKey;not null;default:now()" json:"sent_at"`
}

func (InsightNotification) TableName() string { return "insight_notifications" }
