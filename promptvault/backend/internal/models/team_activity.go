package models

import (
	"encoding/json"
	"time"
)

// TeamActivityLog — append-only лента продуктовых событий внутри команды
// (Phase 14, миграция 000040). Видят все члены команды (viewer+).
//
// Денормализация actor_email/actor_name/target_label — намеренная, чтобы
// feed рендерился без JOIN'ов и переживал удаление user/target. При
// удалении аккаунта actor_email/actor_name anonymize'ятся отдельно
// (см. usecases/user.DeleteAccount).
//
// Tamper-evidence: BEFORE UPDATE триггер в БД запрещает изменение.
// DELETE разрешён — используется retention cron (Pro 90д / Max 365д).
type TeamActivityLog struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	TeamID      uint            `gorm:"not null;index" json:"team_id"`
	ActorID     *uint           `json:"actor_id,omitempty"`
	ActorEmail  string          `gorm:"size:255;not null" json:"actor_email"`
	ActorName   string          `gorm:"size:255" json:"actor_name,omitempty"`
	EventType   string          `gorm:"size:50;not null" json:"event_type"`
	TargetType  string          `gorm:"size:50;not null" json:"target_type"`
	TargetID    *uint           `json:"target_id,omitempty"`
	TargetLabel string          `gorm:"size:500" json:"target_label,omitempty"`
	Metadata    json.RawMessage `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt   time.Time       `gorm:"not null" json:"created_at"`
}

func (TeamActivityLog) TableName() string { return "team_activity_log" }

// Event types — канонические значения event_type.
// Single source of truth, чтобы исключить typo при вызовах activity.Log.
const (
	ActivityPromptCreated     = "prompt.created"
	ActivityPromptUpdated     = "prompt.updated"
	ActivityPromptDeleted     = "prompt.deleted"
	ActivityPromptRestored    = "prompt.restored"
	ActivityCollectionCreated = "collection.created"
	ActivityCollectionUpdated = "collection.updated"
	ActivityCollectionDeleted = "collection.deleted"
	ActivityTagCreated        = "tag.created"
	ActivityTagDeleted        = "tag.deleted"
	ActivityShareCreated      = "share.created"
	ActivityShareRevoked      = "share.revoked"
	ActivityMemberAdded       = "member.added"
	ActivityMemberRemoved     = "member.removed"
	ActivityRoleChanged       = "role.changed"
)

// Target types для TargetType.
const (
	TargetPrompt     = "prompt"
	TargetCollection = "collection"
	TargetTag        = "tag"
	TargetShare      = "share"
	TargetMember     = "member"
)

// Anonymized actor values — используются в AnonymizeActor после удаления user.
const (
	AnonymizedActorEmail = "deleted@anonymized.local"
	AnonymizedActorName  = "(deleted user)"
)
