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

	// CountReferredBy — сколько юзеров зарегистрировалось с referred_by = code (M-7).
	// Используется в GET /api/auth/referral для отображения счётчика приглашённых.
	CountReferredBy(ctx context.Context, code string) (int64, error)

	// GetByReferralCode находит юзера по его ReferralCode (M-7).
	// Используется в webhook/activate для выдачи награды рефереру.
	GetByReferralCode(ctx context.Context, code string) (*models.User, error)

	// MarkReferralRewarded атомарно ставит referral_rewarded_at=now только если
	// он был NULL. Возвращает true если действительно обновил (idempotency).
	// Защищает от повторной выдачи награды при повторных платежах того же рефери.
	MarkReferralRewarded(ctx context.Context, userID uint) (bool, error)

	// ListMaxUsers возвращает ID активных юзеров на тарифе Max (включая max_yearly).
	// Используется analytics.InsightsComputeLoop для ежесуточного пересчёта
	// детерминированных Smart Insights. Ограничение — active (не frozen/deleted).
	ListMaxUsers(ctx context.Context) ([]uint, error)

	// SetInsightEmailsEnabled атомарно меняет users.insight_emails_enabled
	// (Phase 14 M-10). Opt-in по ФЗ-152.
	SetInsightEmailsEnabled(ctx context.Context, userID uint, enabled bool) error
}

// InsightNotificationRepository — лог отправленных email-уведомлений
// по Smart Insights. Используется для rate-limit 1 письмо/неделю.
type InsightNotificationRepository interface {
	// RecentlySent возвращает true, если за последние `within` было отправлено
	// уведомление пары (userID, insightType). Rate-limit защита.
	RecentlySent(ctx context.Context, userID uint, insightType string, within time.Duration) (bool, error)

	// Record вставляет лог-запись факта отправки.
	Record(ctx context.Context, userID uint, insightType string) error
}
