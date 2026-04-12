package apikey

import "errors"

var (
	ErrKeyNotFound    = errors.New("API-ключ не найден")
	ErrMaxKeysReached = errors.New("Достигнут лимит API-ключей")
	ErrNameEmpty      = errors.New("Название ключа обязательно")
	ErrNameTooLong    = errors.New("Название ключа не может быть длиннее 100 символов")
	ErrUnauthorized   = errors.New("Недействительный API-ключ")
)
