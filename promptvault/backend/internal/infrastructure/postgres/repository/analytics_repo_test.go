package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// applyAnalyticsSchemaExtensions реплицирует prod-миграции, которых нет в
// GORM-модели PromptUsageLog: колонка team_id (миграция 000041) и уникальный
// expression-based индекс на user_smart_insights по (user_id, COALESCE(team_id,0),
// insight_type) (миграция 000043). Без этого UsagePerDay/TopPrompts с team-scope
// упадут, а UpsertInsight не сможет использовать ON CONFLICT.
//
// Паттерн повторяет applyAppendOnlyGuard из audit_repo_test.go.
func applyAnalyticsSchemaExtensions(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`ALTER TABLE prompt_usage_log ADD COLUMN IF NOT EXISTS team_id BIGINT`).Error)
	require.NoError(t, db.Exec(`CREATE INDEX IF NOT EXISTS idx_pul_team_id ON prompt_usage_log(team_id)`).Error)
	require.NoError(t, db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_usi_unique
		ON user_smart_insights (user_id, COALESCE(team_id, 0), insight_type)
	`).Error)
}

func newAnalyticsRepoTest(t *testing.T) (repo.AnalyticsRepository, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	applyAnalyticsSchemaExtensions(t, db)
	return NewAnalyticsRepository(db), db
}

// createTestUserWithPlan — userWithPlan. Базовый createTestUser из
// badge_repo_test.go не задаёт plan_id (дефолт 'free'), а для retention
// тестов нужны 'pro'/'max'.
func createTestUserWithPlan(t *testing.T, db *gorm.DB, email, plan string) *models.User {
	t.Helper()
	u := &models.User{
		Email:        email,
		Name:         "Plan User",
		PasswordHash: "irrelevant",
		PlanID:       plan,
	}
	require.NoError(t, db.Create(u).Error)
	return u
}

// insertPromptUsage — raw INSERT в prompt_usage_log. Модель PromptUsageLog
// не содержит поля TeamID (миграция 000041 добавляет колонку в prod, тест
// добавляет её через applyAnalyticsSchemaExtensions), поэтому GORM Create
// пропустил бы колонку. ModelUsed как *string: nil → NULL в БД, "" → "".
func insertPromptUsage(t *testing.T, db *gorm.DB, userID, promptID uint, teamID *uint, modelUsed *string, usedAt time.Time) {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO prompt_usage_log (user_id, prompt_id, team_id, model_used, used_at) VALUES (?, ?, ?, ?, ?)`,
		userID, promptID, teamID, modelUsed, usedAt,
	).Error)
}

// strPtr — удобный helper для nullable string в SQL.
func strPtr(s string) *string { return &s }

// --- GetTrendingPrompts ---

