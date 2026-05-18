package collection

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	badgehttp "promptvault/internal/delivery/http/badge"
	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	analyticsuc "promptvault/internal/usecases/analytics"
	colluc "promptvault/internal/usecases/collection"
)

type Handler struct {
	svc      *colluc.Service
	validate *validator.Validate
	// insights — опциональный hot-refresh кэша Smart Insights после Delete.
	// nil-safe: если не подключён через SetInsightsRecomputer, recompute
	// пропускается и состояние догонит nightly cron loop.
	insights analyticsuc.InsightsRecomputer
}

func NewHandler(svc *colluc.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// SetInsightsRecomputer подключает hot-refresh кэша Smart Insights.
// После DELETE /api/collections/{id} пересчитываются insights типа
// empty_collections в personal scope (teamID=nil). Вызывается из app.go.
func (h *Handler) SetInsightsRecomputer(r analyticsuc.InsightsRecomputer) {
	h.insights = r
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
		"color":        string(c.Color),
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

	// Partial update: nil = «не трогать», sanitized non-nil = новое значение.
	// До правки `Description string` (без указателя) превращал отсутствие поля
	// в пустую строку → silent data loss при PUT с только {name}.
	var name, description, color, icon *string
	if req.Name != nil {
		s := utils.SanitizeString(*req.Name)
		name = &s
	}
	if req.Description != nil {
		s := utils.SanitizeString(*req.Description)
		description = &s
	}
	if req.Color != nil {
		color = req.Color
	}
	if req.Icon != nil {
		icon = req.Icon
	}
	c, err := h.svc.Update(r.Context(), id, userID, name, description, color, icon)
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

	// Hot-refresh Smart Insights кэша: empty_collections мог измениться
	// после удаления коллекции. teamID=nil — personal scope (team-scoped
	// пересчитается на nightly cron). Ошибки swallow — recompute fail
	// не должен ломать DELETE.
	if h.insights != nil {
		if rerr := h.insights.Recompute(r.Context(), userID, nil, []string{models.InsightEmptyCollections}); rerr != nil {
			slog.WarnContext(r.Context(), "collection.delete.insights_recompute_failed",
				"err", rerr, "user_id", userID, "collection_id", id)
		}
	}

	utils.WriteNoContent(w)
}

func parseID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	return uint(id), err
}
