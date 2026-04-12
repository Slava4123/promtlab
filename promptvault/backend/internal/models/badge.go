package models

import "time"

// UserBadge — запись о том, что юзер разблокировал конкретный бейдж из каталога.
// Каталог определяется в usecases/badge/catalog.json (не в БД).
// Операции: INSERT и SELECT. UPDATE не используется — бейдж неотъёмен.
// DELETE используется только через admin revoke (см. usecases/admin).
type UserBadge struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;uniqueIndex:idx_user_badges_user_badge,priority:1" json:"user_id"`
	BadgeID    string    `gorm:"size:50;not null;uniqueIndex:idx_user_badges_user_badge,priority:2" json:"badge_id"`
	UnlockedAt time.Time `gorm:"not null;autoCreateTime" json:"unlocked_at"`
}

func (UserBadge) TableName() string {
	return "user_badges"
}
