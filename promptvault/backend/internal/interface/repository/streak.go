package repository

import (
	"context"

	"promptvault/internal/models"
)

type StreakRepository interface {
	RecordActivity(ctx context.Context, userID uint, today string) error
	GetByUserID(ctx context.Context, userID uint) (*models.UserStreak, error)

	// ListAtRisk — юзеры со streak > minStreak и last_active_date < todayUTC
	// (т.е. сегодня ещё не заходили). M-16: напоминание «не сломай серию».
	// LIMIT 500 — больше за один тик не обрабатываем, хватит на следующий день.
	ListAtRisk(ctx context.Context, todayUTC string, minStreak int) ([]models.UserStreak, error)

	// MarkReminderSent — ставит reminder_sent_on_date = todayUTC (idempotency,
	// не шлём второе письмо в тот же день при повторном тике).
	MarkReminderSent(ctx context.Context, userID uint, todayUTC string) error
}
