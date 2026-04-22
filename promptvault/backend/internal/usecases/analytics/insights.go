package analytics

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	repo "promptvault/internal/interface/repository"
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
	trending, err := s.computeTrend(ctx, userID, teamID, now, 2.0, true, 5)
	if err != nil {
		slog.WarnContext(ctx, "analytics.insights.trending_failed",
			"err", err, "user_id", userID, "team_id", teamID)
	} else if len(trending) > 0 {
		s.upsertSafe(ctx, userID, teamID, models.InsightTrending, trending)
	}

	// 3. DECLINING — uses(last 7d) < 0.5× uses(prev 7d).
	declining, err := s.computeTrend(ctx, userID, teamID, now, 0.5, false, 5)
	if err != nil {
		slog.WarnContext(ctx, "analytics.insights.declining_failed",
			"err", err, "user_id", userID, "team_id", teamID)
	} else if len(declining) > 0 {
		s.upsertSafe(ctx, userID, teamID, models.InsightDeclining, declining)
	}

	// 4. MOST EDITED, POSSIBLE DUPLICATES, ORPHAN TAGS, EMPTY COLLECTIONS —
	//    оставлены заглушками под тикет M8 (Levenshtein через pg_trgm для
	//    duplicates, SQL-агрегации для остальных). Скрыты за
	//    Analytics.ExperimentalInsights feature-flag (Q2) — default false.
	if s.experimentalInsights {
		// Реализация появится в M8.
		_ = now
	}

	return nil
}

// computeTrend возвращает промпты с изменением usage между двумя 7-дневными
// окнами. growing=true → рост >= factor; growing=false → падение <= factor.
// Возвращаемые строки используют Uses как uses(last 7d).
func (s *Service) computeTrend(ctx context.Context, userID uint, teamID *uint, now time.Time, factor float64, growing bool, limit int) ([]trendRow, error) {
	last7 := repo.DateRange{From: now.AddDate(0, 0, -7), To: now}
	prev7 := repo.DateRange{From: now.AddDate(0, 0, -14), To: now.AddDate(0, 0, -7)}

	recent, err := s.analytics.TopPrompts(ctx, userID, teamID, last7, 50)
	if err != nil {
		return nil, err
	}
	previous, err := s.analytics.TopPrompts(ctx, userID, teamID, prev7, 100)
	if err != nil {
		return nil, err
	}

	prevMap := make(map[uint]int64, len(previous))
	for _, p := range previous {
		prevMap[p.PromptID] = p.Uses
	}

	var out []trendRow
	for _, p := range recent {
		prev := prevMap[p.PromptID]
		if growing {
			// Для растущих — учитываем даже если prev==0 (новый тренд).
			if prev == 0 || float64(p.Uses) >= float64(prev)*factor {
				out = append(out, trendRow{
					PromptID:     p.PromptID,
					Title:        p.Title,
					Uses:         p.Uses,
					PreviousUses: prev,
				})
			}
		} else {
			// Для падающих — только если prev был значимый.
			if prev > 0 && float64(p.Uses) <= float64(prev)*factor {
				out = append(out, trendRow{
					PromptID:     p.PromptID,
					Title:        p.Title,
					Uses:         p.Uses,
					PreviousUses: prev,
				})
			}
		}
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

type trendRow struct {
	PromptID     uint   `json:"prompt_id"`
	Title        string `json:"title"`
	Uses         int64  `json:"uses_last_7d"`
	PreviousUses int64  `json:"uses_prev_7d"`
}

// upsertSafe сериализует payload и вызывает UpsertInsight. Ошибки логируются,
// не возвращаются — один insight не должен ломать пересчёт остальных.
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
	}
}

// GetInsights — для HTTP handler /api/analytics/insights.
func (s *Service) GetInsights(ctx context.Context, userID uint, teamID *uint) ([]models.SmartInsight, error) {
	return s.analytics.GetInsights(ctx, userID, teamID)
}
