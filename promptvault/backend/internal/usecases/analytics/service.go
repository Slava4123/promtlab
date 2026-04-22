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
	// notifier — опциональный hook на изменение insights (email/push).
	// Default NoopNotifier.
	notifier InsightsNotifier
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
		notifier:  NoopNotifier{},
	}
}

// SetNowFn переопределяет now (для unit-тестов).
func (s *Service) SetNowFn(fn func() time.Time) { s.nowFn = fn }

// SetExperimentalInsights включает/выключает расчёт 4 заглушечных
// Smart Insight типов (Q2). Вызывается из app.go на основе config.
func (s *Service) SetExperimentalInsights(v bool) { s.experimentalInsights = v }

// SetNotifier заменяет NoopNotifier на реальную реализацию (например,
// EmailInsightsNotifier из infrastructure/email). Вызывать после NewService.
func (s *Service) SetNotifier(n InsightsNotifier) {
	if n == nil {
		s.notifier = NoopNotifier{}
		return
	}
	s.notifier = n
}

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

// RefreshInsightsGated — Max-only force-пересчёт. Обычно инсайты считаются
// раз в сутки в InsightsComputeLoop; этот endpoint позволяет юзеру
// триггернуть пересчёт руками (rate-limit на уровне middleware).
func (s *Service) RefreshInsightsGated(ctx context.Context, userID uint, teamID *uint) ([]models.SmartInsight, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !subscription.IsMax(user.PlanID) {
		return nil, ErrMaxRequired
	}
	if err := s.ComputeInsights(ctx, userID, teamID); err != nil {
		return nil, err
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

// GetPersonalDashboardFiltered — personal dashboard с drill-down по тегу
// и/или коллекции. Если tagID == nil && collectionID == nil — эквивалентно
// GetPersonalDashboard. Иначе usage/top/created/updated/model метрики
// считаются через filter-aware методы AnalyticsRepository.
// ShareViews/TopShared/Quotas не подпадают под drill-down (share-ссылка
// принадлежит юзеру, не tag/collection).
func (s *Service) GetPersonalDashboardFiltered(ctx context.Context, userID uint, requestedRange RangeID, tagID, collectionID *uint) (*PersonalDashboard, error) {
	// Если нет drill-down фильтров — использовать быстрый путь (существующий
	// метод, который callers'ы (MCP, Export) и так используют).
	if tagID == nil && collectionID == nil {
		return s.GetPersonalDashboard(ctx, userID, requestedRange)
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	rng := ClampRange(requestedRange, user.PlanID)
	dr := BuildDateRange(rng, s.nowFn())
	filter := repo.AnalyticsFilter{
		UserID:       userID,
		Range:        dr,
		TagID:        tagID,
		CollectionID: collectionID,
	}

	dashboard := &PersonalDashboard{Range: rng}
	if dashboard.UsagePerDay, err = s.analytics.UsagePerDayFiltered(ctx, filter); err != nil {
		return nil, err
	}
	if dashboard.TopPrompts, err = s.analytics.TopPromptsFiltered(ctx, filter, 10); err != nil {
		return nil, err
	}
	if dashboard.PromptsCreated, err = s.analytics.PromptsCreatedPerDayFiltered(ctx, filter); err != nil {
		return nil, err
	}
	if dashboard.PromptsUpdated, err = s.analytics.PromptsUpdatedPerDayFiltered(ctx, filter); err != nil {
		return nil, err
	}
	// ShareViews не фильтруются drill-down'ом — сам факт share не привязан
	// к tag/collection напрямую. Оставляем полное значение юзера.
	if dashboard.ShareViews, err = s.analytics.ShareViewsPerDay(ctx, userID, dr); err != nil {
		return nil, err
	}
	if dashboard.TopShared, err = s.analytics.TopSharedPrompts(ctx, userID, dr, 10); err != nil {
		return nil, err
	}
	// Quotas остаются глобальными — drill-down их не меняет.
	if s.quotas != nil {
		if summary, qerr := s.quotas.GetUsageSummary(ctx, userID); qerr == nil {
			dashboard.Quotas = summary
		}
	}

	dashboard.TotalsCurrent = Totals{
		Uses:       sumPoints(dashboard.UsagePerDay),
		Created:    sumPoints(dashboard.PromptsCreated),
		Updated:    sumPoints(dashboard.PromptsUpdated),
		ShareViews: sumPoints(dashboard.ShareViews),
	}
	prevFilter := filter
	prevFilter.Range = BuildPreviousRange(rng, s.nowFn())
	if prev, perr := s.analytics.UsagePerDayFiltered(ctx, prevFilter); perr == nil {
		dashboard.TotalsPrevious.Uses = sumPoints(prev)
	}
	if prev, perr := s.analytics.PromptsCreatedPerDayFiltered(ctx, prevFilter); perr == nil {
		dashboard.TotalsPrevious.Created = sumPoints(prev)
	}
	if prev, perr := s.analytics.PromptsUpdatedPerDayFiltered(ctx, prevFilter); perr == nil {
		dashboard.TotalsPrevious.Updated = sumPoints(prev)
	}
	if prev, perr := s.analytics.ShareViewsPerDay(ctx, userID, prevFilter.Range); perr == nil {
		dashboard.TotalsPrevious.ShareViews = sumPoints(prev)
	}
	if rows, merr := s.analytics.UsageByModelFiltered(ctx, filter); merr == nil {
		dashboard.UsageByModel = rows
	}
	ensurePersonalNonNil(dashboard)
	return dashboard, nil
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

	// Totals current — sum уже полученных per-day arrays.
	dashboard.TotalsCurrent = Totals{
		Uses:       sumPoints(dashboard.UsagePerDay),
		Created:    sumPoints(dashboard.PromptsCreated),
		Updated:    sumPoints(dashboard.PromptsUpdated),
		ShareViews: sumPoints(dashboard.ShareViews),
	}

	// Totals previous — вторая серия запросов за равный период до текущего.
	// Если какой-то запрос упадёт — пишем 0 в соответствующее поле (не блокируем).
	prevDR := BuildPreviousRange(rng, s.nowFn())
	if prev, perr := s.analytics.UsagePerDay(ctx, userID, nil, prevDR); perr == nil {
		dashboard.TotalsPrevious.Uses = sumPoints(prev)
	}
	if prev, perr := s.analytics.PromptsCreatedPerDay(ctx, userID, nil, prevDR); perr == nil {
		dashboard.TotalsPrevious.Created = sumPoints(prev)
	}
	if prev, perr := s.analytics.PromptsUpdatedPerDay(ctx, userID, nil, prevDR); perr == nil {
		dashboard.TotalsPrevious.Updated = sumPoints(prev)
	}
	if prev, perr := s.analytics.ShareViewsPerDay(ctx, userID, prevDR); perr == nil {
		dashboard.TotalsPrevious.ShareViews = sumPoints(prev)
	}

	// Segmentation по модели (Phase 14.2 B.7).
	if rows, merr := s.analytics.UsageByModel(ctx, userID, nil, dr); merr == nil {
		dashboard.UsageByModel = rows
	}

	ensurePersonalNonNil(dashboard)
	return dashboard, nil
}

// ensurePersonalNonNil гарантирует что все slice-поля dashboard — пустые [],
// а не nil. GORM Scan возвращает nil для 0 строк, JSON маршалинг тогда даёт
// null — фронт не ожидает null и падает на `.reduce`/`.map`. Нормализуем
// контракт API: массивы всегда есть, пусть даже пустые.
func ensurePersonalNonNil(d *PersonalDashboard) {
	if d.UsagePerDay == nil {
		d.UsagePerDay = []repo.UsagePoint{}
	}
	if d.TopPrompts == nil {
		d.TopPrompts = []repo.PromptUsageRow{}
	}
	if d.PromptsCreated == nil {
		d.PromptsCreated = []repo.UsagePoint{}
	}
	if d.PromptsUpdated == nil {
		d.PromptsUpdated = []repo.UsagePoint{}
	}
	if d.ShareViews == nil {
		d.ShareViews = []repo.UsagePoint{}
	}
	if d.TopShared == nil {
		d.TopShared = []repo.PromptUsageRow{}
	}
	if d.UsageByModel == nil {
		d.UsageByModel = []repo.ModelUsageRow{}
	}
}

func ensureTeamNonNil(d *TeamDashboard) {
	if d.UsagePerDay == nil {
		d.UsagePerDay = []repo.UsagePoint{}
	}
	if d.TopPrompts == nil {
		d.TopPrompts = []repo.PromptUsageRow{}
	}
	if d.PromptsCreated == nil {
		d.PromptsCreated = []repo.UsagePoint{}
	}
	if d.PromptsUpdated == nil {
		d.PromptsUpdated = []repo.UsagePoint{}
	}
	if d.Contributors == nil {
		d.Contributors = []repo.ContributorRow{}
	}
	if d.UsageByModel == nil {
		d.UsageByModel = []repo.ModelUsageRow{}
	}
}

func ensurePromptNonNil(p *PromptAnalytics) {
	if p.UsagePerDay == nil {
		p.UsagePerDay = []repo.UsagePoint{}
	}
	if p.ShareViewsPerDay == nil {
		p.ShareViewsPerDay = []repo.UsagePoint{}
	}
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

	dashboard.TotalsCurrent = Totals{
		Uses:    sumPoints(dashboard.UsagePerDay),
		Created: sumPoints(dashboard.PromptsCreated),
		Updated: sumPoints(dashboard.PromptsUpdated),
	}

	prevDR := BuildPreviousRange(rng, s.nowFn())
	if prev, perr := s.analytics.UsagePerDay(ctx, userID, &tid, prevDR); perr == nil {
		dashboard.TotalsPrevious.Uses = sumPoints(prev)
	}
	if prev, perr := s.analytics.PromptsCreatedPerDay(ctx, userID, &tid, prevDR); perr == nil {
		dashboard.TotalsPrevious.Created = sumPoints(prev)
	}
	if prev, perr := s.analytics.PromptsUpdatedPerDay(ctx, userID, &tid, prevDR); perr == nil {
		dashboard.TotalsPrevious.Updated = sumPoints(prev)
	}

	if rows, merr := s.analytics.UsageByModel(ctx, userID, &tid, dr); merr == nil {
		dashboard.UsageByModel = rows
	}

	ensureTeamNonNil(dashboard)
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

	// Per-prompt usage — отдельный SQL с WHERE prompt_id = ?. Ранее использовали
	// общий UsagePerDay который считал все промпты юзера — это был баг.
	_ = prompt // scope-ссылка оставлена на случай будущих фильтров
	if result.UsagePerDay, err = s.analytics.PromptUsageTimeline(ctx, promptID, dr); err != nil {
		return nil, err
	}
	// Share-просмотры именно этого промпта: JOIN share_links по prompt_id.
	if result.ShareViewsPerDay, err = s.analytics.PromptShareViewsTimeline(ctx, promptID, dr); err != nil {
		return nil, err
	}
	ensurePromptNonNil(result)
	return result, nil
}
