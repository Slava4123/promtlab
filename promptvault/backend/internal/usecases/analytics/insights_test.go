package analytics

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// trackingAnalyticsRepo — фейк AnalyticsRepository, который записывает только
// факт вызова каждого метода, нужного ComputeInsights. Достаточно чтобы
// проверить ветки feature-flag и pg_trgm probe.
type trackingAnalyticsRepo struct {
	calls map[string]int
	mu    sync.Mutex
	// MN-13: failOnUserID — если != 0, вызовы UnusedPrompts для этого юзера
	// возвращают ошибку. Используется для resilience-теста errgroup'а
	// (один failed user не должен блокировать остальные).
	failOnUserID uint
	// insightsAll — Task 6: возвращается из GetInsights. Используется тестами
	// GetInsightsGated для проверки per-type filter'а (repo отдаёт все 7 типов,
	// service фильтрует по plan'у).
	insightsAll []models.SmartInsight
}

func newTrackingRepo() *trackingAnalyticsRepo {
	return &trackingAnalyticsRepo{calls: make(map[string]int)}
}

func (r *trackingAnalyticsRepo) UnusedPrompts(_ context.Context, userID uint, _ *uint, _ time.Time, _ int) ([]repo.PromptUsageRow, error) {
	r.mu.Lock()
	r.calls["UnusedPrompts"]++
	r.mu.Unlock()
	if r.failOnUserID != 0 && userID == r.failOnUserID {
		return nil, errors.New("compute insights failed for user " + uintToStr(userID))
	}
	return []repo.PromptUsageRow{{PromptID: 1, Title: "x", Uses: 0}}, nil
}

func uintToStr(v uint) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
// inc — thread-safe инкремент r.calls[key]. Все методы trackingAnalyticsRepo
// вызываются из errgroup-горутин (compute() параллелит per-user), без mutex
// race detector ловит конкурентную запись в map (Go 1.25.10).
func (r *trackingAnalyticsRepo) inc(key string) {
	r.mu.Lock()
	r.calls[key]++
	r.mu.Unlock()
}

func (r *trackingAnalyticsRepo) GetTrendingPrompts(_ context.Context, _ uint, _ *uint, _ float64, _ bool, _ int) ([]repo.TrendRow, error) {
	r.inc("GetTrendingPrompts")
	return []repo.TrendRow{{PromptID: 1, Title: "x", UsesLast: 5, UsesPrevious: 1}}, nil
}
func (r *trackingAnalyticsRepo) MostEditedPrompts(_ context.Context, _ uint, _ *uint, _ int) ([]repo.PromptUsageRow, error) {
	r.inc("MostEditedPrompts")
	return []repo.PromptUsageRow{{PromptID: 1, Title: "x", Uses: 0}}, nil
}
func (r *trackingAnalyticsRepo) PossibleDuplicates(_ context.Context, _ uint, _ *uint, _ float32, _ int) ([]repo.DuplicatePair, error) {
	r.inc("PossibleDuplicates")
	return []repo.DuplicatePair{{PromptAID: 1, PromptBID: 2, Similarity: 0.9}}, nil
}
func (r *trackingAnalyticsRepo) OrphanTags(_ context.Context, _ uint, _ *uint, _ int) ([]repo.TagRow, error) {
	r.inc("OrphanTags")
	return []repo.TagRow{{TagID: 1, Name: "x"}}, nil
}
func (r *trackingAnalyticsRepo) EmptyCollections(_ context.Context, _ uint, _ *uint, _ int) ([]repo.CollectionRow, error) {
	r.inc("EmptyCollections")
	return []repo.CollectionRow{{CollectionID: 1, Name: "x"}}, nil
}
func (r *trackingAnalyticsRepo) UpsertInsight(_ context.Context, in *models.SmartInsight) error {
	r.inc("UpsertInsight:" + in.InsightType)
	return nil
}

// ----- Не используются ComputeInsights, но нужны для удовлетворения интерфейса. -----