// TestAnalyticsRepo_GetTrendingPrompts проверяет SQL с двумя CTE (last_7d,
// prev_7d). growing=true/factor=2.0 — TRENDING: last >= prev*2 OR prev=NULL
// (новый рост-с-нуля). growing=false/factor=0.5 — DECLINING: prev>0 AND
// last<=prev*0.5.
func TestAnalyticsRepo_GetTrendingPrompts_Growing(t *testing.T) {
	r, db := newAnalyticsRepoTest(t)
	ctx := context.Background()

	u := createTestUser(t, db, "trending-growing@test.local")

	// A — новый промпт (0 uses в prev_7, 5 в last_7) → growing (prev=NULL).
	promptA := createTestPrompt(t, db, u.ID, nil, "A brand new", "", 0)
	// B — 2 uses prev_7, 6 uses last_7 → 6 >= 2*2 = 4 → growing.
	promptB := createTestPrompt(t, db, u.ID, nil, "B surging", "", 0)
	// C — 3 uses prev_7, 3 uses last_7 → 3 < 3*2 → НЕ growing.
	promptC := createTestPrompt(t, db, u.ID, nil, "C steady", "", 0)

	now := time.Now().UTC()
	last := now.Add(-2 * 24 * time.Hour) // внутри last_7d
	prev := now.Add(-10 * 24 * time.Hour) // внутри [-14d, -7d)

	for range 5 {
		insertPromptUsage(t, db, u.ID, promptA.ID, nil, nil, last)
	}
	for range 2 {
		insertPromptUsage(t, db, u.ID, promptB.ID, nil, nil, prev)
	}
	for range 6 {
		insertPromptUsage(t, db, u.ID, promptB.ID, nil, nil, last)
	}
	for range 3 {
		insertPromptUsage(t, db, u.ID, promptC.ID, nil, nil, prev)
	}
	for range 3 {
		insertPromptUsage(t, db, u.ID, promptC.ID, nil, nil, last)
	}

	rows, err := r.GetTrendingPrompts(ctx, u.ID, nil, 2.0, true, 10)
	require.NoError(t, err)

	ids := make(map[uint]repo.TrendRow)
	for _, row := range rows {
		ids[row.PromptID] = row
	}
	assert.Contains(t, ids, promptA.ID, "A brand-new промпт должен быть в trending (prev=NULL)")
	assert.Contains(t, ids, promptB.ID, "B surging промпт должен быть в trending (6 >= 2*2)")
	assert.NotContains(t, ids, promptC.ID, "C steady не должен быть в trending (3 < 3*2)")

	if rowB, ok := ids[promptB.ID]; ok {
		assert.Equal(t, int64(6), rowB.UsesLast, "uses_last должны посчитаться корректно")
		assert.Equal(t, int64(2), rowB.UsesPrevious, "uses_prev должны посчитаться корректно")
	}
}

func TestAnalyticsRepo_GetTrendingPrompts_Declining(t *testing.T) {
	r, db := newAnalyticsRepoTest(t)
	ctx := context.Background()

	u := createTestUser(t, db, "trending-declining@test.local")

	// D — 10 uses prev_7, 3 uses last_7 → 3 <= 10*0.5 = 5 → declining.
	promptD := createTestPrompt(t, db, u.ID, nil, "D falling", "", 0)
	// E — 0 uses prev_7, 5 uses last_7 → prev=NULL → НЕ declining (условие prev>0).
	promptE := createTestPrompt(t, db, u.ID, nil, "E new", "", 0)

	now := time.Now().UTC()
	last := now.Add(-3 * 24 * time.Hour)
	prev := now.Add(-10 * 24 * time.Hour)

	for range 10 {
		insertPromptUsage(t, db, u.ID, promptD.ID, nil, nil, prev)
	}
	for range 3 {
		insertPromptUsage(t, db, u.ID, promptD.ID, nil, nil, last)
	}
	for range 5 {
		insertPromptUsage(t, db, u.ID, promptE.ID, nil, nil, last)
	}

	rows, err := r.GetTrendingPrompts(ctx, u.ID, nil, 0.5, false, 10)
	require.NoError(t, err)

	ids := make(map[uint]struct{})
	for _, row := range rows {
		ids[row.PromptID] = struct{}{}
	}
	assert.Contains(t, ids, promptD.ID, "D должен быть в declining")
	assert.NotContains(t, ids, promptE.ID, "E с prev=NULL не должен быть в declining")
}

func TestAnalyticsRepo_GetTrendingPrompts_EmptyResult(t *testing.T) {
	r, db := newAnalyticsRepoTest(t)
	ctx := context.Background()

	u := createTestUser(t, db, "trending-empty@test.local")
	_ = createTestPrompt(t, db, u.ID, nil, "unused", "", 0)

	rows, err := r.GetTrendingPrompts(ctx, u.ID, nil, 2.0, true, 10)
	require.NoError(t, err)
	assert.Empty(t, rows, "без usage данных trending должен быть пуст")
}

