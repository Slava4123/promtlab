package badge

import (
	"context"
	"errors"
	"log/slog"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	streakuc "promptvault/internal/usecases/streak"
)

// Service — usecase для badges. Инкапсулирует каталог + хук в prompt/collection
// usecases (Evaluate вызывается из prompt.Create, prompt.Update, prompt.IncrementUsage,
// collection.Create). Best-effort pattern: Evaluate никогда не возвращает error,
// все ошибки логируются через slog, основной flow никогда не блокируется.
type Service struct {
	badges  repo.BadgeRepository
	streaks *streakuc.Service

	catalog []Badge
	// byEvent — индекс для быстрого матчинга в Evaluate.
	// ключ: EventType → список бейджей, которые подписаны на этот event.
	byEvent map[EventType][]Badge
	// byID — для быстрого поиска бейджа по ID (для admin grant/revoke в будущем).
	byID map[string]Badge
}

// NewService загружает каталог из embedded JSON, строит индексы и возвращает
// готовый Service. Ошибка при загрузке каталога — fatal (bootstrap failure,
// app.go должен panic'ом остановить старт, как в starter/changelog).
func NewService(badges repo.BadgeRepository, streaks *streakuc.Service) (*Service, error) {
	catalog, err := LoadCatalog()
	if err != nil {
		return nil, err
	}
	byEvent := make(map[EventType][]Badge)
	byID := make(map[string]Badge, len(catalog))
	for _, b := range catalog {
		byID[b.ID] = b
		for _, t := range b.Triggers {
			byEvent[t] = append(byEvent[t], b)
		}
	}
	return &Service{
		badges:  badges,
		streaks: streaks,
		catalog: catalog,
		byEvent: byEvent,
		byID:    byID,
	}, nil
}

// Evaluate проверяет все бейджи, подписанные на event.Type, и разблокирует
// те, условия которых выполнены. Возвращает список newly unlocked для
// включения в ответ mutating API (newly_unlocked_badges).
//
// Гарантии:
//   - Никогда не возвращает error (best-effort, ошибки логируются через slog).
//   - Никогда не блокирует вызов, даже если BadgeRepository падает.
//   - Race-safe: при concurrent Evaluate с тем же event в одну и ту же секунду
//     UNIQUE constraint на (user_id, badge_id) гарантирует что Unlock
//     сработает ровно один раз; остальные вызовы получат ErrBadgeAlreadyUnlocked
//     и тихо пропустят бейдж в newly_unlocked (он уже был unlocked другим
//     вызовом, нет смысла возвращать его дважды как "newly").
//   - Если Evaluate вызван с event без подходящих бейджей в каталоге — no-op.
func (s *Service) Evaluate(ctx context.Context, userID uint, event Event) []Badge {
	candidates := s.byEvent[event.Type]
	if len(candidates) == 0 {
		return nil
	}

	unlocked, err := s.badges.UnlockedIDs(ctx, userID)
	if err != nil {
		slog.Warn("badge.eval.unlocked_ids_failed",
			"user_id", userID,
			"event", event.Type,
			"error", err)
		return nil
	}

	var newly []Badge
	for _, b := range candidates {
		if _, already := unlocked[b.ID]; already {
			continue
		}
		// Для team-бейджей дополнительный short-circuit: если event не содержит
		// TeamID, нет смысла даже проверять условие.
		if b.Condition.Type == CondTeamPromptCount && event.TeamID == nil {
			continue
		}

		ok, err := s.checkCondition(ctx, userID, b.Condition)
		if err != nil {
			slog.Warn("badge.eval.condition_check_failed",
				"user_id", userID,
				"badge_id", b.ID,
				"condition", b.Condition.Type,
				"error", err)
			continue
		}
		if !ok {
			continue
		}

		if err := s.badges.Unlock(ctx, userID, b.ID); err != nil {
			if errors.Is(err, repo.ErrBadgeAlreadyUnlocked) {
				// Race: concurrent Evaluate успел первым. Не ошибка, но и не newly.
				continue
			}
			slog.Warn("badge.eval.unlock_failed",
				"user_id", userID,
				"badge_id", b.ID,
				"error", err)
			continue
		}

		slog.Info("badge.unlock",
			"user_id", userID,
			"badge_id", b.ID,
			"trigger", event.Type)
		newly = append(newly, b)
	}
	return newly
}

