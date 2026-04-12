package streak

import (
	"context"
	"errors"
	"log/slog"
	"time"

	repo "promptvault/internal/interface/repository"
)

type Service struct {
	streaks repo.StreakRepository
}

func NewService(streaks repo.StreakRepository) *Service {
	return &Service{streaks: streaks}
}

// RecordActivity обновляет streak пользователя. Best-effort — ошибки логируются, не пробрасываются.
func (s *Service) RecordActivity(ctx context.Context, userID uint, timezone string) {
	today := todayInTimezone(timezone)
	if err := s.streaks.RecordActivity(ctx, userID, today); err != nil {
		slog.Error("streak.record_activity.failed", "user_id", userID, "error", err)
	}
}

// GetStreak возвращает текущий streak пользователя.
func (s *Service) GetStreak(ctx context.Context, userID uint, timezone string) (*StreakOutput, error) {
	streak, err := s.streaks.GetByUserID(ctx, userID)
	if errors.Is(err, repo.ErrNotFound) {
		return &StreakOutput{
			CurrentStreak:  0,
			LongestStreak:  0,
			LastActiveDate: "",
			ActiveToday:    false,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	today := todayInTimezone(timezone)

	return &StreakOutput{
		CurrentStreak:  streak.CurrentStreak,
		LongestStreak:  streak.LongestStreak,
		LastActiveDate: streak.LastActiveDate,
		ActiveToday:    streak.LastActiveDate == today,
	}, nil
}

func todayInTimezone(tz string) string {
	if tz == "" {
		return time.Now().UTC().Format("2006-01-02")
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		slog.Warn("streak.timezone.invalid", "tz", tz, "error", err, "fallback", "UTC")
		loc = time.UTC
	}
	return time.Now().In(loc).Format("2006-01-02")
}
