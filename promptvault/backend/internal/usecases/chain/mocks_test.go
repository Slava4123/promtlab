// Package chain — мок-инфраструктура для service-level тестов CR-7 part 2.
//
// Реализуем минимум методов всех зависимостей, нужных Service:
//   - ChainRepository: 17 методов (полностью)
//   - PromptRepository: только GetByID (Service использует только его)
//   - TeamRepository: GetMember + ListMembers (для checkReadAccess / isMaxTierForChain)
//   - PlanRepository, QuotaRepository, UserRepository: полные интерфейсы для
//     реального quota.Service (создаём через NewService с mock-ами вместо
//     interface-а — это даёт честный прогон логики IsMaxTierUser/CheckChainQuota).
//
// activity.Service передаём как nil — LogSafe nil-safe (см. activity/service.go:79).
package chain

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- ChainRepository mock ---

type mockChainRepo struct{ mock.Mock }

func (m *mockChainRepo) Create(ctx context.Context, c *models.PromptChain) error {
	args := m.Called(ctx, c)
	if c.ID == 0 {
		c.ID = 1
	}
	return args.Error(0)
}
func (m *mockChainRepo) GetByID(ctx context.Context, id uint) (*models.PromptChain, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PromptChain), args.Error(1)
}
func (m *mockChainRepo) GetByIDWithSteps(ctx context.Context, id uint) (*models.PromptChain, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PromptChain), args.Error(1)
}
func (m *mockChainRepo) ListByUser(ctx context.Context, userID uint, teamIDs []uint, limit, offset int) ([]models.PromptChain, int64, error) {
	args := m.Called(ctx, userID, teamIDs, limit, offset)
	return args.Get(0).([]models.PromptChain), args.Get(1).(int64), args.Error(2)
}
func (m *mockChainRepo) ListByUserWithStats(ctx context.Context, userID uint, teamIDs []uint, limit, offset int) ([]models.PromptChainListRow, int64, error) {
	args := m.Called(ctx, userID, teamIDs, limit, offset)
	return args.Get(0).([]models.PromptChainListRow), args.Get(1).(int64), args.Error(2)
}
func (m *mockChainRepo) Update(ctx context.Context, c *models.PromptChain) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockChainRepo) SoftDelete(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockChainRepo) HasActiveExecutions(ctx context.Context, chainID uint) (bool, error) {
	args := m.Called(ctx, chainID)
	return args.Bool(0), args.Error(1)
}
func (m *mockChainRepo) AddStep(ctx context.Context, s *models.PromptChainStep) error {
	args := m.Called(ctx, s)
	if s.ID == 0 {
		s.ID = uint(time.Now().UnixNano() % 1_000_000)
	}
	return args.Error(0)
}
func (m *mockChainRepo) GetStepByID(ctx context.Context, stepID uint) (*models.PromptChainStep, error) {
	args := m.Called(ctx, stepID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PromptChainStep), args.Error(1)
}
func (m *mockChainRepo) UpdateStep(ctx context.Context, s *models.PromptChainStep) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockChainRepo) RemoveStep(ctx context.Context, stepID uint) error {
	return m.Called(ctx, stepID).Error(0)
}
func (m *mockChainRepo) ListStepsByChain(ctx context.Context, chainID uint) ([]models.PromptChainStep, error) {
	args := m.Called(ctx, chainID)
	return args.Get(0).([]models.PromptChainStep), args.Error(1)
}
func (m *mockChainRepo) ReorderSteps(ctx context.Context, chainID uint, stepIDs []uint) error {
	return m.Called(ctx, chainID, stepIDs).Error(0)
}
func (m *mockChainRepo) CountStepsByChain(ctx context.Context, chainID uint) (int64, error) {
	args := m.Called(ctx, chainID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockChainRepo) CountChainsUsingPrompt(ctx context.Context, promptID uint) (int64, error) {
	args := m.Called(ctx, promptID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockChainRepo) RelinkPromptPredecessors(ctx context.Context, chainID, fromID uint, toID *uint) error {
	return m.Called(ctx, chainID, fromID, toID).Error(0)
}

// InTransaction — выполняем fn(self): mock-репо имитирует tx как no-op обёртку.
// Все mock'и на repo-методах сработают точно так же, как и без транзакции.
// Этого достаточно для unit-тестов: реальная транзакционность тестируется в
// prompt_chain_repo_test.go (testcontainers + RealMigrations).
func (m *mockChainRepo) InTransaction(ctx context.Context, fn func(repo.ChainRepository) error) error {
	return fn(m)
}

func (m *mockChainRepo) CreateExecution(ctx context.Context, e *models.PromptChainExecution) error {
	args := m.Called(ctx, e)
	if e.ID == 0 {
		e.ID = 1
	}
	return args.Error(0)
}
func (m *mockChainRepo) GetExecutionByID(ctx context.Context, execID uint) (*models.PromptChainExecution, error) {
	args := m.Called(ctx, execID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PromptChainExecution), args.Error(1)
}
func (m *mockChainRepo) UpdateExecution(ctx context.Context, e *models.PromptChainExecution) error {
	return m.Called(ctx, e).Error(0)
}
func (m *mockChainRepo) ListExecutionsByChain(ctx context.Context, chainID uint, limit int) ([]models.PromptChainExecution, error) {
	args := m.Called(ctx, chainID, limit)
	return args.Get(0).([]models.PromptChainExecution), args.Error(1)
}

// --- PromptRepository mock (минимум) ---
//
// Service-логика чейнов вызывает только GetByID. Остальные методы
// заглушены (panic если кто-то их вызвал — сигнал что mocking incomplete).

type mockPromptRepo struct{ mock.Mock }

func (m *mockPromptRepo) GetByID(ctx context.Context, id uint) (*models.Prompt, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}
func (m *mockPromptRepo) GetMeta(ctx context.Context, id uint) (*models.Prompt, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

// Остальные методы — чтобы удовлетворить интерфейс repo.PromptRepository.
func (m *mockPromptRepo) Create(ctx context.Context, p *models.Prompt) error { panic("not used") }
func (m *mockPromptRepo) Update(ctx context.Context, p *models.Prompt) error { panic("not used") }
func (m *mockPromptRepo) SoftDelete(ctx context.Context, id uint) error      { panic("not used") }
func (m *mockPromptRepo) List(ctx context.Context, f repo.PromptListFilter) ([]models.Prompt, int64, error) {
	panic("not used")
}
func (m *mockPromptRepo) SetFavorite(ctx context.Context, id uint, fav bool) error {
	panic("not used")
}
func (m *mockPromptRepo) IncrementUsage(ctx context.Context, id uint) error { panic("not used") }
func (m *mockPromptRepo) SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Prompt, error) {
	panic("not used")
}
func (m *mockPromptRepo) UpdateLastUsed(ctx context.Context, id uint) error { panic("not used") }
func (m *mockPromptRepo) ListRecent(ctx context.Context, userID uint, teamID *uint, limit int) ([]models.Prompt, error) {
	panic("not used")
}
func (m *mockPromptRepo) LogUsage(ctx context.Context, userID, promptID uint) error {
	panic("not used")
}
func (m *mockPromptRepo) ListUsageHistory(ctx context.Context, userID uint, teamID *uint, page, pageSize int) ([]models.PromptUsageLog, int64, error) {
	panic("not used")
}
func (m *mockPromptRepo) SuggestByPrefix(ctx context.Context, userID uint, teamID *uint, prefix string, limit int) ([]string, error) {
	panic("not used")
}
func (m *mockPromptRepo) GetPublicBySlug(ctx context.Context, slug string) (*models.Prompt, error) {
	panic("not used")
}
func (m *mockPromptRepo) ListPublic(ctx context.Context, limit int) ([]models.Prompt, error) {
	panic("not used")
}

// --- TeamRepository mock ---

type mockTeamRepo struct{ mock.Mock }

func (m *mockTeamRepo) GetMember(ctx context.Context, teamID, userID uint) (*models.TeamMember, error) {
	args := m.Called(ctx, teamID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TeamMember), args.Error(1)
}
func (m *mockTeamRepo) ListMembers(ctx context.Context, teamID uint) ([]models.TeamMember, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).([]models.TeamMember), args.Error(1)
}

// Остальные методы — заглушки.
func (m *mockTeamRepo) CreateWithOwner(_ context.Context, _ *models.Team, _ uint) error {
	panic("not used")
}
func (m *mockTeamRepo) GetBySlug(_ context.Context, _ string) (*models.Team, error) {
	panic("not used")
}
// Pack T: chain.Service.Create в team-mode дёргает teams.GetByID для резолва
// owner'а команды (для team-pool квоты). Прежняя panic("not used") заменена
// на mock-aware реализацию.
func (m *mockTeamRepo) GetByID(ctx context.Context, id uint) (*models.Team, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}
func (m *mockTeamRepo) ListByUserID(_ context.Context, _ uint) ([]models.Team, error) {
	panic("not used")
}
func (m *mockTeamRepo) ListByUserIDWithRolesAndCounts(_ context.Context, _ uint) ([]models.TeamWithRoleAndCount, error) {
	panic("not used")
}
func (m *mockTeamRepo) ListOwnedTeams(_ context.Context, _ uint) ([]models.Team, error) {
	panic("not used")
}
func (m *mockTeamRepo) Update(_ context.Context, _ *models.Team) error          { panic("not used") }
func (m *mockTeamRepo) Delete(_ context.Context, _ uint) error                   { panic("not used") }
func (m *mockTeamRepo) UpdateMemberRole(_ context.Context, _, _ uint, _ models.TeamRole) error {
	panic("not used")
}
func (m *mockTeamRepo) RemoveMember(_ context.Context, _, _ uint) error { panic("not used") }
func (m *mockTeamRepo) CountMembers(_ context.Context, _ uint) (int, error) {
	panic("not used")
}
func (m *mockTeamRepo) CreateInvitation(_ context.Context, _ *models.TeamInvitation) error {
	panic("not used")
}
func (m *mockTeamRepo) GetInvitationByID(_ context.Context, _ uint) (*models.TeamInvitation, error) {
	panic("not used")
}
func (m *mockTeamRepo) GetPendingInvitation(_ context.Context, _, _ uint) (*models.TeamInvitation, error) {
	panic("not used")
}
func (m *mockTeamRepo) ListPendingByUserID(_ context.Context, _ uint) ([]models.TeamInvitation, error) {
	panic("not used")
}
func (m *mockTeamRepo) ListPendingByTeamID(_ context.Context, _ uint) ([]models.TeamInvitation, error) {
	panic("not used")
}
func (m *mockTeamRepo) UpdateInvitationStatus(_ context.Context, _ uint, _ models.InvitationStatus) error {
	panic("not used")
}
func (m *mockTeamRepo) DeleteInvitation(_ context.Context, _ uint) error { panic("not used") }
func (m *mockTeamRepo) AcceptInvitationTx(_ context.Context, _ uint, _ *models.TeamMember) error {
	panic("not used")
}
func (m *mockTeamRepo) UpdateBranding(_ context.Context, _ uint, _, _, _, _ string) error {
	panic("not used")
}
func (m *mockTeamRepo) UpdateBrandLogoSource(_ context.Context, _ uint, _ string) error {
	panic("not used")
}

// --- PlanRepository mock ---

type mockPlanRepo struct{ mock.Mock }

func (m *mockPlanRepo) GetAll(ctx context.Context) ([]models.SubscriptionPlan, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.SubscriptionPlan), args.Error(1)
}
func (m *mockPlanRepo) GetByID(ctx context.Context, id string) (*models.SubscriptionPlan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SubscriptionPlan), args.Error(1)
}
func (m *mockPlanRepo) GetActive(ctx context.Context) ([]models.SubscriptionPlan, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.SubscriptionPlan), args.Error(1)
}

