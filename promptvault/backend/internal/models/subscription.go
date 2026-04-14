package models

import (
	"encoding/json"
	"time"
)

// SubscriptionPlan — тарифный план (free/pro/max). Хранится в БД,
// кэшируется in-memory (TTL 5 мин). Конвенция: -1 = безлимит.
type SubscriptionPlan struct {
	ID                   string          `gorm:"primaryKey;size:20" json:"id"`
	Name                 string          `gorm:"size:50;not null" json:"name"`
	PriceKop             int             `gorm:"not null;default:0" json:"price_kop"`
	PeriodDays           int             `gorm:"not null;default:30" json:"period_days"`
	MaxPrompts           int             `gorm:"not null;default:50" json:"max_prompts"`
	MaxCollections       int             `gorm:"not null;default:3" json:"max_collections"`
	MaxAIRequestsDaily   int             `gorm:"not null;default:5" json:"max_ai_requests_daily"`
	AIRequestsIsTotal    bool            `gorm:"not null;default:false" json:"ai_requests_is_total"`
	MaxTeams             int             `gorm:"not null;default:1" json:"max_teams"`
	MaxTeamMembers       int             `gorm:"not null;default:3" json:"max_team_members"`
	MaxShareLinks        int             `gorm:"not null;default:2" json:"max_share_links"`
	MaxExtUsesDaily      int             `gorm:"not null;default:5" json:"max_ext_uses_daily"`
	MaxMCPUsesDaily      int             `gorm:"column:max_mcp_uses_daily;not null;default:5" json:"max_mcp_uses_daily"`
	Features             json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"features"`
	SortOrder            int             `gorm:"not null;default:0" json:"sort_order"`
	IsActive             bool            `gorm:"not null;default:true" json:"is_active"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

func (SubscriptionPlan) TableName() string { return "subscription_plans" }

// IsUnlimited проверяет, что лимит = -1 (безлимит).
func IsUnlimited(limit int) bool { return limit == -1 }

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
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`

	Plan SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan,omitzero"`
}

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
// feature_type: "ai", "extension", "mcp". Composite PK (user_id, usage_date, feature_type).
type DailyFeatureUsage struct {
	UserID      uint      `gorm:"primaryKey;not null" json:"user_id"`
	UsageDate   time.Time `gorm:"primaryKey;type:date;not null" json:"usage_date"`
	FeatureType string    `gorm:"primaryKey;size:20;not null" json:"feature_type"`
	Count       int       `gorm:"not null;default:0" json:"count"`
}

func (DailyFeatureUsage) TableName() string { return "daily_feature_usage" }
