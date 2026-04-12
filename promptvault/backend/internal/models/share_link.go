package models

import "time"

type ShareLink struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	PromptID     uint       `gorm:"not null;index" json:"prompt_id"`
	UserID       uint       `gorm:"not null;index" json:"user_id"`
	Token        string     `gorm:"size:64;not null;uniqueIndex" json:"token"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	ViewCount    int        `gorm:"default:0" json:"view_count"`
	LastViewedAt *time.Time `json:"last_viewed_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	Prompt       Prompt     `gorm:"foreignKey:PromptID" json:"-"`
	User         User       `gorm:"foreignKey:UserID" json:"-"`
}
