package models

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

// OAuthClient — регистрация MCP-клиента (Claude.ai и т.п.) через RFC 7591.
// ClientSecretHash может быть пустым для public clients (PKCE-only flow).
type OAuthClient struct {
	ID                       uint           `gorm:"primaryKey" json:"id"`
	ClientID                 string         `gorm:"size:64;not null;uniqueIndex" json:"client_id"`
	ClientSecretHash         string         `gorm:"size:128" json:"-"`
	ClientName               string         `gorm:"size:200;not null" json:"client_name"`
	RedirectURIs             pq.StringArray `gorm:"type:text[];not null" json:"redirect_uris"`
	GrantTypes               pq.StringArray `gorm:"type:text[];not null" json:"grant_types"`
	ResponseTypes            pq.StringArray `gorm:"type:text[];not null" json:"response_types"`
	TokenEndpointAuthMethod  string         `gorm:"size:32;not null;default:none" json:"token_endpoint_auth_method"`
	Scope                    string         `gorm:"size:500;not null" json:"scope"`
	IsDynamic                bool           `gorm:"not null;default:true" json:"is_dynamic"`
	LastUsedAt               *time.Time     `json:"last_used_at,omitempty"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
}

// OAuthAuthorizationCode — short-lived (60 сек) one-time код PKCE-flow.
// CodeHash = SHA256(raw). Raw-код выдан только клиенту.
type OAuthAuthorizationCode struct {
	CodeHash            string         `gorm:"primaryKey;size:64" json:"-"`
	ClientID            string         `gorm:"size:64;not null;index" json:"client_id"`
	UserID              uint           `gorm:"not null;index" json:"user_id"`
	RedirectURI         string         `gorm:"size:512;not null" json:"redirect_uri"`
	CodeChallenge       string         `gorm:"size:128;not null" json:"-"`
	CodeChallengeMethod string         `gorm:"size:16;not null;default:S256" json:"code_challenge_method"`
	Scope               string         `gorm:"size:500;not null" json:"scope"`
	Resource            string         `gorm:"size:512;not null" json:"resource"`
	Policy              json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"policy"`
	ExpiresAt           time.Time      `gorm:"not null;index" json:"expires_at"`
	UsedAt              *time.Time     `json:"used_at,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
}

// OAuthToken — запись access или refresh токена. Access-JWT хранит jti как TokenHash,
// refresh — SHA256(raw_rf_*). ParentTokenID строит цепочку refresh-rotation.
type OAuthToken struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	TokenType     string         `gorm:"size:16;not null" json:"token_type"` // access | refresh
	TokenHash     string         `gorm:"size:64;not null;uniqueIndex" json:"-"`
	ClientID      string         `gorm:"size:64;not null;index" json:"client_id"`
	UserID        uint           `gorm:"not null;index" json:"user_id"`
	Scope         string         `gorm:"size:500;not null" json:"scope"`
	Resource      string         `gorm:"size:512;not null" json:"resource"`
	Policy        json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"policy"`
	ExpiresAt     time.Time      `gorm:"not null;index" json:"expires_at"`
	RevokedAt     *time.Time     `json:"revoked_at,omitempty"`
	ParentTokenID *uint          `json:"parent_token_id,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

// TableName-методы для явного контроля имён таблиц — GORM иначе даёт "oauth_tokens" →
// "o_auth_tokens" через snake_case правило (не всегда корректно для аббревиатур).

func (OAuthClient) TableName() string               { return "oauth_clients" }
func (OAuthAuthorizationCode) TableName() string    { return "oauth_authorization_codes" }
func (OAuthToken) TableName() string                { return "oauth_tokens" }
