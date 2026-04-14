package quota

// QuotaInfo — информация об использовании одного ресурса.
type QuotaInfo struct {
	Used    int  `json:"used"`
	Limit   int  `json:"limit"`
	IsTotal bool `json:"is_total,omitempty"`
}

// UsageSummary — полная сводка использования vs лимитов для текущего юзера.
type UsageSummary struct {
	PlanID        string    `json:"plan_id"`
	Prompts       QuotaInfo `json:"prompts"`
	Collections   QuotaInfo `json:"collections"`
	AIRequests    QuotaInfo `json:"ai_requests"`
	Teams         QuotaInfo `json:"teams"`
	ShareLinks    QuotaInfo `json:"share_links"`
	ExtUsesToday  QuotaInfo `json:"ext_uses_today"`
	MCPUsesToday  QuotaInfo `json:"mcp_uses_today"`
}

// Feature types для daily_feature_usage.
const (
	FeatureAI        = "ai"
	FeatureExtension = "extension"
	FeatureMCP       = "mcp"
)
