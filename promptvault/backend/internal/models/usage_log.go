package models

import "time"

type PromptUsageLog struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	UserID   uint      `gorm:"not null" json:"user_id"`
	PromptID uint      `gorm:"not null" json:"prompt_id"`
	UsedAt   time.Time `gorm:"not null;autoCreateTime" json:"used_at"`
	Prompt   Prompt    `gorm:"foreignKey:PromptID" json:"prompt"`
}

func (PromptUsageLog) TableName() string {
	return "prompt_usage_log"
}
