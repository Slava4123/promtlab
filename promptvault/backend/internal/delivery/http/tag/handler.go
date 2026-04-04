package tag

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	taguc "promptvault/internal/usecases/tag"
)

type Handler struct {
	svc      *taguc.Service
	validate *validator.Validate
}

func NewHandler(svc *taguc.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// GET /api/tags?team_id=123
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	teamID := parseTeamID(r)

	tags, err := h.svc.List(r.Context(), userID, teamID)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, tags)
}

// POST /api/tags
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[CreateRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	tag, err := h.svc.Create(r.Context(), utils.SanitizeString(req.Name), req.Color, userID, req.TeamID)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteCreated(w, tag)
}

func parseTeamID(r *http.Request) *uint {
	s := r.URL.Query().Get("team_id")
	if s == "" {
		return nil
	}
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return nil
	}
	v := uint(id)
	return &v
}

// DELETE /api/tags/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	if err := h.svc.Delete(r.Context(), uint(id), userID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteNoContent(w)
}
