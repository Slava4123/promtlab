package quota

import "promptvault/internal/models"

// QuotaInfo — информация об использовании одного ресурса.
type QuotaInfo struct {
	Used  int `json:"used"`
	Limit int `json:"limit"`
}

// UsageSummary — полная сводка использования vs лимитов для текущего юзера.
type UsageSummary struct {
	PlanID            string    `json:"plan_id"`
	Prompts           QuotaInfo `json:"prompts"`
	Collections       QuotaInfo `json:"collections"`
	Teams             QuotaInfo `json:"teams"`
	ShareLinks        QuotaInfo `json:"share_links"`
	DailySharesToday  QuotaInfo `json:"daily_shares_today"` // Phase 14: fixed-window лимит на создание share-ссылок
	ExtUsesToday      QuotaInfo `json:"ext_uses_today"`
	MCPUsesToday      QuotaInfo `json:"mcp_uses_today"`
	Chains            QuotaInfo `json:"chains"` // Phase 16
}

// TeamUsageSummary — использование ресурсов одной команды против её
// team-pool лимита (Pack T, миграция 000070). Лимиты берутся из плана
// owner'а команды — для всех участников применяется одно и то же значение.
type TeamUsageSummary struct {
	TeamID        uint      `json:"team_id"`
	TeamName      string    `json:"team_name"`
	OwnerPlanID   string    `json:"owner_plan_id"`
	Prompts       QuotaInfo `json:"prompts"`
	Collections   QuotaInfo `json:"collections"`
	Chains        QuotaInfo `json:"chains"`
}

// MN-29: Feature types переехали в models/subscription.go (FeatureType typed
// alias). Здесь оставлены как backward-compat string-константы для callers,
// которые ещё не мигрировали (UpdateDailyUsage принимает string).
const (
	FeatureExtension   = string(models.FeatureExtension)
	FeatureMCP         = string(models.FeatureMCP)
	FeatureShareCreate = string(models.FeatureShareCreate)
)
