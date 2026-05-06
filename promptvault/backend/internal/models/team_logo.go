package models

import "time"

// TeamLogoFile — bytea-хранилище загруженного логотипа команды (1:1 к teams,
// FK CASCADE). Один файл на команду — новая загрузка перезатирает старую через
// ON CONFLICT (team_id). Нулевой риск orphan'ов: каскад при удалении команды.
type TeamLogoFile struct {
	TeamID      uint      `gorm:"primaryKey;column:team_id" json:"team_id"`
	ContentType string    `gorm:"column:content_type;size:32;not null" json:"content_type"`
	SizeBytes   int64     `gorm:"column:size_bytes;not null" json:"size_bytes"`
	SHA256      string    `gorm:"column:sha256;size:64;not null" json:"sha256"`
	Bytes       []byte    `gorm:"column:bytes;not null" json:"-"`
	UploadedAt  time.Time `gorm:"column:uploaded_at" json:"uploaded_at"`
}

func (TeamLogoFile) TableName() string { return "team_logo_files" }
