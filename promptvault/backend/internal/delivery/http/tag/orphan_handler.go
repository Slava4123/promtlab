package tag

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
)

// OrphanAnalyticsRepo — узкий интерфейс на AnalyticsRepository.OrphanTags.
// Service-level wrapping не нужно: orphan-теги — простой list без gating,
// handler сразу зовёт analytics_repo SQL.
type OrphanAnalyticsRepo interface {
	OrphanTags(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.TagRow, error)
}

type OrphanHandler struct {
	analytics OrphanAnalyticsRepo
}

func NewOrphanHandler(analytics OrphanAnalyticsRepo) *OrphanHandler {
	return &OrphanHandler{analytics: analytics}
}

// List — GET /api/tags/orphan. Возвращает теги юзера без активных промптов.
// Limit hard-cap 100 (orphan'ов обычно мало, не нужна пагинация).
func (h *OrphanHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	tags, err := h.analytics.OrphanTags(r.Context(), userID, nil, 100)
	if err != nil {
		slog.Error("tag_orphan.failed", "err", err, "user_id", userID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	type item struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	items := make([]item, 0, len(tags))
	for _, t := range tags {
		items = append(items, item{ID: t.TagID, Name: t.Name})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
}
