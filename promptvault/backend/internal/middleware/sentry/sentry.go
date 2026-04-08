// Package sentry содержит middleware для интеграции Chi router с Sentry SDK.
//
// Handler оборачивает запросы в sentry Hub scope, автоматически ловит panics и
// добавляет request context к events.
//
// UserContext вешает sentry.User{ID} на текущий Hub для authenticated запросов
// (используется после auth middleware для атрибуции ошибок к юзерам).
//
// Оба middleware — no-op если sentry.Init не был вызван (SDK возвращает nil
// Hub), поэтому можно безопасно подключать даже когда SENTRY_ENABLED=false.
package sentry

import (
	"net/http"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"

	authmw "promptvault/internal/middleware/auth"
)

// Handler возвращает middleware, которое создаёт Sentry Hub per-request и
// автоматически перехватывает panics в handler chain.
//
// Repanic: true — после capture события, panic пробрасывается дальше, чтобы
// chimw.Recoverer мог вернуть 500 response клиенту. Это значит sentry Handler
// должен идти ПЕРЕД chimw.Recoverer в middleware chain.
//
// Timeout: 3s — максимум ждать flush при завершении request. Если GlitchTip
// недоступен, SDK не блокирует request handling.
func Handler() func(http.Handler) http.Handler {
	sh := sentryhttp.New(sentryhttp.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         3 * time.Second,
	})
	return sh.Handle
}

// UserContext устанавливает sentry.User{ID} на Hub текущего request'а из
// JWT claims, которые auth middleware положил в context. Вызывается ПОСЛЕ
// authmw.Middleware в protected группах роутов.
//
// Middleware no-op, если:
//   - Hub отсутствует в context (sentry.Init не был вызван);
//   - UserID отсутствует в context (публичный endpoint, не прошедший auth).
func UserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub := sentry.GetHubFromContext(r.Context())
		if hub == nil {
			next.ServeHTTP(w, r)
			return
		}

		userID := authmw.GetUserID(r.Context())
		if userID == 0 {
			next.ServeHTTP(w, r)
			return
		}

		hub.Scope().SetUser(sentry.User{
			ID: strconv.FormatUint(uint64(userID), 10),
		})

		next.ServeHTTP(w, r)
	})
}
