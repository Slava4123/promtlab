package repository

import (
	"context"
	"time"

	"promptvault/internal/models"
)

// StarterRepository — атомарная операция установки starter промптов.
// Внутри одна SQL-транзакция: либо все промпты созданы и юзер помечен
// прошедшим онбординг, либо ничего. Это единственный метод, потому что
// единственный use case фичи — wizard finish.
type StarterRepository interface {
	// InstallTemplates создаёт переданные промпты и проставляет
	// users.onboarding_completed_at. Возвращает время маркировки.
	// prompts может быть пустым — тогда только маркируется юзер.
	InstallTemplates(ctx context.Context, userID uint, prompts []*models.Prompt) (time.Time, error)
}
