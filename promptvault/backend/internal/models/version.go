package models

import "time"

type PromptVersion struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	PromptID      uint      `gorm:"not null;uniqueIndex:idx_prompt_version" json:"prompt_id"`
	VersionNumber uint      `gorm:"not null;uniqueIndex:idx_prompt_version" json:"version_number"`
	Title         string    `gorm:"size:300;not null" json:"title"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	Model         string    `gorm:"size:100" json:"model,omitempty"`
	ChangeNote    string    `gorm:"size:300" json:"change_note,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
