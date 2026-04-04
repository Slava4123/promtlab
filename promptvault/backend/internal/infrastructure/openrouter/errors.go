package openrouter

import "errors"

var (
	ErrUnauthorized        = errors.New("неверный API-ключ OpenRouter")
	ErrRateLimited         = errors.New("превышен лимит запросов OpenRouter")
	ErrInsufficientCredits = errors.New("недостаточно средств на аккаунте OpenRouter")
	ErrModelNotFound       = errors.New("модель не найдена в OpenRouter")
	ErrEmptyResponse       = errors.New("пустой ответ от модели")
)
