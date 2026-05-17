package prompt_insights

import (
	"context"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// Compile-time assertion: *fakeAnalyticsRepo должен удовлетворять
// repo.AnalyticsRepository. Если интерфейс расширится — пакет не скомпилируется,
// и тесты явно покажут, какой метод-стаб нужно добавить.
var _ repo.AnalyticsRepository = (*fakeAnalyticsRepo)(nil)

// --- USAGE metrics ---

func (f *fakeAnalyticsRepo) UsagePerDay(ctx context.Context, userID uint, teamID *uint, r repo.DateRange) ([]repo.UsagePoint, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) TopPrompts(ctx context.Context, userID uint, teamID *uint, r repo.DateRange, limit int) ([]repo.PromptUsageRow, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) GetTrendingPrompts(ctx context.Context, userID uint, teamID *uint, factor float64, growing bool, limit int) ([]repo.TrendRow, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) PromptUsageTimeline(ctx context.Context, promptID uint, r repo.DateRange) ([]repo.UsagePoint, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) PromptShareViewsTimeline(ctx context.Context, promptID uint, r repo.DateRange) ([]repo.UsagePoint, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) UsageByModel(ctx context.Context, userID uint, teamID *uint, r repo.DateRange) ([]repo.ModelUsageRow, error) {
	return nil, nil
}

// --- CREATION activity ---

func (f *fakeAnalyticsRepo) PromptsCreatedPerDay(ctx context.Context, userID uint, teamID *uint, r repo.DateRange) ([]repo.UsagePoint, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) PromptsUpdatedPerDay(ctx context.Context, userID uint, teamID *uint, r repo.DateRange) ([]repo.UsagePoint, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) Contributors(ctx context.Context, teamID uint, r repo.DateRange, limit int) ([]repo.ContributorRow, error) {
	return nil, nil
}

// --- SHARE perf ---

func (f *fakeAnalyticsRepo) ShareViewsPerDay(ctx context.Context, userID uint, r repo.DateRange) ([]repo.UsagePoint, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) TopSharedPrompts(ctx context.Context, userID uint, r repo.DateRange, limit int) ([]repo.PromptUsageRow, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) LogShareView(ctx context.Context, view *models.ShareView) error {
	return nil
}

// --- SMART INSIGHTS ---

func (f *fakeAnalyticsRepo) UpsertInsight(ctx context.Context, insight *models.SmartInsight) error {
	return nil
}

func (f *fakeAnalyticsRepo) GetInsights(ctx context.Context, userID uint, teamID *uint) ([]models.SmartInsight, error) {
	return nil, nil
}

// --- CLEANUP ---

func (f *fakeAnalyticsRepo) DeleteShareViewsOlderThan(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeAnalyticsRepo) CleanupShareViewsByRetention(ctx context.Context) (int64, error) {
	return 0, nil
}

func (f *fakeAnalyticsRepo) CleanupPromptUsageByRetention(ctx context.Context) (int64, error) {
	return 0, nil
}

// --- SMART INSIGHTS M8 ---

func (f *fakeAnalyticsRepo) MostEditedPrompts(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.PromptUsageRow, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) PossibleDuplicates(ctx context.Context, userID uint, teamID *uint, threshold float32, limit int) ([]repo.DuplicatePair, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) OrphanTags(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.TagRow, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) EmptyCollections(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.CollectionRow, error) {
	return nil, nil
}

// --- DRILL-DOWN filter-aware ---

func (f *fakeAnalyticsRepo) UsagePerDayFiltered(ctx context.Context, filter repo.AnalyticsFilter) ([]repo.UsagePoint, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) TopPromptsFiltered(ctx context.Context, filter repo.AnalyticsFilter, limit int) ([]repo.PromptUsageRow, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) PromptsCreatedPerDayFiltered(ctx context.Context, filter repo.AnalyticsFilter) ([]repo.UsagePoint, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) PromptsUpdatedPerDayFiltered(ctx context.Context, filter repo.AnalyticsFilter) ([]repo.UsagePoint, error) {
	return nil, nil
}

func (f *fakeAnalyticsRepo) UsageByModelFiltered(ctx context.Context, filter repo.AnalyticsFilter) ([]repo.ModelUsageRow, error) {
	return nil, nil
}
