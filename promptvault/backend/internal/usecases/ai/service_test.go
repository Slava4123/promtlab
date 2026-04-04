package ai

import (
	"context"
	"fmt"
	"testing"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/infrastructure/openrouter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func testConfig() *config.AIConfig {
	return &config.AIConfig{
		OpenRouterAPIKey: "test-key",
		RateLimitRPM:     100,
		Models: []config.ModelConfig{
			{ID: "test/model", Name: "Test", MaxTokens: 4096},
		},
	}
}

func noopCallback(_ string) error { return nil }

func TestEnhance_Success(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	client.On("Stream",
		mock.Anything,
		mock.MatchedBy(func(req openrouter.ChatRequest) bool {
			return req.Model == "test/model"
		}),
		mock.Anything,
	).Return(nil, nil)

	err := svc.Enhance(context.Background(), EnhanceInput{
		UserID:  1,
		Content: "test prompt",
		Model:   "test/model",
	}, noopCallback)

	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestEnhance_EmptyContent(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	err := svc.Enhance(context.Background(), EnhanceInput{
		UserID:  1,
		Content: "",
		Model:   "test/model",
	}, noopCallback)

	assert.ErrorIs(t, err, ErrEmptyContent)
	client.AssertNotCalled(t, "Stream", mock.Anything, mock.Anything, mock.Anything)
}

func TestEnhance_WhitespaceOnly(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	err := svc.Enhance(context.Background(), EnhanceInput{
		UserID:  1,
		Content: "   ",
		Model:   "test/model",
	}, noopCallback)

	assert.ErrorIs(t, err, ErrEmptyContent)
	client.AssertNotCalled(t, "Stream", mock.Anything, mock.Anything, mock.Anything)
}

func TestEnhance_DisallowedModel(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	err := svc.Enhance(context.Background(), EnhanceInput{
		UserID:  1,
		Content: "test prompt",
		Model:   "unknown/model",
	}, noopCallback)

	assert.ErrorIs(t, err, ErrModelNotFound)
	client.AssertNotCalled(t, "Stream", mock.Anything, mock.Anything, mock.Anything)
}

func TestEnhance_MissingAPIKey(t *testing.T) {
	client := new(mockAIClient)
	cfg := testConfig()
	cfg.OpenRouterAPIKey = ""
	svc := NewService(client, cfg)

	err := svc.Enhance(context.Background(), EnhanceInput{
		UserID:  1,
		Content: "test prompt",
		Model:   "test/model",
	}, noopCallback)

	assert.ErrorIs(t, err, ErrAPIKeyMissing)
	client.AssertNotCalled(t, "Stream", mock.Anything, mock.Anything, mock.Anything)
}

func TestEnhance_ClientError(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	networkErr := fmt.Errorf("network fail")
	client.On("Stream", mock.Anything, mock.Anything, mock.Anything).Return(nil, networkErr)

	err := svc.Enhance(context.Background(), EnhanceInput{
		UserID:  1,
		Content: "test prompt",
		Model:   "test/model",
	}, noopCallback)

	assert.ErrorIs(t, err, networkErr)
	client.AssertExpectations(t)
}

func TestRewrite_WithStyle(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	client.On("Stream",
		mock.Anything,
		mock.MatchedBy(func(req openrouter.ChatRequest) bool {
			return req.Model == "test/model"
		}),
		mock.Anything,
	).Return(nil, nil)

	err := svc.Rewrite(context.Background(), RewriteInput{
		UserID:  1,
		Content: "test prompt",
		Model:   "test/model",
		Style:   "formal",
	}, noopCallback)

	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestVariations_DefaultCount(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	client.On("Stream",
		mock.Anything,
		mock.MatchedBy(func(req openrouter.ChatRequest) bool {
			return req.Model == "test/model"
		}),
		mock.Anything,
	).Return(nil, nil)

	err := svc.Variations(context.Background(), VariationsInput{
		UserID:  1,
		Content: "test prompt",
		Model:   "test/model",
		Count:   0,
	}, noopCallback)

	assert.NoError(t, err)
	client.AssertExpectations(t)
}

// --- Analyze ---

func TestAnalyze_Success(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	client.On("Stream",
		mock.Anything,
		mock.MatchedBy(func(req openrouter.ChatRequest) bool {
			return req.Model == "test/model"
		}),
		mock.Anything,
	).Return(nil, nil)

	err := svc.Analyze(context.Background(), AnalyzeInput{
		UserID:  1,
		Content: "test prompt",
		Model:   "test/model",
	}, noopCallback)

	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestAnalyze_EmptyContent(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	err := svc.Analyze(context.Background(), AnalyzeInput{
		UserID:  1,
		Content: "",
		Model:   "test/model",
	}, noopCallback)

	assert.ErrorIs(t, err, ErrEmptyContent)
	client.AssertNotCalled(t, "Stream", mock.Anything, mock.Anything, mock.Anything)
}

func TestAnalyze_ClientError(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	networkErr := fmt.Errorf("network fail")
	client.On("Stream", mock.Anything, mock.Anything, mock.Anything).Return(nil, networkErr)

	err := svc.Analyze(context.Background(), AnalyzeInput{
		UserID:  1,
		Content: "test prompt",
		Model:   "test/model",
	}, noopCallback)

	assert.ErrorIs(t, err, networkErr)
	client.AssertExpectations(t)
}

// --- Rewrite (extended) ---

func TestRewrite_EmptyContent(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	err := svc.Rewrite(context.Background(), RewriteInput{
		UserID:  1,
		Content: "",
		Model:   "test/model",
		Style:   "formal",
	}, noopCallback)

	assert.ErrorIs(t, err, ErrEmptyContent)
	client.AssertNotCalled(t, "Stream", mock.Anything, mock.Anything, mock.Anything)
}

func TestRewrite_ClientError(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	networkErr := fmt.Errorf("network fail")
	client.On("Stream", mock.Anything, mock.Anything, mock.Anything).Return(nil, networkErr)

	err := svc.Rewrite(context.Background(), RewriteInput{
		UserID:  1,
		Content: "test prompt",
		Model:   "test/model",
		Style:   "formal",
	}, noopCallback)

	assert.ErrorIs(t, err, networkErr)
	client.AssertExpectations(t)
}

// --- Variations (extended) ---

func TestVariations_CustomCount(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	client.On("Stream",
		mock.Anything,
		mock.MatchedBy(func(req openrouter.ChatRequest) bool {
			return req.Model == "test/model"
		}),
		mock.Anything,
	).Return(nil, nil)

	err := svc.Variations(context.Background(), VariationsInput{
		UserID:  1,
		Content: "test prompt",
		Model:   "test/model",
		Count:   5,
	}, noopCallback)

	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestVariations_EmptyContent(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	err := svc.Variations(context.Background(), VariationsInput{
		UserID:  1,
		Content: "",
		Model:   "test/model",
		Count:   3,
	}, noopCallback)

	assert.ErrorIs(t, err, ErrEmptyContent)
	client.AssertNotCalled(t, "Stream", mock.Anything, mock.Anything, mock.Anything)
}

// --- CheckRateLimit ---

func TestCheckRateLimit_Allowed(t *testing.T) {
	cfg := testConfig()
	cfg.RateLimitRPM = 10
	svc := NewService(new(mockAIClient), cfg)

	err := svc.CheckRateLimit(1)
	assert.NoError(t, err)
}

func TestCheckRateLimit_Blocked(t *testing.T) {
	cfg := testConfig()
	cfg.RateLimitRPM = 1
	svc := NewService(new(mockAIClient), cfg)

	err := svc.CheckRateLimit(1)
	assert.NoError(t, err)

	err = svc.CheckRateLimit(1)
	assert.ErrorIs(t, err, ErrRateLimited)
}

// --- Models ---

func TestModels_ReturnsCopy(t *testing.T) {
	client := new(mockAIClient)
	svc := NewService(client, testConfig())

	models := svc.Models()
	assert.Len(t, models, 1)

	// Mutate the returned slice.
	models[0].ID = "mutated/model"

	// Original must be unchanged.
	original := svc.Models()
	assert.Equal(t, "test/model", original[0].ID)
}
