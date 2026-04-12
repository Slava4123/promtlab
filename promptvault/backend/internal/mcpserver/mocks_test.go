package mcpserver

import (
	"context"

	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	promptuc "promptvault/internal/usecases/prompt"
	searchuc "promptvault/internal/usecases/search"
)

// --- context helper ---

func ctxWithUser(userID uint) context.Context {
	return context.WithValue(context.Background(), authmw.UserIDKey, userID)
}

// --- mock services ---

type mockPromptSvc struct{ mock.Mock }

func (m *mockPromptSvc) Create(ctx context.Context, in promptuc.CreateInput) (*models.Prompt, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func (m *mockPromptSvc) GetByID(ctx context.Context, id, userID uint) (*models.Prompt, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func (m *mockPromptSvc) List(ctx context.Context, filter repo.PromptListFilter) ([]models.Prompt, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]models.Prompt), args.Get(1).(int64), args.Error(2)
}

func (m *mockPromptSvc) Update(ctx context.Context, id, userID uint, in promptuc.UpdateInput) (*models.Prompt, error) {
	args := m.Called(ctx, id, userID, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func (m *mockPromptSvc) Delete(ctx context.Context, id, userID uint) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *mockPromptSvc) ListVersions(ctx context.Context, promptID, userID uint, page, pageSize int) ([]models.PromptVersion, int64, error) {
	args := m.Called(ctx, promptID, userID, page, pageSize)
	return args.Get(0).([]models.PromptVersion), args.Get(1).(int64), args.Error(2)
}

type mockCollectionSvc struct{ mock.Mock }

func (m *mockCollectionSvc) List(ctx context.Context, userID uint, teamIDs []uint) ([]models.CollectionWithCount, error) {
	args := m.Called(ctx, userID, teamIDs)
	return args.Get(0).([]models.CollectionWithCount), args.Error(1)
}

func (m *mockCollectionSvc) Create(ctx context.Context, userID uint, name, description, color, icon string, teamID *uint) (*models.Collection, error) {
	args := m.Called(ctx, userID, name, description, color, icon, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Collection), args.Error(1)
}

func (m *mockCollectionSvc) Delete(ctx context.Context, id, userID uint) error {
	return m.Called(ctx, id, userID).Error(0)
}

type mockTagSvc struct{ mock.Mock }

func (m *mockTagSvc) List(ctx context.Context, userID uint, teamID *uint) ([]models.Tag, error) {
	args := m.Called(ctx, userID, teamID)
	return args.Get(0).([]models.Tag), args.Error(1)
}

func (m *mockTagSvc) Create(ctx context.Context, name, color string, userID uint, teamID *uint) (*models.Tag, error) {
	args := m.Called(ctx, name, color, userID, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tag), args.Error(1)
}

type mockSearchSvc struct{ mock.Mock }

func (m *mockSearchSvc) Search(ctx context.Context, userID uint, teamID *uint, query string) (*searchuc.SearchOutput, error) {
	args := m.Called(ctx, userID, teamID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*searchuc.SearchOutput), args.Error(1)
}
