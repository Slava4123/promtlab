package changelog

import (
	"net/http"

	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	changeloguc "promptvault/internal/usecases/changelog"
)

type Handler struct {
	svc *changeloguc.Service
}

func NewHandler(svc *changeloguc.Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/changelog
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	out, err := h.svc.List(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, NewChangelogResponse(out))
}

// POST /api/changelog/seen
func (h *Handler) MarkSeen(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	if err := h.svc.MarkSeen(r.Context(), userID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteNoContent(w)
}
