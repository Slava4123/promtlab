package ai

import "promptvault/internal/infrastructure/config"

type ModelResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
	MaxTokens   int    `json:"max_tokens"`
}

func toModelResponses(models []config.ModelConfig) []ModelResponse {
	out := make([]ModelResponse, len(models))
	for i, m := range models {
		out[i] = ModelResponse{
			ID:          m.ID,
			Name:        m.Name,
			Provider:    m.Provider,
			Description: m.Description,
			MaxTokens:   m.MaxTokens,
		}
	}
	return out
}
