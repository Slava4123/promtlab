package audit

import "context"

// AdminRequestInfo — метаданные админ-запроса, прокидываемые через context
// из middleware/admin.AdminAuditContext. Usecase-слой читает их через
// FromContext при вызове Service.Log.
type AdminRequestInfo struct {
	AdminID   uint
	IP        string
	UserAgent string
}

// adminCtxKey — unexported тип для исключения ключей-коллизий в context
// (идиоматично для Go: см. net/http contextKey).
type adminCtxKey struct{}

// WithContext кладёт AdminRequestInfo в ctx. Вызывается из middleware.
func WithContext(ctx context.Context, info AdminRequestInfo) context.Context {
	return context.WithValue(ctx, adminCtxKey{}, info)
}

// FromContext извлекает AdminRequestInfo. Второй return — false если ctx
// не содержит AdminRequestInfo (middleware не применён к данному route).
func FromContext(ctx context.Context) (AdminRequestInfo, bool) {
	info, ok := ctx.Value(adminCtxKey{}).(AdminRequestInfo)
	return info, ok
}
