package models

import (
	"time"

	"gorm.io/gorm"
)

type Prompt struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	TeamID      *uint          `gorm:"index" json:"team_id,omitempty"`
	Title       string         `gorm:"size:300;not null" json:"title"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Model       string         `gorm:"size:100" json:"model,omitempty"`
	Favorite    bool           `gorm:"default:false" json:"favorite"`
	UsageCount  int            `gorm:"default:0" json:"usage_count"`
	User        User           `gorm:"foreignKey:UserID" json:"-"`
	Tags        []Tag            `gorm:"many2many:prompt_tags" json:"tags,omitempty"`
	Collections []Collection     `gorm:"many2many:prompt_collections" json:"collections,omitempty"`
	Versions    []PromptVersion  `gorm:"foreignKey:PromptID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
