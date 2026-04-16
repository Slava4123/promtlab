package quota

import (
	"context"
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
func (r *fakeUserRepo) Update(context.Context, *models.User) error { panic("not used") }
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

func (r *fakeQuotaRepo) CountPrompts(context.Context, uint) (int64, error) {
	return r.prompts, nil
}
func (r *fakeQuotaRepo) CountCollections(context.Context, uint) (int64, error) {
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
		{"unlimited (-1) always allows", 1_000_000, -1, true},
		{"used < limit allows", 9, 10, true},
		{"used = limit blocks (strict <)", 10, 10, false},
		{"used > limit blocks", 11, 10, false},
		{"zero limit blocks any use", 0, 0, false},
		{"zero used with limit=0 blocks", 0, 0, false},
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
		{"unlimited plan", 9_999_999, -1, false},
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

// --- CheckAIQuota (total vs daily — критично для Free) ---

func TestCheckAIQuota_Total(t *testing.T) {
	user := &models.User{ID: 1, PlanID: "free"}
	plan := &models.SubscriptionPlan{
		ID:                 "free",
		MaxAIRequestsDaily: 5,
		AIRequestsIsTotal:  true,
	}

	cases := []struct {
		name    string
		total   int
		wantErr bool
		quotaT  string
	}{
		{"free 0/5 allows", 0, false, ""},
		{"free 4/5 allows", 4, false, ""},
		{"free 5/5 blocks with ai_total", 5, true, "ai_total"},
		{"free 10/5 blocks with ai_total", 10, true, "ai_total"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := &fakeQuotaRepo{totalUsage: map[string]int{FeatureAI: tc.total}}
			svc := newService(user, plan, q)
			err := svc.CheckAIQuota(context.Background(), 1)
			if tc.wantErr {
				var qe *QuotaExceededError
				if !errors.As(err, &qe) {
					t.Fatalf("want QuotaExceededError, got %v", err)
				}
				if qe.QuotaType != tc.quotaT {
					t.Fatalf("quota_type = %q, want %q", qe.QuotaType, tc.quotaT)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCheckAIQuota_Daily(t *testing.T) {
	user := &models.User{ID: 1, PlanID: "pro"}
	plan := &models.SubscriptionPlan{
		ID:                 "pro",
		MaxAIRequestsDaily: 10,
		AIRequestsIsTotal:  false,
	}

	cases := []struct {
		name    string
		daily   int
		wantErr bool
		quotaT  string
	}{
		{"pro 0/10 allows", 0, false, ""},
		{"pro 9/10 allows", 9, false, ""},
		{"pro 10/10 blocks with ai_daily", 10, true, "ai_daily"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := &fakeQuotaRepo{dailyUsage: map[string]int{FeatureAI: tc.daily}}
			svc := newService(user, plan, q)
			err := svc.CheckAIQuota(context.Background(), 1)
			if tc.wantErr {
				var qe *QuotaExceededError
				if !errors.As(err, &qe) {
					t.Fatalf("want QuotaExceededError, got %v", err)
				}
				if qe.QuotaType != tc.quotaT {
					t.Fatalf("quota_type = %q, want %q", qe.QuotaType, tc.quotaT)
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

// --- Increment ---

func TestIncrementAIUsage_WritesFeatureAI(t *testing.T) {
	user := &models.User{ID: 1, PlanID: "pro"}
	plan := &models.SubscriptionPlan{ID: "pro"}
	q := &fakeQuotaRepo{}
	svc := newService(user, plan, q)

	if err := svc.IncrementAIUsage(context.Background(), 42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.incrementLog) != 1 || q.incrementLog[0].feature != FeatureAI || q.incrementLog[0].userID != 42 {
		t.Fatalf("expected one Increment call for user=42 feature=ai, got %+v", q.incrementLog)
	}
}

// --- GetUsageSummary ---

func TestGetUsageSummary_ReturnsIsTotalForFree(t *testing.T) {
	user := &models.User{ID: 1, PlanID: "free"}
	plan := &models.SubscriptionPlan{
		ID:                 "free",
		MaxPrompts:         50,
		MaxCollections:     3,
		MaxAIRequestsDaily: 5,
		AIRequestsIsTotal:  true,
		MaxTeams:           1,
		MaxShareLinks:      2,
		MaxExtUsesDaily:    5,
		MaxMCPUsesDaily:    5,
	}
	q := &fakeQuotaRepo{
		prompts:     12,
		collections: 2,
		teamsOwned:  1,
		shareLinks:  0,
		totalUsage:  map[string]int{FeatureAI: 3},
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
	if !summary.AIRequests.IsTotal {
		t.Fatalf("AIRequests.IsTotal = false, want true для Free плана")
	}
	if summary.AIRequests.Used != 3 || summary.AIRequests.Limit != 5 {
		t.Fatalf("AIRequests = %+v, want used=3 limit=5", summary.AIRequests)
	}
	if summary.Prompts.Used != 12 || summary.Prompts.Limit != 50 {
		t.Fatalf("Prompts = %+v, want used=12 limit=50", summary.Prompts)
	}
}
