package models

import "time"

// ShareView — запись о просмотре публичной шар-ссылки (Phase 14, миграция 000042).
// Пишется ТОЛЬКО для владельцев тарифа Pro/Max — Free в share_view_log не попадает,
// им остаётся только total view_count в share_links. Проверка tier — на стороне
// usecases/share или delivery/http/share/public.go.
type ShareView struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ShareLinkID     uint      `gorm:"not null;index" json:"share_link_id"`
	ViewedAt        time.Time `gorm:"not null;index" json:"viewed_at"`
	Referer         string    `gorm:"size:500" json:"referer,omitempty"`
	Country         string    `gorm:"size:2" json:"country,omitempty"`
	UserAgentFamily string    `gorm:"size:50" json:"user_agent_family,omitempty"`
}

func (ShareView) TableName() string { return "share_view_log" }
