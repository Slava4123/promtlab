package starter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// Service — application logic для starter templates wizard'а.
//
// Single source of truth для каталога — embedded JSON (см. embed.go).
// Каталог парсится один раз при NewService и хранится в памяти на всё время
// жизни процесса. Никаких БД-запросов на ListCatalog → p95 < 1мс.
type Service struct {
	catalog *Catalog
	starter repo.StarterRepository
	users   repo.UserRepository

	// templatesByID — индекс для O(1) поиска по template_id, строится один
	// раз в NewService. Read-only после построения, поэтому safe для concurrent reads.
	templatesByID map[string]Template
}

func NewService(starter repo.StarterRepository, users repo.UserRepository) (*Service, error) {
	c, err := loadEmbeddedCatalog()
	if err != nil {
		return nil, err
	}
	idx := make(map[string]Template, len(c.Templates))
	for _, t := range c.Templates {
		if _, dup := idx[t.ID]; dup {
			return nil, fmt.Errorf("starter: duplicate template id %q in catalog.json", t.ID)
		}
		idx[t.ID] = t
	}
	return &Service{
		catalog:       c,
		starter:       starter,
		users:         users,
		templatesByID: idx,
	}, nil
}

// ListCatalog возвращает встроенный каталог. Read-only — каталог shared
// across requests, не модифицировать на месте.
func (s *Service) ListCatalog() *Catalog {
	return s.catalog
}

// Install — основная транзакция wizard finish. Одна атомарная операция:
// создаёт промпты-копии в личном workspace юзера + помечает онбординг
// пройденным. Любая ошибка → ничего не закоммичено.
//
// Особенности:
//   - templateIDs может быть пустым — это «Пропустить», маркируем без install
//   - повторный вызов на уже-completed юзере → ErrAlreadyCompleted (409)
//   - неизвестный id → ErrUnknownTemplate (400)
func (s *Service) Install(ctx context.Context, userID uint, templateIDs []string) (*InstallResult, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if user.OnboardingCompletedAt != nil {
		slog.Warn("starter.install.already_completed", "user_id", userID)
		return nil, ErrAlreadyCompleted
	}

	// Дедупликация: API контракт допускает любой массив, валидатор не enforces
	// uniqueness. Без этого "install":["dev-x","dev-x"] создал бы дубли промптов.
	seen := make(map[string]struct{}, len(templateIDs))
	prompts := make([]*models.Prompt, 0, len(templateIDs))
	for _, id := range templateIDs {
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		tpl, ok := s.templatesByID[id]
		if !ok {
			slog.Warn("starter.install.unknown_template", "template_id", id, "user_id", userID)
			return nil, fmt.Errorf("%w: %s", ErrUnknownTemplate, id)
		}
		prompts = append(prompts, &models.Prompt{
			UserID:  userID,
			TeamID:  nil, // starter промпты всегда личные
			Title:   tpl.Title,
			Content: tpl.Content,
			Model:   tpl.Model,
		})
	}

	completedAt, err := s.starter.InstallTemplates(ctx, userID, prompts)
	if err != nil {
		// Conditional UPDATE в repo не нашёл подходящей строки → юзер уже
		// помечен прошедшим онбординг конкурентной транзакцией. Маппим в
		// доменную ErrAlreadyCompleted (HTTP 409), чтобы клиент мог
		// синхронизировать своё локальное состояние через GET /api/auth/me.
		if errors.Is(err, repo.ErrConflict) {
			slog.Warn("starter.install.race_lost", "user_id", userID)
			return nil, ErrAlreadyCompleted
		}
		slog.Error("starter.install.tx_failed", "user_id", userID, "error", err)
		return nil, err
	}

	slog.Info("starter.install", "user_id", userID, "template_count", len(prompts))
	return &InstallResult{
		Prompts:     prompts,
		CompletedAt: completedAt,
	}, nil
}
