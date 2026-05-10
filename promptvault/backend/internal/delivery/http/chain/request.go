package chain

import "encoding/json"

type CreateChainRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=2000"`
	TeamID      *uint  `json:"team_id"`
}

type UpdateChainRequest struct {
	Name        string `json:"name" validate:"max=100"`
	Description string `json:"description" validate:"max=2000"`
}

// AddStepRequest — добавление шага в цепочку. Куда вставить (взаимоисключающие):
//   - AfterStepID:               после указанного prompt-шага
//   - ParentForkID + BranchIndex: как первый шаг указанной ветки fork-шага
//   - ничего: tail-mode — в конец главной линии
//
// StepType: "" / "prompt" — обычный шаг; "fork" — развилка (Max-only).
type AddStepRequest struct {
	// PromptID обязателен для prompt-шагов; для fork-шагов — nil. Валидация
	// корректности связки prompt_id ↔ step_type делается в usecases/chain.
	PromptID         *uint           `json:"prompt_id"`
	Name             string          `json:"name" validate:"max=100"`
	VariableMapping  json.RawMessage `json:"variable_mapping"`
	ManualCheckpoint bool            `json:"manual_checkpoint"`
	StepType         string          `json:"step_type" validate:"omitempty,oneof=prompt fork"`
	Conditions       json.RawMessage `json:"conditions"`
	AfterStepID      *uint           `json:"after_step_id"`
	ParentForkID     *uint           `json:"parent_fork_id"`
	BranchIndex      *int            `json:"branch_index"`
}

type UpdateStepRequest struct {
	Name             string          `json:"name" validate:"max=100"`
	VariableMapping  json.RawMessage `json:"variable_mapping"`
	ManualCheckpoint bool            `json:"manual_checkpoint"`
	StepType         string          `json:"step_type" validate:"omitempty,oneof=prompt fork"`
	Conditions       json.RawMessage `json:"conditions"`
}

type ReorderStepsRequest struct {
	StepIDs []uint `json:"step_ids" validate:"required,min=1,dive,gt=0"`
}

type StartExecutionRequest struct {
	InitialVars json.RawMessage `json:"initial_vars"`
}

type AdvanceStepRequest struct {
	StepOutput string `json:"step_output" validate:"max=200000"`
	// ChosenBranchIndex — обязателен для fork-шагов (0-based индекс ветки).
	// Игнорируется для prompt-шагов.
	ChosenBranchIndex *int `json:"chosen_branch_index"`
}
