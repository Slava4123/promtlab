package ai

import (
	"context"

	"promptvault/internal/infrastructure/openrouter"

	"github.com/stretchr/testify/mock"
)

type mockAIClient struct{ mock.Mock }

func (m *mockAIClient) Stream(ctx context.Context, req openrouter.ChatRequest, cb openrouter.StreamCallback) (*openrouter.Usage, error) {
	args := m.Called(ctx, req, cb)
	usage, _ := args.Get(0).(*openrouter.Usage)
	return usage, args.Error(1)
}
