package prompt

import "errors"

var (
	ErrNotFound           = errors.New("Промпт не найден")
	ErrForbidden          = errors.New("Нет доступа к этому промпту")
	ErrViewerReadOnly     = errors.New("Читатель не может редактировать промпты")
	ErrVersionNotFound    = errors.New("Версия не найдена")
	ErrWorkspaceMismatch  = errors.New("Коллекции и теги должны принадлежать тому же пространству, что и промпт")
)
