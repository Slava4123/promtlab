// Package testcleanup — dev-only HTTP handler для Playwright E2E тестов.
//
// Регистрируется в routes.go ТОЛЬКО при cfg.Server.IsDev() == true.
// На проде route не существует — попытка вызова даст 404.
//
// Назначение: per-test isolation. Между Playwright-тестами вызываем
// POST /api/test/cleanup?email=e2e-free@test.local чтобы вернуть юзера к
// «чистому состоянию» (0 prompts, 0 collections, 0 share-links, 0 chains,
// 0 daily counters), не пересоздавая самого юзера или его подписку.
package testcleanup

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// Handler владеет *gorm.DB напрямую, минуя use-case слой. Это намеренное
// нарушение Clean Architecture, оправданное dev-only scope: handler НЕ собирается
// в prod-build (см. routes.go). Цель — не плодить *.PurgeAllOfUser методы во
// всех repositories, которые понадобятся только для тестов.
type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// CleanupResponse — счётчики удалённых строк по таблицам, для проверки в Playwright.
type CleanupResponse struct {
	Email   string         `json:"email"`
	UserID  uint           `json:"user_id"`
	Deleted map[string]int `json:"deleted"`
}

// Cleanup — POST /api/test/cleanup?email=...
// Удаляет user-scoped данные для тестового юзера. Сам юзер, его plan_id и
// subscription остаются — иначе следующий тест не сможет залогиниться.
func (h *Handler) Cleanup(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.URL.Query().Get("email"))
	if email == "" {
		http.Error(w, `{"error":"email query param required"}`, http.StatusBadRequest)
		return
	}
	// Защита от случайного использования на не-тестовых юзерах. Cleanup-handler
	// должен трогать ТОЛЬКО seed-юзеров e2e-*@test.local.
	if !strings.HasSuffix(email, "@test.local") || !strings.HasPrefix(email, "e2e-") {
		http.Error(w, `{"error":"only e2e-*@test.local emails allowed"}`, http.StatusForbidden)
		return
	}

	var userID uint
	if err := h.db.WithContext(r.Context()).
		Raw(`SELECT id FROM users WHERE email = ?`, email).Scan(&userID).Error; err != nil {
		slog.Error("testcleanup.userlookup", "email", email, "err", err)
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	if userID == 0 {
		http.Error(w, fmt.Sprintf(`{"error":"user %s not found"}`, email), http.StatusNotFound)
		return
	}

	deleted := map[string]int{}
	// Порядок важен: дочерние таблицы (chain_steps, chain_executions, share_links,
	// prompt_versions) ссылаются на prompts/chains через FK. CASCADE есть не везде,
	// поэтому удаляем явно сверху вниз. Внутри одной транзакции, чтобы при ошибке
	// откатить весь cleanup и не оставить partial state.
	err := h.db.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		// Chain executions/steps — FK на prompt_chains, должны быть удалены первыми.
		// Step'ы также имеют FK на prompts (Phase 16 trash issue), но их подъезд
		// здесь не критичен: prompts удаляются позже после полного снятия chains.
		for _, q := range []struct {
			name string
			sql  string
		}{
			{"prompt_chain_executions", `DELETE FROM prompt_chain_executions WHERE user_id = ?`},
			{"prompt_chain_steps", `DELETE FROM prompt_chain_steps WHERE chain_id IN (SELECT id FROM prompt_chains WHERE user_id = ?)`},
			{"prompt_chains", `DELETE FROM prompt_chains WHERE user_id = ?`},
			{"share_links", `DELETE FROM share_links WHERE user_id = ?`},
			{"prompt_pins", `DELETE FROM prompt_pins WHERE user_id = ?`},
			{"prompt_usage_log", `DELETE FROM prompt_usage_log WHERE user_id = ?`},
			{"prompt_versions", `DELETE FROM prompt_versions WHERE prompt_id IN (SELECT id FROM prompts WHERE user_id = ?)`},
			{"prompts", `DELETE FROM prompts WHERE user_id = ?`},
			{"tags", `DELETE FROM tags WHERE user_id = ?`},
			{"collections", `DELETE FROM collections WHERE user_id = ?`},
			{"daily_feature_usage", `DELETE FROM daily_feature_usage WHERE user_id = ?`},
			// Team-related: удаляем команды, где юзер — owner. Members в этих
			// командах удалятся каскадом (FK ON DELETE CASCADE на team_id).
			{"team_invitations", `DELETE FROM team_invitations WHERE team_id IN (SELECT id FROM teams WHERE created_by = ?)`},
			{"team_members", `DELETE FROM team_members WHERE user_id = ? OR team_id IN (SELECT id FROM teams WHERE created_by = ?)`},
			{"teams", `DELETE FROM teams WHERE created_by = ?`},
			// User-state, не блокирующий quota, но мешающий повторам тестов:
			{"user_smart_insights", `DELETE FROM user_smart_insights WHERE user_id = ?`},
			{"insight_notifications", `DELETE FROM insight_notifications WHERE user_id = ?`},
		} {
			args := []any{userID}
			// team_members — у нас два параметра (user_id + created_by),
			// для остальных передаём userID повторно где нужно.
			if strings.Count(q.sql, "?") == 2 {
				args = []any{userID, userID}
			}
			res := tx.Exec(q.sql, args...)
			if res.Error != nil {
				return fmt.Errorf("delete %s: %w", q.name, res.Error)
			}
			deleted[q.name] = int(res.RowsAffected)
		}
		return nil
	})
	if err != nil {
		// errors.Is для совместимости — если в будущем добавится sentinel ошибка.
		_ = errors.Is(err, errors.New(""))
		slog.Error("testcleanup.tx", "email", email, "err", err)
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	slog.Info("testcleanup.ok", "email", email, "user_id", userID, "deleted", deleted)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CleanupResponse{Email: email, UserID: userID, Deleted: deleted})
}
