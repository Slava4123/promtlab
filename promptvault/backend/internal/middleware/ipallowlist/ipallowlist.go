// Package ipallowlist предоставляет HTTP middleware для фильтрации входящих
// запросов по списку разрешённых IP/CIDR. Основное применение — защита
// webhook-эндпойнтов от спуфинга со стороны произвольных клиентов.
package ipallowlist

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
)

// New возвращает middleware, пропускающий только запросы с IP, входящих
// в allowed. Пустой список allowed делает middleware no-op (удобно для dev
// или переходного периода до получения IP-диапазонов от провайдера).
//
// trustForwarded=true включает парсинг X-Forwarded-For: первый IP в цепочке
// считается реальным клиентом. Используется когда приложение стоит за
// nginx/cloudflare. Без trustForwarded middleware проверяет r.RemoteAddr
// (то есть IP непосредственного соединения — за прокси будет IP прокси).
//
// Отказ: 403 Forbidden + slog.Warn с IP/User-Agent/path для observability.
func New(allowed []string, trustForwarded bool) func(http.Handler) http.Handler {
	nets, ips := parse(allowed)
	if len(nets) == 0 && len(ips) == 0 {
		slog.Warn("ipallowlist.disabled", "reason", "empty allowlist — всё пропускается")
		return func(next http.Handler) http.Handler { return next }
	}

	slog.Info("ipallowlist.enabled", "cidrs", len(nets), "ips", len(ips), "trust_forwarded", trustForwarded)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := extractClientIP(r, trustForwarded)
			if clientIP == nil {
				slog.Warn("ipallowlist.denied.unparsable",
					"remote_addr", r.RemoteAddr, "xff", r.Header.Get("X-Forwarded-For"),
					"path", r.URL.Path, "user_agent", r.UserAgent())
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if allowedIP(clientIP, nets, ips) {
				next.ServeHTTP(w, r)
				return
			}

			slog.Warn("ipallowlist.denied",
				"client_ip", clientIP.String(), "path", r.URL.Path, "user_agent", r.UserAgent())
			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}

// parse разбивает входные строки на CIDR-сети и одиночные IP. Невалидные
// значения логируются и пропускаются — лучше запустить сервис с частичным
// allowlist чем не запуститься совсем.
func parse(allowed []string) ([]*net.IPNet, []net.IP) {
	var nets []*net.IPNet
	var ips []net.IP
	for _, raw := range allowed {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		if strings.Contains(s, "/") {
			if _, ipNet, err := net.ParseCIDR(s); err == nil {
				nets = append(nets, ipNet)
				continue
			}
			slog.Warn("ipallowlist.invalid_cidr", "value", s)
			continue
		}
		if ip := net.ParseIP(s); ip != nil {
			ips = append(ips, ip)
			continue
		}
		slog.Warn("ipallowlist.invalid_ip", "value", s)
	}
	return nets, ips
}

// extractClientIP извлекает IP клиента. При trustForwarded=true использует
// первый IP из X-Forwarded-For (реальный клиент перед цепочкой прокси).
// Fallback — r.RemoteAddr без порта.
func extractClientIP(r *http.Request, trustForwarded bool) net.IP {
	if trustForwarded {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// XFF формат: "client, proxy1, proxy2" — берём первый.
			first := strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
			if ip := net.ParseIP(first); ip != nil {
				return ip
			}
		}
	}
	// r.RemoteAddr содержит порт: "1.2.3.4:5678" → извлекаем IP.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr может быть без порта в тестах — пробуем парсить целиком.
		host = r.RemoteAddr
	}
	return net.ParseIP(host)
}

// allowedIP проверяет попадание IP в одну из разрешённых сетей или совпадение
// с одиночным IP. Сравнение работает с IPv4 и IPv6 — net.IP.Equal нормализует
// представление (4-байт vs 16-байт mapped).
func allowedIP(ip net.IP, nets []*net.IPNet, ips []net.IP) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	for _, allowed := range ips {
		if allowed.Equal(ip) {
			return true
		}
	}
	return false
}
