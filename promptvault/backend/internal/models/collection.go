package models

import (
	"time"

	"gorm.io/gorm"
)

type Collection struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	TeamID      *uint          `json:"team_id,omitempty"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description,omitempty"`
	// MJ-26: HexColor + size:7 (после миграции 000065). XSS-защита через
	// BeforeSave validation + БД-constraint, независимо от точки входа.
	Color       HexColor       `gorm:"size:7;default:#8b5cf6" json:"color"`
	Icon        string         `gorm:"size:10" json:"icon,omitempty"`
	User        User           `gorm:"foreignKey:UserID" json:"-"`
	Team        *Team          `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeSave валидирует Color через HexColor regex.
func (c *Collection) BeforeSave(tx *gorm.DB) error {
	return c.Color.Validate()
}

type CollectionWithCount struct {
	Collection
	PromptCount int64 `json:"prompt_count"`
}
