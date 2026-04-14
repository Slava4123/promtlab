package subscription

import (
	"context"
	"log/slog"
	"time"

	repo "promptvault/internal/interface/repository"
)

// ExpirationLoop проверяет подписки с истёкшим периодом каждый час
// и даунгрейдит юзеров на free план.
type ExpirationLoop struct {
	subs     repo.SubscriptionRepository
	interval time.Duration
	stopCh   chan struct{}
}

func NewExpirationLoop(subs repo.SubscriptionRepository, interval time.Duration) *ExpirationLoop {
	return &ExpirationLoop{
		subs:     subs,
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
	}
}
