package mcpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	apikeyuc "promptvault/internal/usecases/apikey"
	oauthsrv "promptvault/internal/usecases/oauth_server"
	"promptvault/internal/pkg/tokens"
)

// OAuthValidator — узкий интерфейс, которого достаточно для валидации OAuth
// access-токенов. Реализуется usecases/oauth_server.Service, мок в тестах.
type OAuthValidator interface {
	ValidateAccessToken(ctx context.Context, raw string) (*oauthsrv.ValidatedAccessToken, error)
}

// APIKeyAuth — middleware для /mcp. Принимает два формата Bearer-токенов:
//   - pvlt_* → статический API-ключ (valide через apikeyuc.Service)
//   - pvoat_* → OAuth 2.1 access token (valide через oauthsrv.Service)
//
// При 401 отдаётся WWW-Authenticate хедер с resource_metadata — per RFC 9728
// §5.1, MCP spec 2025-06-18 §Authorization Server Discovery.
// resourceMetadataURL — полный URL до /.well-known/oauth-protected-resource.
// Пустая строка отключит WWW-Authenticate (для backward-compat со старыми тестами).
func APIKeyAuth(apikeySvc *apikeyuc.Service, oauthSvc OAuthValidator, resourceMetadataURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body to 1 MB to prevent memory exhaustion.
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

			header := r.Header.Get("Authorization")
			if header == "" {
				writeAuthError(w, resourceMetadataURL, "invalid_request", "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				writeAuthError(w, resourceMetadataURL, "invalid_request", "invalid authorization format")
				return
			}

			raw := parts[1]

			// Маршрутизация по префиксу — один round-trip в БД независимо от типа.
			switch {
			case strings.HasPrefix(raw, tokens.PrefixAccessToken):
				if oauthSvc == nil {
					writeAuthError(w, resourceMetadataURL, "invalid_token", "oauth disabled")
					return
				}
				validated, err := oauthSvc.ValidateAccessToken(r.Context(), raw)
				if err != nil {
					writeAuthError(w, resourceMetadataURL, "invalid_token", "unauthorized")
					return
				}
				policy := decodeOAuthPolicy(validated.Policy)
				slog.Debug("mcp.auth.oauth.success",
					"user_id", validated.UserID,
					"client_id", validated.ClientID,
					"scope", validated.Scope,
				)
				ctx := context.WithValue(r.Context(), authmw.UserIDKey, validated.UserID)
				ctx = withKeyPolicy(ctx, &policy)
				next.ServeHTTP(w, r.WithContext(ctx))
				return

			default:
				// Default: API-key (pvlt_*). Остальные префиксы тоже попадут сюда
				// и получат unauthorized от ValidateKey — это нормально.
				result, err := apikeySvc.ValidateKey(r.Context(), raw)
				if err != nil {
					writeAuthError(w, resourceMetadataURL, "invalid_token", "unauthorized")
					return
				}
				policy := result.Policy
				slog.Debug("mcp.auth.apikey.success",
					"user_id", result.UserID,
					"key_id", result.KeyID,
					"read_only", policy.ReadOnly,
					"team_id", policy.TeamID,
					"tools_count", len(policy.AllowedTools),
				)
				ctx := context.WithValue(r.Context(), authmw.UserIDKey, result.UserID)
				ctx = withKeyPolicy(ctx, &policy)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}

// decodeOAuthPolicy парсит JSONB-поле в models.Policy. При ошибке — zero-value
// (полный доступ), чтобы не ломать клиента из-за повреждения данных в БД.
func decodeOAuthPolicy(raw json.RawMessage) models.Policy {
	var p models.Policy
	if len(raw) == 0 {
		return p
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		slog.Warn("mcp.auth.oauth.policy_decode_failed", "error", err)
	}
	return p
}

// writeAuthError возвращает 401 с WWW-Authenticate по RFC 9728 §5.1.
// MCP-клиенты (Claude.ai) парсят этот хедер чтобы найти authorization server.
func writeAuthError(w http.ResponseWriter, resourceMetadataURL, errorCode, msg string) {
	if resourceMetadataURL != "" {
		w.Header().Set("WWW-Authenticate",
			`Bearer realm="promptvault", error="`+errorCode+`", resource_metadata="`+resourceMetadataURL+`"`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
		slog.Debug("mcp.auth.write_error", "error", err)
	}
}
