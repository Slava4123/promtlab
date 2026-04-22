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

// PersonalDashboard — сводка для /api/analytics/personal.
type PersonalDashboard struct {
	Range           RangeID                `json:"range"`
	UsagePerDay     []repo.UsagePoint      `json:"usage_per_day"`
	TopPrompts      []repo.PromptUsageRow  `json:"top_prompts"`
	PromptsCreated  []repo.UsagePoint      `json:"prompts_created_per_day"`
	PromptsUpdated  []repo.UsagePoint      `json:"prompts_updated_per_day"`
	ShareViews      []repo.UsagePoint      `json:"share_views_per_day"`
	TopShared       []repo.PromptUsageRow  `json:"top_shared"`
	Quotas          *quotauc.UsageSummary  `json:"quotas,omitempty"`
}

// TeamDashboard — personal-набор + contributors leaderboard для team scope.
type TeamDashboard struct {
	Range          RangeID                `json:"range"`
	UsagePerDay    []repo.UsagePoint      `json:"usage_per_day"`
	TopPrompts     []repo.PromptUsageRow  `json:"top_prompts"`
	PromptsCreated []repo.UsagePoint      `json:"prompts_created_per_day"`
	PromptsUpdated []repo.UsagePoint      `json:"prompts_updated_per_day"`
	Contributors   []repo.ContributorRow  `json:"contributors"`
}

// PromptAnalytics — per-prompt метрики для /api/analytics/prompts/:id.
type PromptAnalytics struct {
	PromptID         uint              `json:"prompt_id"`
	UsagePerDay      []repo.UsagePoint `json:"usage_per_day"`
	ShareViewsPerDay []repo.UsagePoint `json:"share_views_per_day"`
}
