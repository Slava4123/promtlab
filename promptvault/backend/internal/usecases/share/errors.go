package share

import "errors"

var (
	ErrNotFound       = errors.New("Ссылка не найдена")
	ErrPromptNotFound = errors.New("Промпт не найден")
	ErrForbidden      = errors.New("Нет доступа к этому промпту")
	ErrViewerReadOnly = errors.New("Читатель не может делиться промптами")
)
