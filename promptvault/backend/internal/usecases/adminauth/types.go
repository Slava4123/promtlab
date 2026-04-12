package adminauth

// EnrollResult — данные, возвращаемые при первом вызове Enroll.
// BackupCodes — 10 plaintext-строк, **показываются юзеру ОДИН РАЗ** и
// больше нигде не хранятся (в БД только bcrypt-хеши).
type EnrollResult struct {
	// Secret — base32-encoded TOTP secret. Показывается юзеру для
	// ручного ввода (если QR не отсканировать).
	Secret string `json:"secret"`

	// QRURL — otpauth:// URL для QR-кода (Google Authenticator, 1Password, Authy).
	QRURL string `json:"qr_url"`

	// BackupCodes — 10 одноразовых recovery-кодов в plaintext.
	// Юзер должен сохранить их в безопасном месте (password manager, bank cell).
	BackupCodes []string `json:"backup_codes"`
}

// VerifyResult — возвращается из Verify. UsedBackupCode=true если код
// был распознан как backup (не TOTP). UI может показать баннер
// «осталось N backup кодов».
type VerifyResult struct {
	UsedBackupCode       bool `json:"used_backup_code"`
	RemainingBackupCodes int  `json:"remaining_backup_codes"`
}
