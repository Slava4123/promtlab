package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type analyticsRepo struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) repo.AnalyticsRepository {
	return &analyticsRepo{db: db}
}

// scopeTeam добавляет WHERE team_id IS NULL / team_id = ? в зависимости
// от nil-ности teamID. prefix — имя таблицы/алиаса ("prompt_usage_log",
// "pul", "p", "sl.").
func scopeTeam(q *gorm.DB, column string, teamID *uint) *gorm.DB {
	if teamID == nil {
		return q.Where(column + " IS NULL")
	}
	return q.Where(column+" = ?", *teamID)
}

// --- USAGE metrics ---

func (r *analyticsRepo) UsagePerDay(ctx context.Context, userID uint, teamID *uint, rng repo.DateRange) ([]repo.UsagePoint, error) {
	q := r.db.WithContext(ctx).
		Table("prompt_usage_log").
		Select("date_trunc('day', used_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS count").
		Where("user_id = ? AND used_at >= ? AND used_at < ?", userID, rng.From, rng.To)
	q = scopeTeam(q, "team_id", teamID)

	var points []repo.UsagePoint
	err := q.Group("day").Order("day").Scan(&points).Error
	return points, err
}

func (r *analyticsRepo) TopPrompts(ctx context.Context, userID uint, teamID *uint, rng repo.DateRange, limit int) ([]repo.PromptUsageRow, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	q := r.db.WithContext(ctx).
		Table("prompt_usage_log AS pul").
		Select("pul.prompt_id AS prompt_id, p.title AS title, COUNT(*) AS uses").
		Joins("JOIN prompts p ON p.id = pul.prompt_id").
		Where("pul.user_id = ? AND pul.used_at >= ? AND pul.used_at < ?", userID, rng.From, rng.To).
		Where("p.deleted_at IS NULL")
	q = scopeTeam(q, "pul.team_id", teamID)

	var rows []repo.PromptUsageRow
	err := q.Group("pul.prompt_id, p.title").Order("uses DESC").Limit(limit).Scan(&rows).Error
	return rows, err
}

// UnusedPrompts — промпты с last_used_at < before (или NULL), но
// usage_count > 0 (чтобы отличать от never-used). Сортировка по
// last_used_at ASC NULLS FIRST — сверху те, что давно не трогали.
func (r *analyticsRepo) UnusedPrompts(ctx context.Context, userID uint, teamID *uint, before time.Time, limit int) ([]repo.PromptUsageRow, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	q := r.db.WithContext(ctx).
		Table("prompts").
		Select("id AS prompt_id, title AS title, usage_count AS uses").
		Where("user_id = ?", userID).
		Where("(last_used_at < ? OR last_used_at IS NULL)", before).
		Where("usage_count > 0").
		Where("deleted_at IS NULL")
	q = scopeTeam(q, "team_id", teamID)

	var rows []repo.PromptUsageRow
	err := q.Order("last_used_at ASC NULLS FIRST").Limit(limit).Scan(&rows).Error
	return rows, err
}

// --- CREATION activity ---

func (r *analyticsRepo) PromptsCreatedPerDay(ctx context.Context, userID uint, teamID *uint, rng repo.DateRange) ([]repo.UsagePoint, error) {
	q := r.db.WithContext(ctx).
		Table("prompts").
		Select("date_trunc('day', created_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS count").
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, rng.From, rng.To).
		Where("deleted_at IS NULL")
	q = scopeTeam(q, "team_id", teamID)

	var points []repo.UsagePoint
	err := q.Group("day").Order("day").Scan(&points).Error
	return points, err
}

// PromptsUpdatedPerDay использует prompt_versions как источник обновлений
// (каждая версия = update промпта). JOIN на prompts для team scope.
func (r *analyticsRepo) PromptsUpdatedPerDay(ctx context.Context, userID uint, teamID *uint, rng repo.DateRange) ([]repo.UsagePoint, error) {
	q := r.db.WithContext(ctx).
		Table("prompt_versions AS pv").
		Select("date_trunc('day', pv.created_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS count").
		Joins("JOIN prompts p ON p.id = pv.prompt_id").
		Where("p.user_id = ? AND pv.created_at >= ? AND pv.created_at < ?", userID, rng.From, rng.To).
		Where("p.deleted_at IS NULL")
	q = scopeTeam(q, "p.team_id", teamID)

	var points []repo.UsagePoint
	err := q.Group("day").Order("day").Scan(&points).Error
	return points, err
}

