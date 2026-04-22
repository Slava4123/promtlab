package analytics

import (
	repo "promptvault/internal/interface/repository"
	quotauc "promptvault/internal/usecases/quota"
)

// RangeID — семантический период для dashboard-запросов. UI передаёт
// "7d"/"30d"/"90d"/"365d" в query param — Service clamp'ит по тарифу
// (см. retention.go).
type RangeID string

const (
	Range7d   RangeID = "7d"
	Range30d  RangeID = "30d"
	Range90d  RangeID = "90d"
	Range365d RangeID = "365d"
)

// Totals — агрегированные суммы за период (для "прошлый период" сравнения
// в MetricCard). Отдельный struct вместо расчёта на фронте — единая логика.
type Totals struct {
	Uses       int64 `json:"uses"`
	Created    int64 `json:"created"`
	Updated    int64 `json:"updated"`
	ShareViews int64 `json:"share_views"`
}

// PersonalDashboard — сводка для /api/analytics/personal.
type PersonalDashboard struct {
	Range          RangeID               `json:"range"`
	UsagePerDay    []repo.UsagePoint     `json:"usage_per_day"`
	TopPrompts     []repo.PromptUsageRow `json:"top_prompts"`
	PromptsCreated []repo.UsagePoint     `json:"prompts_created_per_day"`
	PromptsUpdated []repo.UsagePoint     `json:"prompts_updated_per_day"`
	ShareViews     []repo.UsagePoint     `json:"share_views_per_day"`
	TopShared      []repo.PromptUsageRow `json:"top_shared"`
	Quotas         *quotauc.UsageSummary `json:"quotas,omitempty"`
	// Totals для сравнения с предыдущим периодом в метриках сверху.
	TotalsCurrent  Totals                `json:"totals_current"`
	TotalsPrevious Totals                `json:"totals_previous"`
	// UsageByModel — сегментация use'ов по AI-модели (Claude/GPT/etc).
	UsageByModel   []repo.ModelUsageRow  `json:"usage_by_model"`
}

// TeamDashboard — personal-набор + contributors leaderboard для team scope.
type TeamDashboard struct {
	Range          RangeID               `json:"range"`
	UsagePerDay    []repo.UsagePoint     `json:"usage_per_day"`
	TopPrompts     []repo.PromptUsageRow `json:"top_prompts"`
	PromptsCreated []repo.UsagePoint     `json:"prompts_created_per_day"`
	PromptsUpdated []repo.UsagePoint     `json:"prompts_updated_per_day"`
	Contributors   []repo.ContributorRow `json:"contributors"`
	TotalsCurrent  Totals                `json:"totals_current"`
	TotalsPrevious Totals                `json:"totals_previous"`
	UsageByModel   []repo.ModelUsageRow  `json:"usage_by_model"`
}

// PromptAnalytics — per-prompt метрики для /api/analytics/prompts/:id.
type PromptAnalytics struct {
	PromptID         uint              `json:"prompt_id"`
	UsagePerDay      []repo.UsagePoint `json:"usage_per_day"`
	ShareViewsPerDay []repo.UsagePoint `json:"share_views_per_day"`
}
