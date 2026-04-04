package models

import "time"

type EmailVerification struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	Code      string    `gorm:"size:6;not null"`
	Attempts  int       `gorm:"default:0;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}

const MaxVerificationAttempts = 5
