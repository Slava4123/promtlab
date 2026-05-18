package analytics

import "context"

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

// Recompute — реализация InsightsRecomputer через ComputeInsights.
// Тонкая обёртка, чтобы handler'ы могли держать узкий interface вместо
// прямой ссылки на *Service.
func (s *Service) Recompute(ctx context.Context, userID uint, teamID *uint, types []string) error {
	return s.ComputeInsights(ctx, userID, teamID, types)
}
