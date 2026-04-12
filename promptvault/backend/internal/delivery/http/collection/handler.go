package collection

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	badgehttp "promptvault/internal/delivery/http/badge"
	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	colluc "promptvault/internal/usecases/collection"
)

type Handler struct {
	svc      *colluc.Service
	validate *validator.Validate
}

func NewHandler(svc *colluc.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// GET /api/collections?team_id=123
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	var teamIDs []uint
	if tid := r.URL.Query().Get("team_id"); tid != "" {
		id, err := strconv.ParseUint(tid, 10, 32)
		if err != nil {
			httperr.Respond(w, httperr.BadRequest("Неверный team_id"))
			return
		}
		teamIDs = []uint{uint(id)}
	}

	collections, err := h.svc.List(r.Context(), userID, teamIDs)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, collections)
}

// GET /api/collections/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	c, err := h.svc.GetByID(r.Context(), id, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	count, err := h.svc.CountPrompts(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, map[string]any{
		"id":           c.ID,
		"name":         c.Name,
		"description":  c.Description,
		"color":        c.Color,
		"icon":         c.Icon,
		"prompt_count": count,
		"created_at":   c.CreatedAt,
		"updated_at":   c.UpdatedAt,
	})
}

// POST /api/collections
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[CreateRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	c, newBadges, err := h.svc.Create(r.Context(), userID, utils.SanitizeString(req.Name), utils.SanitizeString(req.Description), req.Color, req.Icon, req.TeamID)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteCreated(w, NewCollectionResponse(*c, badgehttp.NewBadgeSummaries(newBadges)))
}

// PUT /api/collections/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	req, err := utils.DecodeAndValidate[UpdateRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	c, err := h.svc.Update(r.Context(), id, userID, utils.SanitizeString(req.Name), utils.SanitizeString(req.Description), req.Color, req.Icon)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, c)
}

// DELETE /api/collections/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteNoContent(w)
}

func parseID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	return uint(id), err
}
