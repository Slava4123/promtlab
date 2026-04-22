package repository

import (
	"context"
	"time"

	"promptvault/internal/models"
)

// TeamActivityFilter — параметры выборки feed'а команды.
//
// Cursor-based пагинация: вместо Page/PageSize используем CursorBefore
// (created_at последней записи предыдущей страницы). Стабильно при
// добавлении новых записей между запросами и быстрее OFFSET'а на
// глубоких страницах (индекс idx_tal_team_created покрывает).
type TeamActivityFilter struct {
	TeamID       uint
	EventType    string     // "" — не фильтровать
	ActorID      *uint      // nil — не фильтровать
	TargetType   string     // "" — не фильтровать
	TargetID     *uint      // nil — не фильтровать
	FromTime     *time.Time // включая
	ToTime       *time.Time // включая
	CursorBefore *time.Time // WHERE created_at < ? (nil — с начала)
	Limit        int        // 1..200, default 50
}

// TeamActivityRepository — append-only repo для team_activity_log.
//
// Update-методов нет — UPDATE запрещён триггером в БД (миграция 000040).
// Delete — только через DeleteOlderThan (retention cleanup cron).
type TeamActivityRepository interface {
	// Log вставляет событие. CreatedAt = NOW() если zero.
	Log(ctx context.Context, event *models.TeamActivityLog) error

	// List — страница по фильтру. nextCursor = nil на последней странице.
	List(ctx context.Context, filter TeamActivityFilter) (events []models.TeamActivityLog, nextCursor *time.Time, err error)

	// ListByTarget — все события конкретного target (для склейки в prompt history).
	// Не использует team_id фильтр — target однозначно идентифицирует ресурс.
	ListByTarget(ctx context.Context, targetType string, targetID uint, limit int) ([]models.TeamActivityLog, error)

	// AnonymizeActor — GDPR-hook при удалении user. Заменяет actor_id=NULL,
	// actor_email/name на константы AnonymizedActor*. Возвращает количество
	// затронутых записей.
	AnonymizeActor(ctx context.Context, userID uint) (int64, error)

	// DeleteOlderThan — cleanup по retention плана команды. Вызывается
	// cron job'ом с правильным before-cutoff для каждого плана.
	DeleteOlderThan(ctx context.Context, teamID uint, before time.Time) (int64, error)

	// CleanupByRetention — массовый cleanup по plan_id владельца команды.
	// Возвращает общее число удалённых строк. Выполняется ежесуточно из
	// usecases/analytics.CleanupLoop.
	CleanupByRetention(ctx context.Context) (int64, error)
}