// Contributors — топ авторов команды по суммарной активности
// (prompts_created + prompts_edited + uses). Raw SQL с тремя
// LEFT JOIN subquery'ями через CTE-подобный шаблон — чище чем
// цепочка GORM-выражений.
func (r *analyticsRepo) Contributors(ctx context.Context, teamID uint, rng repo.DateRange, limit int) ([]repo.ContributorRow, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	const query = `
SELECT
    u.id    AS user_id,
    u.email AS email,
    u.name  AS name,
    COALESCE(created.cnt, 0) AS prompts_created,
    COALESCE(edited.cnt, 0)  AS prompts_edited,
    COALESCE(uses.cnt, 0)    AS uses
FROM users u
INNER JOIN team_members tm ON tm.user_id = u.id AND tm.team_id = ?
LEFT JOIN (
    SELECT user_id, COUNT(*) AS cnt
    FROM prompts
    WHERE team_id = ? AND created_at >= ? AND created_at < ? AND deleted_at IS NULL
    GROUP BY user_id
) created ON created.user_id = u.id
LEFT JOIN (
    SELECT pv.changed_by AS user_id, COUNT(*) AS cnt
    FROM prompt_versions pv
    JOIN prompts p ON p.id = pv.prompt_id
    WHERE p.team_id = ? AND pv.created_at >= ? AND pv.created_at < ? AND p.deleted_at IS NULL
    GROUP BY pv.changed_by
) edited ON edited.user_id = u.id
LEFT JOIN (
    SELECT user_id, COUNT(*) AS cnt
    FROM prompt_usage_log
    WHERE team_id = ? AND used_at >= ? AND used_at < ?
    GROUP BY user_id
) uses ON uses.user_id = u.id
ORDER BY (COALESCE(created.cnt,0) + COALESCE(edited.cnt,0) + COALESCE(uses.cnt,0)) DESC
LIMIT ?`
	var rows []repo.ContributorRow
	err := r.db.WithContext(ctx).Raw(query,
		teamID,
		teamID, rng.From, rng.To,
		teamID, rng.From, rng.To,
		teamID, rng.From, rng.To,
		limit,
	).Scan(&rows).Error
	return rows, err
}

// --- SHARE perf ---

func (r *analyticsRepo) ShareViewsPerDay(ctx context.Context, userID uint, rng repo.DateRange) ([]repo.UsagePoint, error) {
	q := r.db.WithContext(ctx).
		Table("share_view_log AS svl").
		Select("date_trunc('day', svl.viewed_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS count").
		Joins("JOIN share_links sl ON sl.id = svl.share_link_id").
		Where("sl.user_id = ? AND svl.viewed_at >= ? AND svl.viewed_at < ?", userID, rng.From, rng.To)

	var points []repo.UsagePoint
	err := q.Group("day").Order("day").Scan(&points).Error
	return points, err
}

func (r *analyticsRepo) TopSharedPrompts(ctx context.Context, userID uint, rng repo.DateRange, limit int) ([]repo.PromptUsageRow, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	q := r.db.WithContext(ctx).
		Table("share_view_log AS svl").
		Select("sl.prompt_id AS prompt_id, p.title AS title, COUNT(*) AS uses").
		Joins("JOIN share_links sl ON sl.id = svl.share_link_id").
		Joins("JOIN prompts p ON p.id = sl.prompt_id").
		Where("sl.user_id = ? AND svl.viewed_at >= ? AND svl.viewed_at < ?", userID, rng.From, rng.To).
		Where("p.deleted_at IS NULL").
		Group("sl.prompt_id, p.title").
		Order("uses DESC").
		Limit(limit)

	var rows []repo.PromptUsageRow
	err := q.Scan(&rows).Error
	return rows, err
}

func (r *analyticsRepo) LogShareView(ctx context.Context, view *models.ShareView) error {
	return r.db.WithContext(ctx).Create(view).Error
}

// --- SMART INSIGHTS (Max only) ---

