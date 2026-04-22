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
