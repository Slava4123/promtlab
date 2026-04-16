package models

import "time"

type UserStreak struct {
	UserID         uint      `gorm:"primaryKey" json:"user_id"`
	CurrentStreak  int       `gorm:"not null;default:0" json:"current_streak"`
	LongestStreak  int       `gorm:"not null;default:0" json:"longest_streak"`
	LastActiveDate string    `gorm:"type:date;not null" json:"last_active_date"`
	UpdatedAt      time.Time `json:"updated_at"`

	// ReminderSentOn — M-16: дата, когда в последний раз отправляли "не сломай серию"
	// напоминание. Защита от дубликата при повторном тике loop в тот же день.
	ReminderSentOn string `gorm:"column:reminder_sent_on;type:date" json:"-"`
}

func (UserStreak) TableName() string {
	return "user_streaks"
}
