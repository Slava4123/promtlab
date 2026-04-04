package team

import "errors"

var (
	ErrNotFound             = errors.New("Команда не найдена")
	ErrForbidden            = errors.New("Нет доступа к этой команде")
	ErrNotOwner             = errors.New("Только владелец может выполнить это действие")
	ErrUserNotFound         = errors.New("Пользователь не найден")
	ErrAlreadyMember        = errors.New("Пользователь уже в команде")
	ErrAlreadyInvited       = errors.New("Приглашение уже отправлено")
	ErrInvitationNotFound   = errors.New("Приглашение не найдено")
	ErrCannotInviteSelf     = errors.New("Нельзя пригласить себя")
	ErrCannotRemoveOwner    = errors.New("Нельзя удалить владельца команды")
	ErrCannotChangeOwnerRole = errors.New("Нельзя изменить роль владельца")
	ErrInvalidRole          = errors.New("Недопустимая роль")
)
