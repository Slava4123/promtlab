package quota

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	iservice "promptvault/internal/interface/service"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// Service — центральный сервис проверки и инкремента квот.
// Все Check* методы возвращают *QuotaExceededError при превышении лимита
// или nil, если действие разрешено.
type Service struct {
	plans       repo.PlanRepository
	quotas      repo.QuotaRepository
	users       repo.UserRepository
	email       iservice.EmailSender
	frontendURL string
}

func NewService(plans repo.PlanRepository, quotas repo.QuotaRepository, users repo.UserRepository) *Service {
	return &Service{plans: plans, quotas: quotas, users: users}
}

// SetEmailNotifier — опциональный setter для quota-warning email (M-5c).
// Если email==nil или Configured()==false → maybeSendQuotaWarning no-op'ит.
// Отдельный метод (не в NewService) — чтобы не ломать сигнатуру тестов.
func (s *Service) SetEmailNotifier(email iservice.EmailSender, frontendURL string) {
	s.email = email
	s.frontendURL = frontendURL
}

// getPlan загружает план юзера. PlanRepository кэширован (5 мин TTL),
// UserRepository — PK lookup (микросекунды).
func (s *Service) getPlan(ctx context.Context, userID uint) (string, *models.SubscriptionPlan, error) {
	planID, _, plan, err := s.getUserAndPlan(ctx, userID)
	return planID, plan, err
}

// getUserAndPlan — расширенная версия getPlan, возвращает также user, чтобы
// caller мог обратиться к user.LegacyLimit(field) для grandfather-проверок
// (Pack E/F: некоторые юзеры зарегистрированы до изменения тарифа и должны
// сохранить старый лимит). Используется в Check методах где это релевантно
// (CheckPromptQuota, CheckExtensionQuota, CheckMCPQuota).
func (s *Service) getUserAndPlan(ctx context.Context, userID uint) (string, *models.User, *models.SubscriptionPlan, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", nil, nil, err
	}
	plan, err := s.plans.GetByID(ctx, user.PlanID)
	if err != nil {
		return "", nil, nil, err
	}
	return user.PlanID, user, plan, nil
}

// effectiveLimit возвращает максимум из текущего значения плана и
// grandfather-снапшота из users.legacy_quotas. Логика max() защищает
// существующих юзеров от понижения, но не мешает upgrade (если юзер
// перешёл на старший план с большим лимитом — берём его).
//
// Пример: юзер был Free на момент миграции 000068 (legacy max_prompts=50,
// plan теперь 15). Остаётся Free → effective=50 (legacy сохраняет старый
// лимит). Upgrade на Pro → effective=500 (plan больше legacy). Downgrade
// обратно на Free → effective=50 (legacy снова применяется).
//
// Новый юзер (legacy={}): effective = planValue.
// См. миграции 000068+ (Pack E/F).
func effectiveLimit(user *models.User, field string, planValue int) int {
	if v, ok := user.LegacyLimit(field); ok && v > planValue {
		return v
	}
	return planValue
}

// isWithinLimit — проверка used < limit. После миграции 000046 все лимиты —
// неотрицательные числа; legacy sentinel -1 "безлимит" полностью выведен.
func isWithinLimit(used int64, limit int) bool {
	return used < int64(limit)
}

// CheckPromptQuota — проверка лимита ЛИЧНЫХ промптов (team_id IS NULL).
// Командные промпты учитываются отдельно через CheckTeamPromptQuota и не
// расходуют personal-лимит юзера. См. Pack T (миграция 000070).
func (s *Service) CheckPromptQuota(ctx context.Context, userID uint) error {
	planID, user, plan, err := s.getUserAndPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountPersonalPrompts(ctx, userID)
	if err != nil {
		return err
	}
	limit := effectiveLimit(user, "max_prompts", plan.MaxPrompts)
	if !isWithinLimit(used, limit) {
		return newQuotaExceeded("prompts", planID, int(used), limit, "промптов")
	}
	return nil
}

