package models

import "time"

// UserTOTPBackupCode — одноразовый recovery-код для входа когда TOTP-устройство
// недоступно (потерян телефон). CodeHash — bcrypt (как у password), сравнивается
// через bcrypt.CompareHashAndPassword. После использования ставится used_at=NOW
// и больше не принимается.
type UserTOTPBackupCode struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"not null;index" json:"user_id"`
	CodeHash  string     `gorm:"size:128;not null" json:"-"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (UserTOTPBackupCode) TableName() string {
	return "user_totp_backup_codes"
}

// IsUsed — удобный предикат для UI (показать зачёркнутым / скрыть).
func (c *UserTOTPBackupCode) IsUsed() bool {
	return c.UsedAt != nil
}
