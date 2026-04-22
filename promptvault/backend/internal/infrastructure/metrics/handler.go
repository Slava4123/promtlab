package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler возвращает http.Handler для GET /metrics. За feature-flag
// METRICS_ENABLED в config: если false → 404.
//
// Защита пути: IP-allowlist применяется на уровне app.MountRoutes (nginx
// внутренний ingress). Здесь только exposition.
func Handler(enabled bool) http.Handler {
	if !enabled {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.NotFound(w, nil)
		})
	}
	return promhttp.Handler()
}
