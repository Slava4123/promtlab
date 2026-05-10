package quota

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- fakes (in-memory) ---

type fakeUserRepo struct {
	users map[uint]*models.User
}

func (r *fakeUserRepo) GetByID(_ context.Context, id uint) (*models.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return u, nil
}

// Заглушки методов UserRepository, которые quota не вызывает — panic сигнализирует
// о регрессии, если quota однажды начнёт их дёргать.
func (r *fakeUserRepo) Create(context.Context, *models.User) error { panic("not used") }
func (r *fakeUserRepo) GetByEmail(context.Context, string) (*models.User, error) {
	panic("not used")
}
func (r *fakeUserRepo) GetByUsername(context.Context, string) (*models.User, error) {
	panic("not used")
}
func (r *fakeUserRepo) SearchUsers(context.Context, string, int) ([]models.User, error) {
	panic("not used")
}
func (r *fakeUserRepo) Update(context.Context, *models.User) error  { panic("not used") }
func (r *fakeUserRepo) SetPlan(context.Context, uint, string) error { panic("not used") }
func (r *fakeUserRepo) SetQuotaWarningSentOn(context.Context, uint, time.Time) error {
	return nil
}
func (r *fakeUserRepo) TouchLastLogin(context.Context, uint) error {
	return nil
}
func (r *fakeUserRepo) ListInactiveForReengagement(context.Context, time.Time, time.Time, int) ([]models.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) MarkReengagementSent(context.Context, uint) error {
	return nil
}
func (r *fakeUserRepo) CountReferredBy(context.Context, string) (int64, error) { return 0, nil }
func (r *fakeUserRepo) GetByReferralCode(context.Context, string) (*models.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) MarkReferralRewarded(context.Context, uint) (bool, error) {
	return false, nil
}
func (r *fakeUserRepo) ListMaxUsers(context.Context) ([]uint, error) { return nil, nil }
func (r *fakeUserRepo) SetInsightEmailsEnabled(context.Context, uint, bool) error { return nil }

type fakePlanRepo struct {
	plans map[string]*models.SubscriptionPlan
}

func (r *fakePlanRepo) GetByID(_ context.Context, id string) (*models.SubscriptionPlan, error) {
	p, ok := r.plans[id]
	if !ok {
		return nil, errors.New("plan not found")
	}
	return p, nil
}

func (r *fakePlanRepo) GetAll(context.Context) ([]models.SubscriptionPlan, error) {
	panic("not used")
}
func (r *fakePlanRepo) GetActive(context.Context) ([]models.SubscriptionPlan, error) {
	panic("not used")
}

type fakeQuotaRepo struct {
	prompts      int64
	collections  int64
	teamsOwned   int64
	shareLinks   int64
	teamMembers  int
	chains       int64
	stepsByChain map[uint]int64
	dailyUsage   map[string]int // feature_type → count
	totalUsage   map[string]int
	incrementErr error
	incrementLog []incCall
}

type incCall struct {
	userID  uint
	date    time.Time
	feature string
}

