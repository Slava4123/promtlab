// backend/internal/usecases/referral/reward_loop.go
package referral

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"promptvault/internal/infrastructure/metrics"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/pkg/safeloop"
)

// RewardLoop — background обработчик pending'ов реферальных наград.
// Ежечасно SELECT'ит referral_pending_rewards с eligible_at < now,
// вызывает Service.GrantReward и удаляет row на success ИЛИ terminal-skip.
//
// Паттерн скопирован с subscription.RenewalLoop (safeloop + ticker + stop chan):
//   - первый тик прямо на старте (immediate first-run);
//   - каждый тик обёрнут в safeloop.RunWithRecover — panic не убивает loop;
//   - tickOnce extracted для testability (тесты могут гонять без ticker'а).
//
// Terminal vs transient errors:
//   - terminal (Granted / ErrAlreadyRewarded / ErrPaymentRefunded /
//     ErrReferrerMissing) → pending.Delete (нет смысла retry'ить);
//   - transient (DB/transport errors) → pending остаётся для retry на след. тике.
type RewardLoop struct {
	svc      *Service
	pending  repo.ReferralRewardRepository
	interval time.Duration
	batch    int
	nowFn    func() time.Time
	stopCh   chan struct{}
}

// NewRewardLoop создаёт background loop. interval — частота тика
// (рекомендуется 1h в prod, 1m в dev), batch — лимит SELECT (рекомендуется 100,
// чтобы один тик не блокировал DB надолго на bootstrap-всплеске).
func NewRewardLoop(svc *Service, pending repo.ReferralRewardRepository, interval time.Duration, batch int) *RewardLoop {
	return &RewardLoop{
		svc:      svc,
		pending:  pending,
		interval: interval,
		batch:    batch,
		nowFn:    time.Now,
		stopCh:   make(chan struct{}),
	}
}

// SetNowFn — for tests. Позволяет заморозить время в tickOnce, чтобы pending'и
// с eligible_at в "будущем" не попадали в SELECT.
func (l *RewardLoop) SetNowFn(fn func() time.Time) { l.nowFn = fn }

// Start запускает loop в отдельной goroutine. Idempotent не делаем — повторный
// Start спровоцирует два конкурирующих ticker'а (как и в RenewalLoop).
func (l *RewardLoop) Start() {
	slog.Info("referral.reward.loop_started", "interval", l.interval, "batch", l.batch)
	go l.run()
}

// Stop сигналит loop'у остановиться. Текущий tickOnce доработает (его контекст
// независимый), но следующего тика не будет.
func (l *RewardLoop) Stop() { close(l.stopCh) }

func (l *RewardLoop) run() {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()
	// Immediate first-run: не ждать interval после старта, иначе после restart'а
	// сервера pending'ы простоят до часа без обработки.
	safeloop.RunWithRecover("referral_reward", func() { _ = l.tickOnce(context.Background()) })
	for {
		select {
		case <-ticker.C:
			safeloop.RunWithRecover("referral_reward", func() { _ = l.tickOnce(context.Background()) })
		case <-l.stopCh:
			slog.Info("referral.reward.loop_stopped")
			return
		}
	}
}

// tickOnce — обрабатывает один batch eligible pending'ов и возвращает summary.
// Extracted (не inline в run()) для unit-тестов: можно гонять без ticker'а
// и проверять Granted/Skipped*/Errors counters.
//
// Внутренний timeout 5m защищает от висящего ListEligible/GrantReward
// (например, если PG проигрывает дедлок-retry или payment provider тормозит).
func (l *RewardLoop) tickOnce(ctx context.Context) RewardSummary {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	var summary RewardSummary
	now := l.nowFn()
	pendings, err := l.pending.ListEligible(ctx, now, l.batch)
	if err != nil {
		slog.ErrorContext(ctx, "referral.reward.list_failed", "err", err)
		return summary
	}
	for _, p := range pendings {
		grantErr := l.svc.GrantReward(ctx, p.ReferrerID, p.RefereeID, p.PaymentID)
		// terminal=true → удаляем pending (retry бесполезен). terminal=false →
		// оставляем для следующего тика (transient infra error).
		terminal := true
		switch {
		case grantErr == nil:
			summary.Granted++
		case errors.Is(grantErr, ErrAlreadyRewarded):
			summary.SkippedActive++
			metrics.ReferralRewardsSkippedTotal.WithLabelValues("already_rewarded").Inc()
			slog.InfoContext(ctx, "referral.reward.skipped_already_rewarded",
				"referrer_id", p.ReferrerID, "referee_id", p.RefereeID)
		case errors.Is(grantErr, ErrPaymentRefunded):
			summary.SkippedRefund++
			metrics.ReferralRewardsSkippedTotal.WithLabelValues("refunded").Inc()
			slog.InfoContext(ctx, "referral.reward.skipped_refunded",
				"referrer_id", p.ReferrerID, "referee_id", p.RefereeID, "payment_id", p.PaymentID)
		case errors.Is(grantErr, ErrReferrerMissing):
			summary.SkippedDeleted++
			metrics.ReferralRewardsSkippedTotal.WithLabelValues("referrer_deleted").Inc()
			slog.InfoContext(ctx, "referral.reward.skipped_referrer_missing",
				"referrer_id", p.ReferrerID, "referee_id", p.RefereeID)
		default:
			summary.Errors++
			terminal = false
			slog.ErrorContext(ctx, "referral.reward.grant_failed",
				"err", grantErr, "referrer_id", p.ReferrerID, "referee_id", p.RefereeID)
		}
		if terminal {
			if err := l.pending.Delete(ctx, p.ID); err != nil {
				// Delete fail после успешного grant'а — некритично: следующий
				// тик попробует ещё раз; GrantReward вернёт ErrAlreadyRewarded
				// (idempotent на MarkReferralRewarded CAS) и Delete пройдёт.
				slog.ErrorContext(ctx, "referral.reward.delete_failed",
					"err", err, "pending_id", p.ID)
			}
		}
	}
	if summary.Total() > 0 {
		slog.InfoContext(ctx, "referral.reward.tick_summary",
			"granted", summary.Granted,
			"skipped_refund", summary.SkippedRefund,
			"skipped_active", summary.SkippedActive,
			"skipped_deleted", summary.SkippedDeleted,
			"errors", summary.Errors)
	}
	return summary
}
