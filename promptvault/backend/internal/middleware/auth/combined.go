package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

// CombinedAuth принимает и JWT, и API-ключ в заголовке `Authorization: Bearer`.
// Распознавание: если токен начинается с APIKeyPrefix — валидация через apiKeys,
// иначе — через JWT validator. В обоих случаях userID кладётся в контекст под UserIDKey.
//
// Используется на /api/* routes, где Chrome Extension и SPA делят одни и те же endpoints.
func CombinedAuth(jwt TokenValidator, apiKeys APIKeyValidator) func(http.Handler) http.Handler {
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

			token := parts[1]
			var userID uint

			if strings.HasPrefix(token, APIKeyPrefix) {
				uid, keyID, err := apiKeys.ValidateKey(r.Context(), token)
				if err != nil {
					writeJSON(w, http.StatusUnauthorized, "unauthorized")
					return
				}
				slog.Info("apikey.auth.success",
					"user_id", uid,
					"key_id", keyID,
					"x_client", r.Header.Get("X-Client"),
					"path", r.URL.Path,
				)
				userID = uid
			} else {
				claims, err := jwt.ValidateAccessToken(token)
				if err != nil {
					writeJSON(w, http.StatusUnauthorized, "invalid or expired token")
					return
				}
				userID = claims.UserID
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
