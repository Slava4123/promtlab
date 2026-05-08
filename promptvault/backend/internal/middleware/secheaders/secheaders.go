// Package secheaders — middleware для добавления security-relevant
// HTTP response headers ко всем API-ответам.
//
// MJ-12 в REVIEW_2026-05-07.md: до этого fix'а API возвращал ответы без
// X-Frame-Options (clickjacking), X-Content-Type-Options (MIME-sniffing),
// Referrer-Policy, Strict-Transport-Security. SEO-эндпоинт `/p/{slug}`
// рендерит HTML — clickjacking-friendly. Admin /admin/* в SPA без XFO
// позволяет UI redress атаки.
//
// CSP — отдельная задача (требует frontend nonce strategy).
package secheaders

import "net/http"

// Options — настройки middleware. HSTS включается только в prod (для
// localhost dev работа на http без HTTPS).
type Options struct {
	// HSTS — добавлять Strict-Transport-Security. Только для prod (HTTPS).
	HSTS bool
	// HSTSMaxAge — max-age значение (default 1 год).
	HSTSMaxAge int
	// HSTSIncludeSubdomains — добавлять includeSubDomains.
	HSTSIncludeSubdomains bool
}

// New создаёт middleware с указанными опциями.
//
// Headers, которые ставятся всегда:
//   - X-Content-Type-Options: nosniff      — IE/Chrome не угадывает MIME.
//   - X-Frame-Options: DENY                — нельзя iframe-ить страницу.
//   - Referrer-Policy: strict-origin-when-cross-origin — не утекаем path
//     при кросс-доменных запросах.
//
// HSTS — опционально, только в prod через config.Server.IsProd().
func New(opts Options) func(http.Handler) http.Handler {
	maxAge := opts.HSTSMaxAge
	if maxAge == 0 {
		maxAge = 31536000 // 1 год — рекомендация OWASP
	}
	hstsValue := "max-age=" + itoa(maxAge)
	if opts.HSTSIncludeSubdomains {
		hstsValue += "; includeSubDomains"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			if opts.HSTS {
				h.Set("Strict-Transport-Security", hstsValue)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// itoa — мини-replacement для strconv.Itoa без impорта strconv в hot-path.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
