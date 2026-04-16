// Package engagement содержит background loops для email lifecycle
// не связанных напрямую с auth или subscription.
package engagement

import (
	"context"
	"log/slog"
	"time"

	iservice "promptvault/internal/interface/service"
	repo "promptvault/internal/interface/repository"
)

// ReengagementLoop шлёт re-engagement email юзерам, не заходившим 14+ дней (M-5d).
// Частота: раз в день. За раз обрабатывает batch юзеров (защита от массового спама
// при резком росте списка inactive).
//
// Чтобы не слать повторно чаще чем раз в 30 дней, используется reengagement_sent_at.
type ReengagementLoop struct {
	users       repo.UserRepository
	email       iservice.EmailSender
	frontendURL string
	interval    time.Duration
	stopCh      chan struct{}

	// inactiveAfter — сколько дней без входа считается "inactive".
	// cooldown — минимальный интервал между повторными re-engagement.
	// batchSize — сколько юзеров обрабатываем за один тик.
	inactiveAfter time.Duration
	cooldown      time.Duration
	batchSize     int
}

const (
	defaultInactiveAfter = 14 * 24 * time.Hour
	defaultCooldown      = 30 * 24 * time.Hour
	defaultBatchSize     = 50
)

// NewReengagementLoop. email может быть nil или !Configured — loop no-op'ит.
func NewReengagementLoop(users repo.UserRepository, email iservice.EmailSender, frontendURL string, interval time.Duration) *ReengagementLoop {
	return &ReengagementLoop{
		users:         users,
		email:         email,
		frontendURL:   frontendURL,
		interval:      interval,
		stopCh:        make(chan struct{}),
		inactiveAfter: defaultInactiveAfter,
		cooldown:      defaultCooldown,
		batchSize:     defaultBatchSize,
	}
}

func (l *ReengagementLoop) Start() { go l.run() }

func (l *ReengagementLoop) Stop() { close(l.stopCh) }

func (l *ReengagementLoop) run() {
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

func (l *ReengagementLoop) tick() {
	if l.email == nil || !l.email.Configured() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	now := time.Now()
	inactiveBefore := now.Add(-l.inactiveAfter)
	sentBefore := now.Add(-l.cooldown)

	users, err := l.users.ListInactiveForReengagement(ctx, inactiveBefore, sentBefore, l.batchSize)
	if err != nil {
		slog.Error("engagement.reengagement.list_failed", "error", err)
		return
	}
	for _, user := range users {
		if err := l.email.SendReengagement(user.Email, user.Name, l.frontendURL); err != nil {
			slog.Warn("engagement.reengagement.send_failed", "user_id", user.ID, "error", err)
			continue
		}
		if err := l.users.MarkReengagementSent(ctx, user.ID); err != nil {
			slog.Error("engagement.reengagement.mark_sent_failed", "user_id", user.ID, "error", err)
		}
		slog.Info("engagement.reengagement.sent", "user_id", user.ID)
	}
}
