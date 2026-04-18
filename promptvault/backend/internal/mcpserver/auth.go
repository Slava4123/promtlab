package mcpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	authmw "promptvault/internal/middleware/auth"
	apikeyuc "promptvault/internal/usecases/apikey"
)

func APIKeyAuth(svc *apikeyuc.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body to 1 MB to prevent memory exhaustion.
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

			header := r.Header.Get("Authorization")
			if header == "" {
				writeAuthError(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				writeAuthError(w, "invalid authorization format")
				return
			}

			result, err := svc.ValidateKey(r.Context(), parts[1])
			if err != nil {
				// единая ошибка — нет oracle
				writeAuthError(w, "unauthorized")
				return
			}

			policy := result.Policy
			// Debug: каждый MCP-запрос — на проде это до 60 rpm/user. В INFO не хотим.
			slog.Debug("mcp.auth.success",
				"user_id", result.UserID,
				"key_id", result.KeyID,
				"read_only", policy.ReadOnly,
				"team_id", policy.TeamID,
				"tools_count", len(policy.AllowedTools),
				"has_expiry", policy.ExpiresAt != nil,
			)

			ctx := context.WithValue(r.Context(), authmw.UserIDKey, result.UserID)
			ctx = withKeyPolicy(ctx, &policy)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	// Кодируем через encoder — защита от поломки JSON при появлении кавычек в msg.
	if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
		slog.Debug("mcp.auth.write_error", "error", err)
	}
}
