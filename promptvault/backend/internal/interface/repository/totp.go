package repository

import (
	"context"

	"promptvault/internal/models"
)

// TOTPRepository инкапсулирует работу с user_totp и user_totp_backup_codes.
// Используется из usecases/adminauth для enroll/verify flow.
type TOTPRepository interface {
	// UpsertEnrollment создаёт новый TOTP secret или перезаписывает существующий.
	// Перезапись допустима только для неподтверждённых записей (confirmed_at IS NULL).
	// Для подтверждённого TOTP re-enroll требует сначала Delete (политика: смена
	// устройства = полный re-setup со всеми backup codes).
	UpsertEnrollment(ctx context.Context, userID uint, secret string) error

	// GetByUserID возвращает TOTP-запись или repo.ErrNotFound если нет enrollment.
	GetByUserID(ctx context.Context, userID uint) (*models.UserTOTP, error)

	// MarkConfirmed устанавливает confirmed_at = NOW() после успешного verify
	// первым кодом из Authenticator. Идемпотентно.
	MarkConfirmed(ctx context.Context, userID uint) error

	// Delete удаляет TOTP + все backup codes (для disable/reset TOTP flow).
	Delete(ctx context.Context, userID uint) error

	// --- Backup codes ---

	// ReplaceBackupCodes атомарно удаляет все существующие коды юзера и создаёт
	// новые (используется при enroll и regenerate). hashes — bcrypt-хеши.
	ReplaceBackupCodes(ctx context.Context, userID uint, hashes []string) error

	// ListActiveBackupCodes возвращает неиспользованные backup codes для юзера.
	// Нужно для проверки при входе: итерируемся и bcrypt.Compare по каждому.
	ListActiveBackupCodes(ctx context.Context, userID uint) ([]models.UserTOTPBackupCode, error)

	// MarkBackupCodeUsed устанавливает used_at = NOW() по ID кода.
	// Вызывается после успешного bcrypt.Compare.
	MarkBackupCodeUsed(ctx context.Context, codeID uint) error
}
