package prompt_insights

import (
	"context"
	"time"

	repo "promptvault/internal/interface/repository"
)

const (
	insightTypeUnused     = "unused_prompts"
	insightTypeDuplicates = "possible_duplicates"
	insightTypeTrending   = "trending"
	insightTypeDeclining  = "declining"
	insightTypeMostEdited = "most_edited"
)

// PlanGate — DI для проверки тарифа. Реализуется *analytics.Service через адаптер
// в app.go — избегаем циклической зависимости usecases→usecases.
type PlanGate interface {
	InsightsForPlan(planID string) []string
	LookupPlanID(ctx context.Context, userID uint) (string, error)
}

// PromptMerger — узкий интерфейс на repo.PromptRepository.MergeWith.
type PromptMerger interface {
	MergeWith(ctx context.Context, keepID, mergeID, userID uint) error
}

type Service struct {
	analytics repo.AnalyticsRepository
	prompts   PromptMerger
	plans     PlanGate
	nowFn     func() time.Time
}

func NewService(analytics repo.AnalyticsRepository, prompts PromptMerger, plans PlanGate, nowFn func() time.Time) *Service {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &Service{analytics: analytics, prompts: prompts, plans: plans, nowFn: nowFn}
}

func (s *Service) checkAllowed(ctx context.Context, userID uint, insightType string) error {
	planID, err := s.plans.LookupPlanID(ctx, userID)
	if err != nil {
		return err
	}
	for _, t := range s.plans.InsightsForPlan(planID) {
		if t == insightType {
			return nil
		}
	}
	return ErrProRequired
}

// ListUnused — промпты, которые не использовались >= 30 дней. limit clamp [1,100].
func (s *Service) ListUnused(ctx context.Context, userID uint, teamID *uint, limit int) ([]PromptInsightRow, error) {
	if err := s.checkAllowed(ctx, userID, insightTypeUnused); err != nil {
		return nil, err
	}
	limit = clampLimit(limit, 50, 100)
	before := s.nowFn().AddDate(0, 0, -30)
	raws, err := s.analytics.UnusedPrompts(ctx, userID, teamID, before, limit)
	if err != nil {
		return nil, err
	}
	out := make([]PromptInsightRow, 0, len(raws))
	for _, r := range raws {
		out = append(out, PromptInsightRow{PromptID: r.PromptID, Title: r.Title, Uses: int(r.Uses)})
	}
	return out, nil
}

func clampLimit(v, def, max int) int {
	if v <= 0 {
		return def
	}
	if v > max {
		return max
	}
	return v
}
