package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(`{"error":"` + msg + `"}`)); err != nil {
		slog.Error("failed to write auth error response", "error", err)
	}
}

func Middleware(v TokenValidator) func(http.Handler) http.Handler {
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

			claims, err := v.ValidateAccessToken(parts[1])
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) uint {
	id, ok := ctx.Value(UserIDKey).(uint)
	if !ok || id == 0 {
		return 0
	}
	return id
}
