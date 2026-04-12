package badge

import "time"

// EventType — тип события, которое триггерит evaluate бейджей.
// Каждый бейдж в каталоге подписан на один или несколько EventType через
// поле Triggers. badge.Service.Evaluate использует indexed lookup byEvent
// чтобы не прогонять все 11 бейджей на каждом событии.
type EventType string

const (
	EventPromptCreated     EventType = "prompt_created"
	EventPromptUsed        EventType = "prompt_used"
	EventPromptUpdated     EventType = "prompt_updated"
	EventCollectionCreated EventType = "collection_created"
)

// Event — данные события для передачи в Evaluate.
// TeamID используется для short-circuit team/solo условий: event без TeamID
// никогда не будет триггерить team_* бейджи, и наоборот.
// PromptID передаётся для future short-circuit в Переработчике (не в MVP).
type Event struct {
	Type     EventType
	TeamID   *uint
	PromptID uint
}

// ConditionType — тип условия разблокировки бейджа.
// Каждый тип маппится на один метод BadgeRepository.
type ConditionType string

const (
	CondSoloPromptCount      ConditionType = "solo_prompt_count"
	CondTeamPromptCount      ConditionType = "team_prompt_count"
	CondTotalPromptCount     ConditionType = "total_prompt_count"
	CondSoloCollectionCount  ConditionType = "solo_collection_count"
	CondTeamCollectionCount  ConditionType = "team_collection_count"
	CondTotalUsage           ConditionType = "total_usage"
	CondVersionedPromptCount ConditionType = "versioned_prompt_count"
	CondCurrentStreak        ConditionType = "current_streak"
)

// Condition описывает условие разблокировки. Threshold — порог, ≥ которого
// значение засчитывается как выполнение. MinVersions — дополнительный параметр
// только для CondVersionedPromptCount (сколько версий должно быть у промпта).
type Condition struct {
	Type        ConditionType `json:"type"`
	Threshold   int64         `json:"threshold"`
	MinVersions int           `json:"min_versions,omitempty"`
}

// BadgeCategory — визуальная группировка на странице /badges (frontend),
// для бизнес-логики evaluate не используется. Значения согласованы с UI.
type BadgeCategory string

const (
	CategoryPersonal  BadgeCategory = "personal"  // соло-бейджи
	CategoryTeam      BadgeCategory = "team"      // командные
	CategoryMilestone BadgeCategory = "milestone" // общие флагманские
	CategoryStreak    BadgeCategory = "streak"    // стрики
)

// Badge — запись каталога, загруженная из catalog.json. Иммутабельная.
// ID — стабильный ключ, используется как badge_id в user_badges.
type Badge struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Icon        string        `json:"icon"`
	Category    BadgeCategory `json:"category"`
	Triggers    []EventType   `json:"triggers"`
	Condition   Condition     `json:"condition"`
}

// BadgeWithState — бейдж + состояние пользователя для GET /api/badges.
// Для unlocked: Unlocked=true, UnlockedAt=<время>, Progress=Target (финальное).
// Для locked: Unlocked=false, UnlockedAt=nil, Progress=<текущий>, Target=Condition.Threshold.
type BadgeWithState struct {
	Badge
	Unlocked   bool       `json:"unlocked"`
	UnlockedAt *time.Time `json:"unlocked_at,omitempty"`
	Progress   int64      `json:"progress"`
	Target     int64      `json:"target"`
}
