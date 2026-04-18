package apikey

import "errors"

var (
	ErrKeyNotFound     = errors.New("API-ключ не найден")
	ErrMaxKeysReached  = errors.New("Достигнут лимит API-ключей")
	ErrNameEmpty       = errors.New("Название ключа обязательно")
	ErrNameTooLong     = errors.New("Название ключа не может быть длиннее 100 символов")
	ErrUnauthorized    = errors.New("Недействительный API-ключ")
	ErrExpired         = errors.New("API-ключ истёк")
	ErrInvalidToolName = errors.New("Неизвестное имя инструмента в allowed_tools")
	ErrInvalidExpires  = errors.New("expires_at должен быть в будущем")
	ErrScopeDenied     = errors.New("Операция запрещена политикой ключа")
	ErrTeamMismatch    = errors.New("team_id запроса не совпадает с team_id ключа")
)
