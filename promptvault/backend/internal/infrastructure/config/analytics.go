package config

// AnalyticsConfig — настройки Phase 14 аналитики.
type AnalyticsConfig struct {
	// ExperimentalInsights — kill-switch для 4 расширенных типов Smart Insights
	// (most_edited, possible_duplicates, orphan_tags, empty_collections).
	// Default true (Phase 15). Установить ANALYTICS_EXPERIMENTAL_INSIGHTS=false
	// в .env для экстренного отключения без деплоя.
	//
	// possible_duplicates дополнительно требует расширения pg_trgm — при его
	// отсутствии тип тихо пропускается (см. analytics.Service.trgmAvailable).
	ExperimentalInsights bool `koanf:"experimental_insights"`
}
