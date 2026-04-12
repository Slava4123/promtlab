package trash

import "errors"

var (
	ErrNotFound      = errors.New("Элемент не найден в корзине")
	ErrForbidden     = errors.New("Нет доступа к этому элементу")
	ErrViewerReadOnly = errors.New("Читатель не может выполнить это действие")
	ErrInvalidType   = errors.New("Неизвестный тип элемента")
)