func TestAnalyticsRepo_GetTrendingPrompts_TeamScope(t *testing.T) {
	r, db := newAnalyticsRepoTest(t)
	ctx := context.Background()

	u := createTestUser(t, db, "trending-team@test.local")
	team := createTestTeam(t, db, u.ID)
	personalPrompt := createTestPrompt(t, db, u.ID, nil, "personal", "", 0)
	teamPrompt := createTestPrompt(t, db, u.ID, &team.ID, "team", "", 0)

	now := time.Now().UTC()
	last := now.Add(-1 * 24 * time.Hour)

	// personal prompt — 5 uses в last_7d.
	for range 5 {
		insertPromptUsage(t, db, u.ID, personalPrompt.ID, nil, nil, last)
	}
	// team prompt — 3 uses в last_7d с team_id.
	for range 3 {
		insertPromptUsage(t, db, u.ID, teamPrompt.ID, &team.ID, nil, last)
	}

	// Personal scope.
	personalRows, err := r.GetTrendingPrompts(ctx, u.ID, nil, 2.0, true, 10)
	require.NoError(t, err)
	assert.Len(t, personalRows, 1, "в личном скоупе должен быть только personal prompt")
	if len(personalRows) > 0 {
		assert.Equal(t, personalPrompt.ID, personalRows[0].PromptID)
	}

	// Team scope.
	teamRows, err := r.GetTrendingPrompts(ctx, u.ID, &team.ID, 2.0, true, 10)
	require.NoError(t, err)
	assert.Len(t, teamRows, 1, "в team скоупе должен быть только team prompt")
	if len(teamRows) > 0 {
		assert.Equal(t, teamPrompt.ID, teamRows[0].PromptID)
	}
}

// --- CleanupPromptUsageByRetention ---

// TestAnalyticsRepo_CleanupPromptUsageByRetention — DELETE по plan_id юзера.
// Free=30д, Pro/Pro_yearly=90д, Max/Max_yearly=365д. Свежие записи остаются,
// старые — удаляются; подтверждаем счётчик RowsAffected.
func TestAnalyticsRepo_CleanupPromptUsageByRetention(t *testing.T) {
	r, db := newAnalyticsRepoTest(t)
	ctx := context.Background()

	free := createTestUserWithPlan(t, db, "cleanup-free@test.local", "free")
	pro := createTestUserWithPlan(t, db, "cleanup-pro@test.local", "pro")
	max := createTestUserWithPlan(t, db, "cleanup-max@test.local", "max")

	freePrompt := createTestPrompt(t, db, free.ID, nil, "free p", "", 0)
	proPrompt := createTestPrompt(t, db, pro.ID, nil, "pro p", "", 0)
	maxPrompt := createTestPrompt(t, db, max.ID, nil, "max p", "", 0)

	now := time.Now().UTC()
	// Old — за пределами retention.
	oldFree := now.Add(-31 * 24 * time.Hour)
	oldPro := now.Add(-91 * 24 * time.Hour)
	oldMax := now.Add(-366 * 24 * time.Hour)
	// New — внутри retention.
	newFree := now.Add(-1 * 24 * time.Hour)
	newPro := now.Add(-89 * 24 * time.Hour)
	newMax := now.Add(-364 * 24 * time.Hour)

	// Seed old + new для каждого юзера.
	insertPromptUsage(t, db, free.ID, freePrompt.ID, nil, nil, oldFree)
	insertPromptUsage(t, db, free.ID, freePrompt.ID, nil, nil, newFree)
	insertPromptUsage(t, db, pro.ID, proPrompt.ID, nil, nil, oldPro)
	insertPromptUsage(t, db, pro.ID, proPrompt.ID, nil, nil, newPro)
	insertPromptUsage(t, db, max.ID, maxPrompt.ID, nil, nil, oldMax)
	insertPromptUsage(t, db, max.ID, maxPrompt.ID, nil, nil, newMax)

	deleted, err := r.CleanupPromptUsageByRetention(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), deleted, "должны быть удалены 3 старые записи (по одной на каждый план)")

	var count int64
	require.NoError(t, db.Table("prompt_usage_log").Count(&count).Error)
	assert.Equal(t, int64(3), count, "должны остаться 3 свежие записи")

	// Проверяем что конкретно свежие остались.
	var remainingFree int64
	require.NoError(t, db.Table("prompt_usage_log").
		Where("user_id = ?", free.ID).
		Count(&remainingFree).Error)
	assert.Equal(t, int64(1), remainingFree, "у free должна остаться свежая запись")
}