func (r *fakeQuotaRepo) CountPersonalPrompts(context.Context, uint) (int64, error) {
	return r.prompts, nil
}
func (r *fakeQuotaRepo) CountPersonalCollections(context.Context, uint) (int64, error) {
	return r.collections, nil
}
func (r *fakeQuotaRepo) CountTeamPrompts(context.Context, uint) (int64, error) {
	return r.prompts, nil
}
func (r *fakeQuotaRepo) CountTeamCollections(context.Context, uint) (int64, error) {
	return r.collections, nil
}
func (r *fakeQuotaRepo) CountTeamsOwned(context.Context, uint) (int64, error) {
	return r.teamsOwned, nil
}
func (r *fakeQuotaRepo) CountActiveShareLinks(context.Context, uint) (int64, error) {
	return r.shareLinks, nil
}
func (r *fakeQuotaRepo) CountTeamMembers(context.Context, uint) (int, error) {
	return r.teamMembers, nil
}
func (r *fakeQuotaRepo) GetDailyUsage(_ context.Context, _ uint, _ time.Time, feature string) (int, error) {
	return r.dailyUsage[feature], nil
}
func (r *fakeQuotaRepo) GetTotalUsage(_ context.Context, _ uint, feature string) (int, error) {
	return r.totalUsage[feature], nil
}
func (r *fakeQuotaRepo) IncrementDailyUsage(_ context.Context, userID uint, date time.Time, feature string) error {
	if r.incrementErr != nil {
		return r.incrementErr
	}
	r.incrementLog = append(r.incrementLog, incCall{userID: userID, date: date, feature: feature})
	return nil
}
func (r *fakeQuotaRepo) CountPersonalChains(context.Context, uint) (int64, error) {
	return r.chains, nil
}
func (r *fakeQuotaRepo) CountTeamChains(context.Context, uint) (int64, error) {
	return r.chains, nil
}
func (r *fakeQuotaRepo) CountStepsByChain(_ context.Context, chainID uint) (int64, error) {
	return r.stepsByChain[chainID], nil
}
func (r *fakeQuotaRepo) DeleteOldDailyUsage(context.Context, int) (int64, error) {
	return 0, nil
}

func TestEffectiveLimit_NoLegacy(t *testing.T) {
	u := &models.User{}
	got := effectiveLimit(u, "max_prompts", 15)
	if got != 15 {
		t.Errorf("no legacy → expected plan value 15, got %d", got)
	}
}

func TestEffectiveLimit_LegacyHigherThanPlan(t *testing.T) {
	u := &models.User{LegacyQuotas: json.RawMessage(`{"max_prompts": 50}`)}
	// Юзер был на старом Free (50 промптов), новый план Free=15. Должны
	// сохранить ему 50 (grandfather против downgrade).
	got := effectiveLimit(u, "max_prompts", 15)
	if got != 50 {
		t.Errorf("legacy 50 > plan 15 → expected 50, got %d", got)
	}
}

func TestEffectiveLimit_LegacyLowerThanPlan(t *testing.T) {
	u := &models.User{LegacyQuotas: json.RawMessage(`{"max_prompts": 50}`)}
	// Юзер с legacy=50 апгрейднулся на Pro (plan=500). Должен получить 500,
	// а не 50 — legacy не должен ограничивать апгрейд.
	got := effectiveLimit(u, "max_prompts", 500)
	if got != 500 {
		t.Errorf("legacy 50 < plan 500 → expected 500 (upgrade), got %d", got)
	}
}

func TestEffectiveLimit_LegacyEqualToPlan(t *testing.T) {
	u := &models.User{LegacyQuotas: json.RawMessage(`{"max_prompts": 50}`)}
	got := effectiveLimit(u, "max_prompts", 50)
	if got != 50 {
		t.Errorf("legacy == plan → expected 50, got %d", got)
	}
}

// --- helpers ---

func newService(user *models.User, plan *models.SubscriptionPlan, q *fakeQuotaRepo) *Service {
	return NewService(
		&fakePlanRepo{plans: map[string]*models.SubscriptionPlan{plan.ID: plan}},
		q,
		&fakeUserRepo{users: map[uint]*models.User{user.ID: user}},
	)
}

// --- isWithinLimit ---

func TestIsWithinLimit(t *testing.T) {
	cases := []struct {
		name  string
		used  int64
		limit int
		want  bool
	}{
		// Sentinel -1 "unlimited" удалён в Phase 14.3 после миграции 000046 — все лимиты теперь неотрицательные.
		{"used < limit allows", 9, 10, true},
		{"used = limit blocks (strict <)", 10, 10, false},
		{"used > limit blocks", 11, 10, false},
		{"zero limit blocks any use", 0, 0, false},
		{"zero used with limit=0 blocks", 0, 0, false},
		{"large limit still honoured", 9_999_998, 9_999_999, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isWithinLimit(tc.used, tc.limit); got != tc.want {
				t.Fatalf("isWithinLimit(%d, %d) = %v, want %v", tc.used, tc.limit, got, tc.want)
			}
		})
	}
}

