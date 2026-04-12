package streak

import (
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	streakuc "promptvault/internal/usecases/streak"
)

type Handler struct {
	svc *streakuc.Service
}

func NewHandler(svc *streakuc.Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/streaks
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	tz := r.Header.Get("X-Timezone")

	result, err := h.svc.GetStreak(r.Context(), userID, tz)
	if err != nil {
		httperr.Respond(w, httperr.Internal(err))
		return
	}

	utils.WriteOK(w, result)
}
