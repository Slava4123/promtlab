package subscription

import (
	"context"
	"errors"
	"log/slog"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// ReminderLoop шлёт pre-expire напоминания юзерам с auto_renew=false за 3 и 1 день
// до истечения периода (M-5b). auto_renew=true юзерам эти письма не нужны —
// за них RenewalLoop либо продлит, либо пришлёт RenewalFailed.
//
// Stage:
//   1 — отправлено 3-day reminder (period_end ∈ (now, now+3d])
//   2 — отправлено 1-day reminder (period_end ∈ (now, now+1d])
// ExtendPeriod сбрасывает stage=0 при успешном продлении.
type ReminderLoop struct {
	subs     repo.SubscriptionRepository
	users    repo.UserRepository
	notifier ReminderNotifier
	interval time.Duration
	stopCh   chan struct{}
}

// ReminderNotifier — контракт для отправки pre-expire писем.
type ReminderNotifier interface {
	NotifyPreExpireReminder(to, planName string, daysLeft int, endsAt time.Time) error
}

// NewReminderLoop — interval обычно 1-6 часов. users и notifier могут быть nil
// (SMTP не настроен) — loop стартует, но ничего не делает.
func NewReminderLoop(subs repo.SubscriptionRepository, users repo.UserRepository, notifier ReminderNotifier, interval time.Duration) *ReminderLoop {
	return &ReminderLoop{
		subs:     subs,
		users:    users,
		notifier: notifier,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (l *ReminderLoop) Start() { go l.run() }

func (l *ReminderLoop) Stop() { close(l.stopCh) }

func (l *ReminderLoop) run() {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	l.tick() // первый запуск при старте
	for {
		select {
		case <-ticker.C:
			l.tick()
		case <-l.stopCh:
			return
		}
	}
}

func (l *ReminderLoop) tick() {
	if l.notifier == nil || l.users == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	now := time.Now()

	// 1-day reminder имеет приоритет — если юзер впервые попадает в окно уже за
	// день до expire (например, купил подписку на 2 дня в sandbox-тесте), пропустим
	// 3-day stage и сразу перейдём к 2.
	l.sendStage(ctx, now, now.Add(24*time.Hour), 2, 1)
	// 3-day reminder — только тем, кому ещё не отправляли первый stage.
	l.sendStage(ctx, now, now.Add(72*time.Hour), 1, 3)
}

// sendStage обрабатывает одну ступень reminders. minStage — пропускаем подписки,
// у которых pre_expire_stage >= minStage (чтобы не отправлять повторно).
// newStage — значение, выставляемое после успешной отправки.
func (l *ReminderLoop) sendStage(ctx context.Context, now, upTo time.Time, newStage int16, daysLeft int) {
	subs, err := l.subs.ListPreExpiring(ctx, now, upTo, newStage)
	if err != nil {
		slog.Error("subscription.reminder.list_failed", "stage", newStage, "error", err)
		return
	}
	for _, sub := range subs {
		l.notify(ctx, &sub, newStage, daysLeft)
	}
}

func (l *ReminderLoop) notify(ctx context.Context, sub *models.Subscription, stage int16, daysLeft int) {
	user, err := l.users.GetByID(ctx, sub.UserID)
	if err != nil {
		if !errors.Is(err, repo.ErrNotFound) {
			slog.Warn("subscription.reminder.user_fetch_failed",
				"sub_id", sub.ID, "user_id", sub.UserID, "error", err)
		}
		return
	}
	if user == nil || user.Email == "" {
		return
	}
	planName := sub.PlanID
	if sub.Plan.Name != "" {
		planName = sub.Plan.Name
	}
	if err := l.notifier.NotifyPreExpireReminder(user.Email, planName, daysLeft, sub.CurrentPeriodEnd); err != nil {
		slog.Warn("subscription.reminder.send_failed",
			"sub_id", sub.ID, "user_id", sub.UserID, "days_left", daysLeft, "error", err)
		// Не выставляем stage при ошибке — пусть следующий tick повторит.
		return
	}
	if err := l.subs.SetPreExpireStage(ctx, sub.ID, stage); err != nil {
		slog.Error("subscription.reminder.set_stage_failed",
			"sub_id", sub.ID, "stage", stage, "error", err)
		return
	}
	slog.Info("subscription.reminder.sent",
		"sub_id", sub.ID, "user_id", sub.UserID, "days_left", daysLeft, "stage", stage)
}
