package quota

import (
	"context"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// Service — центральный сервис проверки и инкремента квот.
// Все Check* методы возвращают *QuotaExceededError при превышении лимита
// или nil, если действие разрешено.
type Service struct {
	plans  repo.PlanRepository
	quotas repo.QuotaRepository
	users  repo.UserRepository
}

func NewService(plans repo.PlanRepository, quotas repo.QuotaRepository, users repo.UserRepository) *Service {
	return &Service{plans: plans, quotas: quotas, users: users}
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

func isWithinLimit(used int64, limit int) bool {
	return limit == -1 || used < int64(limit)
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

func (s *Service) CheckAIQuota(ctx context.Context, userID uint) error {
	planID, plan, err := s.getPlan(ctx, userID)
	if err != nil {
		return err
	}

	var used int
	if plan.AIRequestsIsTotal {
		used, err = s.quotas.GetTotalUsage(ctx, userID, FeatureAI)
	} else {
		used, err = s.quotas.GetDailyUsage(ctx, userID, time.Now(), FeatureAI)
	}
	if err != nil {
		return err
	}

	if !isWithinLimit(int64(used), plan.MaxAIRequestsDaily) {
		quotaType := "ai_daily"
		if plan.AIRequestsIsTotal {
			quotaType = "ai_total"
		}
		return newQuotaExceeded(quotaType, planID, used, plan.MaxAIRequestsDaily, "AI-запросов")
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

func (s *Service) CheckShareLinkQuota(ctx context.Context, userID uint) error {
	planID, plan, err := s.getPlan(ctx, userID)
	if err != nil {
		return err
	}
	used, err := s.quotas.CountActiveShareLinks(ctx, userID)
	if err != nil {
		return err
	}
	if !isWithinLimit(used, plan.MaxShareLinks) {
		return newQuotaExceeded("share_links", planID, int(used), plan.MaxShareLinks, "публичных ссылок")
	}
	return nil
}

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

func (s *Service) IncrementAIUsage(ctx context.Context, userID uint) error {
	return s.quotas.IncrementDailyUsage(ctx, userID, time.Now(), FeatureAI)
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

	var aiUsed int
	if plan.AIRequestsIsTotal {
		aiUsed, err = s.quotas.GetTotalUsage(ctx, userID, FeatureAI)
	} else {
		aiUsed, err = s.quotas.GetDailyUsage(ctx, userID, now, FeatureAI)
	}
	if err != nil {
		return nil, err
	}

	teams, err := s.quotas.CountTeamsOwned(ctx, userID)
	if err != nil {
		return nil, err
	}
	shares, err := s.quotas.CountActiveShareLinks(ctx, userID)
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

	return &UsageSummary{
		PlanID:       planID,
		Prompts:      QuotaInfo{Used: int(prompts), Limit: plan.MaxPrompts},
		Collections:  QuotaInfo{Used: int(collections), Limit: plan.MaxCollections},
		AIRequests:   QuotaInfo{Used: aiUsed, Limit: plan.MaxAIRequestsDaily, IsTotal: plan.AIRequestsIsTotal},
		Teams:        QuotaInfo{Used: int(teams), Limit: plan.MaxTeams},
		ShareLinks:   QuotaInfo{Used: int(shares), Limit: plan.MaxShareLinks},
		ExtUsesToday: QuotaInfo{Used: extUsed, Limit: plan.MaxExtUsesDaily},
		MCPUsesToday: QuotaInfo{Used: mcpUsed, Limit: plan.MaxMCPUsesDaily},
	}, nil
}
