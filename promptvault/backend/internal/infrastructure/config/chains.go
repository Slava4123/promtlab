package config

// ChainsConfig — Phase 16. Feature flag для Prompt Chains.
// Default false: фича не показывается в UI и не регистрирует API/MCP routes.
// Установить CHAINS_ENABLED=true в .env для включения после QA.
type ChainsConfig struct {
	Enabled bool `koanf:"enabled"`
}
