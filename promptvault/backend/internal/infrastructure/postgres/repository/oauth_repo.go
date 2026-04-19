package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// ---------------------------------------------------------------------------
// OAuth clients
// ---------------------------------------------------------------------------

type oauthClientRepo struct {
	db *gorm.DB
}

func NewOAuthClientRepository(db *gorm.DB) *oauthClientRepo {
	return &oauthClientRepo{db: db}
}

func (r *oauthClientRepo) Create(ctx context.Context, c *models.OAuthClient) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *oauthClientRepo) GetByClientID(ctx context.Context, clientID string) (*models.OAuthClient, error) {
	var c models.OAuthClient
	if err := r.db.WithContext(ctx).Where("client_id = ?", clientID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *oauthClientRepo) UpdateLastUsed(ctx context.Context, clientID string) error {
	return r.db.WithContext(ctx).
		Model(&models.OAuthClient{}).
		Where("client_id = ?", clientID).
		Update("last_used_at", time.Now()).Error
}

func (r *oauthClientRepo) Delete(ctx context.Context, clientID string) error {
	result := r.db.WithContext(ctx).Where("client_id = ?", clientID).Delete(&models.OAuthClient{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// OAuth authorization codes
// ---------------------------------------------------------------------------

type oauthAuthCodeRepo struct {
	db *gorm.DB
}

func NewOAuthAuthorizationCodeRepository(db *gorm.DB) *oauthAuthCodeRepo {
	return &oauthAuthCodeRepo{db: db}
}

func (r *oauthAuthCodeRepo) Create(ctx context.Context, c *models.OAuthAuthorizationCode) error {
	return r.db.WithContext(ctx).Create(c).Error
}

// Consume — атомарная операция: UPDATE ... SET used_at=NOW WHERE used_at IS NULL ...
// RETURNING * (через Scan после Update). Позволяет отловить race replay.
func (r *oauthAuthCodeRepo) Consume(ctx context.Context, codeHash string) (*models.OAuthAuthorizationCode, error) {
	var code models.OAuthAuthorizationCode
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("code_hash = ? AND used_at IS NULL AND expires_at > ?", codeHash, time.Now()).
			First(&code).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repo.ErrNotFound
			}
			return err
		}
		now := time.Now()
		code.UsedAt = &now
		return tx.Model(&code).
			Where("code_hash = ? AND used_at IS NULL", codeHash).
			Update("used_at", now).Error
	})
	if err != nil {
		return nil, err
	}
	return &code, nil
}

func (r *oauthAuthCodeRepo) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.OAuthAuthorizationCode{})
	return result.RowsAffected, result.Error
}

// ---------------------------------------------------------------------------
// OAuth tokens
// ---------------------------------------------------------------------------

type oauthTokenRepo struct {
	db *gorm.DB
}

func NewOAuthTokenRepository(db *gorm.DB) *oauthTokenRepo {
	return &oauthTokenRepo{db: db}
}

func (r *oauthTokenRepo) Create(ctx context.Context, t *models.OAuthToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *oauthTokenRepo) GetByHash(ctx context.Context, hash string) (*models.OAuthToken, error) {
	var t models.OAuthToken
	if err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *oauthTokenRepo) Revoke(ctx context.Context, hash string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.OAuthToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Update("revoked_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

// RevokeChain — рекурсивный CTE. При обнаружении replay'а refresh-токена
// вызываем, чтобы закрыть всю цепочку потомков (refresh rotation breach).
func (r *oauthTokenRepo) RevokeChain(ctx context.Context, parentID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Exec(`
		WITH RECURSIVE chain AS (
			SELECT id FROM oauth_tokens WHERE id = ?
			UNION ALL
			SELECT t.id FROM oauth_tokens t
			INNER JOIN chain c ON t.parent_token_id = c.id
		)
		UPDATE oauth_tokens
		   SET revoked_at = ?
		 WHERE id IN (SELECT id FROM chain) AND revoked_at IS NULL
	`, parentID, now).Error
}

func (r *oauthTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -30)
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", cutoff).
		Delete(&models.OAuthToken{})
	return result.RowsAffected, result.Error
}
