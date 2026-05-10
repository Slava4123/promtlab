package chain

import (
	"encoding/json"
	"unicode/utf8"

	"promptvault/internal/models"
)

// Phase 16 v2 (Tree-canvas): manual fork. Юзер сам выбирает ветку в run-mode,
// больше нет автоматического матчинга output через DSL. Здесь остаются только:
//   - validateForkBranches — проверка структуры branches при сохранении шага
//   - detectCycles         — DFS-проверка отсутствия петель в графе
//   - findBranchByIndex    — по chosen_branch_index в run-mode

// validateForkBranches проверяет JSONB при сохранении fork-шага:
//   - JSON корректный
//   - branches непустой массив
//   - каждый branch имеет Label (≤100 символов), уникальный в шаге
//   - все NextStepID указывают на существующий шаг (validStepIDs)
//   - nil NextStepID допустим — означает «конец цепочки» по этой ветке
func validateForkBranches(raw json.RawMessage, validStepIDs map[uint]struct{}) error {
	if len(raw) == 0 {
		return ErrInvalidConditions
	}
	var conds models.Conditions
	if err := json.Unmarshal(raw, &conds); err != nil {
		return ErrInvalidConditions
	}
	if len(conds.Branches) == 0 {
		return ErrInvalidConditions
	}
	seenLabels := make(map[string]struct{}, len(conds.Branches))
	for _, b := range conds.Branches {
		if b.Label == "" || utf8.RuneCountInString(b.Label) > 100 {
			return ErrInvalidBranchLabel
		}
		if _, dup := seenLabels[b.Label]; dup {
			return ErrDuplicateBranchLabel
		}
		seenLabels[b.Label] = struct{}{}
		if b.NextStepID != nil {
			if _, ok := validStepIDs[*b.NextStepID]; !ok {
				return ErrInvalidNextStep
			}
		}
	}
	return nil
}

// detectCycles — DFS с цветами. Phase 16 v3: граф строится по явным ссылкам:
//
//	StepTypePrompt → next_step_id (если nil — нет исходящих рёбер)
//	StepTypeFork   → все NextStepID из branches
//
// Возвращает ErrCycleInBranches при back-edge.
func detectCycles(steps []models.PromptChainStep) error {
	byID := make(map[uint]*models.PromptChainStep, len(steps))
	for i := range steps {
		byID[steps[i].ID] = &steps[i]
	}

	edges := make(map[uint][]uint, len(steps))
	for i := range steps {
		s := &steps[i]
		switch s.StepType {
		case models.StepTypeFork:
			if len(s.Conditions) == 0 {
				continue
			}
			var conds models.Conditions
			if err := json.Unmarshal(s.Conditions, &conds); err != nil {
				return ErrInvalidConditions
			}
			for _, b := range conds.Branches {
				if b.NextStepID == nil {
					continue
				}
				if _, ok := byID[*b.NextStepID]; !ok {
					return ErrInvalidNextStep
				}
				edges[s.ID] = append(edges[s.ID], *b.NextStepID)
			}
		default:
			if s.NextStepID == nil {
				continue
			}
			if _, ok := byID[*s.NextStepID]; !ok {
				return ErrInvalidNextStep
			}
			edges[s.ID] = append(edges[s.ID], *s.NextStepID)
		}
	}

	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[uint]int, len(steps))
	var dfs func(id uint) error
	dfs = func(id uint) error {
		color[id] = gray
		for _, next := range edges[id] {
			switch color[next] {
			case gray:
				return ErrCycleInBranches
			case white:
				if err := dfs(next); err != nil {
					return err
				}
			}
		}
		color[id] = black
		return nil
	}
	for id := range byID {
		if color[id] == white {
			if err := dfs(id); err != nil {
				return err
			}
		}
	}
	return nil
}

// findBranchByIndex — для AdvanceStep при fork-шаге: возвращает branch по
// 0-based индексу. Используем index, потому что Label может быть переименован,
// но порядок branches стабилен в conditions JSONB.
func findBranchByIndex(raw json.RawMessage, index int) (*models.ConditionBranch, error) {
	if len(raw) == 0 {
		return nil, ErrInvalidConditions
	}
	var conds models.Conditions
	if err := json.Unmarshal(raw, &conds); err != nil {
		return nil, ErrInvalidConditions
	}
	if index < 0 || index >= len(conds.Branches) {
		return nil, ErrChosenBranchNotFound
	}
	return &conds.Branches[index], nil
}
