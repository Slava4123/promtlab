package prompt

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	badgehttp "promptvault/internal/delivery/http/badge"
	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	badgeuc "promptvault/internal/usecases/badge"
	promptuc "promptvault/internal/usecases/prompt"
)

type Handler struct {
	svc      *promptuc.Service
	validate *validator.Validate
}

func NewHandler(svc *promptuc.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// GET /api/prompts
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := repo.PromptListFilter{
		UserID:       userID,
		Query:        q.Get("q"),
		FavoriteOnly: q.Get("favorite") == "true",
		Page:         page,
		PageSize:     pageSize,
	}

	// Team context
	if tid := q.Get("team_id"); tid != "" {
		id, err := strconv.ParseUint(tid, 10, 32)
		if err != nil {
			httperr.Respond(w, httperr.BadRequest("Неверный team_id"))
			return
		}
		filter.TeamIDs = []uint{uint(id)}
	}

	if cid := q.Get("collection_id"); cid != "" {
		id, err := strconv.ParseUint(cid, 10, 32)
		if err != nil {
			httperr.Respond(w, httperr.BadRequest("Неверный collection_id"))
			return
		}
		uid := uint(id)
		filter.CollectionID = &uid
	}

	if tagParam := q.Get("tag_ids"); tagParam != "" {
		for _, p := range strings.Split(tagParam, ",") {
			id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 32)
			if err != nil {
				httperr.Respond(w, httperr.BadRequest("Неверный tag_ids"))
				return
			}
			filter.TagIDs = append(filter.TagIDs, uint(id))
		}
	}

	prompts, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		respondError(w, err)
		return
	}

	ids := make([]uint, len(prompts))
	for i, p := range prompts {
		ids[i] = p.ID
	}
	pinStatuses, err := h.svc.GetPinStatuses(r.Context(), ids, userID)
	if err != nil {
		slog.Error("failed to fetch pin statuses", "error", err, "user_id", userID)
		pinStatuses = make(map[uint]repo.PinStatus)
	}

	utils.WritePaginated(w, NewPromptListResponse(prompts, pinStatuses), total, page, pageSize)
}

// POST /api/prompts
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	ctx := promptuc.ContextWithTimezone(r.Context(), r.Header.Get("X-Timezone"))

	req, err := utils.DecodeAndValidate[CreatePromptRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	p, newBadges, err := h.svc.Create(ctx, promptuc.CreateInput{
		UserID:        userID,
		TeamID:        req.TeamID,
		Title:         strings.TrimSpace(req.Title),
		Content:       strings.TrimSpace(req.Content),
		Model:         req.Model,
		CollectionIDs: req.CollectionIDs,
		TagIDs:        req.TagIDs,
	})
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteCreated(w, h.promptWithPinStatus(r, *p, newBadges...))
}

// GET /api/prompts/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	p, err := h.svc.GetByID(r.Context(), id, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, h.promptWithPinStatus(r, *p))
}

// PUT /api/prompts/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	req, err := utils.DecodeAndValidate[UpdatePromptRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	p, newBadges, err := h.svc.Update(r.Context(), id, userID, promptuc.UpdateInput{
		Title:         trimSpacePtr(req.Title),
		Content:       trimSpacePtr(req.Content),
		Model:         req.Model,
		ChangeNote:    strings.TrimSpace(req.ChangeNote),
		CollectionIDs: req.CollectionIDs,
		TagIDs:        req.TagIDs,
	})
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, h.promptWithPinStatus(r, *p, newBadges...))
}

// DELETE /api/prompts/{id}
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

// POST /api/prompts/{id}/favorite
func (h *Handler) ToggleFavorite(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	p, err := h.svc.ToggleFavorite(r.Context(), id, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, h.promptWithPinStatus(r, *p))
}

// POST /api/prompts/{id}/use
func (h *Handler) IncrementUsage(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	ctx := promptuc.ContextWithTimezone(r.Context(), r.Header.Get("X-Timezone"))
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	newBadges, err := h.svc.IncrementUsage(ctx, id, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, IncrementUsageResponse{
		Message:             "ok",
		NewlyUnlockedBadges: badgehttp.NewBadgeSummaries(newBadges),
	})
}

