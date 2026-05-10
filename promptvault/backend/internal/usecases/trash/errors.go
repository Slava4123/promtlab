package trash

import "errors"

var (
	ErrNotFound      = errors.New("Элемент не найден в корзине")
	ErrForbidden     = errors.New("Нет доступа к этому элементу")
	ErrViewerReadOnly = errors.New("Читатель не может выполнить это действие")
	ErrInvalidType   = errors.New("Неизвестный тип элемента")
	// ErrPromptUsedInChains — Phase 16. Hard-delete заблокирован, потому что
	// промпт используется в одной или более цепочках (FK без CASCADE).
	// User action: убрать из всех цепочек или удалить цепочки сначала.
	ErrPromptUsedInChains = errors.New("Промпт используется в цепочках. Удалите его из всех цепочек, чтобы продолжить.")
)
