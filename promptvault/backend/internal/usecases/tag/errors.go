package tag

import "errors"

var (
	ErrNameEmpty      = errors.New("Название тега обязательно")
	ErrNotFound       = errors.New("Тег не найден")
	ErrForbidden      = errors.New("Нет доступа к этому тегу")
	ErrViewerReadOnly = errors.New("Читатель не может управлять тегами")
)
