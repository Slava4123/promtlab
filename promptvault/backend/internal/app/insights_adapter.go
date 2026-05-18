package app

import (
	"context"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/usecases/analytics"
)

// promptInsightsPlanGate адаптирует analytics.Service + users.Repository к
// PlanGate-интерфейсу из prompt_insights пакета. Изолирует prompt_insights
// от прямой зависимости на analytics service: usecase знает только узкий
// контракт (InsightsForPlan + LookupPlanID), а реализация склеена здесь
// на app-уровне.
//
// Почему адаптер, а не прямой импорт: usecases→usecases циклы — табу в Clean
// Architecture. Один и тот же тариф-маппер живёт в analytics (Phase 14 + ADR-0008),
// и переносить его в общий слой ради одного нового usecase не оправдано.
type promptInsightsPlanGate struct {
	analytics *analytics.Service
	users     repo.UserRepository
}

func (g promptInsightsPlanGate) InsightsForPlan(planID string) []string {
	return g.analytics.InsightsForPlan(planID)
}

func (g promptInsightsPlanGate) LookupPlanID(ctx context.Context, userID uint) (string, error) {
	u, err := g.users.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return u.PlanID, nil
}
