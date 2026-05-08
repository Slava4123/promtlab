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

// DeleteIfMethodsRemain — atomic compare-and-delete для UnlinkProvider.
// Оборачивает в transaction с SELECT FOR UPDATE на linked_accounts(user_id)
// — гарантирует, что между подсчётом и DELETE никто другой не удалил
// одновременно последний linked_account, оставив юзера без login methods.
//
// Логика проверки «остаётся хотя бы один способ войти»:
//   - hasPassword=true: всегда OK (пароль остаётся как login method);
//   - hasPassword=false: count ДОЛЖЕН быть > 1 (после delete останется ≥ 1).
func (r *linkedAccountRepo) DeleteIfMethodsRemain(ctx context.Context, userID uint, provider string, hasPassword bool) (bool, error) {
	var deleted bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// SELECT FOR UPDATE — лочит все linked_accounts юзера; concurrent
		// unlink ждёт COMMIT/ROLLBACK. Pg использует row-level locks.
		var count int64
		if err := tx.Raw(
			"SELECT COUNT(*) FROM linked_accounts WHERE user_id = ? FOR UPDATE",
			userID,
		).Scan(&count).Error; err != nil {
			return err
		}
		// totalMethods после удаления = (count - 1) + (hasPassword ? 1 : 0)
		// Должен быть >= 1.
		remaining := count - 1
		if hasPassword {
			remaining++
		}
		if remaining < 1 {
			return nil // deleted остаётся false → caller вернёт ErrCannotUnlinkLast
		}
		res := tx.Where("user_id = ? AND provider = ?", userID, provider).Delete(&models.LinkedAccount{})
		if res.Error != nil {
			return res.Error
		}
		deleted = res.RowsAffected > 0
		return nil
	})
	return deleted, err
}
