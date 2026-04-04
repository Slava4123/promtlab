package models

import "time"

type LinkedAccount struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"uniqueIndex:idx_user_provider;not null" json:"user_id"`
	Provider   string    `gorm:"uniqueIndex:idx_user_provider;size:20;not null" json:"provider"`
	ProviderID string    `gorm:"size:255;not null;index" json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}
