package analytics

import (
	"context"
	"encoding/json"
	"log/slog"

	"promptvault/internal/models"
)

// ComputeInsights — полный пересчёт детерминистических инсайтов для юзера
// (Max-only). Идемпотентно: UpsertInsight перезапишет существующие записи.
//
// Вызывается:
//   - ежесуточно из InsightsComputeLoop (только для Max-юзеров);
//   - on-demand из HTTP handler /api/analytics/insights/refresh (опционально).
//
// teamID == nil → личный scope, teamID != nil → команда.
//
// M3: repo-fail в каждом расчёте логируется отдельно. Функция возвращает
// nil даже если часть insights fail'нула — это детерминистическое поведение
// идемпотентного пересчёта (следующая итерация попробует снова).
func (s *Service) ComputeInsights(ctx context.Context, userID uint, teamID *uint) error {
	now := s.nowFn()

	// 1. UNUSED PROMPTS — не использовались 30+ дней.
	unused, err := s.analytics.UnusedPrompts(ctx, userID, teamID, now.AddDate(0, 0, -30), 20)
	if err != nil {
		slog.WarnContext(ctx, "analytics.insights.unused_failed",
			"err", err, "user_id", userID, "team_id", teamID)
	} else if len(unused) > 0 {
		s.upsertSafe(ctx, userID, teamID, models.InsightUnusedPrompts, unused)
	}

	// 2. TRENDING — uses(last 7d) > 2× uses(prev 7d).
	// SQL CTE в одном запросе — избегаем 2× TopPrompts + in-memory map.
	trending, err := s.analytics.GetTrendingPrompts(ctx, userID, teamID, 2.0, true, 5)
	if err != nil {
		slog.WarnContext(ctx, "analytics.insights.trending_failed",
			"err", err, "user_id", userID, "team_id", teamID)
	} else if len(trending) > 0 {
		s.upsertSafe(ctx, userID, teamID, models.InsightTrending, trending)
	}

	// 3. DECLINING — uses(last 7d) < 0.5× uses(prev 7d).
	declining, err := s.analytics.GetTrendingPrompts(ctx, userID, teamID, 0.5, false, 5)
	if err != nil {
		slog.WarnContext(ctx, "analytics.insights.declining_failed",
			"err", err, "user_id", userID, "team_id", teamID)
	} else if len(declining) > 0 {
		s.upsertSafe(ctx, userID, teamID, models.InsightDeclining, declining)
	}

	// 4-7. MOST EDITED / POSSIBLE DUPLICATES / ORPHAN TAGS / EMPTY COLLECTIONS.
	// experimentalInsights — kill-switch (default true с Phase 15).
	// possible_duplicates дополнительно требует pg_trgm — если расширение
	// недоступно (managed PG без прав), тип пропускается, остальные работают.
	if s.experimentalInsights {
		// 4. MOST EDITED — по количеству версий.
		edited, err := s.analytics.MostEditedPrompts(ctx, userID, teamID, 5)
		if err != nil {
			slog.WarnContext(ctx, "analytics.insights.most_edited_failed",
				"err", err, "user_id", userID, "team_id", teamID)
		} else if len(edited) > 0 {
			s.upsertSafe(ctx, userID, teamID, models.InsightMostEdited, edited)
		}

		// 5. POSSIBLE DUPLICATES — только если pg_trgm доступен.
		if s.trgmAvailable {
			dups, err := s.analytics.PossibleDuplicates(ctx, userID, teamID, 0.8, 10)
			if err != nil {
				slog.WarnContext(ctx, "analytics.insights.duplicates_failed",
					"err", err, "user_id", userID, "team_id", teamID)
			} else if len(dups) > 0 {
				s.upsertSafe(ctx, userID, teamID, models.InsightPossibleDuplicates, dups)
			}
		} else {
			slog.DebugContext(ctx, "analytics.insights.duplicates_skipped",
				"reason", "pg_trgm_unavailable", "user_id", userID)
		}

		// 6. ORPHAN TAGS — теги без промптов.
		orphans, err := s.analytics.OrphanTags(ctx, userID, teamID, 10)
		if err != nil {
			slog.WarnContext(ctx, "analytics.insights.orphan_tags_failed",
				"err", err, "user_id", userID, "team_id", teamID)
		} else if len(orphans) > 0 {
			s.upsertSafe(ctx, userID, teamID, models.InsightOrphanTags, orphans)
		}

		// 7. EMPTY COLLECTIONS — коллекции без промптов.
		empties, err := s.analytics.EmptyCollections(ctx, userID, teamID, 10)
		if err != nil {
			slog.WarnContext(ctx, "analytics.insights.empty_collections_failed",
				"err", err, "user_id", userID, "team_id", teamID)
		} else if len(empties) > 0 {
			s.upsertSafe(ctx, userID, teamID, models.InsightEmptyCollections, empties)
		}
	}

	return nil
}

// computeTrend удалён — вся логика теперь в SQL (repo.GetTrendingPrompts).
// Возвращаемые строки — repo.TrendRow с полями uses_last_7d / uses_prev_7d.

// upsertSafe сериализует payload и вызывает UpsertInsight. Ошибки логируются,
// не возвращаются — один insight не должен ломать пересчёт остальных.
// После успешного upsert дергает notifier (по умолчанию Noop).
func (s *Service) upsertSafe(ctx context.Context, userID uint, teamID *uint, insightType string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		slog.WarnContext(ctx, "analytics.insights.marshal_failed", "err", err, "type", insightType)
		return
	}
	insight := &models.SmartInsight{
		UserID:      userID,
		TeamID:      teamID,
		InsightType: insightType,
		Payload:     data,
	}
	if err := s.analytics.UpsertInsight(ctx, insight); err != nil {
		slog.WarnContext(ctx, "analytics.insights.upsert_failed", "err", err, "type", insightType)
		return
	}
	if s.notifier != nil {
		s.notifier.OnInsightsChanged(ctx, userID, teamID, insightType, payload)
	}
}

// GetInsights — для HTTP handler /api/analytics/insights.
func (s *Service) GetInsights(ctx context.Context, userID uint, teamID *uint) ([]models.SmartInsight, error) {
	return s.analytics.GetInsights(ctx, userID, teamID)
}