// TestAnalyticsRepo_CleanupPromptUsageByRetention_YearlyPlans проверяет что
// pro_yearly/max_yearly попадают под те же правила что pro/max.
func TestAnalyticsRepo_CleanupPromptUsageByRetention_YearlyPlans(t *testing.T) {
	r, db := newAnalyticsRepoTest(t)
	ctx := context.Background()

	proY := createTestUserWithPlan(t, db, "cleanup-pro-yearly@test.local", "pro_yearly")
	maxY := createTestUserWithPlan(t, db, "cleanup-max-yearly@test.local", "max_yearly")

	proPrompt := createTestPrompt(t, db, proY.ID, nil, "pro_y p", "", 0)
	maxPrompt := createTestPrompt(t, db, maxY.ID, nil, "max_y p", "", 0)

	now := time.Now().UTC()
	insertPromptUsage(t, db, proY.ID, proPrompt.ID, nil, nil, now.Add(-91*24*time.Hour)) // удалится
	insertPromptUsage(t, db, proY.ID, proPrompt.ID, nil, nil, now.Add(-1*24*time.Hour))  // останется
	insertPromptUsage(t, db, maxY.ID, maxPrompt.ID, nil, nil, now.Add(-366*24*time.Hour)) // удалится
	insertPromptUsage(t, db, maxY.ID, maxPrompt.ID, nil, nil, now.Add(-1*24*time.Hour))   // останется

	deleted, err := r.CleanupPromptUsageByRetention(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), deleted, "yearly планы должны подчиняться тем же retention")
}

// --- PromptUsageTimeline ---

// TestAnalyticsRepo_PromptUsageTimeline — per-prompt таймсерия с day-precision.
// Проверяем: WHERE prompt_id = ?, range [From, To), GROUP BY день в UTC,
// ORDER BY day ASC.
func TestAnalyticsRepo_PromptUsageTimeline(t *testing.T) {
	r, db := newAnalyticsRepoTest(t)
	ctx := context.Background()

	u := createTestUser(t, db, "timeline@test.local")
	promptA := createTestPrompt(t, db, u.ID, nil, "A", "", 0)
	promptB := createTestPrompt(t, db, u.ID, nil, "B noise", "", 0)

	now := time.Now().UTC()
	day1 := now.Add(-2 * 24 * time.Hour)
	day2 := now.Add(-1 * 24 * time.Hour)
	outside := now.Add(-30 * 24 * time.Hour) // вне range

	// A: 3 uses в day1, 2 uses в day2, 1 use вне range.
	for range 3 {
		insertPromptUsage(t, db, u.ID, promptA.ID, nil, nil, day1)
	}
	for range 2 {
		insertPromptUsage(t, db, u.ID, promptA.ID, nil, nil, day2)
	}
	insertPromptUsage(t, db, u.ID, promptA.ID, nil, nil, outside)
	// B: 5 uses в day1 — не должен попасть (WHERE prompt_id=A).
	for range 5 {
		insertPromptUsage(t, db, u.ID, promptB.ID, nil, nil, day1)
	}

	rng := repo.DateRange{
		From: now.Add(-7 * 24 * time.Hour),
		To:   now.Add(1 * time.Hour), // немного в будущее, чтобы захватить текущие события
	}
	points, err := r.PromptUsageTimeline(ctx, promptA.ID, rng)
	require.NoError(t, err)

	require.Len(t, points, 2, "должно быть 2 точки (day1, day2)")
	assert.Equal(t, int64(3), points[0].Count, "day1 = 3 uses")
	assert.Equal(t, int64(2), points[1].Count, "day2 = 2 uses")
	assert.True(t, points[0].Day.Before(points[1].Day), "ORDER BY day ASC")
}

