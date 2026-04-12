package models

import (
	"encoding/json"
	"time"
)

// AuditLog — append-only журнал административных действий. Таблица в БД
// защищена через REVOKE UPDATE, DELETE (см. миграцию 000018) — попытка
// UPDATE или DELETE вернёт permission denied, что тестируется в
// audit_repo_test.go (TestAuditLog_UpdateRejected, TestAuditLog_DeleteRejected).
//
// Семантика полей:
//   - AdminID: кто совершил действие (user.id с role='admin')
//   - Action: тип действия, enumeration в usecases/audit/types.go
//     (grant_badge, revoke_badge, freeze_user, reset_password и т.д.)
//   - TargetType/TargetID: над чем совершено (user:42, prompt:10, null для global)
//   - BeforeState/AfterState: diff для восстановимости. Только non-sensitive
//     fields — password_hash и TOTP secrets никогда не попадают сюда.
//   - IP / UserAgent: forensic trail
type AuditLog struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	AdminID     uint            `gorm:"not null;index" json:"admin_id"`
	Action      string          `gorm:"size:50;not null" json:"action"`
	TargetType  string          `gorm:"size:50;not null" json:"target_type"`
	TargetID    *uint           `json:"target_id,omitempty"`
	BeforeState json.RawMessage `gorm:"type:jsonb" json:"before_state,omitempty"`
	AfterState  json.RawMessage `gorm:"type:jsonb" json:"after_state,omitempty"`
	IP          string          `gorm:"type:inet;not null" json:"ip"`
	UserAgent   string          `json:"user_agent,omitempty"`
	CreatedAt   time.Time       `gorm:"not null;index" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_log"
}
