// Package chain — unit-tests для pure-функций cycle detection / fork validation.
//
// CR-7 в REVIEW_2026-05-07.md: usecases/chain/ имел 0 unit-тестов на 1191 LOC
// Phase 16 кода. Эти тесты покрывают самые опасные части без полного
// mocking-а Service:
//   - detectCycles — back-edge → endless run-loop → OOM на сервере;
//   - validateForkBranches — корректность graph-инвариантов перед SaveStep;
//   - findBranchByIndex — boundary check для AdvanceStep при fork.
//
// Service-level RBAC, snapshot semantics и tier-gate тесты — отдельный PR
// (требуют объёмных mock'ов ChainRepository/PromptRepository/TeamRepository).
package chain

import (
	"encoding/json"
	"errors"
	"testing"

	"promptvault/internal/models"
)

// uintPtr — helper для конструирования *uint в тестах.
func uintPtr(v uint) *uint { return &v }

// makePromptStep строит prompt-шаг с явным NextStepID (или nil = конец).
func makePromptStep(id uint, next *uint) models.PromptChainStep {
	return models.PromptChainStep{
		ID:         id,
		StepType:   models.StepTypePrompt,
		PromptID:   uintPtr(id * 100), // dummy promptID — в этих тестах не нужен
		NextStepID: next,
	}
}

// makeForkStep строит fork-шаг с заданными переходами по веткам.
func makeForkStep(t *testing.T, id uint, branches []models.ConditionBranch) models.PromptChainStep {
	t.Helper()
	raw, err := json.Marshal(models.Conditions{Branches: branches})
	if err != nil {
		t.Fatalf("marshal conditions: %v", err)
	}
	return models.PromptChainStep{
		ID:         id,
		StepType:   models.StepTypeFork,
		Conditions: raw,
	}
}

// ----------------------- detectCycles -----------------------

func TestDetectCycles_DAGLinear_OK(t *testing.T) {
	// 1 → 2 → 3 → nil
	steps := []models.PromptChainStep{
		makePromptStep(1, uintPtr(2)),
		makePromptStep(2, uintPtr(3)),
		makePromptStep(3, nil),
	}
	if err := detectCycles(steps); err != nil {
		t.Fatalf("expected no cycle on linear DAG, got %v", err)
	}
}

func TestDetectCycles_DAGWithFork_OK(t *testing.T) {
	// 1 (fork) → 2 или 3; 2 → nil; 3 → nil
	steps := []models.PromptChainStep{
		makeForkStep(t, 1, []models.ConditionBranch{
			{Label: "ok", NextStepID: uintPtr(2)},
			{Label: "fail", NextStepID: uintPtr(3)},
		}),
		makePromptStep(2, nil),
		makePromptStep(3, nil),
	}
	if err := detectCycles(steps); err != nil {
		t.Fatalf("expected no cycle on fork DAG, got %v", err)
	}
}

func TestDetectCycles_BackEdge_Refused(t *testing.T) {
	// 1 → 2 → 3 → 1 (back edge)
	steps := []models.PromptChainStep{
		makePromptStep(1, uintPtr(2)),
		makePromptStep(2, uintPtr(3)),
		makePromptStep(3, uintPtr(1)),
	}
	err := detectCycles(steps)
	if !errors.Is(err, ErrCycleInBranches) {
		t.Fatalf("expected ErrCycleInBranches on back edge, got %v", err)
	}
}

func TestDetectCycles_SelfLoop_Refused(t *testing.T) {
	// 1 → 1 (self-loop)
	steps := []models.PromptChainStep{
		makePromptStep(1, uintPtr(1)),
	}
	err := detectCycles(steps)
	if !errors.Is(err, ErrCycleInBranches) {
		t.Fatalf("expected ErrCycleInBranches on self-loop, got %v", err)
	}
}

