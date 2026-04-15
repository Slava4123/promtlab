package subscription

import (
	"context"
	"errors"
	"log/slog"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// ExpirationNotifier уведомляет юзера о том, что подписка истекла и он переведён
// на free (после исчерпания retry-попыток автопродления или если auto_renew=false).
type ExpirationNotifier interface {
	NotifySubscriptionExpired(to, planName string) error
}

// ExpirationLoop проверяет подписки с истёкшим периодом каждый час
// и даунгрейдит юзеров на free план. Обрабатывает и active (юзер отменил или
// auto_renew=false) и past_due (retry-попытки исчерпаны).
type ExpirationLoop struct {
	subs     repo.SubscriptionRepository
	users    repo.UserRepository
	notifier ExpirationNotifier
	interval time.Duration
	stopCh   chan struct{}
}

// NewExpirationLoop создаёт цикл. users и notifier могут быть nil — тогда email
// не отправляется (dev или SMTP не настроен).
func NewExpirationLoop(subs repo.SubscriptionRepository, users repo.UserRepository, notifier ExpirationNotifier, interval time.Duration) *ExpirationLoop {
	return &ExpirationLoop{
		subs:     subs,
		users:    users,
		notifier: notifier,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (l *ExpirationLoop) Start() {
	go l.run()
}

func (l *ExpirationLoop) Stop() {
	close(l.stopCh)
}

func (l *ExpirationLoop) run() {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	// Первый запуск сразу при старте
	l.expire()

	for {
		select {
		case <-ticker.C:
			l.expire()
		case <-l.stopCh:
			return
		}
	}
}

func (l *ExpirationLoop) expire() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	subs, err := l.subs.ListExpiring(ctx, time.Now())
	if err != nil {
		slog.Error("subscription.expiration.list_failed", "error", err)
		return
	}

	for _, sub := range subs {
		if err := l.subs.ExpireAndDowngrade(ctx, sub.ID, sub.UserID); err != nil {
			slog.Error("subscription.expiration.downgrade_failed",
				"error", err,
				"subscription_id", sub.ID,
				"user_id", sub.UserID,
			)
			continue
		}
		slog.Info("subscription.expired",
			"subscription_id", sub.ID,
			"user_id", sub.UserID,
			"plan_id", sub.PlanID,
		)
		l.notifyExpired(ctx, &sub)
	}
}

func (l *ExpirationLoop) notifyExpired(ctx context.Context, sub *models.Subscription) {
	if l.notifier == nil || l.users == nil {
		return
	}
	user, err := l.users.GetByID(ctx, sub.UserID)
	if err != nil {
		if !errors.Is(err, repo.ErrNotFound) {
			slog.Warn("subscription.expiration.user_fetch_failed",
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
	if err := l.notifier.NotifySubscriptionExpired(user.Email, planName); err != nil {
		slog.Warn("subscription.expiration.notify_email_failed",
			"sub_id", sub.ID, "user_id", sub.UserID, "error", err)
	}
}
