package repository

import (
	"context"
	"time"

	"promptvault/internal/models"
)

// PlanRepository — доступ к тарифным планам. Реализация кэширует
// результаты in-memory с TTL 5 минут.
type PlanRepository interface {
	GetAll(ctx context.Context) ([]models.SubscriptionPlan, error)
	GetByID(ctx context.Context, id string) (*models.SubscriptionPlan, error)
	GetActive(ctx context.Context) ([]models.SubscriptionPlan, error)
}

// SubscriptionRepository — управление подписками.
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetActiveByUserID(ctx context.Context, userID uint) (*models.Subscription, error)
	Update(ctx context.Context, sub *models.Subscription) error

	// ListExpiring возвращает active и past_due подписки с истёкшим периодом
	// (current_period_end < before). Используется expirationLoop для перевода
	// в expired и даунгрейда на free.
	ListExpiring(ctx context.Context, before time.Time) ([]models.Subscription, error)

	// ActivateWithPlanUpdate создаёт/обновляет подписку и устанавливает
	// users.plan_id в одной транзакции.
	ActivateWithPlanUpdate(ctx context.Context, sub *models.Subscription, userID uint, planID string) error

	// CancelAtPeriodEnd помечает подписку для отмены в конце периода.
	CancelAtPeriodEnd(ctx context.Context, subID uint) error

	// ExpireAndDowngrade переводит подписку в expired и users.plan_id в "free".
	ExpireAndDowngrade(ctx context.Context, subID uint, userID uint) error

	// SetRebillId сохраняет RebillId, выданный T-Bank после первого рекуррентного
	// платежа. Используется для последующих /Charge.
	SetRebillId(ctx context.Context, subID uint, rebillID string) error

	// SetAutoRenew управляет автопродлением. false — подписка истечёт без попытки
	// списания; true — renewLoop попытается списать за 3 дня до окончания.
	SetAutoRenew(ctx context.Context, subID uint, autoRenew bool) error

	// ListReadyForRenewal возвращает подписки, готовые к попытке списания:
	//  - active с auto_renew=true, rebill_id<>'' и current_period_end <= before;
	//  - past_due с renewal_attempts < maxAttempts и последней попыткой >= retryAfter назад.
	// Используется renewLoop.
	ListReadyForRenewal(ctx context.Context, before time.Time, retryAfter time.Time, maxAttempts int) ([]models.Subscription, error)

	// ExtendPeriod продлевает подписку на заданный период (для успешного renewal).
	// Сбрасывает renewal_attempts=0 и статус past_due→active.
	ExtendPeriod(ctx context.Context, subID uint, newPeriodEnd time.Time) error

	// RecordRenewalFailure фиксирует неудачную попытку Charge: переводит подписку
	// в past_due (если была active), инкрементирует renewal_attempts и ставит
	// last_renewal_attempt_at=now. Используется renewLoop при ошибке Init/Charge.
	RecordRenewalFailure(ctx context.Context, subID uint) error
}

// PaymentRepository — управление платежами.
type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByExternalID(ctx context.Context, provider, externalID string) (*models.Payment, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*models.Payment, error)
	UpdateStatus(ctx context.Context, id uint, status models.PaymentStatus) error

	// UpdateExternalID обновляет external_id уже сохранённого платежа.
	// Используется для двухфазного сохранения: Payment создаётся до вызова
	// payment.Init() с placeholder external_id, затем обновляется на ID,
	// возвращённый провайдером.
	UpdateExternalID(ctx context.Context, id uint, externalID string) error

	// TransitionStatus атомарно переводит status: expected → next через
	// conditional UPDATE. Возвращает true, если переход произошёл, false
	// если статус уже не expected (другой webhook опередил). Защищает от
	// race conditions при параллельных webhook'ах без явных SELECT FOR UPDATE.
	TransitionStatus(ctx context.Context, id uint, expected, next models.PaymentStatus) (bool, error)

	// LinkSubscription связывает платёж с созданной подпиской после активации.
	// Используется из HandleWebhook: платёж создаётся в Checkout без subscription_id
	// (подписки ещё нет), связка устанавливается после activateSubscription.
	// Без этого невозможно корректно обработать refund конкретной подписки и
	// составить историю платежей за подписку.
	LinkSubscription(ctx context.Context, paymentID, subscriptionID uint) error
}

// QuotaRepository — подсчёт использованных ресурсов для enforcement квот.
type QuotaRepository interface {
	CountPrompts(ctx context.Context, userID uint) (int64, error)
	CountCollections(ctx context.Context, userID uint) (int64, error)
	CountTeamsOwned(ctx context.Context, userID uint) (int64, error)
	CountActiveShareLinks(ctx context.Context, userID uint) (int64, error)
	CountTeamMembers(ctx context.Context, teamID uint) (int, error)
	GetDailyUsage(ctx context.Context, userID uint, date time.Time, featureType string) (int, error)
	GetTotalUsage(ctx context.Context, userID uint, featureType string) (int, error)
	IncrementDailyUsage(ctx context.Context, userID uint, date time.Time, featureType string) error
}
