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
//
// MJ-8: trustProxy gating. Раньше extractIP всегда читал X-Forwarded-For,
// что позволяло admin'у подделывать IP в audit_log (`X-Forwarded-For:
// 8.8.8.8` → forensics-доказательство в логах фейковое). Теперь XFF/X-Real-IP
// читаются только при trustProxy=true (за nginx/CF в prod). В dev/без прокси
// trustProxy=false → IP берётся из r.RemoteAddr (нельзя подделать).
func AdminAuditContext(trustProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := audit.AdminRequestInfo{
				AdminID:   authmw.GetUserID(r.Context()),
				IP:        extractIP(r, trustProxy),
				UserAgent: r.UserAgent(),
			}
			ctx := audit.WithContext(r.Context(), info)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractIP — извлекает client IP. При trustProxy=false принципиально
// игнорирует X-Forwarded-For/X-Real-IP (защита от спуфинга в audit_log).
// При trustProxy=true берёт первый IP из XFF (самый дальний клиент за
// reverse-proxy). Не валидируем prefix — это задача nginx.
func extractIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
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
	}
	// RemoteAddr формата "host:port" — отрежем port.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
