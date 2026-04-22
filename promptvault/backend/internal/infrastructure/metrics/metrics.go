// Package metrics — централизованные Prometheus counters для observability.
// Регистрация через promauto, чтобы случайно не забыть MustRegister.
//
// Counters специально покрывают revenue-sensitive и SRE-critical события:
//   - ShareQuotaIncrementFailed — write прошёл, счётчик не увеличился (revenue leak).
//   - InsightsRefresh{result} — сколько успешных/rate-limited/error пересчётов.
//   - AnalyticsCleanupDeleted{table} — retention cleanup drift.
//
// Endpoint /metrics работает за IP-whitelist, см. NewHandler.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ShareQuotaIncrementFailed — share-link создана, но IncrementDailyUsage
	// упал. Каждое срабатывание = потенциальная потеря контроля над квотой
	// (юзер создал лишнюю ссылку, счётчик не увеличился). SRE alert rule:
	// rate(share_quota_increment_failed_total[5m]) > 0.
	ShareQuotaIncrementFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "share_quota_increment_failed_total",
		Help: "Number of successful share-link creations where daily usage increment failed (revenue leak).",
	})

	// InsightsRefresh — итерации /api/analytics/insights/refresh + background
	// cron. Label result: success | rate_limited | error.
	InsightsRefresh = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "analytics_insights_refresh_total",
		Help: "Number of Smart Insights refresh operations, labelled by outcome.",
	}, []string{"result"})

	// AnalyticsCleanupDeleted — удалённые строки retention cleanup-loop'ом.
	// Label table: team_activity | share_views | prompt_usage.
	AnalyticsCleanupDeleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "analytics_cleanup_deleted_total",
		Help: "Rows deleted by analytics retention cleanup, labelled by table.",
	}, []string{"table"})
)
