package quota

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
}

// Feature types для daily_feature_usage.
const (
	FeatureExtension   = "extension"
	FeatureMCP         = "mcp"
	FeatureShareCreate = "share_create" // Phase 14: дневной счётчик созданных шар-ссылок (fixed window UTC).
)
