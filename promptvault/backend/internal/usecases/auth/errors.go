package auth

import "errors"

// EmailTakenError содержит имя провайдера, через который уже зарегистрирован email
type EmailTakenError struct {
	Provider string
}

func (e *EmailTakenError) Error() string {
	if e.Provider != "" {
		return "Этот email уже используется. Войдите через " + e.Provider + " и добавьте пароль в настройках"
	}
	return "Этот email уже зарегистрирован"
}

var (
	ErrInvalidCredentials = errors.New("Неверный email или пароль")
	ErrInvalidToken       = errors.New("Недействительный токен")
	ErrExpiredToken       = errors.New("Токен истёк")
	ErrUserNotFound       = errors.New("Пользователь не найден")
	ErrUsernameTaken      = errors.New("Это имя пользователя уже занято")

	// Verification
	ErrInvalidCode        = errors.New("Неверный код подтверждения")
	ErrExpiredCode        = errors.New("Код подтверждения истёк")
	ErrTooManyAttempts    = errors.New("Слишком много попыток. Запросите новый код")
	ErrEmailNotVerified   = errors.New("Email не подтверждён")

	ErrCannotUnlinkLast   = errors.New("Нельзя отвязать последний способ входа")
	ErrWrongPassword      = errors.New("Неверный текущий пароль")
	ErrNoPassword         = errors.New("Пароль не установлен. Используйте \"Установить пароль\"")
	ErrPasswordAlreadySet = errors.New("Пароль уже установлен. Используйте \"Изменить пароль\"")

	// OAuth
	ErrOAuthNotConfigured    = errors.New("OAuth-провайдер не настроен")
	ErrOAuthExchangeFailed   = errors.New("Ошибка авторизации через провайдер")
	ErrOAuthProfileFailed    = errors.New("Не удалось получить профиль")
	ErrOAuthStateMismatch    = errors.New("Ошибка безопасности OAuth")

	// Account Linking
	ErrProviderLinkedToOther = errors.New("Этот аккаунт уже привязан к другому пользователю")
	ErrProviderAlreadyLinked = errors.New("Этот провайдер уже привязан")
)
