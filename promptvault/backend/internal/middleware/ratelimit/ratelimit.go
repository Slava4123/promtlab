package ratelimit

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Limiter is a generic sliding window rate limiter keyed by any comparable type.
type Limiter[K comparable] struct {
	mu      sync.Mutex
	window  time.Duration
	limit   int
	entries map[K][]time.Time
	stopCh  chan struct{}
}

// NewLimiter creates a new sliding window rate limiter with the given requests-per-minute limit.
func NewLimiter[K comparable](rpm int) *Limiter[K] {
	return NewLimiterWithWindow[K](rpm, time.Minute)
}

// NewLimiterWithWindow creates a sliding window rate limiter with a custom window duration.
func NewLimiterWithWindow[K comparable](limit int, window time.Duration) *Limiter[K] {
	l := &Limiter[K]{
		window:  window,
		limit:   limit,
		entries: make(map[K][]time.Time),
		stopCh:  make(chan struct{}),
	}
	go l.evictLoop()
	return l
}

// Close stops the background eviction goroutine.
func (l *Limiter[K]) Close() {
	close(l.stopCh)
}

// evictLoop удаляет устаревшие записи каждые 5 минут.
func (l *Limiter[K]) evictLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			cutoff := time.Now().Add(-l.window)
			for key, timestamps := range l.entries {
				valid := timestamps[:0]
				for _, t := range timestamps {
					if t.After(cutoff) {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(l.entries, key)
				} else {
					l.entries[key] = valid
				}
			}
			l.mu.Unlock()
		case <-l.stopCh:
			return
		}
	}
}

// Allow checks if the given key can make a request and records the usage atomically.
func (l *Limiter[K]) Allow(key K) bool {
	if l.limit <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	timestamps := l.entries[key]
	valid := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	if len(valid) == 0 {
		delete(l.entries, key)
	}
	if len(valid) >= l.limit {
		l.entries[key] = valid
		return false
	}
	l.entries[key] = append(valid, now)
	return true
}

// clientIP извлекает реальный IP клиента.
// Берёт первый IP из X-Forwarded-For (установленный ближайшим прокси),
// иначе использует RemoteAddr.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		// Первый IP в цепочке — реальный клиент (если прокси доверенный)
		ip := strings.TrimSpace(strings.SplitN(fwd, ",", 2)[0])
		if ip != "" {
			return ip
		}
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ByIP ограничивает по IP. rpm — макс запросов в минуту.
func ByIP(rpm int) func(http.Handler) http.Handler {
	rl := NewLimiter[string](rpm)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if !rl.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				if _, err := w.Write([]byte(`{"error":"Слишком много запросов. Попробуйте через минуту"}`)); err != nil {
					slog.Error("failed to write rate limit response", "error", err)
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
