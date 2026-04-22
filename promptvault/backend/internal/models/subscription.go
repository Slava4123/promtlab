package models

import (
	"encoding/json"
	"time"
)

// SubscriptionPlan — тарифный план (free/pro/max). Хранится в БД,
// кэшируется in-memory (TTL 5 мин). Все лимиты — конкретные неотрицательные
// числа (миграция 000046 убрала legacy sentinel -1 "безлимит").
type SubscriptionPlan struct {
	ID                   string          `gorm:"primaryKey;size:20" json:"id"`
	Name                 string          `gorm:"size:50;not null" json:"name"`
	PriceKop             int             `gorm:"not null;default:0" json:"price_kop"`
	PeriodDays           int             `gorm:"not null;default:30" json:"period_days"`
	MaxPrompts           int             `gorm:"not null;default:50" json:"max_prompts"`
	MaxCollections       int             `gorm:"not null;default:3" json:"max_collections"`
	MaxTeams             int             `gorm:"not null;default:1" json:"max_teams"`
	MaxTeamMembers       int             `gorm:"not null;default:3" json:"max_team_members"`
	MaxShareLinks        int             `gorm:"not null;default:2" json:"max_share_links"`
	// MaxDailyShares — лимит на СОЗДАНИЕ публичных шар-ссылок в день.
	// Phase 14, миграция 000044. Считается через daily_feature_usage
	// с feature_type='share_create'. Free=10, Pro=100, Max=1000.
	MaxDailyShares       int             `gorm:"not null;default:10" json:"max_daily_shares"`
	MaxExtUsesDaily      int             `gorm:"not null;default:5" json:"max_ext_uses_daily"`
	MaxMCPUsesDaily      int             `gorm:"column:max_mcp_uses_daily;not null;default:5" json:"max_mcp_uses_daily"`
	Features             json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"features"`
	SortOrder            int             `gorm:"not null;default:0" json:"sort_order"`
	IsActive             bool            `gorm:"not null;default:true" json:"is_active"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

func (SubscriptionPlan) TableName() string { return "subscription_plans" }

// SubscriptionStatus — статус подписки.
//
// В v1 используются active, cancelled, expired.
// SubStatusPastDue зарезервирован для v2 (автопродление) — будет назначаться
// при первом неуспехе списания с retry-логикой перед переходом в expired.
// Текущий expirationLoop переводит просроченные active → expired напрямую.
type SubscriptionStatus string

const (
	SubStatusActive    SubscriptionStatus = "active"
	SubStatusPastDue   SubscriptionStatus = "past_due" // reserved for v2 auto-renewal
	SubStatusPaused    SubscriptionStatus = "paused"   // M-6: добровольная пауза 1-3 мес
	SubStatusCancelled SubscriptionStatus = "cancelled"
	SubStatusExpired   SubscriptionStatus = "expired"
)

// Subscription — активная подписка пользователя. Partial unique index
// в БД гарантирует max 1 active/past_due подписку на юзера.
//
// RebillId и AutoRenew — для рекуррентных платежей T-Bank. RebillId выдаётся
// банком после первого успешного платежа (Recurrent=Y в Init); используется
// в /Charge для последующих списаний без 3DS. AutoRenew=false отключает
// автоматическое продление: подписка истечёт без попытки списания.
type Subscription struct {
	ID                 uint               `gorm:"primaryKey" json:"id"`
	UserID             uint               `gorm:"not null" json:"user_id"`
	PlanID             string             `gorm:"size:20;not null" json:"plan_id"`
	Status             SubscriptionStatus `gorm:"size:20;not null;default:active" json:"status"`
	CurrentPeriodStart time.Time          `gorm:"not null" json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `gorm:"not null" json:"current_period_end"`
	CancelAtPeriodEnd  bool               `gorm:"not null;default:false" json:"cancel_at_period_end"`
	CancelledAt        *time.Time         `json:"cancelled_at,omitempty"`
	RebillId           string             `gorm:"column:rebill_id;size:50" json:"-"`
	AutoRenew          bool               `gorm:"not null;default:true" json:"auto_renew"`

	// LastRenewalAttemptAt и RenewalAttempts реализуют retry-политику для
	// past_due подписок: не более 3 попыток Charge за период, следующая — не
	// раньше чем через 24ч после предыдущей. После 3-го фейла expirationLoop
	// переводит в expired + downgrade на free.
	LastRenewalAttemptAt *time.Time `gorm:"column:last_renewal_attempt_at" json:"last_renewal_attempt_at,omitempty"`
	RenewalAttempts      int        `gorm:"column:renewal_attempts;not null;default:0" json:"renewal_attempts"`

	// PreExpireStage — какое pre-expire напоминание уже отправлено:
	//   0 — не отправляли, 1 — 3-day reminder, 2 — 1-day reminder.
	// Сбрасывается в 0 при ExtendPeriod (успешное продление).
	PreExpireStage int16 `gorm:"column:pre_expire_stage;not null;default:0" json:"-"`

	// PausedAt — момент входа в pause (M-6). NULL если подписка не на паузе.
	// При Resume: remaining = current_period_end - paused_at; new_end = now + remaining.
	// PausedUntil — запланированная дата авто-возобновления. ExpirationLoop
	// резюмит подписку автоматически, когда paused_until < now().
	PausedAt    *time.Time `gorm:"column:paused_at" json:"paused_at,omitempty"`
	PausedUntil *time.Time `gorm:"column:paused_until" json:"paused_until,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Plan SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan,omitzero"`
}

// SubscriptionCancellation — запись об отмене подписки с причиной (M-6b exit survey).
// Append-only — если юзер после Resume снова Cancel, создаётся новая запись.
type SubscriptionCancellation struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null" json:"user_id"`
	SubscriptionID uint      `gorm:"not null" json:"subscription_id"`
	PlanID         string    `gorm:"size:20;not null" json:"plan_id"`
	Reason         string    `gorm:"size:30;not null" json:"reason"`
	OtherText      string    `gorm:"type:text;not null;default:''" json:"other_text"`
	CancelledAt    time.Time `gorm:"not null;default:now()" json:"cancelled_at"`
	CreatedAt      time.Time `json:"created_at"`
}

func (SubscriptionCancellation) TableName() string { return "subscription_cancellations" }

// PaymentStatus — статус платежа.
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentSucceeded PaymentStatus = "succeeded"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

// Payment — запись о платеже. Unique index на (provider, external_id)
// обеспечивает idempotency при обработке webhook.
type Payment struct {
	ID              uint            `gorm:"primaryKey" json:"id"`
	UserID          uint            `gorm:"not null" json:"user_id"`
	SubscriptionID  *uint           `json:"subscription_id,omitempty"`
	ExternalID      string          `gorm:"size:100;not null" json:"external_id"`
	IdempotencyKey  string          `gorm:"size:100;not null;uniqueIndex" json:"idempotency_key"`
	AmountKop       int             `gorm:"not null" json:"amount_kop"`
	Currency        string          `gorm:"size:3;not null;default:RUB" json:"currency"`
	Status          PaymentStatus   `gorm:"size:20;not null;default:pending" json:"status"`
	Provider        string          `gorm:"size:20;not null;default:tbank" json:"provider"`
	ProviderData    json.RawMessage `gorm:"type:jsonb" json:"provider_data,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// DailyFeatureUsage — персистентный счётчик дневного использования.
// feature_type: "extension", "mcp". Composite PK (user_id, usage_date, feature_type).
// Исторически существовал "ai" — строки остались в БД от предыдущих версий, но новые не создаются.
type DailyFeatureUsage struct {
	UserID      uint      `gorm:"primaryKey;not null" json:"user_id"`
	UsageDate   time.Time `gorm:"primaryKey;type:date;not null" json:"usage_date"`
	FeatureType string    `gorm:"primaryKey;size:20;not null" json:"feature_type"`
	Count       int       `gorm:"not null;default:0" json:"count"`
}

func (DailyFeatureUsage) TableName() string { return "daily_feature_usage" }
