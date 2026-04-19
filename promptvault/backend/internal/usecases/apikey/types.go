package apikey

import (
	"time"

	"promptvault/internal/models"
)

// KeyPolicy — алиас на shared models.Policy. Определение перенесено в models,
// чтобы переиспользовать тот же scope-тип для OAuth access-токенов без
// циклических импортов (usecases/oauth_server → models.Policy ← usecases/apikey).
type KeyPolicy = models.Policy

type APIKeyInfo struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	KeyPrefix    string     `json:"key_prefix"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	ReadOnly     bool       `json:"read_only"`
	TeamID       *uint      `json:"team_id,omitempty"`
	AllowedTools []string   `json:"allowed_tools,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

type CreateInput struct {
	UserID       uint
	Name         string
	ReadOnly     bool
	TeamID       *uint
	AllowedTools []string
	ExpiresAt    *time.Time
}

type ValidateResult struct {
	UserID uint
	KeyID  uint
	Policy KeyPolicy
}