func (r *trackingAnalyticsRepo) UsagePerDay(context.Context, uint, *uint, repo.DateRange) ([]repo.UsagePoint, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) TopPrompts(context.Context, uint, *uint, repo.DateRange, int) ([]repo.PromptUsageRow, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) PromptUsageTimeline(context.Context, uint, repo.DateRange) ([]repo.UsagePoint, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) PromptShareViewsTimeline(context.Context, uint, repo.DateRange) ([]repo.UsagePoint, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) UsageByModel(context.Context, uint, *uint, repo.DateRange) ([]repo.ModelUsageRow, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) PromptsCreatedPerDay(context.Context, uint, *uint, repo.DateRange) ([]repo.UsagePoint, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) PromptsUpdatedPerDay(context.Context, uint, *uint, repo.DateRange) ([]repo.UsagePoint, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) Contributors(context.Context, uint, repo.DateRange, int) ([]repo.ContributorRow, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) ShareViewsPerDay(context.Context, uint, repo.DateRange) ([]repo.UsagePoint, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) TopSharedPrompts(context.Context, uint, repo.DateRange, int) ([]repo.PromptUsageRow, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) LogShareView(context.Context, *models.ShareView) error {
	return nil
}
func (r *trackingAnalyticsRepo) GetInsights(context.Context, uint, *uint) ([]models.SmartInsight, error) {
	r.inc("GetInsights")
	return r.insightsAll, nil
}
func (r *trackingAnalyticsRepo) DeleteShareViewsOlderThan(context.Context, time.Time) (int64, error) {
	return 0, nil
}
func (r *trackingAnalyticsRepo) CleanupShareViewsByRetention(context.Context) (int64, error) {
	return 0, nil
}
func (r *trackingAnalyticsRepo) CleanupPromptUsageByRetention(context.Context) (int64, error) {
	return 0, nil
}
func (r *trackingAnalyticsRepo) UsagePerDayFiltered(context.Context, repo.AnalyticsFilter) ([]repo.UsagePoint, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) TopPromptsFiltered(context.Context, repo.AnalyticsFilter, int) ([]repo.PromptUsageRow, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) PromptsCreatedPerDayFiltered(context.Context, repo.AnalyticsFilter) ([]repo.UsagePoint, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) PromptsUpdatedPerDayFiltered(context.Context, repo.AnalyticsFilter) ([]repo.UsagePoint, error) {
	return nil, nil
}
func (r *trackingAnalyticsRepo) UsageByModelFiltered(context.Context, repo.AnalyticsFilter) ([]repo.ModelUsageRow, error) {
	return nil, nil
}

// ===== ТЕСТЫ =====

func newServiceForTest(repo *trackingAnalyticsRepo, experimental, trgm bool) *Service {
	s := NewService(repo, nil, nil, nil, nil)
	s.SetExperimentalInsights(experimental)
	s.SetTrgmAvailable(trgm)
	return s
}

func TestComputeInsights_AllSevenTypesWhenTrgmAvailable(t *testing.T) {
	r := newTrackingRepo()
	s := newServiceForTest(r, true, true)

	require.NoError(t, s.ComputeInsights(context.Background(), 42, nil, maxAllInsights))

	// 3 базовых типа
	assert.Equal(t, 1, r.calls["UnusedPrompts"])
	assert.Equal(t, 2, r.calls["GetTrendingPrompts"], "trending + declining = 2 вызова")
	// 4 расширенных типа
	assert.Equal(t, 1, r.calls["MostEditedPrompts"])
	assert.Equal(t, 1, r.calls["PossibleDuplicates"])
	assert.Equal(t, 1, r.calls["OrphanTags"])
	assert.Equal(t, 1, r.calls["EmptyCollections"])
	// upsert по 7 типам
	assert.Equal(t, 1, r.calls["UpsertInsight:"+models.InsightUnusedPrompts])
	assert.Equal(t, 1, r.calls["UpsertInsight:"+models.InsightTrending])
	assert.Equal(t, 1, r.calls["UpsertInsight:"+models.InsightDeclining])
	assert.Equal(t, 1, r.calls["UpsertInsight:"+models.InsightMostEdited])
	assert.Equal(t, 1, r.calls["UpsertInsight:"+models.InsightPossibleDuplicates])
	assert.Equal(t, 1, r.calls["UpsertInsight:"+models.InsightOrphanTags])
	assert.Equal(t, 1, r.calls["UpsertInsight:"+models.InsightEmptyCollections])
}

