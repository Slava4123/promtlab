package quota

import (
	"context"
	"time"

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
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	plan, err := s.plans.GetByID(ctx, user.PlanID)
	if err != nil {
		return "", nil, err
	}
	return user.PlanID, plan, nil
}

// isWithinLimit — проверка used < limit. После миграции 000046 все лимиты —
// неотрицательные числа; legacy sentinel -1 "безлимит" полностью выведен.
func isWithinLimit(used int64, limit int) bool {
	return used < int64(limit)
}

func (s *Service) CheckPromptQuota(ctx context.Context, userID uint) error {
	planID, plan, err := s.getPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountPrompts(ctx, userID)
	if err != nil {
		return err
	}
	if !isWithinLimit(used, plan.MaxPrompts) {
		return newQuotaExceeded("prompts", planID, int(used), plan.MaxPrompts, "промптов")
	}
	return nil
}

func (s *Service) CheckCollectionQuota(ctx context.Context, userID uint) error {
	planID, plan, err := s.getPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountCollections(ctx, userID)
	if err != nil {
		return err
	}
	if !isWithinLimit(used, plan.MaxCollections) {
		return newQuotaExceeded("collections", planID, int(used), plan.MaxCollections, "коллекций")
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
	planID, plan, err := s.getPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.GetDailyUsage(ctx, userID, time.Now(), FeatureExtension)
	if err != nil {
		return err
	}
	if !isWithinLimit(int64(used), plan.MaxExtUsesDaily) {
		return newQuotaExceeded("ext_daily", planID, used, plan.MaxExtUsesDaily, "расширения")
	}
	return nil
}

func (s *Service) CheckMCPQuota(ctx context.Context, userID uint) error {
	planID, plan, err := s.getPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.GetDailyUsage(ctx, userID, time.Now(), FeatureMCP)
	if err != nil {
		return err
	}
	if !isWithinLimit(int64(used), plan.MaxMCPUsesDaily) {
		return newQuotaExceeded("mcp_daily", planID, used, plan.MaxMCPUsesDaily, "MCP-вызовов")
	}
	return nil
}

// CheckChainQuota — Phase 16. Проверяет лимит общего количества цепочек у юзера.
// Считаются только не-soft-deleted (deleted_at IS NULL).
func (s *Service) CheckChainQuota(ctx context.Context, userID uint) error {
	planID, plan, err := s.getPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountChains(ctx, userID)
	if err != nil {
		return err
	}
	if !isWithinLimit(used, plan.MaxChains) {
		return newQuotaExceeded("chains", planID, int(used), plan.MaxChains, "цепочек")
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
func (s *Service) GetUsageSummary(ctx context.Context, userID uint) (*UsageSummary, error) {
	planID, plan, err := s.getPlan(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	prompts, err := s.quotas.CountPrompts(ctx, userID)
	if err != nil {
		return nil, err
	}
	collections, err := s.quotas.CountCollections(ctx, userID)
	if err != nil {
		return nil, err
	}

	teams, err := s.quotas.CountTeamsOwned(ctx, userID)
	if err != nil {
		return nil, err
	}
	extUsed, err := s.quotas.GetDailyUsage(ctx, userID, now, FeatureExtension)
	if err != nil {
		return nil, err
	}
	mcpUsed, err := s.quotas.GetDailyUsage(ctx, userID, now, FeatureMCP)
	if err != nil {
		return nil, err
	}
	chains, err := s.quotas.CountChains(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UsageSummary{
		PlanID:       planID,
		Prompts:      QuotaInfo{Used: int(prompts), Limit: plan.MaxPrompts},
		Collections:  QuotaInfo{Used: int(collections), Limit: plan.MaxCollections},
		Teams:        QuotaInfo{Used: int(teams), Limit: plan.MaxTeams},
		ExtUsesToday: QuotaInfo{Used: extUsed, Limit: plan.MaxExtUsesDaily},
		MCPUsesToday: QuotaInfo{Used: mcpUsed, Limit: plan.MaxMCPUsesDaily},
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

	prompts, err := s.quotas.CountPrompts(ctx, userID)
	if err != nil {
		return nil, err
	}
	collections, err := s.quotas.CountCollections(ctx, userID)
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
