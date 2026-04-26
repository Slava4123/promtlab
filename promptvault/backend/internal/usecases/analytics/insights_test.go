package analytics

import (
	"context"
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
}

func newTrackingRepo() *trackingAnalyticsRepo {
	return &trackingAnalyticsRepo{calls: make(map[string]int)}
}

func (r *trackingAnalyticsRepo) UnusedPrompts(_ context.Context, _ uint, _ *uint, _ time.Time, _ int) ([]repo.PromptUsageRow, error) {
	r.calls["UnusedPrompts"]++
	return []repo.PromptUsageRow{{PromptID: 1, Title: "x", Uses: 0}}, nil
}
func (r *trackingAnalyticsRepo) GetTrendingPrompts(_ context.Context, _ uint, _ *uint, _ float64, _ bool, _ int) ([]repo.TrendRow, error) {
	r.calls["GetTrendingPrompts"]++
	return []repo.TrendRow{{PromptID: 1, Title: "x", UsesLast: 5, UsesPrevious: 1}}, nil
}
func (r *trackingAnalyticsRepo) MostEditedPrompts(_ context.Context, _ uint, _ *uint, _ int) ([]repo.PromptUsageRow, error) {
	r.calls["MostEditedPrompts"]++
	return []repo.PromptUsageRow{{PromptID: 1, Title: "x", Uses: 0}}, nil
}
func (r *trackingAnalyticsRepo) PossibleDuplicates(_ context.Context, _ uint, _ *uint, _ float32, _ int) ([]repo.DuplicatePair, error) {
	r.calls["PossibleDuplicates"]++
	return []repo.DuplicatePair{{PromptAID: 1, PromptBID: 2, Similarity: 0.9}}, nil
}
func (r *trackingAnalyticsRepo) OrphanTags(_ context.Context, _ uint, _ *uint, _ int) ([]repo.TagRow, error) {
	r.calls["OrphanTags"]++
	return []repo.TagRow{{TagID: 1, Name: "x"}}, nil
}
func (r *trackingAnalyticsRepo) EmptyCollections(_ context.Context, _ uint, _ *uint, _ int) ([]repo.CollectionRow, error) {
	r.calls["EmptyCollections"]++
	return []repo.CollectionRow{{CollectionID: 1, Name: "x"}}, nil
}
func (r *trackingAnalyticsRepo) UpsertInsight(_ context.Context, in *models.SmartInsight) error {
	r.calls["UpsertInsight:"+in.InsightType]++
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
	return nil, nil
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

	require.NoError(t, s.ComputeInsights(context.Background(), 42, nil))

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

	require.NoError(t, s.ComputeInsights(context.Background(), 42, nil))

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

	require.NoError(t, s.ComputeInsights(context.Background(), 42, nil))

	// Только 3 базовых типа
	assert.Equal(t, 1, r.calls["UnusedPrompts"])
	assert.Equal(t, 2, r.calls["GetTrendingPrompts"])
	// Расширенные не вызываются
	assert.Equal(t, 0, r.calls["MostEditedPrompts"])
	assert.Equal(t, 0, r.calls["PossibleDuplicates"])
	assert.Equal(t, 0, r.calls["OrphanTags"])
	assert.Equal(t, 0, r.calls["EmptyCollections"])
}
