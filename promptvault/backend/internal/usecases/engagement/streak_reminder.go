package engagement

import (
	"context"
	"log/slog"
	"time"

	iservice "promptvault/internal/interface/service"
	repo "promptvault/internal/interface/repository"
)

// StreakReminderLoop шлёт "не сломай серию" напоминания (M-16).
// Тик раз в день (интервал передаётся); внутри — одна проходка по ListAtRisk
// с batch limit 500. Критерий at-risk: current_streak > minStreak и
// last_active_date < today. Защита от повтора в тот же день через
// reminder_sent_on DATE-колонку в user_streaks (idempotency).
//
// Важно: UTC-based. Полная реализация по user-tz потребовала бы хранить
// timezone per-user и N тиков разных для разных TZ. Для MVP принимаем, что
// 17:00 UTC (≈ 20:00 MSK) — разумное время напоминания для РФ-аудитории.
type StreakReminderLoop struct {
	streaks     repo.StreakRepository
	users       repo.UserRepository
	email       iservice.EmailSender
	frontendURL string
	interval    time.Duration
	minStreak   int
	stopCh      chan struct{}
}

func NewStreakReminderLoop(streaks repo.StreakRepository, users repo.UserRepository, email iservice.EmailSender, frontendURL string, interval time.Duration) *StreakReminderLoop {
	return &StreakReminderLoop{
		streaks:     streaks,
		users:       users,
		email:       email,
		frontendURL: frontendURL,
		interval:    interval,
		minStreak:   3, // ниже 3 не имеет смысла напоминать — серия легко восстановима
		stopCh:      make(chan struct{}),
	}
}

func (l *StreakReminderLoop) Start() { go l.run() }
func (l *StreakReminderLoop) Stop()  { close(l.stopCh) }

func (l *StreakReminderLoop) run() {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()
	l.tick()
	for {
		select {
		case <-ticker.C:
			l.tick()
		case <-l.stopCh:
			return
		}
	}
}

func (l *StreakReminderLoop) tick() {
	if l.email == nil || !l.email.Configured() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	today := time.Now().UTC().Format("2006-01-02")

	atRisk, err := l.streaks.ListAtRisk(ctx, today, l.minStreak)
	if err != nil {
		slog.Error("streak.reminder.list_failed", "error", err)
		return
	}
	for _, s := range atRisk {
		u, err := l.users.GetByID(ctx, s.UserID)
		if err != nil || u == nil || u.Email == "" {
			continue
		}
		if err := l.email.SendStreakReminder(u.Email, u.Name, s.CurrentStreak, l.frontendURL); err != nil {
			slog.Warn("streak.reminder.email_failed", "user_id", s.UserID, "error", err)
			continue
		}
		if err := l.streaks.MarkReminderSent(ctx, s.UserID, today); err != nil {
			slog.Error("streak.reminder.mark_sent_failed", "user_id", s.UserID, "error", err)
		}
	}
}
