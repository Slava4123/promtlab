package audit

import "errors"

// ErrMissingRequestInfo возвращается из audit.Log когда в context нет
// AdminRequestInfo (не был применён AdminAuditContext middleware). Это
// bug — ошибка programmer'а, а не runtime-сценарий. В production такого
// быть не должно: все admin endpoints обёрнуты middleware chain.
var ErrMissingRequestInfo = errors.New("audit: AdminRequestInfo not in context (did you forget AdminAuditContext middleware?)")
