package ai

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/infrastructure/openrouter"
)

// AIClient abstracts the OpenRouter client for testability.
type AIClient interface {
	Stream(ctx context.Context, req openrouter.ChatRequest, cb openrouter.StreamCallback) (*openrouter.Usage, error)
}

// Service handles AI prompt operations.
type Service struct {
	client  AIClient
	models  []config.ModelConfig
	limiter *userLimiter
	apiKey  string
}

// NewService creates a new AI service.
func NewService(client AIClient, cfg *config.AIConfig) *Service {
	svc := &Service{
		client:  client,
		models:  cfg.GetModels(),
		limiter: newUserLimiter(cfg.RateLimitRPM),
		apiKey:  cfg.OpenRouterAPIKey,
	}

	if cfg.OpenRouterAPIKey == "" {
		slog.Warn("AI_OPENROUTER_API_KEY не задан — AI-функции недоступны")
	}

	return svc
}

// Models returns a copy of the available models list.
func (s *Service) Models() []config.ModelConfig {
	out := make([]config.ModelConfig, len(s.models))
	copy(out, s.models)
	return out
}

// CheckRateLimit checks if the user can make a request.
func (s *Service) CheckRateLimit(userID uint) error {
	if !s.limiter.Allow(userID) {
		return ErrRateLimited
	}
	return nil
}

// Enhance improves the quality of a prompt.
func (s *Service) Enhance(ctx context.Context, in EnhanceInput, cb openrouter.StreamCallback) error {
	if err := s.validate(in.Content, in.Model); err != nil {
		return err
	}

	start := time.Now()
	usage, err := s.client.Stream(ctx, openrouter.ChatRequest{
		Model:        in.Model,
		Messages:     s.buildMessages(systemPromptEnhance, in.Content),
		MaxTokens:    s.maxTokensForModel(in.Model),
		Temperature:  s.temperatureForModel(in.Model),
		Reasoning:    s.reasoningForModel(in.Model),
		CacheControl: s.cacheControlForModel(in.Model),
	}, cb)
	s.logUsage("enhance", in.Model, time.Since(start), usage, err)
	return err
}

// Rewrite rewrites a prompt in the specified style.
func (s *Service) Rewrite(ctx context.Context, in RewriteInput, cb openrouter.StreamCallback) error {
	if err := s.validate(in.Content, in.Model); err != nil {
		return err
	}

	start := time.Now()
	usage, err := s.client.Stream(ctx, openrouter.ChatRequest{
		Model:        in.Model,
		Messages:     s.buildMessages(buildRewritePrompt(in.Style), in.Content),
		MaxTokens:    s.maxTokensForModel(in.Model),
		Temperature:  s.temperatureForModel(in.Model),
		Reasoning:    s.reasoningForModel(in.Model),
		CacheControl: s.cacheControlForModel(in.Model),
	}, cb)
	s.logUsage("rewrite", in.Model, time.Since(start), usage, err)
	return err
}

// Analyze analyzes the quality of a prompt.
func (s *Service) Analyze(ctx context.Context, in AnalyzeInput, cb openrouter.StreamCallback) error {
	if err := s.validate(in.Content, in.Model); err != nil {
		return err
	}

	start := time.Now()
	usage, err := s.client.Stream(ctx, openrouter.ChatRequest{
		Model:        in.Model,
		Messages:     s.buildMessages(systemPromptAnalyze, in.Content),
		MaxTokens:    s.maxTokensForModel(in.Model),
		Temperature:  s.temperatureForModel(in.Model),
		Reasoning:    s.reasoningForModel(in.Model),
		CacheControl: s.cacheControlForModel(in.Model),
	}, cb)
	s.logUsage("analyze", in.Model, time.Since(start), usage, err)
	return err
}

// Variations generates multiple variations of a prompt.
func (s *Service) Variations(ctx context.Context, in VariationsInput, cb openrouter.StreamCallback) error {
	if err := s.validate(in.Content, in.Model); err != nil {
		return err
	}

	count := in.Count
	if count <= 0 {
		count = 3
	}

	start := time.Now()
	usage, err := s.client.Stream(ctx, openrouter.ChatRequest{
		Model:        in.Model,
		Messages:     s.buildMessages(buildVariationsPrompt(count), in.Content),
		MaxTokens:    s.maxTokensForModel(in.Model),
		Temperature:  s.temperatureForModel(in.Model),
		Reasoning:    s.reasoningForModel(in.Model),
		CacheControl: s.cacheControlForModel(in.Model),
	}, cb)
	s.logUsage("variations", in.Model, time.Since(start), usage, err)
	return err
}

func (s *Service) logUsage(operation, model string, duration time.Duration, usage *openrouter.Usage, err error) {
	if err != nil {
		return
	}
	if usage == nil {
		slog.Info("ai request completed",
			"operation", operation,
			"model", model,
			"duration_ms", duration.Milliseconds(),
		)
		return
	}
	slog.Info("ai request completed",
		"operation", operation,
		"model", model,
		"duration_ms", duration.Milliseconds(),
		"prompt_tokens", usage.PromptTokens,
		"completion_tokens", usage.CompletionTokens,
		"total_tokens", usage.TotalTokens,
		"cost_usd", usage.Cost,
		"cached_tokens", usage.CachedTokens,
	)
}

func (s *Service) validate(content, model string) error {
	if s.apiKey == "" {
		return ErrAPIKeyMissing
	}
	if strings.TrimSpace(content) == "" {
		return ErrEmptyContent
	}
	if !s.isModelAllowed(model) {
		return ErrModelNotFound
	}
	return nil
}

// findModel returns the model config for the given ID, or nil if not found.
func (s *Service) findModel(id string) *config.ModelConfig {
	for i := range s.models {
		if s.models[i].ID == id {
			return &s.models[i]
		}
	}
	return nil
}

func (s *Service) isModelAllowed(modelID string) bool {
	return s.findModel(modelID) != nil
}

func (s *Service) maxTokensForModel(modelID string) int {
	if m := s.findModel(modelID); m != nil {
		return m.MaxTokens
	}
	return 4096
}

func (s *Service) temperatureForModel(modelID string) *float64 {
	if m := s.findModel(modelID); m != nil {
		return m.Temperature
	}
	return nil
}

func (s *Service) reasoningForModel(modelID string) *openrouter.Reasoning {
	if m := s.findModel(modelID); m != nil && m.ReasoningEffort != "" {
		return &openrouter.Reasoning{Effort: m.ReasoningEffort}
	}
	return nil
}

// buildMessages creates the message list.
func (s *Service) buildMessages(systemPrompt, userContent string) []openrouter.Message {
	return []openrouter.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userContent},
	}
}

// cacheControlForModel returns cache_control for Anthropic models (top-level automatic caching).
func (s *Service) cacheControlForModel(modelID string) *openrouter.CacheControl {
	if strings.HasPrefix(modelID, "anthropic/") {
		return &openrouter.CacheControl{Type: "ephemeral"}
	}
	return nil
}
