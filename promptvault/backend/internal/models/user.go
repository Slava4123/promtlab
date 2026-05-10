package models

import (
	"encoding/json"
	"time"
)

// UserRole — роль пользователя для RBAC. Хранится в БД как VARCHAR(20)
// с CHECK-constraint (см. миграцию 000016), в Go представлена как typed string
// для type-safety при сравнениях.
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// UserStatus — статус аккаунта. active — обычный доступ, frozen — админ
// заблокировал юзера (см. usecases/admin.FreezeUser).
type UserStatus string

const (
	StatusActive UserStatus = "active"
	StatusFrozen UserStatus = "frozen"
)

type User struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	Email                 string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash          string     `gorm:"size:255" json:"-"`
	Name                  string     `gorm:"size:100;not null" json:"name"`
	Username              string     `gorm:"size:30" json:"username,omitempty"`
	AvatarURL             string     `gorm:"size:500" json:"avatar_url,omitempty"`
	EmailVerified         bool       `gorm:"default:false" json:"email_verified"`
	Role                  UserRole   `gorm:"size:20;not null;default:user" json:"role"`
	Status                UserStatus `gorm:"size:20;not null;default:active" json:"status"`
	PlanID                string     `gorm:"size:20;not null;default:free" json:"plan_id"`
	TokenNonce            string     `gorm:"size:64" json:"-"`
	OnboardingCompletedAt *time.Time `json:"onboarding_completed_at,omitempty"`
	LastChangelogSeenAt   *time.Time `json:"last_changelog_seen_at,omitempty"`

	// Email lifecycle tracking (M-5).
	// WelcomeSentAt — чтобы повторный verify не отправил welcome дважды.
	// LastLoginAt — триггер для re-engagement email (M-5d).
	// ReengagementSentAt — чтобы не слать re-engagement чаще раза в 30 дней.
	// QuotaWarningSentOn — DATE: последний день когда отправили 80%-warning.
	//   Сравниваем с today (user tz) чтобы не спамить повторно внутри суток.
	WelcomeSentAt        *time.Time `gorm:"column:welcome_sent_at" json:"-"`
	LastLoginAt          *time.Time `gorm:"column:last_login_at" json:"-"`
	ReengagementSentAt   *time.Time `gorm:"column:reengagement_sent_at" json:"-"`
	QuotaWarningSentOn   *time.Time `gorm:"column:quota_warning_sent_on;type:date" json:"-"`

	// M-7 Referral.
	// ReferralCode — уникальный 8-символьный код юзера (делится с друзьями).
	// ReferredBy — код пригласившего, nullable (был ли юзер приглашён).
	// ReferralRewardedAt — момент выдачи награды пригласившему (idempotency).
	ReferralCode       string     `gorm:"column:referral_code;size:16;uniqueIndex" json:"referral_code"`
	ReferredBy         string     `gorm:"column:referred_by;size:16" json:"-"`
	ReferralRewardedAt *time.Time `gorm:"column:referral_rewarded_at" json:"-"`

	// Phase 14 M-10 opt-in по ФЗ-152. Default false — юзер должен явно
	// включить email-уведомления по Smart Insights в настройках.
	InsightEmailsEnabled bool `gorm:"column:insight_emails_enabled;not null;default:false" json:"insight_emails_enabled"`

	// LegacyQuotas — grandfather-снапшот старых лимитов для юзеров,
	// зарегистрированных до изменения тарифа. Ключи = имена колонок
	// SubscriptionPlan ("max_prompts", "max_ext_uses_daily", ...). Значение
	// JSON-числа. При проверке квоты Service.effectiveLimit использует
	// значение отсюда если есть, иначе из плана. Пустой {} = новый юзер
	// получает текущие лимиты плана. См. миграции 000068+ (Pack E/F).
	LegacyQuotas json.RawMessage `gorm:"column:legacy_quotas;type:jsonb;not null;default:'{}'" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	LinkedAccounts []LinkedAccount `gorm:"foreignKey:UserID" json:"linked_accounts,omitempty"`
}

// NewUser — конструктор User с обязательными полями (email, name, passwordHash)
// и явными дефолтами для критичных enum-полей. MN-32: до этого callers
// собирали через struct literal — легко промахнуться (Status, Role default'ы).
//
// passwordHash может быть пустым — для OAuth-only регистрации.
// referredBy — referral code пригласившего, "" если без реферала.
func NewUser(email, name, passwordHash, referredBy string) *User {
	return &User{
		Email:                email,
		Name:                 name,
		PasswordHash:         passwordHash,
		Role:                 RoleUser,
		Status:               StatusActive,
		PlanID:               "free",
		EmailVerified:        false,
		ReferredBy:           referredBy,
		InsightEmailsEnabled: false,
	}
}

func (u *User) HasPassword() bool {
	return u.PasswordHash != ""
}

// IsAdmin — удобный предикат для middleware/handlers. Использовать ТОЛЬКО
// в combination с re-check из БД (не доверять claim из JWT в security-sensitive
// ситуациях; см. middleware/admin.RequireAdmin).
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsActive — статус аккаунта, должен проверяться при login.
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// LegacyLimit возвращает grandfather-лимит для указанного поля плана,
// если он сохранён в users.legacy_quotas. ok=false → юзер получает
// актуальный лимит из своего плана (новые регистрации, или поля без
// grandfather'а). Безопасно к пустому/невалидному JSON — возвращает
// (0, false), не паникует.
//
// field соответствует имени колонки SubscriptionPlan, например
// "max_prompts", "max_ext_uses_daily", "max_mcp_uses_daily".
func (u *User) LegacyLimit(field string) (int, bool) {
	if len(u.LegacyQuotas) == 0 {
		return 0, false
	}
	var m map[string]json.Number
	if err := json.Unmarshal(u.LegacyQuotas, &m); err != nil {
		return 0, false
	}
	v, ok := m[field]
	if !ok {
		return 0, false
	}
	n, err := v.Int64()
	if err != nil {
		return 0, false
	}
	return int(n), true
}
