package prompt

import "errors"

var (
	ErrNotFound        = errors.New("Промпт не найден")
	ErrForbidden       = errors.New("Нет доступа к этому промпту")
	ErrViewerReadOnly  = errors.New("Читатель не может редактировать промпты")
	ErrVersionNotFound = errors.New("Версия не найдена")
)
