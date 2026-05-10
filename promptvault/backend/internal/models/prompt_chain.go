package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PromptChain — последовательность промптов с output→input маппингом.
// Phase 16, миграция 000053. Soft-delete через gorm.DeletedAt.
type PromptChain struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	TeamID      *uint          `gorm:"index" json:"team_id,omitempty"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Description string         `gorm:"type:text;not null;default:''" json:"description"`
	User        User           `gorm:"foreignKey:UserID" json:"-"`
	Team        *Team          `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Steps       []PromptChainStep `gorm:"foreignKey:ChainID;constraint:OnDelete:CASCADE" json:"steps,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// PromptChainStepPreview — облегчённое представление шага для list-эндпойнта
// /api/chains. Содержит только тип и позицию первых N шагов цепочки — без
// контента промптов, conditions, variable_mapping. Используется для рендера
// mini-graph на карточке цепочки в UI (Phase 16 UI polish).
type PromptChainStepPreview struct {
	Position int    `json:"position"`
	StepType string `json:"step_type"`
}

// PromptChainListRow — расширенное представление цепочки для list-эндпойнта,
// агрегирующее статистику в одном SELECT через LATERAL (см. chain_repo.
// ListByUserWithStats). Не персистентный тип: GORM не сохраняет его.
//
// StepsPreview ограничен первыми ChainStepsPreviewLimit шагами; UI может
// нарисовать первые 4-5 шагов и показать "+N more" для длинных цепочек.
type PromptChainListRow struct {
	PromptChain
	StepCount      int                      `json:"step_count"`
	HasBranching   bool                     `json:"has_branching"`
	SavedRunsCount int                      `json:"saved_runs_count"`
	StepsPreview   []PromptChainStepPreview `json:"steps_preview"`
}

// ChainStepsPreviewLimit — максимальное число шагов в StepsPreview.
// 5 покрывает «компактное» отображение в UI; для длинных цепочек +N бейдж.
const ChainStepsPreviewLimit = 5

// PromptChainStep — шаг в цепочке. Position сохранён как стабильная сортировка
// для UI (Order BY position в репозитории), но логика переходов после миграции
// 000056 строится на NextStepID, не на position+1.
//
// VariableMapping JSONB — отображение {var_name: VariableSource}; десериализуется
// в map[string]VariableSource в сервисе через ParseVariableMapping.
type PromptChainStep struct {
	ID       uint  `gorm:"primaryKey" json:"id"`
	ChainID  uint  `gorm:"not null;index" json:"chain_id"`
	Position int   `gorm:"not null" json:"position"`
	// PromptID — для prompt-шага обязателен; для fork-шага NULL (fork — это
	// контейнер с ветками, без своего промпта). Валидация в usecases/chain.
	PromptID *uint           `gorm:"index" json:"prompt_id,omitempty"`
	Name     string          `gorm:"size:255;not null;default:''" json:"name"`
	VariableMapping  json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"variable_mapping"`
	ManualCheckpoint bool            `gorm:"not null;default:false" json:"manual_checkpoint"`
	// StepType='prompt' — обычный шаг. 'fork' — ветвление, переход через
	// Conditions.Branches (юзер выбирает в run-mode). См. миграции 000054, 000055.
	StepType   string          `gorm:"size:20;not null;default:'prompt'" json:"step_type"`
	Conditions json.RawMessage `gorm:"type:jsonb" json:"conditions,omitempty"`
	// NextStepID — явный переход для prompt-шагов (миграция 000056). nil = конец
	// ветки/цепочки. Игнорируется для fork-шагов — у них переход через
	// Conditions.Branches[chosenIdx].NextStepID.
	NextStepID *uint     `gorm:"index" json:"next_step_id,omitempty"`
	Prompt     *Prompt   `gorm:"foreignKey:PromptID" json:"prompt,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Допустимые значения PromptChainStep.StepType.
// Phase 16 v2 (Tree-canvas): 'conditional' переименован в 'fork' — manual choice
// юзером вместо DSL-эвалюатора. CHECK constraint обновлён в миграции 000055.
const (
	StepTypePrompt = "prompt"
	StepTypeFork   = "fork"
)

// Validate проверяет sum-type инварианты PromptChainStep.
//
// MJ-28: до этого fix'а ограничения держались только в комментариях.
// Step имеет дискриминатор StepType и две группы полей:
//   - StepType=prompt → PromptID != nil, Conditions == nil
//   - StepType=fork   → PromptID == nil, Conditions != nil (branches)
// Без validation миграция/ручной hotfix мог создать «гибридный» шаг,
// и run-loop поведение становилось неопределённым.
func (s *PromptChainStep) Validate() error {
	switch s.StepType {
	case StepTypePrompt:
		if s.PromptID == nil {
			return fmt.Errorf("prompt_chain_step: prompt step requires prompt_id")
		}
		if len(s.Conditions) > 0 {
			return fmt.Errorf("prompt_chain_step: prompt step must not have conditions")
		}
	case StepTypeFork:
		if s.PromptID != nil {
			return fmt.Errorf("prompt_chain_step: fork step must not have prompt_id (got %d)", *s.PromptID)
		}
		if len(s.Conditions) == 0 {
			return fmt.Errorf("prompt_chain_step: fork step requires conditions branches")
		}
	default:
		return fmt.Errorf("prompt_chain_step: unknown step_type %q", s.StepType)
	}
	return nil
}

// BeforeSave — GORM hook, вызывает Validate перед каждым INSERT/UPDATE.
func (s *PromptChainStep) BeforeSave(_ *gorm.DB) error { return s.Validate() }

// Conditions — структура prompt_chain_steps.conditions JSONB.
// Только при StepType='fork'. Каждая branch имеет читаемый Label («Если код OK»,
// «Если критический баг») — юзер выбирает по нему в run-mode.
type Conditions struct {
	Branches []ConditionBranch `json:"branches"`
}

type ConditionBranch struct {
	// Label — отображается юзеру в run-mode как кнопка выбора пути.
	// Обязателен, ≤200 символов, уникален в пределах одного fork-шага.
	Label string `json:"label"`
	// NextStepID — id PromptChainStep, на который перейти при выборе этой ветки.
	// nil — конец цепочки по этой ветке (юзер выбрал «закончить здесь»).
	NextStepID *uint `json:"next_step_id,omitempty"`
}

// PromptChainExecution — запуск цепочки. Status проходит in_progress → completed
// (или → abandoned по TTL/cleanup loop). ChainSnapshot JSONB — заморозка структуры
// на момент старта (Decision §2.2): редактирование цепочки в процессе run не
// ломает выполнение. Variables — initial_vars от пользователя; StepOutputs —
// {step_id: output_text} накопительно.
type PromptChainExecution struct {
	ID            uint                 `gorm:"primaryKey" json:"id"`
	ChainID       uint                 `gorm:"not null;index" json:"chain_id"`
	UserID        uint                 `gorm:"not null;index" json:"user_id"`
	CurrentStep   int                  `gorm:"not null;default:1" json:"current_step"`
	Variables     json.RawMessage      `gorm:"type:jsonb;not null;default:'{}'" json:"variables"`
	StepOutputs   json.RawMessage      `gorm:"type:jsonb;not null;default:'{}'" json:"step_outputs"`
	ChainSnapshot json.RawMessage      `gorm:"type:jsonb;not null" json:"chain_snapshot"`
	// MN-33: typed Status вместо raw string. Захватывает invalid-state
	// присваивания на compile-time + параллель с CHECK constraint в миграции 000053.
	Status        ChainExecutionStatus `gorm:"size:20;not null;default:in_progress" json:"status"`
	StartedAt     time.Time            `gorm:"not null;default:now()" json:"started_at"`
	CompletedAt   *time.Time           `json:"completed_at,omitempty"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// ChainExecutionStatus — статус выполнения цепочки.
