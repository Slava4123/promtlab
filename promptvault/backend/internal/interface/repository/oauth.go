package repository

import (
	"context"

	"promptvault/internal/models"
)

// OAuthClientRepository — регистрация клиентов (RFC 7591 Dynamic + статическая).
type OAuthClientRepository interface {
	Create(ctx context.Context, client *models.OAuthClient) error
	GetByClientID(ctx context.Context, clientID string) (*models.OAuthClient, error)
	UpdateLastUsed(ctx context.Context, clientID string) error
	Delete(ctx context.Context, clientID string) error
}

// OAuthAuthorizationCodeRepository — 60-секундные one-time коды PKCE.
type OAuthAuthorizationCodeRepository interface {
	Create(ctx context.Context, code *models.OAuthAuthorizationCode) error
	// Consume атомарно помечает код использованным и возвращает запись.
	// ErrNotFound если код уже использован или не существует.
	Consume(ctx context.Context, codeHash string) (*models.OAuthAuthorizationCode, error)
	// DeleteExpired вызывается cron'ом для чистки.
	DeleteExpired(ctx context.Context) (int64, error)
}

// OAuthTokenRepository — access (JWT) + refresh (opaque).
type OAuthTokenRepository interface {
	Create(ctx context.Context, token *models.OAuthToken) error
	GetByHash(ctx context.Context, hash string) (*models.OAuthToken, error)
	// Revoke помечает revoked_at=NOW. Не удаляет для audit.
	Revoke(ctx context.Context, hash string) error
	// RevokeChain помечает revoked_at для всех потомков через parent_token_id
	// (используется при обнаружении replay refresh).
	RevokeChain(ctx context.Context, parentID uint) error
	// DeleteExpired чистит истёкшие >= 30 дней назад.
	DeleteExpired(ctx context.Context) (int64, error)
}
