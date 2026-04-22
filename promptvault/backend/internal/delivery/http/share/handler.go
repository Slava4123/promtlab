package share

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	shareuc "promptvault/internal/usecases/share"
)

type Handler struct {
	svc *shareuc.Service
}

func NewHandler(svc *shareuc.Service) *Handler {
	return &Handler{svc: svc}
}

// GetPublic handles GET /api/s/{token} — public, no auth.
func (h *Handler) GetPublic(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		httperr.Respond(w, httperr.BadRequest("Токен обязателен"))
		return
	}

	info, err := h.svc.GetPublicPrompt(r.Context(), token, shareuc.ViewMeta{
		Referer:   r.Referer(),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	utils.WriteOK(w, toPublicPromptResponse(info))
}

// Create handles POST /api/prompts/{id}/share — protected.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	promptID, err := parsePromptID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID промпта"))
		return
	}

	info, created, err := h.svc.CreateOrGet(r.Context(), promptID, userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	resp := toShareLinkResponse(info)
	if created {
		utils.WriteCreated(w, resp)
	} else {
		utils.WriteOK(w, resp)
	}
}

// Get handles GET /api/prompts/{id}/share — protected.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	promptID, err := parsePromptID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID промпта"))
		return
	}

	info, err := h.svc.GetByPromptID(r.Context(), promptID, userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	utils.WriteOK(w, toShareLinkResponse(info))
}

// Delete handles DELETE /api/prompts/{id}/share — protected.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	promptID, err := parsePromptID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID промпта"))
		return
	}

	if err := h.svc.Deactivate(r.Context(), promptID, userID); err != nil {
		respondError(w, r, err)
		return
	}

	utils.WriteNoContent(w)
}

func parsePromptID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