func TestDetectCycles_ForkBackEdge_Refused(t *testing.T) {
	// 1 → 2 (fork) → 1 (back-edge через ветку)
	steps := []models.PromptChainStep{
		makePromptStep(1, uintPtr(2)),
		makeForkStep(t, 2, []models.ConditionBranch{
			{Label: "loop", NextStepID: uintPtr(1)},
		}),
	}
	err := detectCycles(steps)
	if !errors.Is(err, ErrCycleInBranches) {
		t.Fatalf("expected ErrCycleInBranches on fork→back edge, got %v", err)
	}
}

func TestDetectCycles_NextStepNotFound_Refused(t *testing.T) {
	// 1 → 999 (orphan link)
	steps := []models.PromptChainStep{
		makePromptStep(1, uintPtr(999)),
	}
	err := detectCycles(steps)
	if !errors.Is(err, ErrInvalidNextStep) {
		t.Fatalf("expected ErrInvalidNextStep on orphan next, got %v", err)
	}
}

func TestDetectCycles_ForkInvalidJSON_Refused(t *testing.T) {
	// Fork-шаг с corrupt JSON в Conditions.
	steps := []models.PromptChainStep{
		{
			ID:         1,
			StepType:   models.StepTypeFork,
			Conditions: json.RawMessage(`{not-json}`),
		},
	}
	err := detectCycles(steps)
	if !errors.Is(err, ErrInvalidConditions) {
		t.Fatalf("expected ErrInvalidConditions on bad JSON, got %v", err)
	}
}

func TestDetectCycles_DiamondPattern_OK(t *testing.T) {
	//        1 (fork)
	//       /  \
	//      2    3
	//       \  /
	//        4
	steps := []models.PromptChainStep{
		makeForkStep(t, 1, []models.ConditionBranch{
			{Label: "left", NextStepID: uintPtr(2)},
			{Label: "right", NextStepID: uintPtr(3)},
		}),
		makePromptStep(2, uintPtr(4)),
		makePromptStep(3, uintPtr(4)),
		makePromptStep(4, nil),
	}
	if err := detectCycles(steps); err != nil {
		t.Fatalf("expected no cycle on diamond DAG (multiple paths to same node OK), got %v", err)
	}
}

// ----------------------- validateForkBranches -----------------------

func TestValidateForkBranches_OK(t *testing.T) {
	raw, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{
			{Label: "ok", NextStepID: uintPtr(2)},
			{Label: "fail", NextStepID: uintPtr(3)},
		},
	})
	valid := map[uint]struct{}{2: {}, 3: {}}
	if err := validateForkBranches(raw, valid); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateForkBranches_NilNextStep_OK(t *testing.T) {
	// nil NextStepID = «конец цепочки по этой ветке» — должно проходить.
	raw, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{
			{Label: "done", NextStepID: nil},
		},
	})
	if err := validateForkBranches(raw, map[uint]struct{}{}); err != nil {
		t.Fatalf("expected nil for nil NextStepID, got %v", err)
	}
}

func TestValidateForkBranches_EmptyJSON_Refused(t *testing.T) {
	if err := validateForkBranches(nil, map[uint]struct{}{}); !errors.Is(err, ErrInvalidConditions) {
		t.Fatalf("expected ErrInvalidConditions on nil raw, got %v", err)
	}
}

func TestValidateForkBranches_BadJSON_Refused(t *testing.T) {
	if err := validateForkBranches(json.RawMessage(`not-json`), map[uint]struct{}{}); !errors.Is(err, ErrInvalidConditions) {
		t.Fatalf("expected ErrInvalidConditions on bad JSON, got %v", err)
	}
}

func TestValidateForkBranches_NoBranches_Refused(t *testing.T) {
	raw, _ := json.Marshal(models.Conditions{Branches: []models.ConditionBranch{}})
	if err := validateForkBranches(raw, map[uint]struct{}{}); !errors.Is(err, ErrInvalidConditions) {
		t.Fatalf("expected ErrInvalidConditions on empty branches, got %v", err)
	}
}