// UpsertInsight использует ON CONFLICT на expression-based unique index
// idx_usi_unique (user_id, COALESCE(team_id, 0), insight_type). GORM
// clause.OnConflict.Columns не умеет expression-targets, поэтому Raw SQL.
func (r *analyticsRepo) UpsertInsight(ctx context.Context, insight *models.SmartInsight) error {
	const query = `
INSERT INTO user_smart_insights (user_id, team_id, insight_type, payload, computed_at)
VALUES (?, ?, ?, ?, NOW())
ON CONFLICT (user_id, COALESCE(team_id, 0), insight_type)
DO UPDATE SET payload = EXCLUDED.payload, computed_at = EXCLUDED.computed_at`
	return r.db.WithContext(ctx).Exec(query,
		insight.UserID, insight.TeamID, insight.InsightType, insight.Payload,
	).Error
}

func (r *analyticsRepo) GetInsights(ctx context.Context, userID uint, teamID *uint) ([]models.SmartInsight, error) {
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	q = scopeTeam(q, "team_id", teamID)
	var insights []models.SmartInsight
	err := q.Order("computed_at DESC").Find(&insights).Error
	return insights, err
}

// --- CLEANUP (cron) ---

func (r *analyticsRepo) DeleteShareViewsOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("viewed_at < ?", before).
		Delete(&models.ShareView{})
	return res.RowsAffected, res.Error
}

// CleanupShareViewsByRetention — per-plan retention одним SQL. Free не
// пишет в share_view_log по задумке, но покрывается fallback (30 дней).
// Явные whitelist-значения plan_id вместо LIKE 'pro%'/'max%' — исключаем
// ложное срабатывание на будущих plan_id (H3).
func (r *analyticsRepo) CleanupShareViewsByRetention(ctx context.Context) (int64, error) {
	const query = `
DELETE FROM share_view_log svl
USING share_links sl, users u
WHERE svl.share_link_id = sl.id
  AND sl.user_id = u.id
  AND (
    (u.plan_id = 'free' AND svl.viewed_at < NOW() - INTERVAL '30 days')
    OR (u.plan_id IN ('pro', 'pro_yearly') AND svl.viewed_at < NOW() - INTERVAL '90 days')
    OR (u.plan_id IN ('max', 'max_yearly') AND svl.viewed_at < NOW() - INTERVAL '365 days')
  )`
	res := r.db.WithContext(ctx).Exec(query)
	return res.RowsAffected, res.Error
}

// GetTrendingPrompts — растущие/падающие промпты одним SQL-запросом с двумя CTE.
// Раньше считалось через 2× TopPrompts + сравнение в Go map'е — O(prompts) по памяти
// и 2 отдельных round-trip. Теперь — один запрос, БД делает merge естественно.
//
// growing=true: берём промпты где last_7d >= prev_7d * factor (factor=2.0 для TRENDING).
//               Новые промпты (prev=0) тоже считаем trending (тренд-с-нуля).
// growing=false: берём промпты где prev > 0 AND last_7d <= prev_7d * factor
//                (factor=0.5 для DECLINING — падение минимум вдвое).
func (r *analyticsRepo) GetTrendingPrompts(
	ctx context.Context,
	userID uint,
	teamID *uint,
	factor float64,
	growing bool,
	limit int,
) ([]repo.TrendRow, error) {
	if limit < 1 || limit > 50 {
		limit = 5
	}
	teamFilter := "pul.team_id IS NULL"
	args := []any{userID}
	if teamID != nil {
		teamFilter = "pul.team_id = ?"
		args = append(args, *teamID)
		// user_id + team_id для second CTE (prev_7d) — тот же набор аргументов.
	}
	// Подготовим where-clauses для both CTE.
	// Первый CTE: last 7 days; второй: [-14d, -7d).
	var whereClause string
	if growing {
		whereClause = "l.uses >= COALESCE(p.uses, 0) * ? OR p.uses IS NULL"
	} else {
		whereClause = "p.uses > 0 AND l.uses <= p.uses * ?"
	}

	query := `
WITH last_7 AS (
  SELECT pul.prompt_id AS prompt_id, COUNT(*)::bigint AS uses
  FROM prompt_usage_log pul
  WHERE pul.user_id = ? AND ` + teamFilter + `
    AND pul.used_at >= NOW() - INTERVAL '7 days'
  GROUP BY pul.prompt_id
),
prev_7 AS (
  SELECT pul.prompt_id AS prompt_id, COUNT(*)::bigint AS uses
  FROM prompt_usage_log pul
  WHERE pul.user_id = ? AND ` + teamFilter + `
    AND pul.used_at >= NOW() - INTERVAL '14 days'
    AND pul.used_at < NOW() - INTERVAL '7 days'
  GROUP BY pul.prompt_id
)
SELECT l.prompt_id, pr.title, l.uses AS uses_last, COALESCE(p.uses, 0) AS uses_previous
FROM last_7 l
JOIN prompts pr ON pr.id = l.prompt_id AND pr.deleted_at IS NULL
LEFT JOIN prev_7 p ON p.prompt_id = l.prompt_id
WHERE ` + whereClause + `
ORDER BY l.uses DESC
LIMIT ?`

	// Аргументы: userID (first CTE) + [teamID] + userID (second CTE) + [teamID] + factor + limit
	bindArgs := []any{userID}
	if teamID != nil {
		bindArgs = append(bindArgs, *teamID)
	}
	bindArgs = append(bindArgs, userID)
	if teamID != nil {
		bindArgs = append(bindArgs, *teamID)
	}
	bindArgs = append(bindArgs, factor, limit)
	_ = args // consumed above; left for clarity

	var rows []repo.TrendRow
	err := r.db.WithContext(ctx).Raw(query, bindArgs...).Scan(&rows).Error
	return rows, err
}

