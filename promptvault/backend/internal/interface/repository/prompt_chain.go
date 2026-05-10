package repository

import (
	"context"

	"promptvault/internal/models"
)

// ChainRepository — персистентный слой для цепочек (Phase 16). Интерфейс,
// реализация в infrastructure/postgres/repository. Все методы принимают ctx первым
// аргументом и не утечкают gorm-типы.
type ChainRepository interface {
	// Цепочка целиком
	Create(ctx context.Context, c *models.PromptChain) error
	GetByID(ctx context.Context, id uint) (*models.PromptChain, error)
	GetByIDWithSteps(ctx context.Context, id uint) (*models.PromptChain, error)
	ListByUser(ctx context.Context, userID uint, teamIDs []uint, limit, offset int) ([]models.PromptChain, int64, error)
	// ListByUserWithStats — расширенный list для UI: каждая цепочка идёт со
	// step_count, has_branching, saved_runs_count и steps_preview (первые
	// models.ChainStepsPreviewLimit шагов). Один SELECT через LATERAL,
	// pattern из team_repo (MN-38) — без N+1.
	ListByUserWithStats(ctx context.Context, userID uint, teamIDs []uint, limit, offset int) ([]models.PromptChainListRow, int64, error)
	Update(ctx context.Context, c *models.PromptChain) error
	SoftDelete(ctx context.Context, id uint) error
	HasActiveExecutions(ctx context.Context, chainID uint) (bool, error)

	// Шаги
	AddStep(ctx context.Context, s *models.PromptChainStep) error
	GetStepByID(ctx context.Context, stepID uint) (*models.PromptChainStep, error)
	UpdateStep(ctx context.Context, s *models.PromptChainStep) error
	RemoveStep(ctx context.Context, stepID uint) error
	ListStepsByChain(ctx context.Context, chainID uint) ([]models.PromptChainStep, error)
	// ReorderSteps выполняется в одной транзакции с DEFERRED uniqueness check.
	// stepIDs передаются в желаемом порядке; репо проставляет position=1..N.
	ReorderSteps(ctx context.Context, chainID uint, stepIDs []uint) error
	CountStepsByChain(ctx context.Context, chainID uint) (int64, error)
	CountChainsUsingPrompt(ctx context.Context, promptID uint) (int64, error)
	// RelinkPromptPredecessors — UPDATE prompt_chain_steps SET next_step_id=$toID
	// WHERE chain_id=$chainID AND next_step_id=$fromID. Используется в Service.RemoveStep
	// чтобы «зашить» граф после удаления промежуточного шага.
	RelinkPromptPredecessors(ctx context.Context, chainID, fromID uint, toID *uint) error

	// InTransaction — атомарное выполнение нескольких mutate-операций.
	// Внутри fn получает tx-bound repo (та же ChainRepository), любая ошибка
	// откатывает транзакцию. Используется в Service.AddStep чтобы insert-шага +
	// update-anchor + cycle-check не оставляли граф в полу-битом состоянии.
	InTransaction(ctx context.Context, fn func(ChainRepository) error) error

	// Запуски
	CreateExecution(ctx context.Context, e *models.PromptChainExecution) error
	GetExecutionByID(ctx context.Context, execID uint) (*models.PromptChainExecution, error)
	UpdateExecution(ctx context.Context, e *models.PromptChainExecution) error
	ListExecutionsByChain(ctx context.Context, chainID uint, limit int) ([]models.PromptChainExecution, error)
}
