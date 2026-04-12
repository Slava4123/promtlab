package models

import "time"

type UserStreak struct {
	UserID         uint      `gorm:"primaryKey" json:"user_id"`
	CurrentStreak  int       `gorm:"not null;default:0" json:"current_streak"`
	LongestStreak  int       `gorm:"not null;default:0" json:"longest_streak"`
	LastActiveDate string    `gorm:"type:date;not null" json:"last_active_date"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (UserStreak) TableName() string {
	return "user_streaks"
}