// UsageByModel — сегментация use'ов по model_used. Группируем NULL в пустую
// строку на Go-стороне — на выходе UI покажет его как "Без модели".
func (r *analyticsRepo) UsageByModel(ctx context.Context, userID uint, teamID *uint, rng repo.DateRange) ([]repo.ModelUsageRow, error) {
	q := r.db.WithContext(ctx).
		Table("prompt_usage_log").
		Select("COALESCE(model_used, '') AS model, COUNT(*) AS uses").
		Where("user_id = ? AND used_at >= ? AND used_at < ?", userID, rng.From, rng.To)
	q = scopeTeam(q, "team_id", teamID)

	var rows []repo.ModelUsageRow
	err := q.Group("model").Order("uses DESC").Scan(&rows).Error
	return rows, err
}

// PromptUsageTimeline — per-prompt использование по дням. WHERE prompt_id = ?
// (общий UsagePerDay считает все промпты юзера).
func (r *analyticsRepo) PromptUsageTimeline(ctx context.Context, promptID uint, rng repo.DateRange) ([]repo.UsagePoint, error) {
	var points []repo.UsagePoint
	err := r.db.WithContext(ctx).
		Table("prompt_usage_log").
		Select("date_trunc('day', used_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS count").
		Where("prompt_id = ? AND used_at >= ? AND used_at < ?", promptID, rng.From, rng.To).
		Group("day").Order("day").
		Scan(&points).Error
	return points, err
}

// PromptShareViewsTimeline — per-prompt просмотры share-ссылок по дням.
// JOIN с share_links для фильтра по prompt_id.
func (r *analyticsRepo) PromptShareViewsTimeline(ctx context.Context, promptID uint, rng repo.DateRange) ([]repo.UsagePoint, error) {
	var points []repo.UsagePoint
	err := r.db.WithContext(ctx).
		Table("share_view_log AS svl").
		Select("date_trunc('day', svl.viewed_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS count").
		Joins("JOIN share_links sl ON sl.id = svl.share_link_id").
		Where("sl.prompt_id = ? AND svl.viewed_at >= ? AND svl.viewed_at < ?", promptID, rng.From, rng.To).
		Group("day").Order("day").
		Scan(&points).Error
	return points, err
}

// CleanupPromptUsageByRetention — retention prompt_usage_log по plan_id юзера.
// Free=30д, Pro=90д, Max=365д. Зеркало CleanupShareViewsByRetention.
func (r *analyticsRepo) CleanupPromptUsageByRetention(ctx context.Context) (int64, error) {
	const query = `
DELETE FROM prompt_usage_log pul
USING users u
WHERE pul.user_id = u.id
  AND (
    (u.plan_id = 'free' AND pul.used_at < NOW() - INTERVAL '30 days')
    OR (u.plan_id IN ('pro', 'pro_yearly') AND pul.used_at < NOW() - INTERVAL '90 days')
    OR (u.plan_id IN ('max', 'max_yearly') AND pul.used_at < NOW() - INTERVAL '365 days')
  )`
	res := r.db.WithContext(ctx).Exec(query)
	return res.RowsAffected, res.Error
}

