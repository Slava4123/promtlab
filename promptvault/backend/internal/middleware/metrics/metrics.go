// Package metrics предоставляет HTTP middleware для экспонирования
// Prometheus-метрик уровня приложения: rate, errors, duration (RED method)
// плюс in-flight gauge.
//
// Метрики регистрируются один раз через promauto на уровне пакета.
// Endpoint /metrics обслуживается отдельным handler'ом из
// internal/infrastructure/metrics/handler.go (тот же registry).
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// httpRequestsTotal — общее число обработанных HTTP-запросов.
	// Labels: method (GET/POST/...), path (Chi route pattern, не literal URL),
	// status (HTTP status code).
	//
	// Cardinality: контролируется нормализацией path до route pattern
	// (chi.RouteContext.RoutePattern), без user-IDs в URL.
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests handled, partitioned by method, route and status.",
	}, []string{"method", "path", "status"})

	// httpRequestDuration — длительность обработки запроса в секундах.
	// Histogram buckets — Prometheus default (5ms..10s) — покрывает типичный
	// web latency profile.
	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds, partitioned by method, route and status.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	// httpRequestsInFlight — текущее число одновременных запросов в обработке.
	// Поднимается в начале handler, падает в defer.
	httpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Current number of HTTP requests being processed.",
	})
)

// Middleware инструментирует HTTP requests Prometheus-метриками.
// Должна быть зарегистрирована ПОСЛЕ chi.NewRouter() и ВНЕ
// группы routes (чтобы покрывать все handlers, включая /metrics).
//
// Path label — это chi route pattern (например "/api/prompts/{id}"),
// извлекаемый ПОСЛЕ next.ServeHTTP — Chi заполняет RouteContext во время
// dispatch'а. Для unmatched (404) routes Chi отвечает 404 ДО вызова
// middleware chain — такие запросы НЕ записываются в metrics. Это
// сознательный выбор: random 404 paths иначе создавали бы cardinality
// explosion. NotFoundHandler можно зарегистрировать через r.NotFound()
// если потребуется учёт.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		next.ServeHTTP(ww, r)

		method := r.Method
		status := strconv.Itoa(ww.Status())
		path := routePattern(r)
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)
	})
}

// routePattern возвращает Chi route pattern (например "/api/prompts/{id}")
// если запрос совпал с зарегистрированным route, иначе "not_found".
// Без нормализации каждый user-ID в URL создавал бы новую timeseries —
// метрики разрослись бы в неуправляемый размер.
func routePattern(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		if p := rctx.RoutePattern(); p != "" {
			return p
		}
	}
	return "not_found"
}