func TestComputeInsights_SkipsDuplicatesWhenTrgmUnavailable(t *testing.T) {
	r := newTrackingRepo()
	s := newServiceForTest(r, true, false) // experimental=true, trgm=false

	require.NoError(t, s.ComputeInsights(context.Background(), 42, nil, maxAllInsights))

	// PossibleDuplicates НЕ вызван
	assert.Equal(t, 0, r.calls["PossibleDuplicates"])
	assert.Equal(t, 0, r.calls["UpsertInsight:"+models.InsightPossibleDuplicates])
	// Остальные расширенные работают
	assert.Equal(t, 1, r.calls["MostEditedPrompts"])
	assert.Equal(t, 1, r.calls["OrphanTags"])
	assert.Equal(t, 1, r.calls["EmptyCollections"])
}

func TestComputeInsights_KillSwitchOff_OnlyThreeBaseTypes(t *testing.T) {
	r := newTrackingRepo()
	s := newServiceForTest(r, false, true) // kill-switch on

	require.NoError(t, s.ComputeInsights(context.Background(), 42, nil, maxAllInsights))

	// Только 3 базовых типа
	assert.Equal(t, 1, r.calls["UnusedPrompts"])
	assert.Equal(t, 2, r.calls["GetTrendingPrompts"])
	// Расширенные не вызываются
	assert.Equal(t, 0, r.calls["MostEditedPrompts"])
	assert.Equal(t, 0, r.calls["PossibleDuplicates"])
	assert.Equal(t, 0, r.calls["OrphanTags"])
	assert.Equal(t, 0, r.calls["EmptyCollections"])
}

// TestLookupPlanID_NoRecursion — регрессионный тест на bug, найденный в self-review:
// global sed заменил `s.users.GetByID` → `s.lookupPlanID` внутри самого helper'а,
// что вызывало infinite recursion → stack overflow в production. Тест дёргает
// helper напрямую и проверяет что он завершается без OOM/stack overflow.
func TestLookupPlanID_NoRecursion(t *testing.T) {
	users := &fakeUsersForLookup{user: &models.User{ID: 42, PlanID: "max"}}
	s := &Service{users: users}

	// 1. Без callback — должен пойти в DB.
	plan, err := s.lookupPlanID(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, "max", plan)
	assert.Equal(t, 1, users.calls)

	// 2. С callback который возвращает ok=true — DB не дёргается.
	s.SetPlanFromCtx(func(_ context.Context) (string, bool) { return "pro", true })
	plan2, err2 := s.lookupPlanID(context.Background(), 42)
	require.NoError(t, err2)
	assert.Equal(t, "pro", plan2)
	assert.Equal(t, 1, users.calls, "callback hit — DB не вызвана повторно")

	// 3. Callback вернул ok=false — fallback на DB.
	s.SetPlanFromCtx(func(_ context.Context) (string, bool) { return "", false })
	plan3, err3 := s.lookupPlanID(context.Background(), 42)
	require.NoError(t, err3)
	assert.Equal(t, "max", plan3)
	assert.Equal(t, 2, users.calls)
}

// fakeUsersForLookup — минимальный stub для регрессионного теста recursion.
type fakeUsersForLookup struct {
	user  *models.User
	calls int
}

