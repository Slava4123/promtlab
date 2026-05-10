package chain

import "errors"

// Доменные ошибки usecases/chain. Маппятся в HTTP-коды в delivery/http/chain/errors.go.
var (
	ErrNotFound                  = errors.New("Цепочка не найдена")
	ErrStepNotFound              = errors.New("Шаг не найден")
	ErrExecutionNotFound         = errors.New("Запуск не найден")
	ErrForbidden                 = errors.New("Нет доступа к этой цепочке")
	ErrViewerReadOnly            = errors.New("Читатель не может редактировать цепочки")
	ErrInvalidName               = errors.New("Имя цепочки обязательно")
	ErrInvalidDescription        = errors.New("Слишком длинное описание (макс. 2000 символов)")
	ErrInvalidVariableMapping    = errors.New("Некорректный variable_mapping")
	ErrPromptNotFound            = errors.New("Промпт не найден")
	ErrEmptyChain                = errors.New("Цепочка не содержит шагов")
	ErrExecutionAlreadyCompleted = errors.New("Запуск уже завершён или прерван")
	ErrChainHasActiveExecutions  = errors.New("У цепочки есть активные запуски")
	ErrConcurrentAdvance         = errors.New("Запуск был обновлён параллельно — обновите страницу и попробуйте снова")
	ErrInvalidStepPosition       = errors.New("Некорректная позиция шага")
	ErrReorderMismatch           = errors.New("Список step_ids не соответствует шагам цепочки")

	// Phase 16 v2 (Tree-canvas, manual fork). DSL evaluator удалён — юзер
	// сам выбирает ветку в run-mode через chosen_branch_index.
	ErrForkRequiresMax      = errors.New("Развилки доступны только на тарифе Max")
	ErrInvalidConditions    = errors.New("Некорректная структура branches")
	ErrInvalidForkStep      = errors.New("Fork-шаг должен иметь branches с непустыми label")
	ErrInvalidBranchLabel   = errors.New("Название ветки обязательно (макс. 100 символов)")
	ErrDuplicateBranchLabel = errors.New("Названия веток должны быть уникальны")
	ErrInvalidNextStep      = errors.New("Ветка ссылается на несуществующий шаг")
	ErrCycleInBranches      = errors.New("Обнаружен цикл в ветках развилки")
	ErrChooseBranchRequired = errors.New("Для fork-шага нужно указать chosen_branch_index")
	ErrChosenBranchNotFound = errors.New("Выбранная ветка не существует в текущем шаге")

	// Phase 16 v3 (Inline-tree editor, миграция 000056). Гарантии корректности
	// графа при добавлении шага через after_step_id / parent_fork_id.
	ErrCannotInsertAfterFork = errors.New("После шага-развилки можно вставить только в одну из веток (укажите parent_fork_id и branch_index)")
	ErrParentNotFork         = errors.New("Указанный parent_fork_id не является развилкой")
	ErrInsertForkLosesTail   = errors.New("Развилку нельзя вставлять в середине ветки — старый хвост ветки потеряется. Вставьте развилку в конец ветки.")

	// MoveStepUp / MoveStepDown — reorder шагов в линейной части.
	ErrCannotMoveFork        = errors.New("Развилку перетаскивать нельзя — пересоздайте")
	ErrCannotMoveAtBoundary  = errors.New("Шаг уже на границе — двигать некуда")
)
