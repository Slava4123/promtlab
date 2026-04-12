package mcpserver

import (
	"context"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	promptuc "promptvault/internal/usecases/prompt"
	searchuc "promptvault/internal/usecases/search"
)

type PromptService interface {
	Create(ctx context.Context, in promptuc.CreateInput) (*models.Prompt, error)
	GetByID(ctx context.Context, id, userID uint) (*models.Prompt, error)
	List(ctx context.Context, filter repo.PromptListFilter) ([]models.Prompt, int64, error)
	Update(ctx context.Context, id, userID uint, in promptuc.UpdateInput) (*models.Prompt, error)
	Delete(ctx context.Context, id, userID uint) error
	ListVersions(ctx context.Context, promptID, userID uint, page, pageSize int) ([]models.PromptVersion, int64, error)
}

type CollectionService interface {
	List(ctx context.Context, userID uint, teamIDs []uint) ([]models.CollectionWithCount, error)
	Create(ctx context.Context, userID uint, name, description, color, icon string, teamID *uint) (*models.Collection, error)
	Delete(ctx context.Context, id, userID uint) error
}

type TagService interface {
	List(ctx context.Context, userID uint, teamID *uint) ([]models.Tag, error)
	Create(ctx context.Context, name, color string, userID uint, teamID *uint) (*models.Tag, error)
}

type SearchService interface {
	Search(ctx context.Context, userID uint, teamID *uint, query string) (*searchuc.SearchOutput, error)
}
