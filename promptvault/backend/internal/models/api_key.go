package models

import "time"

type APIKey struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null;index" json:"user_id"`
	Name       string     `gorm:"size:100;not null" json:"name"`
	KeyPrefix  string     `gorm:"size:20;not null" json:"key_prefix"`
	KeyHash    string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
