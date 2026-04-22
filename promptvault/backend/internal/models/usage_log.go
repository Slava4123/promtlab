package models

import "time"

type PromptUsageLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	PromptID  uint      `gorm:"not null" json:"prompt_id"`
	UsedAt    time.Time `gorm:"not null;autoCreateTime" json:"used_at"`
	// ModelUsed — копия prompts.model на момент использования. NULL для
	// старых записей и промптов без указанной модели. Phase 14.2 для
	// segmentation по AI-моделям.
	ModelUsed string `gorm:"size:50" json:"model_used,omitempty"`
	Prompt    Prompt `gorm:"foreignKey:PromptID" json:"prompt"`
}

func (PromptUsageLog) TableName() string {
	return "prompt_usage_log"
}
