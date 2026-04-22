package models

import (
	"encoding/json"
	"time"
)

// SmartInsight — кэш детерминированных инсайтов для Max-пользователей
// (Phase 14, миграция 000043). Вычисляются daily cron (analytics/insights.go),
// читаются эндпоинтом /api/analytics/insights и UI insights-panel.tsx.
//
// TeamID nullable: личный инсайт (team_id=NULL) или командный (team_id=X).
// Один актуальный набор на (user_id, team_id, insight_type) —
// UNIQUE constraint в БД через COALESCE(team_id, 0).
type SmartInsight struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	UserID      uint            `gorm:"not null;index" json:"user_id"`
	TeamID      *uint           `json:"team_id,omitempty"`
	InsightType string          `gorm:"size:50;not null" json:"insight_type"`
	Payload     json.RawMessage `gorm:"type:jsonb;not null" json:"payload"`
	ComputedAt  time.Time       `gorm:"not null" json:"computed_at"`
}

func (SmartInsight) TableName() string { return "user_smart_insights" }

// InsightType values — канонические ключи. Используются cron job'ом
// (analytics/insights.go) и UI (insights-panel.tsx).
const (
	InsightUnusedPrompts      = "unused_prompts"
	InsightTrending           = "trending"
	InsightDeclining          = "declining"
	InsightMostEdited         = "most_edited"
	InsightPossibleDuplicates = "possible_duplicates"
	InsightOrphanTags         = "orphan_tags"
	InsightEmptyCollections   = "empty_collections"
)