// List возвращает все бейджи каталога с текущим состоянием пользователя для
// GET /api/badges. Для разблокированных бейджей Progress=Target. Для заблокированных
// вычисляется фактический прогресс из тех же aggregation-методов, что использует
// Evaluate. При ошибке агрегирования Progress ставится в 0, а в лог идёт warn.
func (s *Service) List(ctx context.Context, userID uint) ([]BadgeWithState, error) {
	// Один запрос к user_badges вместо N через UnlockedIDs.
	userBadges, err := s.badges.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	unlockedMap := make(map[string]models.UserBadge, len(userBadges))
	for _, ub := range userBadges {
		unlockedMap[ub.BadgeID] = ub
	}

	result := make([]BadgeWithState, 0, len(s.catalog))
	for _, b := range s.catalog {
		state := BadgeWithState{
			Badge:  b,
			Target: b.Condition.Threshold,
		}
		if ub, ok := unlockedMap[b.ID]; ok {
			state.Unlocked = true
			unlockedAt := ub.UnlockedAt
			state.UnlockedAt = &unlockedAt
			state.Progress = b.Condition.Threshold
			result = append(result, state)
			continue
		}

		progress, err := s.currentProgress(ctx, userID, b.Condition)
		if err != nil {
			slog.Warn("badge.list.progress_failed",
				"user_id", userID,
				"badge_id", b.ID,
				"error", err)
			progress = 0
		}
		// Clamp на Target, чтобы фронт никогда не показывал 12/10 для locked.
		// (такая ситуация возможна если юзер выполнил условие, но evaluate
		// не успел сработать — например, прогресс пришёл из других источников).
		if progress > b.Condition.Threshold {
			progress = b.Condition.Threshold
		}
		state.Progress = progress
		result = append(result, state)
	}

	// Стабильный порядок: по catalog order (как определено в catalog.json).
	// Не сортируем по alphabet или дате — порядок каталога семантический
	// (от простого к сложному), важен для UX.
	return result, nil
}

// GetByID возвращает бейдж из каталога по ID.
// Используется admin.GrantBadge / RevokeBadge для валидации badge_id
// (убедиться, что бейдж существует в каталоге) и для возврата полного
// определения в ответе API.
// Возвращает (Badge{}, false) если badge_id не найден.
func (s *Service) GetByID(id string) (Badge, bool) {
	b, ok := s.byID[id]
	return b, ok
}

// Unlock напрямую разблокирует бейдж у юзера (admin grant).
// В отличие от Evaluate не проверяет условие — админ решает, что юзер
// заслужил бейдж. Возвращает repo.ErrBadgeAlreadyUnlocked если уже есть.
func (s *Service) Unlock(ctx context.Context, userID uint, badgeID string) error {
	return s.badges.Unlock(ctx, userID, badgeID)
}

// Revoke удаляет бейдж у юзера (admin revoke). Идемпотентно: если записи
// нет — возвращает nil. Факт revoke должен фиксироваться в audit_log
// вызывающей стороной (admin.Service.RevokeBadge).
func (s *Service) Revoke(ctx context.Context, userID uint, badgeID string) error {
	return s.badges.DeleteByUserAndBadge(ctx, userID, badgeID)
}

// checkCondition проверяет, выполнено ли условие бейджа для юзера.
func (s *Service) checkCondition(ctx context.Context, userID uint, c Condition) (bool, error) {
	current, err := s.currentProgress(ctx, userID, c)
	if err != nil {
		return false, err
	}
	return current >= c.Threshold, nil
}

// currentProgress возвращает текущее значение метрики для condition type.
// Единая точка раскатки всех ConditionType в конкретные aggregation-методы.
func (s *Service) currentProgress(ctx context.Context, userID uint, c Condition) (int64, error) {
	switch c.Type {
	case CondSoloPromptCount:
		return s.badges.CountSoloPrompts(ctx, userID)
	case CondTeamPromptCount:
		return s.badges.CountTeamPrompts(ctx, userID)
	case CondTotalPromptCount:
		return s.badges.CountAllPrompts(ctx, userID)
	case CondSoloCollectionCount:
		return s.badges.CountSoloCollections(ctx, userID)
	case CondTeamCollectionCount:
		return s.badges.CountTeamCollections(ctx, userID)
	case CondTotalUsage:
		return s.badges.SumUsage(ctx, userID)
	case CondVersionedPromptCount:
		return s.badges.CountVersionedPrompts(ctx, userID, c.MinVersions)
	case CondCurrentStreak:
		return s.currentStreak(ctx, userID)
	}
	return 0, ErrUnknownCondition
}

// currentStreak читает текущий streak из streakuc.Service.
// Если streaks == nil (тесты/специальный режим) — возвращает 0.
// Timezone в этом контексте не важен — streak уже посчитан при последнем
// RecordActivity, а сюда приходит только current_streak как число.
func (s *Service) currentStreak(ctx context.Context, userID uint) (int64, error) {
	if s.streaks == nil {
		return 0, nil
	}
	out, err := s.streaks.GetStreak(ctx, userID, "")
	if err != nil {
		return 0, err
	}
	if out == nil {
		return 0, nil
	}
	return int64(out.CurrentStreak), nil
}

