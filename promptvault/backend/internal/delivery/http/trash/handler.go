package trash

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	analyticsuc "promptvault/internal/usecases/analytics"
	trashuc "promptvault/internal/usecases/trash"
)

type Handler struct {
	svc *trashuc.Service
	// insights — опциональный hot-refresh кэша Smart Insights после Restore.
	// nil-safe: если не подключён через SetInsightsRecomputer, recompute
	// пропускается и состояние догонит nightly cron loop.
	insights analyticsuc.InsightsRecomputer
}

func NewHandler(svc *trashuc.Service) *Handler {
	return &Handler{svc: svc}
}

// SetInsightsRecomputer подключает hot-refresh кэша Smart Insights.
// После POST /api/trash/prompt/{id}/restore пересчитываются все 7 типов
// в personal scope (teamID=nil): восстановленный промпт может попасть в
// unused/possible_duplicates/trending/declining/most_edited, его теги в
// orphan_tags, его коллекция в empty_collections. Вызывается из app.go.
func (h *Handler) SetInsightsRecomputer(r analyticsuc.InsightsRecomputer) {
	h.insights = r
}

// GET /api/trash
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	teamIDs := parseTeamIDs(r)

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	prompts, total, err := h.svc.ListDeletedPrompts(r.Context(), userID, teamIDs, page, pageSize)
	if err != nil {
		respondError(w, err)
		return
	}

	collections, err := h.svc.ListDeletedCollections(r.Context(), userID, teamIDs)
	if err != nil {
		respondError(w, err)
		return
	}

	type listResponse struct {
		Prompts     utils.PaginatedResponse[TrashPromptResponse] `json:"prompts"`
		Collections []TrashCollectionResponse                     `json:"collections"`
	}

	utils.WriteOK(w, listResponse{
		Prompts: utils.PaginatedResponse[TrashPromptResponse]{
			Items:    NewTrashPromptListResponse(prompts),
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			HasMore:  int64(page*pageSize) < total,
		},
		Collections: NewTrashCollectionListResponse(collections),
	})
}

// GET /api/trash/count
func (h *Handler) Count(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	teamIDs := parseTeamIDs(r)

	counts, err := h.svc.Count(r.Context(), userID, teamIDs)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, counts)
}

// POST /api/trash/{type}/{id}/restore
func (h *Handler) Restore(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	itemType, id, err := parseTypeAndID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if err := h.svc.Restore(r.Context(), itemType, id, userID); err != nil {
		respondError(w, err)
		return
	}

	// Hot-refresh Smart Insights кэша: восстановленный промпт возвращается
	// в библиотеку → может аффектить все 7 типов (он сам мог стать unused/
	// most_edited/trending/declining/possible_duplicates; его теги — orphan;
	// его коллекция перестаёт быть пустой). Recompute только для prompt:
	// collection restore не меняет insights (метрики строятся на промптах).
	// teamID=nil — personal scope. Ошибки swallow — recompute fail не должен
	// ломать RESTORE.
	if h.insights != nil && itemType == trashuc.TypePrompt {
		types := []string{
			models.InsightUnusedPrompts,
			models.InsightPossibleDuplicates,
			models.InsightTrending,
			models.InsightDeclining,
			models.InsightMostEdited,
			models.InsightOrphanTags,
			models.InsightEmptyCollections,
		}
		if rerr := h.insights.Recompute(r.Context(), userID, nil, types); rerr != nil {
			slog.WarnContext(r.Context(), "trash.restore.insights_recompute_failed",
				"err", rerr, "user_id", userID, "prompt_id", id)
		}
	}

	utils.WriteOK(w, map[string]string{"status": "restored"})
}

// DELETE /api/trash/{type}/{id}
func (h *Handler) PermanentDelete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	itemType, id, err := parseTypeAndID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if err := h.svc.PermanentDelete(r.Context(), itemType, id, userID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteNoContent(w)
}

// DELETE /api/trash
func (h *Handler) Empty(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	teamIDs := parseTeamIDs(r)

	deleted, err := h.svc.EmptyTrash(r.Context(), userID, teamIDs)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, map[string]int64{"deleted": deleted})
}

// ---------- helpers ----------

func parseTeamIDs(r *http.Request) []uint {
	tid := r.URL.Query().Get("team_id")
	if tid == "" {
		return nil
	}
	id, err := strconv.ParseUint(tid, 10, 32)
	if err != nil {
		return nil
	}
	return []uint{uint(id)}
}

func parseTypeAndID(r *http.Request) (trashuc.ItemType, uint, error) {
	t := trashuc.ItemType(chi.URLParam(r, "type"))
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return "", 0, trashuc.ErrInvalidType
	}
	switch t {
	case trashuc.TypePrompt, trashuc.TypeCollection:
		return t, uint(id), nil
	default:
		return "", 0, trashuc.ErrInvalidType
	}
}
