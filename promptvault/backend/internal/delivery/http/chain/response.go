package chain

import (
	"encoding/json"
	"time"

	"promptvault/internal/models"
)

type ChainResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	TeamID      *uint     `json:"team_id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewChainResponse(c models.PromptChain) ChainResponse {
	return ChainResponse{
		ID:          c.ID,
		UserID:      c.UserID,
		TeamID:      c.TeamID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

type StepResponse struct {
	ID       uint   `json:"id"`
	ChainID  uint   `json:"chain_id"`
	Position int    `json:"position"`
	// PromptID — nil для fork-шагов (контейнер без своего промпта).
	PromptID *uint  `json:"prompt_id,omitempty"`
	Name     string `json:"name"`
	VariableMapping  json.RawMessage `json:"variable_mapping"`
	ManualCheckpoint bool            `json:"manual_checkpoint"`
	StepType         string          `json:"step_type"`
	Conditions       json.RawMessage `json:"conditions,omitempty"`
	// Phase 16 v3: явный переход для prompt-шагов. nil = конец ветки/цепочки.
	NextStepID *uint     `json:"next_step_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	// Phase 16 v2: preloaded prompt — title для отображения в Canvas-узлах,
	// content для hover-preview. omitempty — на случай soft-deleted prompt.
	Prompt *PromptSummary `json:"prompt,omitempty"`
}

// PromptSummary — облегчённое представление промпта для отображения в шаге.
// Не возвращаем full prompt (теги/коллекции/версии) — только то что показываем в UI.
type PromptSummary struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func NewStepResponse(s models.PromptChainStep) StepResponse {
	resp := StepResponse{
		ID:               s.ID,
		ChainID:          s.ChainID,
		Position:         s.Position,
		PromptID:         s.PromptID,
		Name:             s.Name,
		VariableMapping:  s.VariableMapping,
		ManualCheckpoint: s.ManualCheckpoint,
		StepType:         s.StepType,
		Conditions:       s.Conditions,
		NextStepID:       s.NextStepID,
		CreatedAt:        s.CreatedAt,
	}
	if s.Prompt != nil {
		resp.Prompt = &PromptSummary{
			ID:      s.Prompt.ID,
			Title:   s.Prompt.Title,
			Content: s.Prompt.Content,
		}
	}
	return resp
}

type ChainDetailResponse struct {
	ChainResponse
	Steps []StepResponse `json:"steps"`
}

func NewChainDetailResponse(c models.PromptChain) ChainDetailResponse {
	steps := make([]StepResponse, len(c.Steps))
	for i, s := range c.Steps {
		steps[i] = NewStepResponse(s)
	}
	return ChainDetailResponse{
		ChainResponse: NewChainResponse(c),
		Steps:         steps,
	}
}

type ChainListResponse struct {
	Items []ChainListItem `json:"items"`
	Total int64           `json:"total"`
	Limit int             `json:"limit"`
	Offset int            `json:"offset"`
}

// ChainListItem — расширенное представление цепочки для list-эндпойнта /api/chains.
// Phase 16 UI polish: содержит агрегатную статистику + steps_preview для рендера
// mini-graph на карточке цепочки в UI.
type ChainListItem struct {
	ChainResponse
	StepCount      int                `json:"step_count"`
	HasBranching   bool               `json:"has_branching"`
	SavedRunsCount int                `json:"saved_runs_count"`
	StepsPreview   []ChainStepPreview `json:"steps_preview"`
}

// ChainStepPreview — облегчённое представление шага для mini-graph: только
// position и step_type. Не содержит контента промпта, conditions, mapping.
type ChainStepPreview struct {
	Position int    `json:"position"`
	StepType string `json:"step_type"`
}

func NewChainListItem(row models.PromptChainListRow) ChainListItem {
	preview := make([]ChainStepPreview, len(row.StepsPreview))
	for i, p := range row.StepsPreview {
		preview[i] = ChainStepPreview{Position: p.Position, StepType: p.StepType}
	}
	return ChainListItem{
		ChainResponse:  NewChainResponse(row.PromptChain),
		StepCount:      row.StepCount,
		HasBranching:   row.HasBranching,
		SavedRunsCount: row.SavedRunsCount,
		StepsPreview:   preview,
	}
}

// ExecutionSummary — компактное представление execution для списка истории.
// БЕЗ chain_snapshot/step_outputs/variables (могут быть мегабайтными JSON'ами).
// Полная инфа доступна через GET /api/executions/{exec_id}.
type ExecutionSummary struct {
	ID          uint       `json:"id"`
	ChainID     uint       `json:"chain_id"`
	UserID      uint       `json:"user_id"`
	CurrentStep int        `json:"current_step"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func NewExecutionSummary(e models.PromptChainExecution) ExecutionSummary {
	return ExecutionSummary{
		ID:          e.ID,
		ChainID:     e.ChainID,
		UserID:      e.UserID,
		CurrentStep: e.CurrentStep,
		Status:      string(e.Status),
		StartedAt:   e.StartedAt,
		CompletedAt: e.CompletedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

type ExecutionListResponse struct {
	Items []ExecutionSummary `json:"items"`
	Limit int                `json:"limit"`
}

type ExecutionResponse struct {
	ID            uint            `json:"id"`
	ChainID       uint            `json:"chain_id"`
	UserID        uint            `json:"user_id"`
	CurrentStep   int             `json:"current_step"`
	Variables     json.RawMessage `json:"variables"`
	StepOutputs   json.RawMessage `json:"step_outputs"`
	ChainSnapshot json.RawMessage `json:"chain_snapshot"`
	Status        string          `json:"status"`
	StartedAt     time.Time       `json:"started_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func NewExecutionResponse(e models.PromptChainExecution) ExecutionResponse {
	return ExecutionResponse{
		ID:            e.ID,
		ChainID:       e.ChainID,
		UserID:        e.UserID,
		CurrentStep:   e.CurrentStep,
		Variables:     e.Variables,
		StepOutputs:   e.StepOutputs,
		ChainSnapshot: e.ChainSnapshot,
		Status:        string(e.Status),
		StartedAt:     e.StartedAt,
		CompletedAt:   e.CompletedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}
