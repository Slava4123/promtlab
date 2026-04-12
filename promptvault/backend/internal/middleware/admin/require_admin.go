package admin

import (
	"context"
	"log/slog"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
)

// UserLookup — узкий интерфейс того, что нужно RequireAdmin от user-слоя.
// Объявлен здесь (consumer side) чтобы не завязываться на usecases/user.
type UserLookup interface {
	GetByID(ctx context.Context, id uint) (*models.User, error)
}

// RequireAdmin — middleware, который проверяет role='admin' в БД на каждый
// admin request. Два этапа защиты:
//  1. userID берётся из ctx (установлен authmw.Middleware выше по цепочке).
//  2. Re-check из БД: даже если JWT claim говорит role=admin, мы делаем
//     свежий SELECT role FROM users WHERE id=? — защита от stale JWT
//     в случае demote'а админа до обычного юзера.
//
// ВАЖНО: RequireAdmin НЕ проверяет TOTP freshness. Это отдельная задача
// для более чувствительных operations (revoke_badge, reset_password) — их
// защита делается на уровне usecase/handler, не middleware (нужен больше
// granular контроль и возможность вернуть 401 с retry_totp hint).
func RequireAdmin(users UserLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := authmw.GetUserID(r.Context())
			if userID == 0 {
				httperr.Respond(w, httperr.Unauthorized("требуется авторизация"))
				return
			}

			u, err := users.GetByID(r.Context(), userID)
			if err != nil {
				if err == repo.ErrNotFound {
					httperr.Respond(w, httperr.Unauthorized("пользователь не найден"))
					return
				}
				slog.Error("admin.require_admin.user_lookup_failed", "user_id", userID, "error", err)
				httperr.Respond(w, httperr.Internal(err))
				return
			}

			if !u.IsAdmin() {
				slog.Warn("admin.require_admin.denied", "user_id", userID, "role", u.Role)
				httperr.Respond(w, httperr.Forbidden("доступ только для администраторов"))
				return
			}

			if !u.IsActive() {
				slog.Warn("admin.require_admin.frozen", "user_id", userID)
				httperr.Respond(w, httperr.Forbidden("аккаунт заблокирован"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
