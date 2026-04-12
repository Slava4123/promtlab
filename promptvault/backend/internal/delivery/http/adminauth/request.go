package adminauth

// ConfirmEnrollmentRequest — POST /api/admin/totp/verify-enrollment
// Юзер вводит первый 6-значный код из Authenticator после сканирования QR.
// Backup codes НЕ принимаются на этом этапе — enrollment confirm требует
// доказательства что secret корректно добавлен в Authenticator (т.е.
// клиент способен сгенерировать TOTP из того же secret).
type ConfirmEnrollmentRequest struct {
	Code string `json:"code" validate:"required,len=6,numeric"`
}
