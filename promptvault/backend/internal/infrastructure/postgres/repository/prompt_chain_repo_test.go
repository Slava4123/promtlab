package repository

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// setupChainTest поднимает testcontainer + добавляет DEFERRABLE constraint на
// (chain_id, position), которого нет в AutoMigrate (только в SQL-миграции 000053).
// Без этого тесты ReorderSteps не воспроизведут реальное поведение прод-схемы.
func setupChainTest(t *testing.T) (repo.ChainRepository, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.Exec(`
		ALTER TABLE prompt_chain_steps
		ADD CONSTRAINT uq_prompt_chain_steps_position
		UNIQUE (chain_id, position) DEFERRABLE INITIALLY DEFERRED
	`).Error)
	return NewChainRepository(db), db
}

func createTestChain(t *testing.T, db *gorm.DB, userID uint, name string) *models.PromptChain {
	t.Helper()
	c := &models.PromptChain{UserID: userID, Name: name}
	require.NoError(t, db.Create(c).Error)
	return c
}

func createTestStep(t *testing.T, db *gorm.DB, chainID, promptID uint, position int) *models.PromptChainStep {
	t.Helper()
	s := &models.PromptChainStep{
		ChainID:         chainID,
		PromptID:        &promptID,
		Position:        position,
		VariableMapping: json.RawMessage(`{}`),
	}
	require.NoError(t, db.Create(s).Error)
	return s
}

func TestChainRepo_Create_Get(t *testing.T) {
	r, db := setupChainTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "chain-create@test.local")

	c := &models.PromptChain{UserID: u.ID, Name: "PRD Generator"}
	require.NoError(t, r.Create(ctx, c))
	assert.NotZero(t, c.ID)

	got, err := r.GetByID(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, "PRD Generator", got.Name)
	assert.Equal(t, u.ID, got.UserID)
}

func TestChainRepo_GetByID_NotFound(t *testing.T) {
	r, _ := setupChainTest(t)
	_, err := r.GetByID(context.Background(), 99999)
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func TestChainRepo_GetByIDWithSteps_OrdersByPosition(t *testing.T) {
	r, db := setupChainTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "chain-with-steps@test.local")
	p := createTestPrompt(t, db, u.ID, nil, "Step prompt", "content", 0)
	c := createTestChain(t, db, u.ID, "Chain")
	createTestStep(t, db, c.ID, p.ID, 3)
	createTestStep(t, db, c.ID, p.ID, 1)
	createTestStep(t, db, c.ID, p.ID, 2)

	got, err := r.GetByIDWithSteps(ctx, c.ID)
	require.NoError(t, err)
	require.Len(t, got.Steps, 3)
	assert.Equal(t, 1, got.Steps[0].Position)
	assert.Equal(t, 2, got.Steps[1].Position)
	assert.Equal(t, 3, got.Steps[2].Position)
}

func TestChainRepo_ListByUser_PaginatesAndCounts(t *testing.T) {
	r, db := setupChainTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "chain-list@test.local")
	for range 5 {
		createTestChain(t, db, u.ID, "Chain")
	}

	chains, total, err := r.ListByUser(ctx, u.ID, nil, 2, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, chains, 2)

	chains, total, err = r.ListByUser(ctx, u.ID, nil, 2, 4)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, chains, 1)
}

// ReorderSteps критический сценарий — при простой UPDATE position
// без DEFERRABLE constraint реверс положений вызывает temporary
// duplicate (e.g. при swap [1,2] → [2,1] в середине транзакции).
// Тест верифицирует, что DEFERRED constraint позволяет корректный COMMIT.
func TestChainRepo_ReorderSteps_HandlesDuplicateMidTransaction(t *testing.T) {
	r, db := setupChainTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "chain-reorder@test.local")
	p := createTestPrompt(t, db, u.ID, nil, "Step", "content", 0)
	c := createTestChain(t, db, u.ID, "Chain")
	s1 := createTestStep(t, db, c.ID, p.ID, 1)
	s2 := createTestStep(t, db, c.ID, p.ID, 2)
	s3 := createTestStep(t, db, c.ID, p.ID, 3)

	require.NoError(t, r.ReorderSteps(ctx, c.ID, []uint{s3.ID, s1.ID, s2.ID}))

	steps, err := r.ListStepsByChain(ctx, c.ID)
	require.NoError(t, err)
	require.Len(t, steps, 3)
	assert.Equal(t, s3.ID, steps[0].ID)
	assert.Equal(t, s1.ID, steps[1].ID)
	assert.Equal(t, s2.ID, steps[2].ID)
}

