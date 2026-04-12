package models

import "time"

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
	DefaultModel          string     `gorm:"size:100;default:anthropic/claude-sonnet-4" json:"default_model"`
	TokenNonce            string     `gorm:"size:64" json:"-"`
	OnboardingCompletedAt *time.Time `json:"onboarding_completed_at,omitempty"`
	LastChangelogSeenAt   *time.Time `json:"last_changelog_seen_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`

	LinkedAccounts []LinkedAccount `gorm:"foreignKey:UserID" json:"linked_accounts,omitempty"`
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
