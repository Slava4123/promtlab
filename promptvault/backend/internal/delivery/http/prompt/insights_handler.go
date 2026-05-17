package prompt

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	httperr "promptvault/internal/delivery/http/errors"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/usecases/prompt_insights"
)

// InsightsService — узкий интерфейс на *prompt_insights.Service.
// Описывает только методы, нужные HTTP-слою. Позволяет подменять fake
// в тестах handlers без нагрузки реального usecase.
type InsightsService interface {
	ListUnused(ctx context.Context, userID uint, teamID *uint, limit int) ([]prompt_insights.PromptInsightRow, error)
	ListDuplicates(ctx context.Context, userID uint, teamID *uint, limit int) ([]prompt_insights.DuplicatePair, error)
	ListTrending(ctx context.Context, userID uint, teamID *uint, limit int) ([]prompt_insights.PromptInsightRow, error)
	ListDeclining(ctx context.Context, userID uint, teamID *uint, limit int) ([]prompt_insights.PromptInsightRow, error)
	ListMostEdited(ctx context.Context, userID uint, teamID *uint, limit int) ([]prompt_insights.PromptInsightRow, error)
	MergePrompts(ctx context.Context, userID, keepID, mergeID uint) error
}

// InsightsHandler — 5 GET endpoints для prompt insights + (B9) Merge POST.
// Pro-gating живёт в usecase (B3-B5): handler только транслирует sentinel
// errors в HTTP-коды.
type InsightsHandler struct {
	svc InsightsService
}

func NewInsightsHandler(svc InsightsService) *InsightsHandler {
	return &InsightsHandler{svc: svc}
}

// Unused — GET /api/prompts/insights/unused?team_id=&limit=
func (h *InsightsHandler) Unused(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	teamID := parseTeamID(r)
	limit := parseLimit(r, 50)
	rows, err := h.svc.ListUnused(r.Context(), userID, teamID, limit)
	if err != nil {
		respondInsightsError(w, r, err)
		return
	}
	slog.Info("prompt_insights.requested", "type", "unused", "user_id", userID, "items_count", len(rows))
	writeItems(w, rows)
}

// Duplicates — GET /api/prompts/insights/duplicates?team_id=&limit=
func (h *InsightsHandler) Duplicates(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	teamID := parseTeamID(r)
	limit := parseLimit(r, 20)
	pairs, err := h.svc.ListDuplicates(r.Context(), userID, teamID, limit)
	if err != nil {
		respondInsightsError(w, r, err)
		return
	}
	slog.Info("prompt_insights.requested", "type", "duplicates", "user_id", userID, "items_count", len(pairs))
	writeItems(w, pairs)
}

// Trending — GET /api/prompts/insights/trending?team_id=&limit=
func (h *InsightsHandler) Trending(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	teamID := parseTeamID(r)
	limit := parseLimit(r, 10)
	rows, err := h.svc.ListTrending(r.Context(), userID, teamID, limit)
	if err != nil {
		respondInsightsError(w, r, err)
		return
	}
	slog.Info("prompt_insights.requested", "type", "trending", "user_id", userID, "items_count", len(rows))
	writeItems(w, rows)
}

// Declining — GET /api/prompts/insights/declining?team_id=&limit=
func (h *InsightsHandler) Declining(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	teamID := parseTeamID(r)
	limit := parseLimit(r, 10)
	rows, err := h.svc.ListDeclining(r.Context(), userID, teamID, limit)
	if err != nil {
		respondInsightsError(w, r, err)
		return
	}
	slog.Info("prompt_insights.requested", "type", "declining", "user_id", userID, "items_count", len(rows))
	writeItems(w, rows)
}

// MostEdited — GET /api/prompts/insights/most-edited?team_id=&limit=
func (h *InsightsHandler) MostEdited(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	teamID := parseTeamID(r)
	limit := parseLimit(r, 10)
	rows, err := h.svc.ListMostEdited(r.Context(), userID, teamID, limit)
	if err != nil {
		respondInsightsError(w, r, err)
		return
	}
	slog.Info("prompt_insights.requested", "type", "most_edited", "user_id", userID, "items_count", len(rows))
	writeItems(w, rows)
}

// Merge — POST /api/prompts/{id}/merge-with/{other_id}.
// id остаётся, other_id soft-удаляется. Возвращает {kept_id, merged_id}.
func (h *InsightsHandler) Merge(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	keepID, err := parsePathUint(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный id"))
		return
	}
	mergeID, err := parsePathUint(r, "other_id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный other_id"))
		return
	}
	if err := h.svc.MergePrompts(r.Context(), userID, keepID, mergeID); err != nil {
		respondInsightsError(w, r, err)
		return
	}
	slog.Info("prompt_insights.merge", "user_id", userID, "kept_id", keepID, "merged_id", mergeID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]uint{"kept_id": keepID, "merged_id": mergeID})
}

// parsePathUint — парсит chi URL-param как uint32.
func parsePathUint(r *http.Request, name string) (uint, error) {
	s := chi.URLParam(r, name)
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

// parseTeamID — парсит ?team_id= в *uint, nil на ошибку/пусто (personal scope).
func parseTeamID(r *http.Request) *uint {
	q := r.URL.Query().Get("team_id")
	if q == "" {
		return nil
	}
	id, err := strconv.ParseUint(q, 10, 32)
	if err != nil {
		return nil
	}
	u := uint(id)
	return &u
}

// parseLimit — парсит ?limit=, fallback на def при пустом/невалидном.
// Реальный clamp [1,max] делает usecase.clampLimit.
func parseLimit(r *http.Request, def int) int {
	q := r.URL.Query().Get("limit")
	if q == "" {
		return def
	}
	v, err := strconv.Atoi(q)
	if err != nil || v <= 0 {
		return def
	}
	return v
}

// writeItems — JSON envelope {"items": [...]} (consistent с analytics endpoints).
func writeItems[T any](w http.ResponseWriter, items []T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]any{"items": items}); err != nil {
		slog.Error("prompt_insights.encode_failed", "err", err)
	}
}

// respondInsightsError — маппинг доменных ошибок prompt_insights → HTTP.
//   - ErrProRequired   → 402 с plan="pro" (паттерн как в analytics для upgrade prompt)
//   - ErrPromptsNotOwned → 404 (используется в Merge handler, B9)
//   - ErrSamePrompt    → 400 (используется в Merge handler, B9)
//   - default          → 500 (с захватом в Sentry через RespondWithRequest).
func respondInsightsError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, prompt_insights.ErrProRequired):
		// Pattern совпадает с analytics.respondTierRequired: 402 с feature/plan
		// для frontend upgrade prompt (CTA на /pricing).
		httperr.RespondQuotaError(w, "premium_feature", 0, 0, "pro",
			"Фича доступна на тарифе Pro. Обновите план на /pricing.")
	case errors.Is(err, prompt_insights.ErrPromptsNotOwned):
		httperr.Respond(w, httperr.NotFound("Промпт не найден"))
	case errors.Is(err, prompt_insights.ErrSamePrompt):
		httperr.Respond(w, httperr.BadRequest("Нельзя объединить промпт сам с собой"))
	default:
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
	}
}
