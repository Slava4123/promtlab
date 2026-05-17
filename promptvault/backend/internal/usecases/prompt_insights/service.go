package prompt_insights

import (
	"context"
	"errors"
	"slices"
	"time"

	repo "promptvault/internal/interface/repository"

	"gorm.io/gorm"
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
	if slices.Contains(s.plans.InsightsForPlan(planID), insightType) {
		return nil
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

const duplicateSimilarityThreshold = 0.85

// ListDuplicates — пары похожих по pg_trgm промптов. threshold = 0.85 (consistent
// с InsightsCompute из analytics service).
func (s *Service) ListDuplicates(ctx context.Context, userID uint, teamID *uint, limit int) ([]DuplicatePair, error) {
	if err := s.checkAllowed(ctx, userID, insightTypeDuplicates); err != nil {
		return nil, err
	}
	limit = clampLimit(limit, 20, 50)
	raws, err := s.analytics.PossibleDuplicates(ctx, userID, teamID, duplicateSimilarityThreshold, limit)
	if err != nil {
		return nil, err
	}
	out := make([]DuplicatePair, 0, len(raws))
	for _, r := range raws {
		out = append(out, DuplicatePair{
			PromptA:    PromptInsightRow{PromptID: r.PromptAID, Title: r.PromptATitle},
			PromptB:    PromptInsightRow{PromptID: r.PromptBID, Title: r.PromptBTitle},
			Similarity: float64(r.Similarity),
		})
	}
	return out, nil
}

// ListTrending — промпты с ростом использования >2× за неделю (factor=2.0, growing=true).
func (s *Service) ListTrending(ctx context.Context, userID uint, teamID *uint, limit int) ([]PromptInsightRow, error) {
	return s.listTrendDirection(ctx, userID, teamID, insightTypeTrending, 2.0, true, limit)
}

// ListDeclining — промпты с падением >2× (factor=0.5 = «текущее ≤ половины предыдущего»).
func (s *Service) ListDeclining(ctx context.Context, userID uint, teamID *uint, limit int) ([]PromptInsightRow, error) {
	return s.listTrendDirection(ctx, userID, teamID, insightTypeDeclining, 0.5, false, limit)
}

func (s *Service) listTrendDirection(ctx context.Context, userID uint, teamID *uint, kind string, factor float64, growing bool, limit int) ([]PromptInsightRow, error) {
	if err := s.checkAllowed(ctx, userID, kind); err != nil {
		return nil, err
	}
	limit = clampLimit(limit, 10, 50)
	raws, err := s.analytics.GetTrendingPrompts(ctx, userID, teamID, factor, growing, limit)
	if err != nil {
		return nil, err
	}
	out := make([]PromptInsightRow, 0, len(raws))
	for _, r := range raws {
		out = append(out, PromptInsightRow{PromptID: r.PromptID, Title: r.Title, Uses: int(r.UsesLast)})
	}
	return out, nil
}

// ListMostEdited — промпты с >=2 версиями (HAVING COUNT > 1 в SQL).
func (s *Service) ListMostEdited(ctx context.Context, userID uint, teamID *uint, limit int) ([]PromptInsightRow, error) {
	if err := s.checkAllowed(ctx, userID, insightTypeMostEdited); err != nil {
		return nil, err
	}
	limit = clampLimit(limit, 10, 50)
	raws, err := s.analytics.MostEditedPrompts(ctx, userID, teamID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]PromptInsightRow, 0, len(raws))
	for _, r := range raws {
		out = append(out, PromptInsightRow{PromptID: r.PromptID, Title: r.Title, Uses: int(r.Uses)})
	}
	return out, nil
}

// MergePrompts soft-delete'ит mergeID, сохраняя keepID. Проверка ownership
// и same-id case: handler не нуждается в gorm-knowledge — мы транслируем
// gorm.ErrRecordNotFound в ErrPromptsNotOwned.
func (s *Service) MergePrompts(ctx context.Context, userID, keepID, mergeID uint) error {
	if keepID == mergeID {
		return ErrSamePrompt
	}
	err := s.prompts.MergeWith(ctx, keepID, mergeID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrPromptsNotOwned
	}
	return err
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