// CheckTeamPromptQuota — проверка лимита промптов в команде. Лимит берётся
// из плана owner'а команды (план owner'а определяет «силу» всей команды).
// Pack T: позволяет Free участнику в Pro команде создавать промпты против
// общего пула, а не своего personal-лимита 15.
func (s *Service) CheckTeamPromptQuota(ctx context.Context, teamID, ownerUserID uint) error {
	planID, plan, err := s.getPlan(ctx, ownerUserID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountTeamPrompts(ctx, teamID)
	if err != nil {
		return err
	}
	if !isWithinLimit(used, plan.MaxTeamPrompts) {
		return newQuotaExceeded("team_prompts", planID, int(used), plan.MaxTeamPrompts, "промптов команды")
	}
	return nil
}

// CheckCollectionQuota — проверка лимита ЛИЧНЫХ коллекций (team_id IS NULL).
func (s *Service) CheckCollectionQuota(ctx context.Context, userID uint) error {
	planID, plan, err := s.getPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountPersonalCollections(ctx, userID)
	if err != nil {
		return err
	}
	if !isWithinLimit(used, plan.MaxCollections) {
		return newQuotaExceeded("collections", planID, int(used), plan.MaxCollections, "коллекций")
	}
	return nil
}

// CheckTeamCollectionQuota — проверка лимита коллекций в команде. Pool для
// всей команды, лимит из плана owner'а.
func (s *Service) CheckTeamCollectionQuota(ctx context.Context, teamID, ownerUserID uint) error {
	planID, plan, err := s.getPlan(ctx, ownerUserID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountTeamCollections(ctx, teamID)
	if err != nil {
		return err
	}
	if !isWithinLimit(used, plan.MaxTeamCollections) {
		return newQuotaExceeded("team_collections", planID, int(used), plan.MaxTeamCollections, "коллекций команды")
	}
	return nil
}

func (s *Service) CheckTeamQuota(ctx context.Context, userID uint) error {
	planID, plan, err := s.getPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountTeamsOwned(ctx, userID)
	if err != nil {
		return err
	}
	if !isWithinLimit(used, plan.MaxTeams) {
		return newQuotaExceeded("teams", planID, int(used), plan.MaxTeams, "команд")
	}
	return nil
}

// CheckTeamMemberQuota проверяет квоту участников по плану владельца команды.
func (s *Service) CheckTeamMemberQuota(ctx context.Context, teamID uint, ownerUserID uint) error {
	planID, plan, err := s.getPlan(ctx, ownerUserID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountTeamMembers(ctx, teamID)
	if err != nil {
		return err
	}
	if !isWithinLimit(int64(used), plan.MaxTeamMembers) {
		return newQuotaExceeded("team_members", planID, used, plan.MaxTeamMembers, "участников команды")
	}
	return nil
}

// Phase 16-Y: CheckShareLinkQuota и CheckDailyShareCreation удалены.
// Share-ссылки теперь живут по TTL (миграция 000061), активный count и
// daily-create счётчики не используются. Анти-абуз — общий per-user
// rate-limit (byUser(120/min)) на уровне HTTP middleware.

func (s *Service) CheckExtensionQuota(ctx context.Context, userID uint) error {
	planID, user, plan, err := s.getUserAndPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.GetDailyUsage(ctx, userID, time.Now(), FeatureExtension)
	if err != nil {
		return err
	}
	limit := effectiveLimit(user, "max_ext_uses_daily", plan.MaxExtUsesDaily)
	if !isWithinLimit(int64(used), limit) {
		return newQuotaExceeded("ext_daily", planID, used, limit, "расширения")
	}
	return nil
}

func (s *Service) CheckMCPQuota(ctx context.Context, userID uint) error {
	planID, user, plan, err := s.getUserAndPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.GetDailyUsage(ctx, userID, time.Now(), FeatureMCP)
	if err != nil {
		return err
	}
	limit := effectiveLimit(user, "max_mcp_uses_daily", plan.MaxMCPUsesDaily)
	if !isWithinLimit(int64(used), limit) {
		return newQuotaExceeded("mcp_daily", planID, used, limit, "MCP-вызовов")
	}
	return nil
}

// CheckChainQuota — Phase 16. Проверяет лимит ЛИЧНЫХ цепочек юзера
// (team_id IS NULL). Командные цепочки — через CheckTeamChainQuota.
// Считаются только не-soft-deleted (deleted_at IS NULL).
func (s *Service) CheckChainQuota(ctx context.Context, userID uint) error {
	planID, plan, err := s.getPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountPersonalChains(ctx, userID)
	if err != nil {
		return err
	}
	if !isWithinLimit(used, plan.MaxChains) {
		return newQuotaExceeded("chains", planID, int(used), plan.MaxChains, "цепочек")
	}
	return nil
}

// GetTeamUsageSummary — usage всех ресурсов команды против её team-pool лимита.
// Лимиты берутся из плана owner'а команды (team.CreatedBy). Caller передаёт
// готовый team — quota.Service не имеет TeamRepository в зависимостях, чтобы
// не менять signature NewService и не ломать тесты.
//
// Используется в GET /api/teams/{id}/usage и в settings/subscription для
// отображения «Использование команд» юзера.
func (s *Service) GetTeamUsageSummary(ctx context.Context, team *models.Team) (*TeamUsageSummary, error) {
	owner, err := s.users.GetByID(ctx, team.CreatedBy)
	if err != nil {
		return nil, err
	}
	plan, err := s.plans.GetByID(ctx, owner.PlanID)
	if err != nil {
		return nil, err
	}

	var (
		prompts, collections, chains int64
	)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		prompts, err = s.quotas.CountTeamPrompts(gctx, team.ID)
		return
	})
	g.Go(func() (err error) {
		collections, err = s.quotas.CountTeamCollections(gctx, team.ID)
		return
	})
	g.Go(func() (err error) {
		chains, err = s.quotas.CountTeamChains(gctx, team.ID)
		return
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &TeamUsageSummary{
		TeamID:      team.ID,
		TeamName:    team.Name,
		OwnerPlanID: owner.PlanID,
		Prompts:     QuotaInfo{Used: int(prompts), Limit: plan.MaxTeamPrompts},
		Collections: QuotaInfo{Used: int(collections), Limit: plan.MaxTeamCollections},
		Chains:      QuotaInfo{Used: int(chains), Limit: plan.MaxTeamChains},
	}, nil
}

// CheckTeamChainQuota — Pack T. Лимит цепочек в команде. Pool для всей
// команды, считается против плана owner'а.
func (s *Service) CheckTeamChainQuota(ctx context.Context, teamID, ownerUserID uint) error {
	planID, plan, err := s.getPlan(ctx, ownerUserID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountTeamChains(ctx, teamID)
	if err != nil {
		return err
	}
	if !isWithinLimit(used, plan.MaxTeamChains) {
		return newQuotaExceeded("team_chains", planID, int(used), plan.MaxTeamChains, "цепочек команды")
	}
	return nil
}

// IsMaxTierUser — Phase B (Conditional Chains). True если планы 'max' / 'max_yearly'.
// Используется в chain.Service для гейта conditional шагов (Max-only фича).
// Безопасный default false при ошибках чтения плана.
func (s *Service) IsMaxTierUser(ctx context.Context, userID uint) bool {
	planID, _, err := s.getPlan(ctx, userID)
	if err != nil {
		return false
	}
	return planID == "max" || planID == "max_yearly"
}

// CheckChainStepQuota — Phase 16. Лимит шагов внутри одной цепочки. Вызывается
// перед AddStep. plan.MaxStepsPerChain действует на уровне chain, не user.
func (s *Service) CheckChainStepQuota(ctx context.Context, userID uint, chainID uint) error {
	planID, plan, err := s.getPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountStepsByChain(ctx, chainID)
	if err != nil {
		return err
	}
	if !isWithinLimit(used, plan.MaxStepsPerChain) {
		return newQuotaExceeded("chain_steps", planID, int(used), plan.MaxStepsPerChain, "шагов в цепочке")
	}
	return nil
}

func (s *Service) IncrementExtensionUsage(ctx context.Context, userID uint) error {
	return s.quotas.IncrementDailyUsage(ctx, userID, time.Now(), FeatureExtension)
}

func (s *Service) IncrementMCPUsage(ctx context.Context, userID uint) error {
	return s.quotas.IncrementDailyUsage(ctx, userID, time.Now(), FeatureMCP)
}

// GetUsageSummary возвращает полную сводку использования для /api/subscription/usage.
//
// MJ-20: 6 SELECT'ов идут параллельно через errgroup вместо sequential.
// Раньше: 6 round-trip'ов serializable; на VPS PG локально ~1-2ms каждый
// = +6-12ms baseline на каждый /api/subscription/usage запрос. Теперь
// параллельно — общий latency = max(individual), не sum.
func (s *Service) GetUsageSummary(ctx context.Context, userID uint) (*UsageSummary, error) {
	planID, user, plan, err := s.getUserAndPlan(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	var (
		prompts, collections, teams, chains int64
		extUsed, mcpUsed                    int
	)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		// Pack T: usage-summary показывает PERSONAL ресурсы юзера.
		// Командные пулы — отдельная страница «Команда → Использование» (TODO).
		prompts, err = s.quotas.CountPersonalPrompts(gctx, userID)
		return
	})
	g.Go(func() (err error) {
		collections, err = s.quotas.CountPersonalCollections(gctx, userID)
		return
	})
	g.Go(func() (err error) {
		teams, err = s.quotas.CountTeamsOwned(gctx, userID)
		return
	})
	g.Go(func() (err error) {
		extUsed, err = s.quotas.GetDailyUsage(gctx, userID, now, FeatureExtension)
		return
	})
	g.Go(func() (err error) {
		mcpUsed, err = s.quotas.GetDailyUsage(gctx, userID, now, FeatureMCP)
		return
	})
	g.Go(func() (err error) {
		chains, err = s.quotas.CountPersonalChains(gctx, userID)
		return
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &UsageSummary{
		PlanID:       planID,
		Prompts:      QuotaInfo{Used: int(prompts), Limit: effectiveLimit(user, "max_prompts", plan.MaxPrompts)},
		Collections:  QuotaInfo{Used: int(collections), Limit: plan.MaxCollections},
		Teams:        QuotaInfo{Used: int(teams), Limit: plan.MaxTeams},
		ExtUsesToday: QuotaInfo{Used: extUsed, Limit: effectiveLimit(user, "max_ext_uses_daily", plan.MaxExtUsesDaily)},
		MCPUsesToday: QuotaInfo{Used: mcpUsed, Limit: effectiveLimit(user, "max_mcp_uses_daily", plan.MaxMCPUsesDaily)},
		Chains:       QuotaInfo{Used: int(chains), Limit: plan.MaxChains},
	}, nil
}

// DowngradePreview — превышения лимитов целевого плана (M-10).
// Phase 16-Y: OverShares убран — share-ссылки теперь живут по TTL без
// active-count, downgrade на share не влияет (свежие ссылки на новом плане
// получат default TTL 30d, существующие доживут свой срок).
// Поле Over — абсолютное превышение (used - limit), 0 если в пределах.
type DowngradePreview struct {
	TargetPlanID    string `json:"target_plan_id"`
	CurrentPlanID   string `json:"current_plan_id"`
	OverPrompts     int    `json:"over_prompts"`
	OverCollections int    `json:"over_collections"`
	OverTeams       int    `json:"over_teams"`
}

// HasOverages возвращает true если хотя бы один ресурс превышает лимит target-плана.
// Удобно для UI — не нужно разбирать каждое поле отдельно.
func (p *DowngradePreview) HasOverages() bool {
	return p.OverPrompts > 0 || p.OverCollections > 0 || p.OverTeams > 0
}

// GetDowngradePreview считает, сколько ресурсов у юзера превышает лимиты
// target-плана (M-10). Вызывается перед POST /downgrade, чтобы UI показал
// warning "У вас 55 промптов, на Free лимит 50 — 5 самых старых будут архивированы".
func (s *Service) GetDowngradePreview(ctx context.Context, userID uint, targetPlanID string) (*DowngradePreview, error) {
	currentPlanID, _, err := s.getPlan(ctx, userID)
	if err != nil {
		return nil, err
	}
	targetPlan, err := s.plans.GetByID(ctx, targetPlanID)
	if err != nil {
		return nil, err
	}

	// Pack T: downgrade preview оценивает ЛИЧНЫЕ ресурсы (не команд) —
	// командные пулы остаются у команды независимо от плана участника.
	prompts, err := s.quotas.CountPersonalPrompts(ctx, userID)
	if err != nil {
		return nil, err
	}
	collections, err := s.quotas.CountPersonalCollections(ctx, userID)
	if err != nil {
		return nil, err
	}
	teams, err := s.quotas.CountTeamsOwned(ctx, userID)
	if err != nil {
		return nil, err
	}

	over := func(used int64, limit int) int {
		diff := int(used) - limit
		if diff < 0 {
			return 0
		}
		return diff
	}

	return &DowngradePreview{
		TargetPlanID:    targetPlan.ID,
		CurrentPlanID:   currentPlanID,
		OverPrompts:     over(prompts, targetPlan.MaxPrompts),
		OverCollections: over(collections, targetPlan.MaxCollections),
		OverTeams:       over(teams, targetPlan.MaxTeams),
	}, nil
}
