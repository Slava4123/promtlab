package models

import "time"

type FeedbackType string

const (
	FeedbackBug     FeedbackType = "bug"
	FeedbackFeature FeedbackType = "feature"
	FeedbackOther   FeedbackType = "other"
)

// IsValid возвращает true, если значение допустимо.
// MN-35: parity с FeedbackStatus.IsValid — раньше usecases/feedback.Submit
// делал inline switch без переиспользуемой проверки.
func (t FeedbackType) IsValid() bool {
	switch t {
	case FeedbackBug, FeedbackFeature, FeedbackOther:
		return true
	}
	return false
}

// FeedbackStatus — статус отзыва для admin-обработки.
// new (default)  — отзыв ещё не прочитан админом.
// read           — админ открыл detail и пометил как прочитанный.
// archived       — убран из основного списка (не удалено, можно вернуть).
type FeedbackStatus string

const (
	FeedbackStatusNew      FeedbackStatus = "new"
	FeedbackStatusRead     FeedbackStatus = "read"
	FeedbackStatusArchived FeedbackStatus = "archived"
)

// IsValid возвращает true, если значение допустимо как target в API.
// Используется для валидации PATCH /admin/feedbacks/:id/status.
func (s FeedbackStatus) IsValid() bool {
	switch s {
	case FeedbackStatusNew, FeedbackStatusRead, FeedbackStatusArchived:
		return true
	}
	return false
}

type Feedback struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	Type      FeedbackType   `gorm:"type:feedback_type;not null;default:'other'" json:"type"`
	Status    FeedbackStatus `gorm:"type:feedback_status;not null;default:'new'" json:"status"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	PageURL   string         `gorm:"size:2000" json:"page_url"`
	CreatedAt time.Time      `json:"created_at"`
}
