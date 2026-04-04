package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type linkedAccountRepo struct {
	db *gorm.DB
}

func NewLinkedAccountRepository(db *gorm.DB) *linkedAccountRepo {
	return &linkedAccountRepo{db: db}
}

func (r *linkedAccountRepo) Create(ctx context.Context, la *models.LinkedAccount) error {
	return r.db.WithContext(ctx).Create(la).Error
}

func (r *linkedAccountRepo) GetByUserID(ctx context.Context, userID uint) ([]models.LinkedAccount, error) {
	var accounts []models.LinkedAccount
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts).Error
	return accounts, err
}

func (r *linkedAccountRepo) GetByProviderID(ctx context.Context, provider, providerID string) (*models.LinkedAccount, error) {
	var la models.LinkedAccount
	err := r.db.WithContext(ctx).Where("provider = ? AND provider_id = ?", provider, providerID).First(&la).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repo.ErrNotFound
	}
	return &la, err
}

func (r *linkedAccountRepo) Delete(ctx context.Context, userID uint, provider string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND provider = ?", userID, provider).Delete(&models.LinkedAccount{}).Error
}

func (r *linkedAccountRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.LinkedAccount{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