// --- CheckPromptQuota ---

func TestCheckPromptQuota(t *testing.T) {
	cases := []struct {
		name       string
		used       int64
		maxPrompts int
		wantErr    bool
	}{
		{"free within limit", 49, 50, false},
		{"free at limit", 50, 50, true},
		{"free over limit", 51, 50, true},
		{"max with large finite limit", 9_998, 9_999, false},
	}
	user := &models.User{ID: 1, PlanID: "free"}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan := &models.SubscriptionPlan{ID: "free", MaxPrompts: tc.maxPrompts}
			svc := newService(user, plan, &fakeQuotaRepo{prompts: tc.used})
			err := svc.CheckPromptQuota(context.Background(), 1)
			if tc.wantErr {
				var qe *QuotaExceededError
				if !errors.As(err, &qe) {
					t.Fatalf("want QuotaExceededError, got %v", err)
				}
				if qe.QuotaType != "prompts" || qe.Used != int(tc.used) || qe.Limit != tc.maxPrompts {
					t.Fatalf("unexpected error fields: %+v", qe)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// --- CheckExtensionQuota / CheckMCPQuota — единая логика, один smoke-test ---

func TestCheckExtensionAndMCPQuota(t *testing.T) {
	user := &models.User{ID: 1, PlanID: "free"}
	plan := &models.SubscriptionPlan{
		ID:              "free",
		MaxExtUsesDaily: 5,
		MaxMCPUsesDaily: 5,
	}
	q := &fakeQuotaRepo{dailyUsage: map[string]int{
		FeatureExtension: 5, // at limit
		FeatureMCP:       4, // within limit
	}}
	svc := newService(user, plan, q)

	err := svc.CheckExtensionQuota(context.Background(), 1)
	var ext *QuotaExceededError
	if !errors.As(err, &ext) || ext.QuotaType != "ext_daily" {
		t.Fatalf("extension: want QuotaExceeded ext_daily, got %v", err)
	}

	if err := svc.CheckMCPQuota(context.Background(), 1); err != nil {
		t.Fatalf("mcp: unexpected error %v", err)
	}
}

// --- MN-9: Boundary cases ---

// Zero limit (например, plan ставит MaxChains=0 → создание любой цепочки заблокировано).
func TestCheckChainQuota_ZeroLimit_AnyUsageBlocks(t *testing.T) {
	user := &models.User{ID: 1, PlanID: "free"}
	plan := &models.SubscriptionPlan{ID: "free", MaxChains: 0}
	q := &fakeQuotaRepo{chains: 0}
	svc := newService(user, plan, q)

	err := svc.CheckChainQuota(context.Background(), 1)
	var qe *QuotaExceededError
	if !errors.As(err, &qe) {
		t.Fatalf("expected QuotaExceededError на limit=0, got %v", err)
	}
	if qe.QuotaType != "chains" || qe.Limit != 0 {
		t.Fatalf("expected chains/limit=0, got %+v", qe)
	}
}

// Exactly at limit (used == limit) — границу не пересекаем; isWithinLimit strict <.
func TestCheckPromptQuota_ExactlyAtLimit_Blocks(t *testing.T) {
	user := &models.User{ID: 1, PlanID: "pro"}
	plan := &models.SubscriptionPlan{ID: "pro", MaxPrompts: 100}
	q := &fakeQuotaRepo{prompts: 100}
	svc := newService(user, plan, q)

	err := svc.CheckPromptQuota(context.Background(), 1)
	var qe *QuotaExceededError
	if !errors.As(err, &qe) {
		t.Fatalf("expected QuotaExceeded на used=limit=100, got %v", err)
	}
	if qe.Used != 100 || qe.Limit != 100 {
		t.Fatalf("expected used=100 limit=100, got %+v", qe)
	}
}

// Огромный used (потенциальный uint overflow) — int64 holds 2^63-1; не должно crash.
func TestCheckPromptQuota_HugeUsage_StillRejects(t *testing.T) {
	user := &models.User{ID: 1, PlanID: "pro"}
	plan := &models.SubscriptionPlan{ID: "pro", MaxPrompts: 100}
	q := &fakeQuotaRepo{prompts: 1_000_000_000_000} // 10^12
	svc := newService(user, plan, q)

	err := svc.CheckPromptQuota(context.Background(), 1)
	var qe *QuotaExceededError
	if !errors.As(err, &qe) {
		t.Fatalf("expected QuotaExceeded на огромном used, got %v", err)
	}
	if qe.Used <= qe.Limit {
		t.Fatalf("expected used > limit, got used=%d limit=%d", qe.Used, qe.Limit)
	}
}

// CheckChainStepQuota по chainID — разные цепочки имеют независимые счётчики.
func TestCheckChainStepQuota_PerChainIndependent(t *testing.T) {
	user := &models.User{ID: 1, PlanID: "pro"}
	plan := &models.SubscriptionPlan{ID: "pro", MaxStepsPerChain: 10}
	q := &fakeQuotaRepo{stepsByChain: map[uint]int64{
		100: 10, // at limit
		200: 5,  // within limit
	}}
	svc := newService(user, plan, q)

	if err := svc.CheckChainStepQuota(context.Background(), 1, 100); err == nil {
		t.Fatal("chain 100 at limit — expected error")
	}
	if err := svc.CheckChainStepQuota(context.Background(), 1, 200); err != nil {
		t.Fatalf("chain 200 within limit — unexpected error %v", err)
	}
}

// IsMaxTierUser возвращает false при ошибке загрузки плана (defence-in-depth).
func TestIsMaxTierUser_PlanLoadError_False(t *testing.T) {
	user := &models.User{ID: 1, PlanID: "nonexistent"} // план не зарегистрирован
	plan := &models.SubscriptionPlan{ID: "max"}
	svc := newService(user, plan, &fakeQuotaRepo{})

	if got := svc.IsMaxTierUser(context.Background(), 1); got {
		t.Fatal("ожидался false при невозможности загрузить план юзера, got true")
	}
}

// --- GetUsageSummary ---

func TestGetUsageSummary_ReturnsAllCounters(t *testing.T) {
	user := &models.User{ID: 1, PlanID: "free"}
	plan := &models.SubscriptionPlan{
		ID:              "free",
		MaxPrompts:      50,
		MaxCollections:  3,
		MaxTeams:        1,
		MaxExtUsesDaily: 5,
		MaxMCPUsesDaily: 5,
	}
	q := &fakeQuotaRepo{
		prompts:     12,
		collections: 2,
		teamsOwned:  1,
		dailyUsage:  map[string]int{FeatureExtension: 1, FeatureMCP: 0},
	}
	svc := newService(user, plan, q)

	summary, err := svc.GetUsageSummary(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.PlanID != "free" {
		t.Fatalf("plan_id = %q, want free", summary.PlanID)
	}
	if summary.Prompts.Used != 12 || summary.Prompts.Limit != 50 {
		t.Fatalf("Prompts = %+v, want used=12 limit=50", summary.Prompts)
	}
	if summary.ExtUsesToday.Used != 1 || summary.ExtUsesToday.Limit != 5 {
		t.Fatalf("ExtUsesToday = %+v, want used=1 limit=5", summary.ExtUsesToday)
	}
	if summary.MCPUsesToday.Used != 0 || summary.MCPUsesToday.Limit != 5 {
		t.Fatalf("MCPUsesToday = %+v, want used=0 limit=5", summary.MCPUsesToday)
	}
}