// CHECK constraint в миграции 000053 enforces те же значения на БД-уровне.
type ChainExecutionStatus string

// Допустимые значения PromptChainExecution.Status.
const (
	ChainExecutionStatusInProgress ChainExecutionStatus = "in_progress"
	ChainExecutionStatusCompleted  ChainExecutionStatus = "completed"
	ChainExecutionStatusAbandoned  ChainExecutionStatus = "abandoned"
)

// IsValid возвращает true если значение — допустимый статус.
func (s ChainExecutionStatus) IsValid() bool {
	switch s {
	case ChainExecutionStatusInProgress, ChainExecutionStatusCompleted, ChainExecutionStatusAbandoned:
		return true
	}
	return false
}

// VariableSource — источник значения переменной в шаге цепочки.
//   Type="manual"      → значение вводит юзер вручную в run-mode
//   Type="step_output" → берётся output предыдущего шага по StepID
//   Type="chain_var"   → берётся из chain-level Variables по VarName
type VariableSource struct {
	Type    string  `json:"type"`
	StepID  *uint   `json:"step_id,omitempty"`
	VarName *string `json:"var_name,omitempty"`
}

// Допустимые значения VariableSource.Type (валидация в usecases/chain).
const (
	VariableSourceManual     = "manual"
	VariableSourceStepOutput = "step_output"
	VariableSourceChainVar   = "chain_var"
)

// ChainSnapshot — структура цепочки, замороженная в момент StartExecution.
// Сериализуется в PromptChainExecution.ChainSnapshot JSONB. AdvanceStep работает
// со snapshot, игнорируя текущее состояние prompt_chains/prompt_chain_steps.
//
// PromptContents — map[promptID]content. Содержит контент промптов на момент Start,
// чтобы рендеринг не зависел от последующих изменений Prompt.Content. Если промпт
// удалён между Start и AdvanceStep — execution всё равно завершится корректно.
type ChainSnapshot struct {
	Chain          PromptChain       `json:"chain"`
	Steps          []PromptChainStep `json:"steps"`
	PromptContents map[uint]string   `json:"prompt_contents"`
}