// --- QuotaRepository mock ---

type mockQuotaRepo struct{ mock.Mock }

func (m *mockQuotaRepo) CountPersonalPrompts(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockQuotaRepo) CountPersonalCollections(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockQuotaRepo) CountPersonalChains(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockQuotaRepo) CountTeamPrompts(ctx context.Context, teamID uint) (int64, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockQuotaRepo) CountTeamCollections(ctx context.Context, teamID uint) (int64, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockQuotaRepo) CountTeamChains(ctx context.Context, teamID uint) (int64, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockQuotaRepo) CountTeamsOwned(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockQuotaRepo) CountActiveShareLinks(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockQuotaRepo) CountTeamMembers(ctx context.Context, teamID uint) (int, error) {
	args := m.Called(ctx, teamID)
	return args.Int(0), args.Error(1)
}
func (m *mockQuotaRepo) GetDailyUsage(ctx context.Context, userID uint, date time.Time, featureType string) (int, error) {
	args := m.Called(ctx, userID, date, featureType)
	return args.Int(0), args.Error(1)
}
func (m *mockQuotaRepo) GetTotalUsage(ctx context.Context, userID uint, featureType string) (int, error) {
	args := m.Called(ctx, userID, featureType)
	return args.Int(0), args.Error(1)
}
func (m *mockQuotaRepo) IncrementDailyUsage(ctx context.Context, userID uint, date time.Time, featureType string) error {
	return m.Called(ctx, userID, date, featureType).Error(0)
}
func (m *mockQuotaRepo) CountStepsByChain(ctx context.Context, chainID uint) (int64, error) {
	args := m.Called(ctx, chainID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockQuotaRepo) DeleteOldDailyUsage(ctx context.Context, olderThanDays int) (int64, error) {
	args := m.Called(ctx, olderThanDays)
	return args.Get(0).(int64), args.Error(1)
}

// --- UserRepository mock (минимум для quota.Service) ---

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) GetByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// Все остальные UserRepository-методы — заглушки.
func (m *mockUserRepo) Create(_ context.Context, _ *models.User) error { panic("not used") }
func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*models.User, error) {
	panic("not used")
}
func (m *mockUserRepo) GetByUsername(_ context.Context, _ string) (*models.User, error) {
	panic("not used")
}
func (m *mockUserRepo) SearchUsers(_ context.Context, _ string, _ int) ([]models.User, error) {
	panic("not used")
}
func (m *mockUserRepo) Update(_ context.Context, _ *models.User) error { panic("not used") }
func (m *mockUserRepo) SetPlan(_ context.Context, _ uint, _ string) error {
	panic("not used")
}
func (m *mockUserRepo) SetQuotaWarningSentOn(_ context.Context, _ uint, _ time.Time) error {
	panic("not used")
}
func (m *mockUserRepo) TouchLastLogin(_ context.Context, _ uint) error { panic("not used") }
func (m *mockUserRepo) ListInactiveForReengagement(_ context.Context, _, _ time.Time, _ int) ([]models.User, error) {
	panic("not used")
}
func (m *mockUserRepo) MarkReengagementSent(_ context.Context, _ uint) error {
	panic("not used")
}
func (m *mockUserRepo) CountReferredBy(_ context.Context, _ string) (int64, error) {
	panic("not used")
}
func (m *mockUserRepo) GetByReferralCode(_ context.Context, _ string) (*models.User, error) {
	panic("not used")
}
func (m *mockUserRepo) MarkReferralRewarded(_ context.Context, _ uint) (bool, error) {
	panic("not used")
}
func (m *mockUserRepo) ListPaidUsers(_ context.Context) ([]uint, error) { panic("not used") }
func (m *mockUserRepo) SetInsightEmailsEnabled(_ context.Context, _ uint, _ bool) error {
	panic("not used")
}
