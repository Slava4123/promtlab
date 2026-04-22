package activity

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// Service — запись и чтение team activity log.
//
// Write-path (Log/LogSafe) вызывается хуками из других usecases:
// prompt/collection/share/team/user. LogSafe глотает ошибки, чтобы fail
// логирования не ломал основную операцию.
//
// Read-path (ListByTeam/GetPromptHistory) — для HTTP handler и MCP tool.
type Service struct {
	activity repo.TeamActivityRepository
	users    repo.UserRepository
}

func NewService(activity repo.TeamActivityRepository, users repo.UserRepository) *Service {
	return &Service{activity: activity, users: users}
}

// Log записывает событие. Если ActorEmail пуст и ActorID != 0 — подтягивает
// email/name из UserRepository. ActorEmail не может быть пустым в БД (NOT NULL),
// поэтому при невозможности резолва возвращает ErrMissingActor.
func (s *Service) Log(ctx context.Context, e Event) error {
	if e.TeamID == 0 {
		return ErrMissingTeam
	}
	if e.EventType == "" || e.TargetType == "" {
		return ErrMissingEventType
	}

	if e.ActorEmail == "" && e.ActorID != 0 {
		user, err := s.users.GetByID(ctx, e.ActorID)
		if err != nil {
			return fmt.Errorf("resolve actor %d: %w", e.ActorID, err)
		}
		e.ActorEmail = user.Email
		e.ActorName = user.Name
	}
	if e.ActorEmail == "" {
		return ErrMissingActor
	}

	meta, err := marshalMetadata(e.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	var actorID *uint
	if e.ActorID != 0 {
		actorID = &e.ActorID
	}

	return s.activity.Log(ctx, &models.TeamActivityLog{
		TeamID:      e.TeamID,
		ActorID:     actorID,
		ActorEmail:  e.ActorEmail,
		ActorName:   e.ActorName,
		EventType:   e.EventType,
		TargetType:  e.TargetType,
		TargetID:    e.TargetID,
		TargetLabel: e.TargetLabel,
		Metadata:    meta,
	})
}

// LogSafe — как Log, но проглатывает ошибку с slog.Warn. Используется в хуках
// других usecases, где log-fail не должен ломать основной flow. Nil-safe:
// при вызове на nil-сервисе (если activity не подключена) — no-op.
func (s *Service) LogSafe(ctx context.Context, e Event) {
	if s == nil {
		return
	}
	if err := s.Log(ctx, e); err != nil {
		slog.WarnContext(ctx, "activity log failed",
			"err", err,
			"event_type", e.EventType,
			"target_type", e.TargetType,
		)
	}
}

// ListByTeam — feed команды с cursor-пагинацией (для /api/teams/:id/activity).
func (s *Service) ListByTeam(ctx context.Context, filter repo.TeamActivityFilter) ([]models.TeamActivityLog, *time.Time, error) {
	return s.activity.List(ctx, filter)
}

// GetPromptHistory — все события, связанные с промптом (для /api/prompts/:id/history).
// Склейка с prompt_versions — на стороне handler.
func (s *Service) GetPromptHistory(ctx context.Context, promptID uint, limit int) ([]models.TeamActivityLog, error) {
	return s.activity.ListByTarget(ctx, models.TargetPrompt, promptID, limit)
}

// AnonymizeActor — GDPR hook для user.DeleteAccount. Заменяет actor_id=NULL,
// actor_email/name на константы models.AnonymizedActor*. Возвращает количество
// затронутых строк.
func (s *Service) AnonymizeActor(ctx context.Context, userID uint) (int64, error) {
	return s.activity.AnonymizeActor(ctx, userID)
}
