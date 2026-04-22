package analytics

import (
	"context"
	"log/slog"
	"time"

	"promptvault/internal/infrastructure/metrics"
	repo "promptvault/internal/interface/repository"
)

// CleanupLoop — ежесуточный retention cleanup для team_activity_log и
// share_view_log. Один loop объединяет оба, чтобы не плодить cron-ов;
// запросы к БД независимые и быстрые.
//
// Паттерн — зеркало trash.PurgeLoop: Ticker + stopCh, первый запуск сразу.
type CleanupLoop struct {
	activity  repo.TeamActivityRepository
	analytics repo.AnalyticsRepository
	interval  time.Duration
	stopCh    chan struct{}
}

func NewCleanupLoop(activity repo.TeamActivityRepository, analytics repo.AnalyticsRepository, interval time.Duration) *CleanupLoop {
	return &CleanupLoop{
		activity:  activity,
		analytics: analytics,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

func (l *CleanupLoop) Start() { go l.run() }
func (l *CleanupLoop) Stop()  { close(l.stopCh) }

func (l *CleanupLoop) run() {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	l.cleanup()
	for {
		select {
		case <-ticker.C:
			l.cleanup()
		case <-l.stopCh:
			return
		}
	}
}

func (l *CleanupLoop) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if n, err := l.activity.CleanupByRetention(ctx); err != nil {
		slog.Error("analytics.cleanup.activity.failed", "error", err)
	} else if n > 0 {
		slog.Info("analytics.cleanup.activity", "deleted", n)
		metrics.AnalyticsCleanupDeleted.WithLabelValues("team_activity").Add(float64(n))
	}

	if n, err := l.analytics.CleanupShareViewsByRetention(ctx); err != nil {
		slog.Error("analytics.cleanup.share_views.failed", "error", err)
	} else if n > 0 {
		slog.Info("analytics.cleanup.share_views", "deleted", n)
		metrics.AnalyticsCleanupDeleted.WithLabelValues("share_views").Add(float64(n))
	}

	if n, err := l.analytics.CleanupPromptUsageByRetention(ctx); err != nil {
		slog.Error("analytics.cleanup.prompt_usage.failed", "error", err)
	} else if n > 0 {
		slog.Info("analytics.cleanup.prompt_usage", "deleted", n)
		metrics.AnalyticsCleanupDeleted.WithLabelValues("prompt_usage").Add(float64(n))
	}
}
