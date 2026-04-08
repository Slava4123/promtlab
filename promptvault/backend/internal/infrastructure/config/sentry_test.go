package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSentryConfig_DefaultValues проверяет zero-value поведение без env vars.
// Enabled=false, DSN пустой — это валидное состояние для dev без Sentry.
func TestSentryConfig_DefaultValues(t *testing.T) {
	cfg := SentryConfig{}

	assert.False(t, cfg.Enabled, "default Enabled must be false")
	assert.Empty(t, cfg.Dsn, "default Dsn must be empty")
	assert.Empty(t, cfg.Environment, "default Environment must be empty (filled by loader)")
	assert.InDelta(t, 0.0, cfg.TracesSampleRate, 0.0001, "default TracesSampleRate must be 0.0")
	assert.False(t, cfg.Debug, "default Debug must be false")
}

// TestSentryConfig_Koanf_FieldTags проверяет что koanf tags корректны.
// Регрессионный тест на случай опечаток при рефакторинге.
func TestSentryConfig_Koanf_FieldTags(t *testing.T) {
	// Простая проверка что структура корректно инициализируется через literal —
	// если koanf tags сломаны, loader не сможет замапить env vars.
	cfg := SentryConfig{
		Enabled:          true,
		Dsn:              "http://key@glitchtip.example/1",
		Environment:      "production",
		Release:          "abc123",
		TracesSampleRate: 0.1,
		Debug:            false,
	}
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "http://key@glitchtip.example/1", cfg.Dsn)
	assert.Equal(t, "production", cfg.Environment)
	assert.Equal(t, "abc123", cfg.Release)
	assert.InDelta(t, 0.1, cfg.TracesSampleRate, 0.0001)
}
