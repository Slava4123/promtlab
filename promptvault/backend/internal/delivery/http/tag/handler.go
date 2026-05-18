package tag

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	analyticsuc "promptvault/internal/usecases/analytics"
	taguc "promptvault/internal/usecases/tag"
)

type Handler struct {
	svc      *taguc.Service
	validate *validator.Validate
	// insights — опциональный hot-refresh кэша Smart Insights после Delete.
	// nil-safe: если не подключён через SetInsightsRecomputer, recompute
	// пропускается и состояние догонит nightly cron loop.
	insights analyticsuc.InsightsRecomputer
}

func NewHandler(svc *taguc.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// SetInsightsRecomputer подключает hot-refresh кэша Smart Insights.
// После DELETE /api/tags/{id} пересчитываются insights типа orphan_tags
// в personal scope (teamID=nil). Вызывается из app.go.
func (h *Handler) SetInsightsRecomputer(r analyticsuc.InsightsRecomputer) {
	h.insights = r
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

	// Hot-refresh Smart Insights кэша: только что созданный тег ещё не
	// прикреплён к промптам → orphan_tags гарантированно увеличивается.
	// teamID=nil — personal scope (team-scoped пересчитается на nightly cron).
	// Ошибки swallow — recompute fail не должен ломать CREATE.
	if h.insights != nil {
		if rerr := h.insights.Recompute(r.Context(), userID, nil, []string{models.InsightOrphanTags}); rerr != nil {
			slog.WarnContext(r.Context(), "tag.create.insights_recompute_failed",
				"err", rerr, "user_id", userID, "tag_id", tag.ID)
		}
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

	// Hot-refresh Smart Insights кэша: orphan_tags мог измениться после
	// удаления тега. teamID=nil — personal scope (team-scoped пересчитается
	// на nightly cron). Ошибки swallow — recompute fail не ломает DELETE.
	if h.insights != nil {
		if rerr := h.insights.Recompute(r.Context(), userID, nil, []string{models.InsightOrphanTags}); rerr != nil {
			slog.WarnContext(r.Context(), "tag.delete.insights_recompute_failed",
				"err", rerr, "user_id", userID, "tag_id", id)
		}
	}

	utils.WriteNoContent(w)
}
