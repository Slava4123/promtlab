package models

import (
	"slices"
	"time"
)

// Policy — scope-параметры токена/ключа, переиспользуются между API-ключами
// (pvlt_*) и OAuth access-токенами. Zero-value означает полный доступ — это
// backward-compat для ключей, созданных до миграции 000035.
type Policy struct {
	ReadOnly     bool       `json:"read_only"`
	TeamID       *uint      `json:"team_id,omitempty"`
	AllowedTools []string   `json:"allowed_tools,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// IsToolAllowed возвращает true, если tool разрешён текущим whitelist'ом.
// nil/пустой AllowedTools означает "все разрешены".
func (p *Policy) IsToolAllowed(toolName string) bool {
	if p == nil || len(p.AllowedTools) == 0 {
		return true
	}
	return slices.Contains(p.AllowedTools, toolName)
}
