package analytics

import (
	"context"
	"log/slog"
	"time"

	"promptvault/internal/infrastructure/metrics"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/pkg/safeloop"
)

// CleanupLoop — ежесуточный retention cleanup для team_activity_log,
// share_view_log и Phase 16-Y expired share_links. Один loop объединяет
// все, чтобы не плодить cron-ов; запросы к БД независимые и быстрые.
//
// Паттерн — зеркало trash.PurgeLoop: Ticker + stopCh, первый запуск сразу.
type CleanupLoop struct {
	activity   repo.TeamActivityRepository
	analytics  repo.AnalyticsRepository
	shareLinks repo.ShareLinkRepository
	interval   time.Duration
	// shareLinkGrace — сколько хранить просроченные ссылки перед hard DELETE.
	// 30 дней — даёт юзерам грейс-период, в течение которого фронт показывает
	// 410 Gone «срок истёк» вместо 404.
	shareLinkGrace time.Duration
	stopCh         chan struct{}
}

func NewCleanupLoop(activity repo.TeamActivityRepository, analytics repo.AnalyticsRepository, shareLinks repo.ShareLinkRepository, interval time.Duration) *CleanupLoop {
	return &CleanupLoop{
		activity:       activity,
		analytics:      analytics,
		shareLinks:     shareLinks,
		interval:       interval,
		shareLinkGrace: 30 * 24 * time.Hour,
		stopCh:         make(chan struct{}),
	}
}

func (l *CleanupLoop) Start() { go l.run() }
func (l *CleanupLoop) Stop()  { close(l.stopCh) }

func (l *CleanupLoop) run() {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	safeloop.RunWithRecover("analytics_cleanup", l.cleanup)
	for {
		select {
		case <-ticker.C:
			safeloop.RunWithRecover("analytics_cleanup", l.cleanup)
		case <-l.stopCh:
			return
		}
	}
}

func (l *CleanupLoop) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	metrics.AnalyticsCleanupRuns.Inc()

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

	if l.shareLinks != nil {
		if n, err := l.shareLinks.CleanupExpired(ctx, l.shareLinkGrace); err != nil {
			slog.Error("analytics.cleanup.share_links.expired.failed", "error", err)
		} else if n > 0 {
			slog.Info("analytics.cleanup.share_links.expired", "deleted", n, "grace", l.shareLinkGrace)
			metrics.AnalyticsCleanupDeleted.WithLabelValues("share_links_expired").Add(float64(n))
		}
	}
}
