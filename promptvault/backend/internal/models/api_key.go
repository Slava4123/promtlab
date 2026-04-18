package models

import (
	"time"

	"github.com/lib/pq"
)

type APIKey struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	KeyPrefix    string         `gorm:"size:20;not null" json:"key_prefix"`
	KeyHash      string         `gorm:"size:64;not null;uniqueIndex" json:"-"`
	LastUsedAt   *time.Time     `json:"last_used_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	ReadOnly     bool           `gorm:"not null;default:false" json:"read_only"`
	TeamID       *uint          `json:"team_id,omitempty"`
	AllowedTools pq.StringArray `gorm:"type:text[]" json:"allowed_tools,omitempty"`
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
}