func (f *fakeUsersForLookup) GetByID(_ context.Context, _ uint) (*models.User, error) {
	f.calls++
	return f.user, nil
}
func (f *fakeUsersForLookup) Create(context.Context, *models.User) error { panic("unused") }
func (f *fakeUsersForLookup) GetByEmail(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLookup) GetByUsername(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLookup) SearchUsers(context.Context, string, int) ([]models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLookup) Update(context.Context, *models.User) error  { panic("unused") }
func (f *fakeUsersForLookup) SetPlan(context.Context, uint, string) error { panic("unused") }
func (f *fakeUsersForLookup) SetQuotaWarningSentOn(context.Context, uint, time.Time) error {
	panic("unused")
}
func (f *fakeUsersForLookup) TouchLastLogin(context.Context, uint) error { panic("unused") }
func (f *fakeUsersForLookup) ListInactiveForReengagement(context.Context, time.Time, time.Time, int) ([]models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLookup) MarkReengagementSent(context.Context, uint) error { panic("unused") }
func (f *fakeUsersForLookup) CountReferredBy(context.Context, string) (int64, error) {
	panic("unused")
}
func (f *fakeUsersForLookup) GetByReferralCode(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLookup) MarkReferralRewarded(context.Context, uint) (bool, error) {
	panic("unused")
}
func (f *fakeUsersForLookup) ListMaxUsers(context.Context) ([]uint, error) { panic("unused") }
func (f *fakeUsersForLookup) SetInsightEmailsEnabled(context.Context, uint, bool) error {
	panic("unused")
}

// upsertedTypes возвращает список insight-типов, для которых был
// UpsertInsight (используется в Task 5 для проверки allowed-фильтра).
func (r *trackingAnalyticsRepo) upsertedTypes() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, 0, len(r.calls))
	const prefix = "UpsertInsight:"
	for k, v := range r.calls {
		if v > 0 && len(k) > len(prefix) && k[:len(prefix)] == prefix {
			out = append(out, k[len(prefix):])
		}
	}
	return out
}

// TestComputeInsights_FiltersByAllowed — Pricing Iteration v3 Task 5:
// ComputeInsights должен считать и upsert'ить только те типы, что переданы
// в allowed []string. Остальные skip'аются без SQL-запроса.
//
// Сценарий: allowed = Pro набор (unused + duplicates). Trending/declining/
// most_edited/orphan_tags/empty_collections — НЕ должны попасть в upsert.
func TestComputeInsights_FiltersByAllowed(t *testing.T) {
	r := newTrackingRepo()
	s := newServiceForTest(r, true, true) // experimental=true, trgm=true
	s.SetNowFn(func() time.Time { return time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC) })

	// Allowed только unused + duplicates (Pro набор).
	err := s.ComputeInsights(context.Background(), 42, nil, []string{
		models.InsightUnusedPrompts,
		models.InsightPossibleDuplicates,
	})
	require.NoError(t, err)

	upserted := r.upsertedTypes()
	require.ElementsMatch(t, []string{
		models.InsightUnusedPrompts,
		models.InsightPossibleDuplicates,
	}, upserted, "trending/declining/most_edited/orphan_tags/empty_collections должны быть skip'нуты — не в allowed list")

	// Дополнительно: репо-методы НЕ должны быть вызваны для не-allowed типов.
	assert.Equal(t, 0, r.calls["GetTrendingPrompts"], "trending/declining не вызваны — не в allowed")
	assert.Equal(t, 0, r.calls["MostEditedPrompts"], "most_edited не вызван — не в allowed")
	assert.Equal(t, 0, r.calls["OrphanTags"], "orphan_tags не вызван — не в allowed")
	assert.Equal(t, 0, r.calls["EmptyCollections"], "empty_collections не вызван — не в allowed")
}

