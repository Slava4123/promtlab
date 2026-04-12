package feedback

import "errors"

var (
	ErrInvalidType   = errors.New("недопустимый тип обратной связи")
	ErrMessageTooLong = errors.New("сообщение слишком длинное")
)
