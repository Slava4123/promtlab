package models

import "time"

type FeedbackType string

const (
	FeedbackBug     FeedbackType = "bug"
	FeedbackFeature FeedbackType = "feature"
	FeedbackOther   FeedbackType = "other"
)

type Feedback struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	UserID    uint         `gorm:"not null" json:"user_id"`
	Type      FeedbackType `gorm:"type:feedback_type;not null;default:'other'" json:"type"`
	Message   string       `gorm:"type:text;not null" json:"message"`
	PageURL   string       `gorm:"size:2000" json:"page_url"`
	CreatedAt time.Time    `json:"created_at"`
}