func TestValidateForkBranches_EmptyLabel_Refused(t *testing.T) {
	raw, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{{Label: "", NextStepID: nil}},
	})
	if err := validateForkBranches(raw, map[uint]struct{}{}); !errors.Is(err, ErrInvalidBranchLabel) {
		t.Fatalf("expected ErrInvalidBranchLabel on empty label, got %v", err)
	}
}

func TestValidateForkBranches_TooLongLabel_Refused(t *testing.T) {
	// 101 символ — выше лимита 100.
	long := make([]byte, 101)
	for i := range long {
		long[i] = 'A'
	}
	raw, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{{Label: string(long), NextStepID: nil}},
	})
	if err := validateForkBranches(raw, map[uint]struct{}{}); !errors.Is(err, ErrInvalidBranchLabel) {
		t.Fatalf("expected ErrInvalidBranchLabel on >100 chars, got %v", err)
	}
}

func TestValidateForkBranches_CyrillicLabelExactly100_OK(t *testing.T) {
	// 100 кириллических рун — на границе лимита, должны пройти (rune-aware).
	runes := make([]rune, 100)
	for i := range runes {
		runes[i] = 'я'
	}
	raw, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{{Label: string(runes), NextStepID: nil}},
	})
	if err := validateForkBranches(raw, map[uint]struct{}{}); err != nil {
		t.Fatalf("expected OK on 100 Cyrillic runes (rune-aware count), got %v", err)
	}
}

func TestValidateForkBranches_DuplicateLabel_Refused(t *testing.T) {
	raw, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{
			{Label: "ok", NextStepID: nil},
			{Label: "ok", NextStepID: nil},
		},
	})
	if err := validateForkBranches(raw, map[uint]struct{}{}); !errors.Is(err, ErrDuplicateBranchLabel) {
		t.Fatalf("expected ErrDuplicateBranchLabel, got %v", err)
	}
}

func TestValidateForkBranches_InvalidNextStep_Refused(t *testing.T) {
	raw, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{{Label: "ok", NextStepID: uintPtr(999)}},
	})
	valid := map[uint]struct{}{1: {}, 2: {}}
	if err := validateForkBranches(raw, valid); !errors.Is(err, ErrInvalidNextStep) {
		t.Fatalf("expected ErrInvalidNextStep on 999, got %v", err)
	}
}

// ----------------------- findBranchByIndex -----------------------

func TestFindBranchByIndex_Valid(t *testing.T) {
	raw, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{
			{Label: "first", NextStepID: uintPtr(10)},
			{Label: "second", NextStepID: uintPtr(20)},
		},
	})
	b, err := findBranchByIndex(raw, 1)
	if err != nil {
		t.Fatalf("expected OK, got %v", err)
	}
	if b.Label != "second" {
		t.Fatalf("expected 'second' branch, got %q", b.Label)
	}
}

func TestFindBranchByIndex_Negative_Refused(t *testing.T) {
	raw, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{{Label: "x", NextStepID: nil}},
	})
	if _, err := findBranchByIndex(raw, -1); !errors.Is(err, ErrChosenBranchNotFound) {
		t.Fatalf("expected ErrChosenBranchNotFound on negative index, got %v", err)
	}
}

func TestFindBranchByIndex_OutOfRange_Refused(t *testing.T) {
	raw, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{{Label: "x", NextStepID: nil}},
	})
	if _, err := findBranchByIndex(raw, 5); !errors.Is(err, ErrChosenBranchNotFound) {
		t.Fatalf("expected ErrChosenBranchNotFound on index 5, got %v", err)
	}
}

func TestFindBranchByIndex_EmptyJSON_Refused(t *testing.T) {
	if _, err := findBranchByIndex(nil, 0); !errors.Is(err, ErrInvalidConditions) {
		t.Fatalf("expected ErrInvalidConditions on nil raw, got %v", err)
	}
}

func TestFindBranchByIndex_BadJSON_Refused(t *testing.T) {
	if _, err := findBranchByIndex(json.RawMessage(`not-json`), 0); !errors.Is(err, ErrInvalidConditions) {
		t.Fatalf("expected ErrInvalidConditions on bad JSON, got %v", err)
	}
}
