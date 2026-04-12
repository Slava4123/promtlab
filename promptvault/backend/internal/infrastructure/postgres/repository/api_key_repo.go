package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type apiKeyRepo struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *apiKeyRepo {
	return &apiKeyRepo{db: db}
}

func (r *apiKeyRepo) Create(ctx context.Context, key *models.APIKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

func (r *apiKeyRepo) ListByUserID(ctx context.Context, userID uint) ([]models.APIKey, error) {
	var keys []models.APIKey
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&keys).Error
	return keys, err
}

func (r *apiKeyRepo) GetByHash(ctx context.Context, hash string) (*models.APIKey, error) {
	var key models.APIKey
	if err := r.db.WithContext(ctx).Where("key_hash = ?", hash).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepo) Delete(ctx context.Context, id, userID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.APIKey{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *apiKeyRepo) UpdateLastUsed(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", time.Now()).Error
}

func (r *apiKeyRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.APIKey{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
