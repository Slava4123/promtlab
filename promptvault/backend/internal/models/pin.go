package models

import "time"

type PromptPin struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PromptID   uint      `gorm:"not null" json:"prompt_id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	IsTeamWide bool      `gorm:"not null;default:false" json:"is_team_wide"`
	PinnedAt   time.Time `gorm:"not null;autoCreateTime" json:"pinned_at"`
}
