package trash

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	trashuc "promptvault/internal/usecases/trash"
)

type Handler struct {
	svc *trashuc.Service
}

func NewHandler(svc *trashuc.Service) *Handler {
	return &Handler{svc: svc}
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