func TestChainRepo_SoftDelete_CascadesViaActiveScope(t *testing.T) {
	r, db := setupChainTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "chain-softdelete@test.local")
	c := createTestChain(t, db, u.ID, "Chain")

	require.NoError(t, r.SoftDelete(ctx, c.ID))

	_, err := r.GetByID(ctx, c.ID)
	assert.ErrorIs(t, err, repo.ErrNotFound, "GORM scope скрывает soft-deleted")
}

func TestChainRepo_HasActiveExecutions(t *testing.T) {
	r, db := setupChainTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "chain-active-exec@test.local")
	c := createTestChain(t, db, u.ID, "Chain")

	has, err := r.HasActiveExecutions(ctx, c.ID)
	require.NoError(t, err)
	assert.False(t, has)

	exec := &models.PromptChainExecution{
		ChainID:       c.ID,
		UserID:        u.ID,
		ChainSnapshot: json.RawMessage(`{}`),
		Status:        models.ChainExecutionStatusInProgress,
	}
	require.NoError(t, db.Create(exec).Error)

	has, err = r.HasActiveExecutions(ctx, c.ID)
	require.NoError(t, err)
	assert.True(t, has)

	exec.Status = models.ChainExecutionStatusCompleted
	require.NoError(t, db.Save(exec).Error)

	has, err = r.HasActiveExecutions(ctx, c.ID)
	require.NoError(t, err)
	assert.False(t, has)
}

// Phase 16 v2 (Tree-canvas): проверка fork-шага с label-based branches.
func TestChainRepo_ForkStep_RoundTrip(t *testing.T) {
	r, db := setupChainTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "chain-fork@test.local")
	p := createTestPrompt(t, db, u.ID, nil, "Step", "content", 0)
	c := createTestChain(t, db, u.ID, "Chain")

	// Создаём prompt-шаг (как target для branch).
	target := createTestStep(t, db, c.ID, p.ID, 1)

	// Fork step с branches: одна ведёт на target, другая — конец цепочки.
	conditionsJSON := json.RawMessage(`{
        "branches": [
            {"label": "Если критический баг", "next_step_id": ` + uintToString(target.ID) + `},
            {"label": "Если всё OK", "next_step_id": null}
        ]
    }`)

	forkStep := &models.PromptChainStep{
		ChainID:         c.ID,
		Position:        2,
		StepType:        models.StepTypeFork,
		Conditions:      conditionsJSON,
		VariableMapping: json.RawMessage(`{}`),
	}
	require.NoError(t, r.AddStep(ctx, forkStep))

	// Round-trip через GetByIDWithSteps.
	got, err := r.GetByIDWithSteps(ctx, c.ID)
	require.NoError(t, err)
	require.Len(t, got.Steps, 2)

	var savedFork *models.PromptChainStep
	for i := range got.Steps {
		if got.Steps[i].StepType == models.StepTypeFork {
			savedFork = &got.Steps[i]
		}
	}
	require.NotNil(t, savedFork, "fork step должен быть сохранён")
	assert.NotEmpty(t, savedFork.Conditions, "conditions JSONB не должен быть пустым")

	// Verify content survived round-trip.
	var conds models.Conditions
	require.NoError(t, json.Unmarshal(savedFork.Conditions, &conds))
	require.Len(t, conds.Branches, 2)
	assert.Equal(t, "Если критический баг", conds.Branches[0].Label)
	assert.Equal(t, target.ID, *conds.Branches[0].NextStepID)
	assert.Equal(t, "Если всё OK", conds.Branches[1].Label)
	assert.Nil(t, conds.Branches[1].NextStepID, "ветка ведущая в конец без next_step_id")
}

// uintToString — helper для форматирования id в JSON inline без strconv import.
func uintToString(v uint) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

func TestChainRepo_CountChainsUsingPrompt(t *testing.T) {
	r, db := setupChainTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "chain-count-using@test.local")
	p := createTestPrompt(t, db, u.ID, nil, "Step", "content", 0)
	c1 := createTestChain(t, db, u.ID, "Chain 1")
	c2 := createTestChain(t, db, u.ID, "Chain 2")

	createTestStep(t, db, c1.ID, p.ID, 1)
	createTestStep(t, db, c1.ID, p.ID, 2) // тот же prompt в той же цепочке
	createTestStep(t, db, c2.ID, p.ID, 1)

	count, err := r.CountChainsUsingPrompt(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count, "DISTINCT по chain — 2, не 3")

	// Soft-deleted цепочки не считаются.
	require.NoError(t, r.SoftDelete(ctx, c1.ID))
	count, err = r.CountChainsUsingPrompt(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
