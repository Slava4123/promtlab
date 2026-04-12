package apikey

import "time"

type APIKeyInfo struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type ValidateResult struct {
	UserID uint
	KeyID  uint
}
