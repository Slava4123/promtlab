package models

import "time"

// UserTOTP — TOTP secret пользователя. Одна запись на юзера (PRIMARY KEY user_id).
// ConfirmedAt=nil означает что enrollment начат, но ещё не подтверждён первым
// кодом из Authenticator — такую запись можно безопасно перезаписать при re-enroll.
// Secret хранится в plaintext (как password не может быть хеширован —
// нам надо заново генерировать TOTP при каждой проверке).
type UserTOTP struct {
	UserID      uint       `gorm:"primaryKey" json:"user_id"`
	Secret      string     `gorm:"size:64;not null" json:"-"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (UserTOTP) TableName() string {
	return "user_totp"
}

// IsConfirmed — удобный предикат для проверки «прошёл ли юзер enrollment flow».
func (t *UserTOTP) IsConfirmed() bool {
	return t.ConfirmedAt != nil
}