// TestComputeInsights_EmptyAllowedNoOp — пустой allowed list → no-op (нет
// SQL-запросов, нет upsert'ов). Используется loop'ом для Free юзеров
// (хотя Task 7 заменит на ListPaidUsers, чтобы Free не доходили сюда).
func TestComputeInsights_EmptyAllowedNoOp(t *testing.T) {
	r := newTrackingRepo()
	s := newServiceForTest(r, true, true)

	require.NoError(t, s.ComputeInsights(context.Background(), 42, nil, nil))

	assert.Empty(t, r.upsertedTypes(), "пустой allowed → ни одного upsert'а")
	assert.Equal(t, 0, r.calls["UnusedPrompts"])
	assert.Equal(t, 0, r.calls["GetTrendingPrompts"])
	assert.Equal(t, 0, r.calls["MostEditedPrompts"])
	assert.Equal(t, 0, r.calls["PossibleDuplicates"])
	assert.Equal(t, 0, r.calls["OrphanTags"])
	assert.Equal(t, 0, r.calls["EmptyCollections"])
}

// TestInsightsForPlan — Pricing Iteration v3 Task 4: helper для маппинга
// plan_id → разрешённый набор insight типов. Free/unknown → nil, Pro → 2
// housekeeping типа (teaser), Max → все 7. Решение зафиксировано в ADR-0008.
func TestInsightsForPlan(t *testing.T) {
	cases := []struct {
		plan string
		want []string
	}{
		{"free", nil},
		{"pro", []string{models.InsightUnusedPrompts, models.InsightPossibleDuplicates}},
		{"pro_yearly", []string{models.InsightUnusedPrompts, models.InsightPossibleDuplicates}},
		{"max", allInsightTypes()},
		{"max_yearly", allInsightTypes()},
		{"unknown", nil},
	}
	for _, tc := range cases {
		t.Run(tc.plan, func(t *testing.T) {
			got := insightsForPlan(tc.plan)
			require.ElementsMatch(t, tc.want, got)
		})
	}
}

// helper для теста — все 7 типов.
func allInsightTypes() []string {
	return []string{
		models.InsightUnusedPrompts,
		models.InsightTrending,
		models.InsightDeclining,
		models.InsightMostEdited,
		models.InsightPossibleDuplicates,
		models.InsightOrphanTags,
		models.InsightEmptyCollections,
	}
}

// TestService_GetInsightsGated_PerTypeFilter — Pricing Iteration v3 Task 6:
// GetInsightsGated больше не master-gate (Pro/Free → 402), а per-type filter:
//   - Free → ErrProRequired (HTTP 402) — upgrade prompt.
//   - Pro/pro_yearly → 2 типа (unused + duplicates) teaser.
//   - Max/max_yearly → все 7 типов.
//
// Repo возвращает все 7 типов; service фильтрует по тарифу.
func TestService_GetInsightsGated_PerTypeFilter(t *testing.T) {
	ctx := context.Background()
	allTypes := []models.SmartInsight{
		{InsightType: models.InsightUnusedPrompts},
		{InsightType: models.InsightTrending},
		{InsightType: models.InsightDeclining},
		{InsightType: models.InsightMostEdited},
		{InsightType: models.InsightPossibleDuplicates},
		{InsightType: models.InsightOrphanTags},
		{InsightType: models.InsightEmptyCollections},
	}
	cases := []struct {
		plan          string
		wantErr       error
		wantTypeCount int
	}{
		{"free", ErrProRequired, 0},
		{"pro", nil, 2},
		{"pro_yearly", nil, 2},
		{"max", nil, 7},
		{"max_yearly", nil, 7},
	}
	for _, tc := range cases {
		t.Run(tc.plan, func(t *testing.T) {
			repo := &trackingAnalyticsRepo{
				calls:       make(map[string]int),
				insightsAll: allTypes,
			}
			users := &fakeUsersForLookup{user: &models.User{ID: 1, PlanID: tc.plan}}
			svc := NewService(repo, nil, nil, users, nil)
			// Не выставляем planFromCtx — lookupPlanID идёт через users.GetByID.
			insights, err := svc.GetInsightsGated(ctx, 1, nil)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, insights, tc.wantTypeCount)
		})
	}
}
