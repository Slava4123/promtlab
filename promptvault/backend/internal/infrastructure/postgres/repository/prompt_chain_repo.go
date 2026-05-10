package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type chainRepo struct {
	db *gorm.DB
}

func NewChainRepository(db *gorm.DB) *chainRepo {
	return &chainRepo{db: db}
}

func (r *chainRepo) Create(ctx context.Context, c *models.PromptChain) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *chainRepo) GetByID(ctx context.Context, id uint) (*models.PromptChain, error) {
	var c models.PromptChain
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *chainRepo) GetByIDWithSteps(ctx context.Context, id uint) (*models.PromptChain, error) {
	var c models.PromptChain
	err := r.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		// Phase 16 v2: nested preload — подтягиваем Prompt каждого шага для отображения
		// title/content в Canvas-узлах (вместо «Промпт #ID»).
		Preload("Steps.Prompt").
		First(&c, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *chainRepo) ListByUser(ctx context.Context, userID uint, teamIDs []uint, limit, offset int) ([]models.PromptChain, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.PromptChain{})
	if len(teamIDs) > 0 {
		q = q.Where("team_id IN ?", teamIDs)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL", userID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var chains []models.PromptChain
	if err := q.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&chains).Error; err != nil {
		return nil, 0, err
	}
	return chains, total, nil
}

// ListByUserWithStats — расширенный list для UI с агрегатной статистикой
// (step_count, has_branching, saved_runs_count, steps_preview) в одном SELECT
// через LATERAL. Без N+1 — pattern из team_repo (MN-38).
//
// steps_preview ограничен models.ChainStepsPreviewLimit первых шагов
// (по position), для длинных цепочек UI рисует "+N more".
func (r *chainRepo) ListByUserWithStats(ctx context.Context, userID uint, teamIDs []uint, limit, offset int) ([]models.PromptChainListRow, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.PromptChain{})
	if len(teamIDs) > 0 {
		q = q.Where("team_id IN ?", teamIDs)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL", userID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type rawRow struct {
		models.PromptChain
		StepCount        int             `gorm:"column:step_count"`
		HasBranching     bool            `gorm:"column:has_branching"`
		SavedRunsCount   int             `gorm:"column:saved_runs_count"`
		StepsPreviewJSON json.RawMessage `gorm:"column:steps_preview_json"`
	}

	rawSQL := `
SELECT c.*,
       COALESCE(cs.step_count, 0)        AS step_count,
       COALESCE(cs.has_branching, false) AS has_branching,
       COALESCE(cr.runs_count, 0)        AS saved_runs_count,
       COALESCE(sp.steps_preview_json, '[]'::json) AS steps_preview_json
FROM prompt_chains c
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS step_count,
           BOOL_OR(step_type = 'fork') AS has_branching
    FROM prompt_chain_steps
    WHERE chain_id = c.id
) cs ON true
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS runs_count
    FROM prompt_chain_executions
    WHERE chain_id = c.id AND status = ?
) cr ON true
LEFT JOIN LATERAL (
    SELECT json_agg(json_build_object('position', position, 'step_type', step_type) ORDER BY position) AS steps_preview_json
    FROM (
        SELECT position, step_type
        FROM prompt_chain_steps
        WHERE chain_id = c.id
        ORDER BY position
        LIMIT ?
    ) s
) sp ON true
WHERE c.deleted_at IS NULL
`
	args := []any{string(models.ChainExecutionStatusCompleted), models.ChainStepsPreviewLimit}

	if len(teamIDs) > 0 {
		rawSQL += " AND c.team_id IN ?"
		args = append(args, teamIDs)
	} else {
		rawSQL += " AND c.user_id = ? AND c.team_id IS NULL"
		args = append(args, userID)
	}
	rawSQL += " ORDER BY c.updated_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	var raws []rawRow
	if err := r.db.WithContext(ctx).Raw(rawSQL, args...).Scan(&raws).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]models.PromptChainListRow, len(raws))
	for i, raw := range raws {
		var preview []models.PromptChainStepPreview
		if len(raw.StepsPreviewJSON) > 0 && string(raw.StepsPreviewJSON) != "null" {
			if err := json.Unmarshal(raw.StepsPreviewJSON, &preview); err != nil {
				// Defensive: при corrupt JSON отдаём пустой preview, но не валим весь list.
				preview = nil
			}
		}
		if preview == nil {
			preview = []models.PromptChainStepPreview{}
		}
		rows[i] = models.PromptChainListRow{
			PromptChain:    raw.PromptChain,
			StepCount:      raw.StepCount,
			HasBranching:   raw.HasBranching,
			SavedRunsCount: raw.SavedRunsCount,
			StepsPreview:   preview,
		}
	}
	return rows, total, nil
}

func (r *chainRepo) Update(ctx context.Context, c *models.PromptChain) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *chainRepo) SoftDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.PromptChain{}).Error
}

func (r *chainRepo) HasActiveExecutions(ctx context.Context, chainID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.PromptChainExecution{}).
		Where("chain_id = ? AND status = ?", chainID, models.ChainExecutionStatusInProgress).
		Count(&count).Error
	return count > 0, err
}

