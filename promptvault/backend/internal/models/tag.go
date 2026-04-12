package models

import "time"

type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Color     string    `gorm:"size:7;default:#6366f1" json:"color"`
	UserID    uint      `gorm:"index" json:"user_id"`
	TeamID    *uint     `gorm:"index" json:"team_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
