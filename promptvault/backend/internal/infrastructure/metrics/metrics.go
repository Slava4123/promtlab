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

	// InsightsTeamRun — итерации team-scope расчёта Smart Insights.
	// Label result: success | error (получение списка команд или ComputeInsights упал).
	// Phase 15: добавлено вместе с TeamRepository.ListOwnedTeams.
	InsightsTeamRun = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "analytics_insights_loop_team_run_total",
		Help: "Team-scope Smart Insights compute iterations, labelled by outcome.",
	}, []string{"result"})

	// AnalyticsCleanupRuns — каждый тик retention cleanup-loop'а, инкрементится
	// в начале cleanup() независимо от того, удалено что-то или нет.
	// Используется в alert CleanupLoopStalled чтобы отличить «loop помер» от
	// «нечего удалять» (свежий prod, retention 90d не достигнут).
	AnalyticsCleanupRuns = promauto.NewCounter(prometheus.CounterOpts{
		Name: "analytics_cleanup_runs_total",
		Help: "Number of retention cleanup loop iterations (regardless of rows deleted).",
	})

	// InsightsLoopRuns — каждый тик insights compute-loop'а, инкрементится
	// в начале compute() независимо от наличия Max-юзеров.
	// Используется в alert InsightsComputeLoopStalled чтобы отличить
	// «loop помер» от «нет Max-юзеров на проде».
	InsightsLoopRuns = promauto.NewCounter(prometheus.CounterOpts{
		Name: "analytics_insights_loop_runs_total",
		Help: "Number of Smart Insights compute loop iterations (regardless of users processed).",
	})

	// ChainsCreated — создание новой Prompt Chain. Phase 16.
	// Label scope: personal | team.
	ChainsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "chains_created_total",
		Help: "Number of created Prompt Chains, labelled by scope (personal vs team).",
	}, []string{"scope"})

	// ChainExecutionsStarted — запуск цепочки (StartExecution). Phase 16.
	ChainExecutionsStarted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "chain_executions_started_total",
		Help: "Number of started Prompt Chain executions.",
	})

	// ChainExecutionsCompleted — переход execution в финальный статус.
	// Label status: completed | abandoned (TTL cleanup).
	ChainExecutionsCompleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "chain_executions_completed_total",
		Help: "Number of finalized Prompt Chain executions, labelled by status.",
	}, []string{"status"})

	// Phase 16-X. Branding logo upload (bytea storage) — лимит-критичный path.

	// TeamBrandingLogoUploads — попытки загрузить логотип через POST /branding/logo.
	// Label result: success|too_large|bad_format|forbidden|other.
	// «forbidden» покрывает Max-gate + не-owner; «other» — внутренние ошибки.
	TeamBrandingLogoUploads = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "team_branding_logo_uploads_total",
		Help: "Team branding logo upload attempts, labelled by validation result.",
	}, []string{"result"})

	// TeamBrandingLogoSizeBytes — гистограмма размеров загруженных файлов
	// (только success). Helps capacity planning и обнаружения юзеров около лимита.
	TeamBrandingLogoSizeBytes = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "team_branding_logo_size_bytes",
		Help:    "Successfully uploaded logo file sizes in bytes.",
		Buckets: []float64{10_000, 50_000, 100_000, 250_000, 500_000, 1_048_576},
	})

	// TeamBrandingLogoServe — попадания в GET /branding/logo.
	// Label cache_hit: hit (304) | miss (200). hit_rate = hit / (hit+miss).
	TeamBrandingLogoServe = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "team_branding_logo_serve_total",
		Help: "Team branding logo serve responses, labelled by cache hit (304 vs 200).",
	}, []string{"cache_hit"})

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

	InsightsTeamRun.WithLabelValues("success").Add(0)
	InsightsTeamRun.WithLabelValues("error").Add(0)

	ChainsCreated.WithLabelValues("personal").Add(0)
	ChainsCreated.WithLabelValues("team").Add(0)
	ChainExecutionsCompleted.WithLabelValues("completed").Add(0)
	ChainExecutionsCompleted.WithLabelValues("abandoned").Add(0)

	// Phase 16-X branding logo metrics — zero-init все labels чтобы absent_over_time
	// в alerts не давал false-positive на тихом проде.
	TeamBrandingLogoUploads.WithLabelValues("success").Add(0)
	TeamBrandingLogoUploads.WithLabelValues("too_large").Add(0)
	TeamBrandingLogoUploads.WithLabelValues("bad_format").Add(0)
	TeamBrandingLogoUploads.WithLabelValues("forbidden").Add(0)
	TeamBrandingLogoUploads.WithLabelValues("other").Add(0)
	TeamBrandingLogoServe.WithLabelValues("hit").Add(0)
	TeamBrandingLogoServe.WithLabelValues("miss").Add(0)
}
