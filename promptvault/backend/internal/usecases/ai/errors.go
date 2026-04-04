package ai

import "errors"

var (
	ErrRateLimited   = errors.New("Превышен лимит запросов. Попробуйте позже")
	ErrModelNotFound = errors.New("Модель не найдена")
	ErrEmptyContent  = errors.New("Содержимое промпта не может быть пустым")
	ErrAPIKeyMissing = errors.New("AI-сервис не настроен")
)
