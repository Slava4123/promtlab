package repository

import (
	"context"
	"time"

	"promptvault/internal/models"
)

// AuditLogFilter — параметры для выборки записей audit_log в админ-панель.
// Nil-указатели и пустые строки означают «не фильтровать по этому полю».
type AuditLogFilter struct {
	AdminID    *uint
	Action     string
	TargetType string
	TargetID   *uint
	FromTime   *time.Time
	ToTime     *time.Time
	Page       int
	PageSize   int
}

// AuditRepository — append-only repo для audit_log.
// ВАЖНО: нет Update / Delete методов. Даже если кто-то добавит — REVOKE
// в миграции 000018 гарантирует что INSERT работает, а UPDATE/DELETE вернут
// permission denied на уровне БД. Тестируется в audit_repo_test.go.
type AuditRepository interface {
	// Log вставляет новую запись audit. created_at заполняется PG через DEFAULT NOW().
	Log(ctx context.Context, entry *models.AuditLog) error

	// List возвращает страницу записей по фильтру с общим count.
	// Упорядочено по created_at DESC (свежие сверху).
	List(ctx context.Context, filter AuditLogFilter) ([]models.AuditLog, int64, error)
}
