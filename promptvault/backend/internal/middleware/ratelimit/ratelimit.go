package ratelimit

import (
	"hash/fnv"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// numShards — количество шардов для снижения lock contention под высоким
// параллелизмом (P-9). 16 — разумный баланс: 16 × mutex + 16 × map overhead
// незначителен vs. серьёзное снижение contention при сотнях concurrent rps.
// Должно быть степенью 2 для дешёвого модуло через битовую маску.
const numShards = 16
const shardMask = numShards - 1

// shard — отдельная секция limiter'а со своим mutex'ом и map'ом.
type shard[K comparable] struct {
	mu      sync.Mutex
	entries map[K][]time.Time
}

// Limiter is a generic sliding window rate limiter keyed by any comparable type.
//
// P-9: внутренне sharded на numShards секций. Разные ключи с высокой
// вероятностью попадают в разные шарды → разные mutex'ы → нет contention.
// В прежней one-mutex версии при 1000 concurrent rps один долгий lookup мог
// заблокировать все остальные запросы.
type Limiter[K comparable] struct {
	shards [numShards]shard[K]
	window time.Duration
	limit  int
	hash   func(K) uint64
	stopCh chan struct{}
}

// UintHash — identity hash для uint-ключей (userID).
func UintHash(k uint) uint64 { return uint64(k) }

// StringHash — FNV-1a для string-ключей (IP).
func StringHash(s string) uint64 {
	h := fnv.New64a()
	// Write не может ошибиться на []byte.
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

// NewLimiter creates a new sliding window rate limiter with the given requests-per-minute limit.
func NewLimiter[K comparable](rpm int, hash func(K) uint64) *Limiter[K] {
	return NewLimiterWithWindow[K](rpm, time.Minute, hash)
}

// NewLimiterWithWindow creates a sliding window rate limiter with a custom window duration.
func NewLimiterWithWindow[K comparable](limit int, window time.Duration, hash func(K) uint64) *Limiter[K] {
	l := &Limiter[K]{
		window: window,
		limit:  limit,
		hash:   hash,
		stopCh: make(chan struct{}),
	}
	for i := range l.shards {
		l.shards[i].entries = make(map[K][]time.Time)
	}
	go l.evictLoop()
	return l
}

// shardFor возвращает shard для данного ключа через хеш-функцию (конструктор).
func (l *Limiter[K]) shardFor(key K) *shard[K] {
	return &l.shards[l.hash(key)&shardMask]
}

// Close stops the background eviction goroutine.
func (l *Limiter[K]) Close() {
	close(l.stopCh)
}

// evictLoop удаляет устаревшие записи каждые 5 минут. Проходит по всем шардам
// по очереди (не лочит все одновременно — чтобы в конце-концов evict не
// заблокировал весь трафик).
func (l *Limiter[K]) evictLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cutoff := time.Now().Add(-l.window)
			for i := range l.shards {
				sh := &l.shards[i]
				sh.mu.Lock()
				for key, timestamps := range sh.entries {
					valid := timestamps[:0]
					for _, t := range timestamps {
						if t.After(cutoff) {
							valid = append(valid, t)
						}
					}
					if len(valid) == 0 {
						delete(sh.entries, key)
					} else {
						sh.entries[key] = valid
					}
				}
				sh.mu.Unlock()
			}
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
	sh := l.shardFor(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	timestamps := sh.entries[key]
	valid := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	if len(valid) == 0 {
		delete(sh.entries, key)
	}
	if len(valid) >= l.limit {
		sh.entries[key] = valid
		return false
	}
	sh.entries[key] = append(valid, now)
	return true
}

// clientIP извлекает IP клиента. При trustProxy=true читает X-Forwarded-For
// (первый IP в цепочке) и X-Real-IP — только если backend за доверенным reverse-proxy,
// который затирает incoming-значения этих заголовков. Иначе клиент может подделать
// XFF и обойти rate-limit, ставя на каждый запрос новое значение.
func clientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip := strings.TrimSpace(strings.SplitN(fwd, ",", 2)[0])
			if ip != "" {
				return ip
			}
		}
		if ip := r.Header.Get("X-Real-IP"); ip != "" {
			return strings.TrimSpace(ip)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ByUserID ограничивает по ID аутентифицированного пользователя (из context).
// Должен применяться ПОСЛЕ auth middleware, который помещает userID в context.
func ByUserID(rpm int, getUserID func(r *http.Request) uint) func(http.Handler) http.Handler {
	rl := NewLimiter[uint](rpm, UintHash)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := getUserID(r)
			if userID == 0 {
				next.ServeHTTP(w, r)
				return
			}
			if !rl.Allow(userID) {
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

// ByIP ограничивает по IP. rpm — макс запросов в минуту.
// trustProxy должен быть true только за доверенным reverse-proxy (см. clientIP).
func ByIP(rpm int, trustProxy bool) func(http.Handler) http.Handler {
	rl := NewLimiter[string](rpm, StringHash)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r, trustProxy)
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
