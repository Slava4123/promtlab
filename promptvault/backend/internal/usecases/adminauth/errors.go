package adminauth

import "errors"

var (
	// ErrNotAdmin — попытка enroll TOTP / admin action от юзера с role='user'.
	ErrNotAdmin = errors.New("only admin users can enroll TOTP")

	// ErrInvalidCode — неверный 6-значный код из Authenticator (или backup code).
	ErrInvalidCode = errors.New("неверный код")

	// ErrTOTPNotEnrolled — у юзера нет TOTP enrollment. Верификация невозможна.
	ErrTOTPNotEnrolled = errors.New("TOTP не настроен для этого пользователя")

	// ErrTOTPAlreadyConfirmed — попытка Verify enrollment когда он уже confirmed.
	// Обычно значит что юзер прошёл enroll flow и пытается повторно verify
	// тот же secret — UI должен вести его в обычный login-TOTP flow.
	ErrTOTPAlreadyConfirmed = errors.New("TOTP уже подтверждён, используйте обычный вход")

	// ErrGenerateFailed — низкоуровневая ошибка генерации TOTP secret (rare).
	ErrGenerateFailed = errors.New("не удалось сгенерировать TOTP secret")
)
