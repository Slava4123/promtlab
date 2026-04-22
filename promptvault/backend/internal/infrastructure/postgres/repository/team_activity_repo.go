package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type teamActivityRepo struct {
	db *gorm.DB
}

func NewTeamActivityRepository(db *gorm.DB) repo.TeamActivityRepository {
	return &teamActivityRepo{db: db}
}

func (r *teamActivityRepo) Log(ctx context.Context, event *models.TeamActivityLog) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *teamActivityRepo) List(ctx context.Context, f repo.TeamActivityFilter) ([]models.TeamActivityLog, *time.Time, error) {
	limit := f.Limit
	if limit < 1 || limit > 200 {
		limit = 50
	}

	q := r.db.WithContext(ctx).Model(&models.TeamActivityLog{}).Where("team_id = ?", f.TeamID)
	if f.EventType != "" {
		q = q.Where("event_type = ?", f.EventType)
	}
	if f.ActorID != nil {
		q = q.Where("actor_id = ?", *f.ActorID)
	}
	if f.TargetType != "" {
		q = q.Where("target_type = ?", f.TargetType)
	}
	if f.TargetID != nil {
		q = q.Where("target_id = ?", *f.TargetID)
	}
	if f.FromTime != nil {
		q = q.Where("created_at >= ?", *f.FromTime)
	}
	if f.ToTime != nil {
		q = q.Where("created_at <= ?", *f.ToTime)
	}
	if f.CursorBefore != nil {
		q = q.Where("created_at < ?", *f.CursorBefore)
	}

	// Fetch limit+1 — простой sentinel для определения "есть ли ещё страницы".
	var events []models.TeamActivityLog
	if err := q.Order("created_at DESC").Limit(limit + 1).Find(&events).Error; err != nil {
		return nil, nil, err
	}

	var next *time.Time
	if len(events) > limit {
		cursor := events[limit-1].CreatedAt
		next = &cursor
		events = events[:limit]
	}
	return events, next, nil
}

func (r *teamActivityRepo) ListByTarget(ctx context.Context, targetType string, targetID uint, limit int) ([]models.TeamActivityLog, error) {
	if limit < 1 || limit > 500 {
		limit = 100
	}
	var events []models.TeamActivityLog
	err := r.db.WithContext(ctx).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

// AnonymizeActor заменяет actor_id/email/name на deleted-константы во всех
// записях. Row-level триггер в БД (prevent_team_activity_log_mutation)
// разрешает UPDATE только actor_* полей — запрос пройдёт валидацию.
func (r *teamActivityRepo) AnonymizeActor(ctx context.Context, userID uint) (int64, error) {
	res := r.db.WithContext(ctx).
		Model(&models.TeamActivityLog{}).
		Where("actor_id = ?", userID).
		Updates(map[string]any{
			"actor_id":    nil,
			"actor_email": models.AnonymizedActorEmail,
			"actor_name":  models.AnonymizedActorName,
		})
	return res.RowsAffected, res.Error
}

// DeleteOlderThan — cleanup по retention. Вызывается cron-job'ом.
// Триггер на UPDATE, DELETE разрешён полностью.
func (r *teamActivityRepo) DeleteOlderThan(ctx context.Context, teamID uint, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("team_id = ? AND created_at < ?", teamID, before).
		Delete(&models.TeamActivityLog{})
	return res.RowsAffected, res.Error
}

// CleanupByRetention — один SQL DELETE для всех команд с JOIN на users
// для определения плана владельца. Free=30д, Pro=90д, Max=365д.
// Предпочтительнее чем цикл по командам: один запрос вместо N.
// Явные whitelist-значения plan_id вместо LIKE 'pro%'/'max%' — исключаем
// ложное срабатывание на будущих plan_id (H3).
func (r *teamActivityRepo) CleanupByRetention(ctx context.Context) (int64, error) {
	const query = `
DELETE FROM team_activity_log tal
USING teams t, users u
WHERE tal.team_id = t.id
  AND t.created_by = u.id
  AND (
    (u.plan_id = 'free' AND tal.created_at < NOW() - INTERVAL '30 days')
    OR (u.plan_id IN ('pro', 'pro_yearly') AND tal.created_at < NOW() - INTERVAL '90 days')
    OR (u.plan_id IN ('max', 'max_yearly') AND tal.created_at < NOW() - INTERVAL '365 days')
  )`
	res := r.db.WithContext(ctx).Exec(query)
	return res.RowsAffected, res.Error
}
