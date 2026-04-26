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

// init zero-инициализирует все ожидаемые label combinations CounterVec'ов.
//
// Без этого series появляется только после первого .Inc() — а пока loop
// видит пустой набор работы (total=0), он не зовёт Inc() ни на success,
// ни на error. В Prometheus при этом метрика отсутствует целиком, и
// alert `absent_over_time(...)[48h]` через час становится firing —
// false positive на тихом проде.
//
// .Add(0) создаёт series со значением 0 без побочных эффектов (idempotent
// относительно последующих Inc), absent_over_time возвращает 0,
// increase() остаётся корректным.
func init() {
	InsightsRefresh.WithLabelValues("success").Add(0)
	InsightsRefresh.WithLabelValues("rate_limited").Add(0)
	InsightsRefresh.WithLabelValues("error").Add(0)

	AnalyticsCleanupDeleted.WithLabelValues("team_activity").Add(0)
	AnalyticsCleanupDeleted.WithLabelValues("share_views").Add(0)
	AnalyticsCleanupDeleted.WithLabelValues("prompt_usage").Add(0)
}
