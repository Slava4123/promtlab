package apikey

import (
	"slices"
	"time"
)

// KeyPolicy — scope-параметры API-ключа. Копия из БД, живёт в ctx на время запроса.
// Zero-value (ReadOnly=false, всё остальное nil) означает полный доступ — backward-compat.
type KeyPolicy struct {
	ReadOnly     bool
	TeamID       *uint
	AllowedTools []string
	ExpiresAt    *time.Time
}

// IsToolAllowed возвращает true, если tool разрешён текущим whitelist'ом.
// nil/пустой AllowedTools означает "все разрешены".
func (p *KeyPolicy) IsToolAllowed(toolName string) bool {
	if p == nil || len(p.AllowedTools) == 0 {
		return true
	}
	return slices.Contains(p.AllowedTools, toolName)
}

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
