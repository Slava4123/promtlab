package config

// SentryConfig — параметры интеграции с Sentry-compatible backend (GlitchTip).
//
// Enabled — feature flag для gradual rollout. При false инициализация не
// происходит, код в zero-cost no-op режиме.
//
// Dsn — публичный Data Source Name из GlitchTip UI (Project Settings → Client Keys).
//
// Environment — "development" / "production" / "staging". Используется для
// фильтрации в UI.
//
// Release — версия приложения. Обычно GITHUB_SHA из CI/CD (см. .github/workflows/deploy.yml).
//
// TracesSampleRate — доля request'ов, для которых создаются performance
// transactions (0.0-1.0). 0.0 = performance monitoring выключен, 0.1 = 10%
// семплирование (рекомендуется для prod).
//
// Debug — включает stderr логи sentry-go SDK. Включать только для troubleshoot'а.
type SentryConfig struct {
	Enabled          bool    `koanf:"enabled"`
	Dsn              string  `koanf:"dsn"`
	Environment      string  `koanf:"environment"`
	Release          string  `koanf:"release"`
	TracesSampleRate float64 `koanf:"traces_sample_rate"`
	Debug            bool    `koanf:"debug"`
}
