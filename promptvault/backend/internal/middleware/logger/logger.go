package logger

import (
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

// Middleware logs each HTTP request with method, path, status code, duration
// и trace_id (если OpenTelemetry активен) — для cross-correlation между
// Loki logs и Tempo traces в Grafana.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		fields := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"bytes", ww.BytesWritten(),
		}
		// trace_id propagated через otelchi middleware (Phase 16 Этап 3).
		// IsValid()=false если SDK no-op (Telemetry.Enabled=false).
		if traceID := trace.SpanFromContext(r.Context()).SpanContext().TraceID(); traceID.IsValid() {
			fields = append(fields, "trace_id", traceID.String())
		}
		slog.Info("http", fields...)
	})
}
