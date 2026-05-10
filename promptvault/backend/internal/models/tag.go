package models

import (
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:50;not null" json:"name"`
	Color     HexColor       `gorm:"size:7;default:#6366f1" json:"color"`
	UserID    uint           `gorm:"index" json:"user_id"`
	TeamID    *uint          `gorm:"index" json:"team_id,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	// MN-34: parity с Collection.DeletedAt — Tag сейчас удаляются hard
	// (DeleteOrphans + Delete напрямую). Soft-delete позволяет:
	//  1. Восстановить случайно удалённый тег вместе с привязкой к промптам.
	//  2. Audit-trail: видим кто и когда удалил тег команды.
	// Миграция 000066 добавляет deleted_at column + partial index.
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeSave валидирует Color через HexColor regex.
// MJ-26: защищает от XSS через MCP `create_tag` (раньше Service.Create
// принимал любую строку без regex; HTTP-validator стоял только на handler-уровне).
func (t *Tag) BeforeSave(tx *gorm.DB) error {
	return t.Color.Validate()
}
