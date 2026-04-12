package apikey

import (
	"time"

	apikeyuc "promptvault/internal/usecases/apikey"
)

type APIKeyResponse struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreatedAPIKeyResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	KeyPrefix string    `json:"key_prefix"`
	CreatedAt time.Time `json:"created_at"`
}

type ListResponse struct {
	Keys    []APIKeyResponse `json:"keys"`
	MaxKeys int              `json:"max_keys"`
}

func toAPIKeyResponse(info apikeyuc.APIKeyInfo) APIKeyResponse {
	return APIKeyResponse{
		ID:         info.ID,
		Name:       info.Name,
		KeyPrefix:  info.KeyPrefix,
		LastUsedAt: info.LastUsedAt,
		CreatedAt:  info.CreatedAt,
	}
}
