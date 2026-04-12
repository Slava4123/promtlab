package adminauth

import (
	adminauthuc "promptvault/internal/usecases/adminauth"
)

// EnrollResponse — POST /api/admin/totp/enroll
// ВАЖНО: BackupCodes показываются ОДИН РАЗ. После этого responsы не содержат
// их даже для того же юзера. Юзер должен сохранить в password manager.
type EnrollResponse struct {
	Secret      string   `json:"secret"`
	QRURL       string   `json:"qr_url"`
	BackupCodes []string `json:"backup_codes"`
}

// ConfirmEnrollmentResponse — ответ после успешного verify первого кода.
type ConfirmEnrollmentResponse struct {
	Confirmed bool `json:"confirmed"`
}

// RegenerateBackupCodesResponse — POST /api/admin/totp/backup-codes/regenerate.
// Старые коды инвалидируются, новые показываются один раз.
type RegenerateBackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
}

// StatusResponse — GET /api/admin/totp/status (опционально, может пригодиться фронту
// для показа «TOTP настроен / не настроен» без триггера enroll).
type StatusResponse struct {
	Enrolled  bool `json:"enrolled"`
	Confirmed bool `json:"confirmed"`
}

// NewEnrollResponse конвертит usecase result в transport DTO.
func NewEnrollResponse(r *adminauthuc.EnrollResult) EnrollResponse {
	return EnrollResponse{
		Secret:      r.Secret,
		QRURL:       r.QRURL,
		BackupCodes: r.BackupCodes,
	}
}
