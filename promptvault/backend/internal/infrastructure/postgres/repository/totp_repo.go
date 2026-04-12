package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type totpRepo struct {
	db *gorm.DB
}

func NewTOTPRepository(db *gorm.DB) repo.TOTPRepository {
	return &totpRepo{db: db}
}

func (r *totpRepo) UpsertEnrollment(ctx context.Context, userID uint, secret string) error {
	// UPSERT по user_id (primary key). При существующей записи перезаписываем
	// secret + updated_at; confirmed_at сбрасывается в NULL (новый enroll =
	// требует повторного verify первым кодом).
	sql := `
		INSERT INTO user_totp (user_id, secret, confirmed_at, created_at, updated_at)
		VALUES (?, ?, NULL, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			secret = EXCLUDED.secret,
			confirmed_at = NULL,
			updated_at = NOW()
	`
	return r.db.WithContext(ctx).Exec(sql, userID, secret).Error
}

func (r *totpRepo) GetByUserID(ctx context.Context, userID uint) (*models.UserTOTP, error) {
	var t models.UserTOTP
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *totpRepo) MarkConfirmed(ctx context.Context, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.UserTOTP{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{
			"confirmed_at": now,
			"updated_at":   now,
		}).Error
}

func (r *totpRepo) Delete(ctx context.Context, userID uint) error {
	// CASCADE гарантируется через ON DELETE CASCADE в user_totp_backup_codes.user_id,
	// но в реале backup_codes.user_id ссылается на users(id), не на user_totp(user_id).
	// Поэтому чистим вручную в транзакции.
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserTOTPBackupCode{}).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", userID).Delete(&models.UserTOTP{}).Error
	})
}

func (r *totpRepo) ReplaceBackupCodes(ctx context.Context, userID uint, hashes []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Полная замена: удалить все существующие + вставить новые.
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserTOTPBackupCode{}).Error; err != nil {
			return err
		}
		if len(hashes) == 0 {
			return nil
		}
		codes := make([]models.UserTOTPBackupCode, 0, len(hashes))
		for _, h := range hashes {
			codes = append(codes, models.UserTOTPBackupCode{UserID: userID, CodeHash: h})
		}
		return tx.Create(&codes).Error
	})
}

func (r *totpRepo) ListActiveBackupCodes(ctx context.Context, userID uint) ([]models.UserTOTPBackupCode, error) {
	var codes []models.UserTOTPBackupCode
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND used_at IS NULL", userID).
		Order("created_at ASC").
		Find(&codes).Error
	return codes, err
}

func (r *totpRepo) MarkBackupCodeUsed(ctx context.Context, codeID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.UserTOTPBackupCode{}).
		Where("id = ? AND used_at IS NULL", codeID).
		Update("used_at", time.Now()).Error
}
