// Package safeloop предоставляет общий helper для оборачивания фоновых
// background loops и fire-and-forget горутин в `defer recover()`.
//
// Мотивация — CR-6 в REVIEW_2026-05-07.md: 8 loops (insights/cleanup/
// renewal/expiration/reminder/purge/reengagement/streak_reminder) +
// несколько fire-and-forget горутин в проекте имели идентичный паттерн
// без recover():
//
//	for {
//	    select {
//	    case <-ticker.C:
//	        l.compute()    // panic тут убивает loop НАВСЕГДА (до restart)
//	    case <-l.stopCh:
//	        return
//	    }
//	}
//
// Один panic в template-render, JSON-corrupt step_outputs или nil-deref
// убивал, например, RenewalLoop → подписки T-Bank не продлевались, потеря
// revenue без алерта (loop сам себя не пишет в логи). А goroutine'ы
// без recover (oauth.touchLastLogin, team.invite_email) роняли весь
// сервер — Effective Go: «A panic in any goroutine kills the whole
// program; recover() in a defer at the top of the goroutine ensures a
// panicking goroutine cannot take the rest of the server down.»
//
// Использование:
//
//	for {
//	    select {
//	    case <-ticker.C:
//	        safeloop.RunWithRecover("insights_compute", l.compute)
//	    case <-l.stopCh:
//	        return
//	    }
//	}
//
// или
//
//	go safeloop.RunWithRecover("oauth_touch_last_login", func() {
//	    s.users.TouchLastLogin(ctx, userID)
//	})
package safeloop

import (
	"log/slog"
	"runtime/debug"

	"promptvault/internal/infrastructure/metrics"
)

// RunWithRecover вызывает fn под `defer recover()`. При panic'е логирует
// событие `loop.panic` с именем loop'а и stack trace, инкрементит
// Prometheus counter promptvault_loop_panics_total{loop=name}.
//
// Не возвращает значение — для loop'ов это не нужно (следующий тик пойдёт
// независимо). Для fire-and-forget паттернов возврат тоже не используется.
func RunWithRecover(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("loop.panic",
				"loop", name,
				"recover", r,
				"stack", string(debug.Stack()))
			metrics.LoopPanicsTotal.WithLabelValues(name).Inc()
		}
	}()
	fn()
}
