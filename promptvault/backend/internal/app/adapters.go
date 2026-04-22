package app

import (
	"context"

	apikeyuc "promptvault/internal/usecases/apikey"
	colluc "promptvault/internal/usecases/collection"
	promptuc "promptvault/internal/usecases/prompt"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// Здесь собраны узкие адаптеры между usecases-Service и middleware/mcpserver
// интерфейсами. Вынесены из app.go, чтобы главный DI-файл содержал только
// сборку графа. Поведение адаптеров намеренно тривиальное: отбрасывают
// newly_unlocked_badges или проксируют в ValidateKey.

// apiKeyValidatorAdapter — authmw.APIKeyValidator для *apikeyuc.Service.
type apiKeyValidatorAdapter struct {
	svc *apikeyuc.Service
}

func (a *apiKeyValidatorAdapter) ValidateKey(ctx context.Context, rawKey string) (userID uint, keyID uint, err error) {
	result, err := a.svc.ValidateKey(ctx, rawKey)
	if err != nil {
		return 0, 0, err
	}
	return result.UserID, result.KeyID, nil
}

// mcpPromptAdapter — mcpserver.PromptService поверх *promptuc.Service.
// Скрывает newly_unlocked_badges — MCP-клиенты не показывают toast-UI.
type mcpPromptAdapter struct {
	*promptuc.Service
}

func (a *mcpPromptAdapter) Create(ctx context.Context, in promptuc.CreateInput) (*models.Prompt, error) {
	p, _, err := a.Service.Create(ctx, in)
	return p, err
}

func (a *mcpPromptAdapter) Update(ctx context.Context, id, userID uint, in promptuc.UpdateInput) (*models.Prompt, error) {
	p, _, err := a.Service.Update(ctx, id, userID, in)
	return p, err
}

func (a *mcpPromptAdapter) RevertToVersion(ctx context.Context, promptID, userID, versionID uint) (*models.Prompt, error) {
	p, _, err := a.Service.RevertToVersion(ctx, promptID, userID, versionID)
	return p, err
}

func (a *mcpPromptAdapter) IncrementUsage(ctx context.Context, id, userID uint) error {
	_, err := a.Service.IncrementUsage(ctx, id, userID)
	return err
}

// mcpCollectionAdapter — симметричный адаптер для *colluc.Service.
type mcpCollectionAdapter struct {
	*colluc.Service
}

func (a *mcpCollectionAdapter) Create(ctx context.Context, userID uint, name, description, color, icon string, teamID *uint) (*models.Collection, error) {
	c, _, err := a.Service.Create(ctx, userID, name, description, color, icon, teamID)
	return c, err
}

// adminHealthAdapter — узкий HealthCounter для adminhttp.Handler.
type adminHealthAdapter struct {
	repo repo.AdminRepository
}

func (a *adminHealthAdapter) CountUsers(ctx context.Context) (total, admins, active, frozen int64, err error) {
	return a.repo.CountUsers(ctx)
}
