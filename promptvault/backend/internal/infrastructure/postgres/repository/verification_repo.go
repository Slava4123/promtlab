package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type verificationRepo struct {
	db *gorm.DB
}

func NewVerificationRepository(db *gorm.DB) repo.VerificationRepository {
	return &verificationRepo{db: db}
}

func (r *verificationRepo) Create(ctx context.Context, v *models.EmailVerification) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *verificationRepo) GetByUserID(ctx context.Context, userID uint) (*models.EmailVerification, error) {
	var v models.EmailVerification
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *verificationRepo) IncrementAttempts(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&models.EmailVerification{}).Where("id = ?", id).
		UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error
}

func (r *verificationRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.EmailVerification{}).Error
}