func TestAnalyticsRepo_PromptUsageTimeline_EmptyRange(t *testing.T) {
	r, db := newAnalyticsRepoTest(t)
	ctx := context.Background()

	u := createTestUser(t, db, "timeline-empty@test.local")
	p := createTestPrompt(t, db, u.ID, nil, "P", "", 0)

	now := time.Now().UTC()
	// Seed запись сегодня, но запрашиваем диапазон 30 дней назад.
	insertPromptUsage(t, db, u.ID, p.ID, nil, nil, now)

	rng := repo.DateRange{
		From: now.Add(-60 * 24 * time.Hour),
		To:   now.Add(-30 * 24 * time.Hour),
	}
	points, err := r.PromptUsageTimeline(ctx, p.ID, rng)
	require.NoError(t, err)
	assert.Empty(t, points, "вне диапазона — пустой результат")
}

// --- UsageByModel ---

// TestAnalyticsRepo_UsageByModel — сегментация по model_used. NULL в БД
// мапится в "" через COALESCE в SQL. ORDER BY uses DESC.
func TestAnalyticsRepo_UsageByModel(t *testing.T) {
	r, db := newAnalyticsRepoTest(t)
	ctx := context.Background()

	u := createTestUser(t, db, "bymodel@test.local")
	p := createTestPrompt(t, db, u.ID, nil, "P", "", 0)

	now := time.Now().UTC()
	in := now.Add(-1 * 24 * time.Hour)

	// gpt-4: 3 раза, claude: 2 раза, NULL: 1 раз, outside: 5 раз (не в range).
	for range 3 {
		insertPromptUsage(t, db, u.ID, p.ID, nil, strPtr("gpt-4"), in)
	}
	for range 2 {
		insertPromptUsage(t, db, u.ID, p.ID, nil, strPtr("claude-sonnet-4"), in)
	}
	insertPromptUsage(t, db, u.ID, p.ID, nil, nil, in) // NULL model

	outside := now.Add(-30 * 24 * time.Hour)
	for range 5 {
		insertPromptUsage(t, db, u.ID, p.ID, nil, strPtr("gpt-4"), outside)
	}

	rng := repo.DateRange{
		From: now.Add(-7 * 24 * time.Hour),
		To:   now.Add(1 * time.Hour),
	}
	rows, err := r.UsageByModel(ctx, u.ID, nil, rng)
	require.NoError(t, err)
	require.Len(t, rows, 3, "должно быть 3 различных model-значения")

	// Проверяем ORDER BY uses DESC.
	assert.Equal(t, "gpt-4", rows[0].Model)
	assert.Equal(t, int64(3), rows[0].Uses)
	assert.Equal(t, "claude-sonnet-4", rows[1].Model)
	assert.Equal(t, int64(2), rows[1].Uses)
	assert.Equal(t, "", rows[2].Model, "NULL model_used должен мапиться в пустую строку")
	assert.Equal(t, int64(1), rows[2].Uses)
}

func TestAnalyticsRepo_UsageByModel_TeamScope(t *testing.T) {
	r, db := newAnalyticsRepoTest(t)
	ctx := context.Background()

	u := createTestUser(t, db, "bymodel-team@test.local")
	team := createTestTeam(t, db, u.ID)
	personal := createTestPrompt(t, db, u.ID, nil, "personal", "", 0)
	teamP := createTestPrompt(t, db, u.ID, &team.ID, "team", "", 0)

	now := time.Now().UTC()
	in := now.Add(-1 * 24 * time.Hour)

	insertPromptUsage(t, db, u.ID, personal.ID, nil, strPtr("gpt-4"), in)
	insertPromptUsage(t, db, u.ID, teamP.ID, &team.ID, strPtr("claude-sonnet-4"), in)

	rng := repo.DateRange{
		From: now.Add(-7 * 24 * time.Hour),
		To:   now.Add(1 * time.Hour),
	}

	personalRows, err := r.UsageByModel(ctx, u.ID, nil, rng)
	require.NoError(t, err)
	require.Len(t, personalRows, 1)
	assert.Equal(t, "gpt-4", personalRows[0].Model)

	teamRows, err := r.UsageByModel(ctx, u.ID, &team.ID, rng)
	require.NoError(t, err)
	require.Len(t, teamRows, 1)
	assert.Equal(t, "claude-sonnet-4", teamRows[0].Model)
}