// GET /api/prompts/{id}/versions
func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	versions, total, err := h.svc.ListVersions(r.Context(), id, userID, page, pageSize)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WritePaginated(w, NewVersionListResponse(versions), total, page, pageSize)
}

// POST /api/prompts/{id}/revert/{versionId}
func (h *Handler) RevertToVersion(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	versionID, err := parseVersionID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID версии"))
		return
	}

	p, newBadges, err := h.svc.RevertToVersion(r.Context(), id, userID, versionID)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, h.promptWithPinStatus(r, *p, newBadges...))
}

// POST /api/prompts/{id}/pin
func (h *Handler) TogglePin(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	var req PinRequest
	if r.ContentLength > 0 {
		if err := utils.DecodeJSON(r, &req); err != nil {
			httperr.Respond(w, httperr.BadRequest(err.Error()))
			return
		}
	}

	result, err := h.svc.TogglePin(r.Context(), promptuc.PinInput{
		PromptID: id,
		UserID:   userID,
		TeamWide: req.TeamWide,
	})
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, result)
}

// GET /api/prompts/pinned
func (h *Handler) ListPinned(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	q := r.URL.Query()

	var teamID *uint
	if tid := q.Get("team_id"); tid != "" {
		id, err := strconv.ParseUint(tid, 10, 32)
		if err != nil {
			httperr.Respond(w, httperr.BadRequest("Неверный team_id"))
			return
		}
		uid := uint(id)
		teamID = &uid
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	prompts, err := h.svc.ListPinned(r.Context(), userID, teamID, limit)
	if err != nil {
		respondError(w, err)
		return
	}

	ids := make([]uint, len(prompts))
	for i, p := range prompts {
		ids[i] = p.ID
	}
	ps, err := h.svc.GetPinStatuses(r.Context(), ids, userID)
	if err != nil {
		slog.Error("failed to fetch pin statuses", "error", err, "user_id", userID)
		ps = make(map[uint]repo.PinStatus)
	}

	utils.WriteOK(w, map[string]any{
		"items": NewPromptListResponse(prompts, ps),
		"total": len(prompts),
	})
}

// GET /api/prompts/recent
func (h *Handler) ListRecent(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	q := r.URL.Query()

	var teamID *uint
	if tid := q.Get("team_id"); tid != "" {
		id, err := strconv.ParseUint(tid, 10, 32)
		if err != nil {
			httperr.Respond(w, httperr.BadRequest("Неверный team_id"))
			return
		}
		uid := uint(id)
		teamID = &uid
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	prompts, err := h.svc.ListRecent(r.Context(), userID, teamID, limit)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, map[string]any{
		"items": NewPromptListResponse(prompts, nil),
		"total": len(prompts),
	})
}

// GET /api/prompts/history
func (h *Handler) ListHistory(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	q := r.URL.Query()

	var teamID *uint
	if tid := q.Get("team_id"); tid != "" {
		id, err := strconv.ParseUint(tid, 10, 32)
		if err != nil {
			httperr.Respond(w, httperr.BadRequest("Неверный team_id"))
			return
		}
		uid := uint(id)
		teamID = &uid
	}

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := h.svc.ListHistory(r.Context(), userID, teamID, page, pageSize)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WritePaginated(w, NewUsageLogListResponse(logs), total, page, pageSize)
}

// promptWithPinStatus возвращает PromptResponse с актуальным pin status.
// Variadic newBadges — опциональный список разблокированных в этом запросе
// бейджей; если не пусто, попадает в поле NewlyUnlockedBadges (omitempty).
func (h *Handler) promptWithPinStatus(r *http.Request, p models.Prompt, newBadges ...badgeuc.Badge) PromptResponse {
	userID := authmw.GetUserID(r.Context())
	ps, err := h.svc.GetPinStatuses(r.Context(), []uint{p.ID}, userID)
	if err != nil {
		slog.Error("failed to fetch pin statuses", "error", err, "user_id", userID)
		ps = make(map[uint]repo.PinStatus)
	}
	resp := NewPromptResponse(p, ps[p.ID])
	if len(newBadges) > 0 {
		resp.NewlyUnlockedBadges = badgehttp.NewBadgeSummaries(newBadges)
	}
	return resp
}

func trimSpacePtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	return &v
}

func parseID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	return uint(id), err
}

func parseVersionID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "versionId"), 10, 32)
	return uint(id), err
}
