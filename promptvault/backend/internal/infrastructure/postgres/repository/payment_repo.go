package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type paymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *paymentRepo {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepo) GetByExternalID(ctx context.Context, provider, externalID string) (*models.Payment, error) {
	var p models.Payment
	err := r.db.WithContext(ctx).
		Where("provider = ? AND external_id = ?", provider, externalID).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) GetByIdempotencyKey(ctx context.Context, key string) (*models.Payment, error) {
	var p models.Payment
	err := r.db.WithContext(ctx).
		Where("idempotency_key = ?", key).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) UpdateStatus(ctx context.Context, id uint, status models.PaymentStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *paymentRepo) UpdateExternalID(ctx context.Context, id uint, externalID string) error {
	return r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id = ?", id).
		Update("external_id", externalID).Error
}

func (r *paymentRepo) TransitionStatus(ctx context.Context, id uint, expected, next models.PaymentStatus) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id = ? AND status = ?", id, expected).
		Update("status", next)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *paymentRepo) LinkSubscription(ctx context.Context, paymentID, subscriptionID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id = ?", paymentID).
		Update("subscription_id", subscriptionID).Error
}
