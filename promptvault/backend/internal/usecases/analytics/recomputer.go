package analytics

import (
	"context"
	"log/slog"
)

// slogWarnIgnored — однострочник для error-логирования без падения.
// Recompute допускает частичные сбои (DeleteInsight упал — Compute попробует UPSERT).
func slogWarnIgnored(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

// InsightsRecomputer — DI seam для mutations, которые меняют состояние кэша
// Smart Insights. handler делает delete tag → вызывает Recompute с конкретными
// типами. *Service реализует этот интерфейс через ComputeInsights — без
// циклической зависимости usecases↔handlers.
//
// Использование (см. tag/collection/prompt handlers, insights_handler):
//
//	if h.insights != nil {
//	    _ = h.insights.Recompute(ctx, userID, nil, []string{"orphan_tags"})
//	}
//
// Все вызовы — с teamID=nil (personal scope). Team-scoped insights
// пересчитываются на cron loop (out of scope для inline hot-refresh).
// Ошибки swallow'аются silently (handler не должен падать из-за recompute).
type InsightsRecomputer interface {
	Recompute(ctx context.Context, userID uint, teamID *uint, types []string) error
}

// Recompute — реализация InsightsRecomputer для hot-refresh после mutations.
//
// В отличие от cron'ного ComputeInsights (который только UPSERT'ит non-empty
// результат — by design, чтобы редкие зеро-значения не перезатирали полезные
// кэшированные данные), Recompute сначала DELETE'ит row для каждого type,
// потом запускает ComputeInsights. Это даёт корректный hot-refresh: если
// после mutation в DB не осталось orphan-тегов — карточка действительно
// исчезает (а не висит со stale count).
func (s *Service) Recompute(ctx context.Context, userID uint, teamID *uint, types []string) error {
	for _, t := range types {
		if err := s.analytics.DeleteInsight(ctx, userID, teamID, t); err != nil {
			// Лог, но не падаем — ComputeInsights ниже всё равно перепишет
			// активные row через UPSERT. Худший сценарий — старый stale остаётся.
			slogWarnIgnored(ctx, "analytics.recompute.delete_failed",
				"err", err, "user_id", userID, "type", t)
		}
	}
	return s.ComputeInsights(ctx, userID, teamID, types)
}
