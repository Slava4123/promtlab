package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type shareLinkRepo struct {
	db *gorm.DB
}

func NewShareLinkRepository(db *gorm.DB) *shareLinkRepo {
	return &shareLinkRepo{db: db}
}

func (r *shareLinkRepo) Create(ctx context.Context, link *models.ShareLink) error {
	return r.db.WithContext(ctx).Create(link).Error
}

func (r *shareLinkRepo) GetByToken(ctx context.Context, token string) (*models.ShareLink, error) {
	var link models.ShareLink
	// Phase 16-Y: фильтр по expires_at делается в usecase (для 410 Gone vs 404).
	// Здесь только token + is_active. Revoked (is_active=false) → 404 одинаково
	// с несуществующим — это manual-отзыв, не TTL-просрочка.
	err := r.db.WithContext(ctx).
		Preload("Prompt.Tags").
		Preload("Prompt.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "avatar_url")
		}).
		Where("share_links.token = ? AND share_links.is_active = true", token).
		First(&link).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &link, nil
}

func (r *shareLinkRepo) GetActiveByPromptID(ctx context.Context, promptID uint) (*models.ShareLink, error) {
	var link models.ShareLink
	err := r.db.WithContext(ctx).
		Where("prompt_id = ? AND is_active = true", promptID).
		First(&link).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &link, nil
}

func (r *shareLinkRepo) Deactivate(ctx context.Context, promptID uint) error {
	result := r.db.WithContext(ctx).
		Model(&models.ShareLink{}).
		Where("prompt_id = ? AND is_active = true", promptID).
		Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *shareLinkRepo) IncrementViewCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.ShareLink{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"view_count":     gorm.Expr("view_count + 1"),
			"last_viewed_at": time.Now(),
		}).Error
}

// CleanupExpired — Phase 16-Y. Hard DELETE для ссылок, чьи expires_at < now()-grace.
// expires_at IS NULL — это «бессрочные» ссылки (Max-эксклюзив), их не трогаем.
// На больших объёмах батчинг можно добавить через DELETE ... LIMIT (Postgres
// требует CTE + ID для этого), но при ≤1M строк это излишне.
func (r *shareLinkRepo) CleanupExpired(ctx context.Context, grace time.Duration) (int64, error) {
	cutoff := time.Now().Add(-grace)
	res := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", cutoff).
		Delete(&models.ShareLink{})
	return res.RowsAffected, res.Error
}
