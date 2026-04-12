package search

import (
	"net/http"
	"strconv"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	searchuc "promptvault/internal/usecases/search"
)

type Handler struct {
	svc *searchuc.Service
}

func NewHandler(svc *searchuc.Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/search?q=&team_id=
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	query := r.URL.Query().Get("q")

	var teamID *uint
	if tid := r.URL.Query().Get("team_id"); tid != "" {
		id, err := strconv.ParseUint(tid, 10, 32)
		if err != nil {
			httperr.Respond(w, httperr.BadRequest("Неверный team_id"))
			return
		}
		v := uint(id)
		teamID = &v
	}

	result, err := h.svc.Search(r.Context(), userID, teamID, query)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, result)
}

// GET /api/search/suggest?q=&team_id=
func (h *Handler) Suggest(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	prefix := r.URL.Query().Get("q")

	var teamID *uint
	if tid := r.URL.Query().Get("team_id"); tid != "" {
		id, err := strconv.ParseUint(tid, 10, 32)
		if err != nil {
			httperr.Respond(w, httperr.BadRequest("Неверный team_id"))
			return
		}
		v := uint(id)
		teamID = &v
	}

	result, err := h.svc.Suggest(r.Context(), userID, teamID, prefix)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, result)
}
