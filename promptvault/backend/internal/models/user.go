package models

import "time"

type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Email         string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash  string    `gorm:"size:255" json:"-"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	Username      string    `gorm:"size:30" json:"username,omitempty"`
	AvatarURL     string    `gorm:"size:500" json:"avatar_url,omitempty"`
	EmailVerified bool      `gorm:"default:false" json:"email_verified"`
	DefaultModel  string    `gorm:"size:100;default:anthropic/claude-sonnet-4" json:"default_model"`
	TokenNonce    string    `gorm:"size:64" json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	LinkedAccounts []LinkedAccount `gorm:"foreignKey:UserID" json:"linked_accounts,omitempty"`
}

func (u *User) HasPassword() bool {
	return u.PasswordHash != ""
}
