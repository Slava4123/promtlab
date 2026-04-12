package admin

import (
	"net"
	"net/http"
	"strings"

	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/usecases/audit"
)

// AdminAuditContext — middleware, который извлекает admin_id / ip / user_agent
// и кладёт в context через audit.WithContext. Usecase-слой читает эти данные
// через audit.FromContext при вызове Service.Log.
//
// Должен быть применён ПОСЛЕ authmw.Middleware и RequireAdmin в цепочке —
// он полагается на userID в ctx.
func AdminAuditContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := audit.AdminRequestInfo{
			AdminID:   authmw.GetUserID(r.Context()),
			IP:        extractIP(r),
			UserAgent: r.UserAgent(),
		}
		ctx := audit.WithContext(r.Context(), info)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractIP — извлекает client IP с учётом X-Forwarded-For (reverse proxy).
// Берёт первый IP из списка (самый дальний клиент). Fallback — r.RemoteAddr.
// Не валидируем prefix — это задача реверс-прокси.
func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// "a, b, c" — берём первый.
		if idx := strings.Index(xff, ","); idx >= 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	// RemoteAddr формата "host:port" — отрежем port.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
