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
	err := r.db.WithContext(ctx).
		Preload("Prompt.Tags").
		Preload("Prompt.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "avatar_url")
		}).
		Where("share_links.token = ? AND share_links.is_active = true AND (share_links.expires_at IS NULL OR share_links.expires_at > ?)", token, time.Now()).
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
