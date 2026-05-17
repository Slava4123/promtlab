package collection

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
)

// EmptyAnalyticsRepo — узкий интерфейс на AnalyticsRepository.EmptyCollections.
// Service-level wrapping не нужно: empty-коллекции — простой list без gating,
// handler сразу зовёт analytics_repo SQL.
type EmptyAnalyticsRepo interface {
	EmptyCollections(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.CollectionRow, error)
}

type EmptyHandler struct {
	analytics EmptyAnalyticsRepo
}

func NewEmptyHandler(analytics EmptyAnalyticsRepo) *EmptyHandler {
	return &EmptyHandler{analytics: analytics}
}

// List — GET /api/collections/empty. Возвращает коллекции юзера без активных промптов.
// Limit hard-cap 100 (empty-коллекций обычно мало, не нужна пагинация).
func (h *EmptyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	rows, err := h.analytics.EmptyCollections(r.Context(), userID, nil, 100)
	if err != nil {
		slog.Error("collection_empty.failed", "err", err, "user_id", userID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	type item struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	items := make([]item, 0, len(rows))
	for _, row := range rows {
		items = append(items, item{ID: row.CollectionID, Name: row.Name})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
}
