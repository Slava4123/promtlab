package collection

import "errors"

var (
	ErrNotFound       = errors.New("Коллекция не найдена")
	ErrForbidden      = errors.New("Нет доступа к этой коллекции")
	ErrViewerReadOnly = errors.New("Читатель не может редактировать коллекции")
)