func (r *chainRepo) AddStep(ctx context.Context, s *models.PromptChainStep) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *chainRepo) GetStepByID(ctx context.Context, stepID uint) (*models.PromptChainStep, error) {
	var s models.PromptChainStep
	if err := r.db.WithContext(ctx).First(&s, stepID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *chainRepo) UpdateStep(ctx context.Context, s *models.PromptChainStep) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *chainRepo) RemoveStep(ctx context.Context, stepID uint) error {
	return r.db.WithContext(ctx).Where("id = ?", stepID).Delete(&models.PromptChainStep{}).Error
}

func (r *chainRepo) ListStepsByChain(ctx context.Context, chainID uint) ([]models.PromptChainStep, error) {
	var steps []models.PromptChainStep
	err := r.db.WithContext(ctx).
		Where("chain_id = ?", chainID).
		Order("position ASC").
		Find(&steps).Error
	return steps, err
}

// ReorderSteps использует уникальный constraint uq_prompt_chain_steps_position.
// MJ-40: миграция 000062 переключила constraint на INITIALLY IMMEDIATE
// (глобальная отложенность была overhead на 99% операций). Здесь явно
// ставим SET CONSTRAINTS ... DEFERRED для текущей transaction — несколько
// UPDATE position допустимы, проверка уникальности отложена до COMMIT.
func (r *chainRepo) ReorderSteps(ctx context.Context, chainID uint, stepIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET CONSTRAINTS uq_prompt_chain_steps_position DEFERRED").Error; err != nil {
			return err
		}
		for i, stepID := range stepIDs {
			res := tx.Model(&models.PromptChainStep{}).
				Where("id = ? AND chain_id = ?", stepID, chainID).
				Update("position", i+1)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return repo.ErrNotFound
			}
		}
		return nil
	})
}

// RelinkPromptPredecessors переключает все next_step_id, ссылавшиеся на fromID,
// на toID (или NULL). Один UPDATE — атомарно.
func (r *chainRepo) RelinkPromptPredecessors(ctx context.Context, chainID, fromID uint, toID *uint) error {
	return r.db.WithContext(ctx).Model(&models.PromptChainStep{}).
		Where("chain_id = ? AND next_step_id = ?", chainID, fromID).
		Update("next_step_id", toID).Error
}

// InTransaction оборачивает fn в gorm-транзакцию. Внутри fn получает
// tx-bound chainRepo с тем же интерфейсом — все его методы работают на tx.
func (r *chainRepo) InTransaction(ctx context.Context, fn func(repo.ChainRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&chainRepo{db: tx})
	})
}

func (r *chainRepo) CountStepsByChain(ctx context.Context, chainID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.PromptChainStep{}).
		Where("chain_id = ?", chainID).
		Count(&count).Error
	return count, err
}

// CountChainsUsingPrompt — нужен для блокировки удаления промпта, используемого
// в активных (не soft-deleted) цепочках. Возвращает 409 Conflict из usecases/prompt.
func (r *chainRepo) CountChainsUsingPrompt(ctx context.Context, promptID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("prompt_chain_steps").
		Joins("JOIN prompt_chains ON prompt_chains.id = prompt_chain_steps.chain_id AND prompt_chains.deleted_at IS NULL").
		Where("prompt_chain_steps.prompt_id = ?", promptID).
		Distinct("prompt_chains.id").
		Count(&count).Error
	return count, err
}

func (r *chainRepo) CreateExecution(ctx context.Context, e *models.PromptChainExecution) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *chainRepo) GetExecutionByID(ctx context.Context, execID uint) (*models.PromptChainExecution, error) {
	var e models.PromptChainExecution
	if err := r.db.WithContext(ctx).First(&e, execID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

// UpdateExecution использует optimistic lock через updated_at: UPDATE проходит,
// только если updated_at в БД совпадает с e.UpdatedAt из памяти. Защита от race
// в AdvanceStep (двойной клик «Далее» / параллельные MCP-клиенты).
// При несовпадении возвращается repo.ErrConflict.
func (r *chainRepo) UpdateExecution(ctx context.Context, e *models.PromptChainExecution) error {
	expectedAt := e.UpdatedAt
	newAt := time.Now()
	res := r.db.WithContext(ctx).Model(&models.PromptChainExecution{}).
		Where("id = ? AND updated_at = ?", e.ID, expectedAt).
		Updates(map[string]any{
			"current_step":   e.CurrentStep,
			"variables":      e.Variables,
			"step_outputs":   e.StepOutputs,
			"chain_snapshot": e.ChainSnapshot,
			"status":         e.Status,
			"completed_at":   e.CompletedAt,
			"updated_at":     newAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		// Либо строка не найдена, либо updated_at сместился — для caller'а это
		// одно: нужно перечитать execution и повторить операцию.
		return repo.ErrConflict
	}
	e.UpdatedAt = newAt
	return nil
}

func (r *chainRepo) ListExecutionsByChain(ctx context.Context, chainID uint, limit int) ([]models.PromptChainExecution, error) {
	var executions []models.PromptChainExecution
	err := r.db.WithContext(ctx).
		Where("chain_id = ?", chainID).
		Order("started_at DESC").
		Limit(limit).
		Find(&executions).Error
	return executions, err
}
