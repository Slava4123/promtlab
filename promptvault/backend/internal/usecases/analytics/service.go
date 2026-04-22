package analytics

import (
	"context"
	"errors"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	quotauc "promptvault/internal/usecases/quota"
	"promptvault/internal/usecases/subscription"
	"promptvault/internal/usecases/teamcheck"
)

// Service — агрегации для dashboard-страниц и per-prompt analytics.
//
// Retention (Free 7 / Pro 90 / Max 365) enforce'ится внутри: requestedRange
// clamp'ится по тарифу юзера. Quota summary подмешивается в PersonalDashboard
// через QuotaService, если задан.
type Service struct {
	analytics repo.AnalyticsRepository
	prompts   repo.PromptRepository
	teams     repo.TeamRepository
	users     repo.UserRepository
	quotas    *quotauc.Service
	// nowFn — для тестируемости времени. Default time.Now.
	nowFn func() time.Time
	// experimentalInsights включает 4 неготовых Smart Insight типа (Q2).
	// Default false, toggle через Analytics.ExperimentalInsights.
	experimentalInsights bool
}

func NewService(
	analytics repo.AnalyticsRepository,
	prompts repo.PromptRepository,
	teams repo.TeamRepository,
	users repo.UserRepository,
	quotas *quotauc.Service,
) *Service {
	return &Service{
		analytics: analytics,
		prompts:   prompts,
		teams:     teams,
		users:     users,
		quotas:    quotas,
		nowFn:     time.Now,
	}
}

// SetNowFn переопределяет now (для unit-тестов).
func (s *Service) SetNowFn(fn func() time.Time) { s.nowFn = fn }

// SetExperimentalInsights включает/выключает расчёт 4 заглушечных
// Smart Insight типов (Q2). Вызывается из app.go на основе config.
func (s *Service) SetExperimentalInsights(v bool) { s.experimentalInsights = v }

// GetInsightsGated — проверка плана + чтение insights. Free/Pro получают
// ErrMaxRequired. Логика плана вынесена из handler'а в service (H5).
func (s *Service) GetInsightsGated(ctx context.Context, userID uint, teamID *uint) ([]models.SmartInsight, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !subscription.IsMax(user.PlanID) {
		return nil, ErrMaxRequired
	}
	return s.analytics.GetInsights(ctx, userID, teamID)
}

// ExportGate возвращает nil если юзеру доступен export (Pro+), иначе
// ErrProRequired. Handler вызывает перед streaming'ом CSV (H5).
func (s *Service) ExportGate(ctx context.Context, userID uint) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !subscription.IsPaid(user.PlanID) {
		return ErrProRequired
	}
	return nil
}

// GetPersonalDashboard — personal scope (team_id IS NULL).
func (s *Service) GetPersonalDashboard(ctx context.Context, userID uint, requestedRange RangeID) (*PersonalDashboard, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	rng := ClampRange(requestedRange, user.PlanID)
	dr := BuildDateRange(rng, s.nowFn())

	dashboard := &PersonalDashboard{Range: rng}
	if dashboard.UsagePerDay, err = s.analytics.UsagePerDay(ctx, userID, nil, dr); err != nil {
		return nil, err
	}
	if dashboard.TopPrompts, err = s.analytics.TopPrompts(ctx, userID, nil, dr, 10); err != nil {
		return nil, err
	}
	if dashboard.PromptsCreated, err = s.analytics.PromptsCreatedPerDay(ctx, userID, nil, dr); err != nil {
		return nil, err
	}
	if dashboard.PromptsUpdated, err = s.analytics.PromptsUpdatedPerDay(ctx, userID, nil, dr); err != nil {
		return nil, err
	}
	if dashboard.ShareViews, err = s.analytics.ShareViewsPerDay(ctx, userID, dr); err != nil {
		return nil, err
	}
	if dashboard.TopShared, err = s.analytics.TopSharedPrompts(ctx, userID, dr, 10); err != nil {
		return nil, err
	}
	if s.quotas != nil {
		summary, qerr := s.quotas.GetUsageSummary(ctx, userID)
		if qerr == nil {
			dashboard.Quotas = summary
		}
	}
	return dashboard, nil
}

// GetTeamDashboard — team scope. Проверяет membership (viewer+ достаточно для чтения).
func (s *Service) GetTeamDashboard(ctx context.Context, userID, teamID uint, requestedRange RangeID) (*TeamDashboard, error) {
	// Чтобы читать team analytics — достаточно быть членом команды (любая роль).
	tid := teamID
	if err := teamcheck.RequireMembership(ctx, s.teams, []uint{teamID}, userID); err != nil {
		if errors.Is(err, teamcheck.ErrForbidden) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	rng := ClampRange(requestedRange, user.PlanID)
	dr := BuildDateRange(rng, s.nowFn())

	dashboard := &TeamDashboard{Range: rng}
	if dashboard.UsagePerDay, err = s.analytics.UsagePerDay(ctx, userID, &tid, dr); err != nil {
		return nil, err
	}
	if dashboard.TopPrompts, err = s.analytics.TopPrompts(ctx, userID, &tid, dr, 10); err != nil {
		return nil, err
	}
	if dashboard.PromptsCreated, err = s.analytics.PromptsCreatedPerDay(ctx, userID, &tid, dr); err != nil {
		return nil, err
	}
	if dashboard.PromptsUpdated, err = s.analytics.PromptsUpdatedPerDay(ctx, userID, &tid, dr); err != nil {
		return nil, err
	}
	if dashboard.Contributors, err = s.analytics.Contributors(ctx, tid, dr, 10); err != nil {
		return nil, err
	}
	return dashboard, nil
}

// GetPromptAnalytics — per-prompt страница /api/analytics/prompts/:id.
// Проверка доступа: юзер владелец или член team промпта.
func (s *Service) GetPromptAnalytics(ctx context.Context, promptID, userID uint) (*PromptAnalytics, error) {
	prompt, err := s.prompts.GetByID(ctx, promptID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	// Личный промпт — только владелец. Командный — любой член.
	if prompt.TeamID == nil {
		if prompt.UserID != userID {
			return nil, ErrForbidden
		}
	} else {
		if _, err := s.teams.GetMember(ctx, *prompt.TeamID, userID); err != nil {
			return nil, ErrForbidden
		}
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	rng := ClampRange(Range90d, user.PlanID) // prompt page — фиксированное окно с clamp
	dr := BuildDateRange(rng, s.nowFn())

	result := &PromptAnalytics{PromptID: promptID}

	// Per-prompt usage: reuse TopPrompts семантики через фильтр? Репа даёт
	// только top-N, но не per-prompt timeline. Используем UsagePerDay с
	// teamID из промпта и фильтруем при чтении (для MVP — можно через
	// Raw SQL в репе, сейчас пропускаем и даём общий UsagePerDay).
	var scope *uint
	if prompt.TeamID != nil {
		scope = prompt.TeamID
	}
	if result.UsagePerDay, err = s.analytics.UsagePerDay(ctx, prompt.UserID, scope, dr); err != nil {
		return nil, err
	}
	// Timeline share-просмотров владельца — фильтр именно по этой ссылке не
	// реализован; отдаём общий ShareViewsPerDay. Уточним в AnalyticsRepository
	// позже (метод PerShareLinkViews) если понадобится point-accurate.
	if result.ShareViewsPerDay, err = s.analytics.ShareViewsPerDay(ctx, prompt.UserID, dr); err != nil {
		return nil, err
	}
	return result, nil
}
