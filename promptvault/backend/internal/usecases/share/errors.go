package share

import "errors"

var (
	ErrNotFound       = errors.New("Ссылка не найдена")
	ErrPromptNotFound = errors.New("Промпт не найден")
	ErrForbidden      = errors.New("Нет доступа к этому промпту")
	ErrViewerReadOnly = errors.New("Читатель не может делиться промптами")
	// ErrLinkExpired — ссылка просрочена (expires_at < now).
	// Phase 16-Y: HTTP-уровень мапит на 410 Gone; страница «ссылка истекла».
	ErrLinkExpired = errors.New("Срок действия ссылки истёк")
)
