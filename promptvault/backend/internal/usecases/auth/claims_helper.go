package auth

import (
	"context"

	repo "promptvault/internal/interface/repository"
)

// claimsCtxKey — приватный alias для ключа ctx, в котором middleware/auth
// хранит *Claims. Здесь дублируем, чтобы не создать import-cycle
// (middleware/auth → usecases/auth).
type claimsCtxKey string

const ClaimsContextKey claimsCtxKey = "claims"

// FromContext возвращает Claims, помещённые в ctx middleware/auth.
// Для legacy-JWT (без поля PlanID) Claims.PlanID == "" — caller сам решает,
// делать ли fallback на DB.
//
// Так как middleware/auth и usecases/auth — два разных пакета с разным
// типом contextKey, реальный ключ передаётся через WithContext / FromContext
// взаимно: middleware ставит claims по своему key, чтение тут происходит
// через установленный middleware key. Чтобы их синхронизировать без cyclic
// import, приёмное место (analytics) делает context.Value(authmw.ClaimsKey).
//
// Эта функция — для тестов / non-HTTP кода, который сам кладёт Claims по
// ClaimsContextKey ниже.
func FromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return c, ok && c != nil
}

// WithContext — для тестов / встроенных вызовов: положить Claims в ctx
// под ключ ClaimsContextKey пакета auth (не middleware/auth.ClaimsKey).
func WithContext(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, ClaimsContextKey, c)
}

// PlanIDLookup — узкий интерфейс для DB-fallback'а.
type PlanIDLookup interface {
	PlanIDOf(ctx context.Context, userID uint) (string, error)
}

// PlanIDOrFallback пробует прочитать PlanID из claims (либо authmw.ClaimsKey,
// либо локального ClaimsContextKey). Если не найдено или поле пустое —
// делает DB lookup через users repo.
//
// authmwKey — context key, под которым middleware кладёт *Claims. Передаётся
// извне, чтобы избежать cyclic import middleware/auth → usecases/auth.
func PlanIDOrFallback(ctx context.Context, users repo.UserRepository, userID uint, authmwKey any) (string, error) {
	if c, ok := ctx.Value(authmwKey).(*Claims); ok && c != nil && c.PlanID != "" {
		return c.PlanID, nil
	}
	if c, ok := FromContext(ctx); ok && c.PlanID != "" {
		return c.PlanID, nil
	}
	user, err := users.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return user.PlanID, nil
}
