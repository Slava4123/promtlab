package admin

import "errors"

var (
	// ErrCannotFreezeSelf — админ не может заморозить свой же аккаунт.
	// Защита от случайного self-lockout.
	ErrCannotFreezeSelf = errors.New("нельзя заморозить свой аккаунт")

	// ErrCannotRevokeSelfRole — админ не может забрать у себя role=admin через
	// admin API (откат только через CLI / прямой SQL).
	ErrCannotRevokeSelfRole = errors.New("нельзя снять роль admin с себя")

	// ErrUserNotFound — целевой юзер не найден.
	ErrUserNotFound = errors.New("пользователь не найден")

	// ErrBadgeNotFound — badge_id не существует в каталоге.
	ErrBadgeNotFound = errors.New("бейдж не найден")

	// ErrBadgeAlreadyUnlocked — попытка grant уже разблокированного бейджа.
	ErrBadgeAlreadyUnlocked = errors.New("бейдж уже разблокирован")

	// ErrInvalidStatus — передан неизвестный статус юзера.
	ErrInvalidStatus = errors.New("неверный статус пользователя")

	// ErrInvalidTier — передан несуществующий plan_id.
	ErrInvalidTier = errors.New("неверный тарифный план")

	// ErrEmailNotConfigured — попытка reset password при отключённом SMTP.
	ErrEmailNotConfigured = errors.New("SMTP не настроен, невозможно отправить email")
)
