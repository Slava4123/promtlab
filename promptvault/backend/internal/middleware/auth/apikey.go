package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

// APIKeyPrefix — префикс API-ключей PromptVault (см. usecases/apikey).
// Используется CombinedAuth для распознавания типа токена в заголовке.
const APIKeyPrefix = "pvlt_"

// APIKeyValidator — узкий интерфейс, который middleware ожидает от apikey usecase.
// Конкретная реализация (*apikeyuc.Service) оборачивается адаптером в app.go,
// чтобы middleware не зависел от usecases package.
type APIKeyValidator interface {
	ValidateKey(ctx context.Context, rawKey string) (userID uint, keyID uint, err error)
}

// APIKeyAuth — middleware аутентификации по заголовку `Authorization: Bearer pvlt_<key>`.
// При успехе кладёт userID в контекст под ключом UserIDKey (тот же, что JWT middleware).
func APIKeyAuth(v APIKeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				writeJSON(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], BearerScheme) {
				writeJSON(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			userID, keyID, err := v.ValidateKey(r.Context(), parts[1])
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			slog.Info("apikey.auth.success", "user_id", userID, "key_id", keyID)

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
