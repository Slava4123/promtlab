package config

// AnalyticsConfig — настройки Phase 14 аналитики.
type AnalyticsConfig struct {
	// ExperimentalInsights включает расчёт 4 неготовых Smart Insight типов
	// (most_edited, possible_duplicates, orphan_tags, empty_collections).
	// Default false — Phase 14 релизится с 3 рабочими типами
	// (unused, trending, declining). Доделка — follow-up тикет M8.
	ExperimentalInsights bool `koanf:"experimental_insights"`
}
