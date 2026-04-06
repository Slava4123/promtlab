package config

type AIConfig struct {
	OpenRouterAPIKey    string        `koanf:"openrouter_api_key"`
	OpenRouterBaseURL   string        `koanf:"openrouter_base_url"`
	OpenRouterTimeoutSec int          `koanf:"openrouter_timeout_seconds"`
	RateLimitRPM        int           `koanf:"rate_limit_rpm"`
	Models              []ModelConfig `koanf:"models"`
}

type ModelConfig struct {
	ID              string   `koanf:"id"               json:"id"`
	Name            string   `koanf:"name"             json:"name"`
	Provider        string   `koanf:"provider"         json:"provider"`
	Description     string   `koanf:"description"      json:"description"`
	MaxTokens       int      `koanf:"max_tokens"       json:"max_tokens"`
	Temperature     *float64 `koanf:"temperature"      json:"temperature,omitempty"`
	ReasoningEffort string   `koanf:"reasoning_effort" json:"reasoning_effort,omitempty"`
}

// DefaultModels returns the curated list of models for prompt enhancement.
func DefaultModels() []ModelConfig {
	temp04 := 0.4

	return []ModelConfig{
		{
			ID:          "anthropic/claude-sonnet-4",
			Name:        "Claude Sonnet 4",
			Provider:    "anthropic",
			Description: "Лучшая модель для работы с текстом",
			MaxTokens:   8192,
			Temperature: &temp04,
		},
	}
}

// GetModels returns configured models or defaults.
func (c *AIConfig) GetModels() []ModelConfig {
	if len(c.Models) > 0 {
		return c.Models
	}
	return DefaultModels()
}