// --- SMART INSIGHTS M8 (за feature-flag experimentalInsights в Service) ---

// MostEditedPrompts — промпты с наибольшим числом версий (prompt_versions).
// LIMIT ограничивает выдачу; WHERE фильтрует team/personal scope + soft-delete.
func (r *analyticsRepo) MostEditedPrompts(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.PromptUsageRow, error) {
	if limit < 1 || limit > 50 {
		limit = 5
	}
	q := r.db.WithContext(ctx).
		Table("prompts AS p").
		Select("p.id AS prompt_id, p.title AS title, COUNT(pv.id) AS uses").
		Joins("JOIN prompt_versions pv ON pv.prompt_id = p.id").
		Where("p.user_id = ? AND p.deleted_at IS NULL", userID)
	q = scopeTeam(q, "p.team_id", teamID)

	var rows []repo.PromptUsageRow
	err := q.Group("p.id, p.title").
		Having("COUNT(pv.id) > 1"). // >1 т.к. версия 1 = исходный промпт
		Order("uses DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// PossibleDuplicates — пары похожих промптов через pg_trgm.similarity().
// Защита от O(n²): WHERE a.updated_at >= NOW() - INTERVAL '30 days' ограничивает
// candidate pool, threshold и LIMIT режут выхлоп.
// a.id < b.id избавляет от симметричных дубликатов (A-B и B-A).
func (r *analyticsRepo) PossibleDuplicates(ctx context.Context, userID uint, teamID *uint, threshold float32, limit int) ([]repo.DuplicatePair, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	if threshold <= 0 || threshold > 1 {
		threshold = 0.8
	}
	teamFilter := "a.team_id IS NULL AND b.team_id IS NULL"
	args := []any{userID, userID}
	if teamID != nil {
		teamFilter = "a.team_id = ? AND b.team_id = ?"
		args = append(args, *teamID, *teamID)
	}
	query := `
SELECT a.id AS prompt_a_id, a.title AS prompt_a_title,
       b.id AS prompt_b_id, b.title AS prompt_b_title,
       similarity(a.content, b.content) AS similarity
FROM prompts a
JOIN prompts b ON a.id < b.id
WHERE a.user_id = ? AND b.user_id = ?
  AND a.deleted_at IS NULL AND b.deleted_at IS NULL
  AND ` + teamFilter + `
  AND a.updated_at >= NOW() - INTERVAL '30 days'
  AND similarity(a.content, b.content) >= ?
ORDER BY similarity DESC
LIMIT ?`
	args = append(args, threshold, limit)

	var rows []repo.DuplicatePair
	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error
	return rows, err
}

// OrphanTags — теги без ни одного промпта через LEFT JOIN + IS NULL.
func (r *analyticsRepo) OrphanTags(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.TagRow, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	q := r.db.WithContext(ctx).
		Table("tags AS t").
		Select("t.id AS tag_id, t.name AS name").
		Joins("LEFT JOIN prompt_tags pt ON pt.tag_id = t.id").
		Where("t.user_id = ? AND pt.tag_id IS NULL", userID)
	q = scopeTeam(q, "t.team_id", teamID)

	var rows []repo.TagRow
	err := q.Group("t.id, t.name").Limit(limit).Scan(&rows).Error
	return rows, err
}

// EmptyCollections — коллекции без промптов. Аналогично OrphanTags через
// LEFT JOIN prompt_collections.
func (r *analyticsRepo) EmptyCollections(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.CollectionRow, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	q := r.db.WithContext(ctx).
		Table("collections AS c").
		Select("c.id AS collection_id, c.name AS name").
		Joins("LEFT JOIN prompt_collections pc ON pc.collection_id = c.id").
		Where("c.user_id = ? AND c.deleted_at IS NULL AND pc.collection_id IS NULL", userID)
	q = scopeTeam(q, "c.team_id", teamID)

	var rows []repo.CollectionRow
	err := q.Group("c.id, c.name").Limit(limit).Scan(&rows).Error
	return rows, err
}
