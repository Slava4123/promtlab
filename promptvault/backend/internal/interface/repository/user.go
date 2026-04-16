package repository

import (
	"context"
	"time"

	"promptvault/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	SearchUsers(ctx context.Context, query string, limit int) ([]models.User, error)
	Update(ctx context.Context, user *models.User) error

	// SetQuotaWarningSentOn выставляет users.quota_warning_sent_on = date
	// атомарным UPDATE (M-5c). Используется quota.Service чтобы не слать
	// warning повторно в ту же дату.
	SetQuotaWarningSentOn(ctx context.Context, userID uint, date time.Time) error

	// TouchLastLogin обновляет last_login_at=now (M-5d).
	// Вызывается из auth.Login и OAuth callbacks.
	TouchLastLogin(ctx context.Context, userID uint) error

	// ListInactiveForReengagement возвращает юзеров для re-engagement email (M-5d).
	// Критерии: verified, active, last_login_at < inactiveBefore И
	// (reengagement_sent_at IS NULL ИЛИ reengagement_sent_at < sentBefore).
	// Для защиты от спама от cron'а при большой базе — LIMIT batch.
	ListInactiveForReengagement(ctx context.Context, inactiveBefore, sentBefore time.Time, limit int) ([]models.User, error)

	// MarkReengagementSent выставляет reengagement_sent_at=now.
	MarkReengagementSent(ctx context.Context, userID uint) error
}
