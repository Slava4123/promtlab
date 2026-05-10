package quota

import (
	"context"
	"log/slog"
	"time"

	"promptvault/internal/infrastructure/metrics"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/pkg/safeloop"
)

// CleanupLoop — ежесуточный retention cleanup для daily_feature_usage.
// Read-path использует только сегодняшний день (GetDailyUsage с date=now),
// всё остальное — мёртвый balast. Без cleanup таблица растёт линейно по
// юзерам × фичам × дни (на 10К юзеров × 2 фичи = ~7M строк/год).
//
// Default retention = 30 дней. Это с запасом: даже при анализе подобных
// trends читается только текущий день, 30д даёт debug-window для recent
// incidents без раздувания.
//
// Паттерн полностью повторяет analytics.CleanupLoop: Ticker + stopCh,
// первый запуск сразу при Start (чтобы не ждать сутки на старте сервера).
type CleanupLoop struct {
	quotas        repo.QuotaRepository
	interval      time.Duration
	retentionDays int
	stopCh        chan struct{}
}

// DefaultRetentionDays — 30 дней. Менять только если есть конкретная
// бизнес-причина (например, юр-комплаенс на хранение или дольше debug-window).
const DefaultRetentionDays = 30

func NewCleanupLoop(quotas repo.QuotaRepository, interval time.Duration) *CleanupLoop {
	return &CleanupLoop{
		quotas:        quotas,
		interval:      interval,
		retentionDays: DefaultRetentionDays,
		stopCh:        make(chan struct{}),
	}
}

func (l *CleanupLoop) Start() { go l.run() }
func (l *CleanupLoop) Stop()  { close(l.stopCh) }

func (l *CleanupLoop) run() {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	safeloop.RunWithRecover("quota_cleanup", l.cleanup)
	for {
		select {
		case <-ticker.C:
			safeloop.RunWithRecover("quota_cleanup", l.cleanup)
		case <-l.stopCh:
			return
		}
	}
}

func (l *CleanupLoop) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	metrics.QuotaCleanupRuns.Inc()

	n, err := l.quotas.DeleteOldDailyUsage(ctx, l.retentionDays)
	if err != nil {
		slog.Error("quota.cleanup.daily_usage.failed", "error", err)
		return
	}
	if n > 0 {
		slog.Info("quota.cleanup.daily_usage", "deleted", n, "retention_days", l.retentionDays)
		metrics.QuotaCleanupRowsDeleted.Add(float64(n))
	}
}
