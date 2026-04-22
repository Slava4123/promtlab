package prompt

import (
	"context"
	"errors"
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
	activityuc "promptvault/internal/usecases/activity"
	badgeuc "promptvault/internal/usecases/badge"
	promptuc "promptvault/internal/usecases/prompt"
	quotauc "promptvault/internal/usecases/quota"
)

type Handler struct {
	svc      *promptuc.Service
	quotas   *quotauc.Service
	validate *validator.Validate
	// activity — опциональный source для GetHistory (Phase 14).
	// Nil → GetHistory возвращает только versions, без team activity событий.
	activity *activityuc.Service
	// users — lookup для ChangedByEmail/ChangedByName в VersionResponse.
	users repo.UserRepository
}

func NewHandler(svc *promptuc.Service, quotas *quotauc.Service) *Handler {
	return &Handler{svc: svc, quotas: quotas, validate: validator.New()}
}

// SetHistoryDeps подключает activity service и users repo для GET /api/prompts/:id/history.
// Опционально: если не вызван — endpoint вернёт только versions без actor info.
func (h *Handler) SetHistoryDeps(activity *activityuc.Service, users repo.UserRepository) {
	h.activity = activity
	h.users = users
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
		IsPublic:      req.IsPublic,
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
		IsPublic:      req.IsPublic,
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

	// Extension quota check (по заголовку X-Client-Source)
	isExtension := r.Header.Get("X-Client-Source") == "extension"
	if isExtension && h.quotas != nil {
		if qErr := h.quotas.CheckExtensionQuota(ctx, userID); qErr != nil {
			var qe *quotauc.QuotaExceededError
			if errors.As(qErr, &qe) {
				httperr.RespondQuotaError(w, qe.QuotaType, qe.Used, qe.Limit, qe.PlanID, qe.Message)
				return
			}
			httperr.Respond(w, httperr.Internal(qErr))
			return
		}
	}

	newBadges, err := h.svc.IncrementUsage(ctx, id, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	// Extension usage increment
	if isExtension && h.quotas != nil {
		_ = h.quotas.IncrementExtensionUsage(ctx, userID)
	}

	utils.WriteOK(w, IncrementUsageResponse{
		Message:             "ok",
		NewlyUnlockedBadges: badgehttp.NewBadgeSummaries(newBadges),
	})
}

// GetPublic — GET /api/public/prompts/{slug} (no auth).
// Только is_public=true промпты. Страница /p/:slug на фронте использует этот
// endpoint. Без pin-статусов — это public view, ничего лично-pinned не возвращаем.
func (h *Handler) GetPublic(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(chi.URLParam(r, "slug"))
	if slug == "" {
		httperr.Respond(w, httperr.BadRequest("slug is required"))
		return
	}
	p, err := h.svc.GetPublicBySlug(r.Context(), slug)
	if err != nil {
		respondError(w, err)
		return
	}
	utils.WriteOK(w, NewPromptResponse(*p))
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

	responses := NewVersionListResponse(versions)
	_ = h.enrichVersionsWithActors(r.Context(), versions, responses)
	utils.WritePaginated(w, responses, total, page, pageSize)
}

// enrichVersionsWithActors — добавляет ChangedByEmail/Name в response items.
// N ≤ 100 (pageSize cap), per-row lookup приемлем для MVP.
// Nil-safe: если h.users не подключён — no-op.
//
// M6: возвращает true, если хотя бы один lookup fail'нулся (используется
// для флага actors_partial в ответе, чтобы UI мог показать «Неизвестный
// автор — данные неполные»).
func (h *Handler) enrichVersionsWithActors(ctx context.Context, versions []models.PromptVersion, responses []VersionResponse) bool {
	if h.users == nil {
		return false
	}
	actorIDs := make(map[uint]struct{})
	for _, v := range versions {
		if v.ChangedBy != nil {
			actorIDs[*v.ChangedBy] = struct{}{}
		}
	}
	actorMap := make(map[uint]*models.User, len(actorIDs))
	var partial bool
	var failCount int
	for uid := range actorIDs {
		u, err := h.users.GetByID(ctx, uid)
		if err != nil {
			if !partial {
				slog.WarnContext(ctx, "prompt.history.actor_lookup_failed",
					"err", err, "actor_id", uid)
			}
			partial = true
			failCount++
			continue
		}
		actorMap[uid] = u
	}
	if partial {
		slog.WarnContext(ctx, "prompt.history.actors_partial",
			"fail_count", failCount, "total", len(actorIDs))
	}
	for i := range responses {
		if responses[i].ChangedByID != nil {
			if u, ok := actorMap[*responses[i].ChangedByID]; ok {
				responses[i].ChangedByEmail = u.Email
				responses[i].ChangedByName = u.Name
			}
		}
	}
	return partial
}

// GetHistory — GET /api/prompts/{id}/history (Phase 14).
// Склейка prompt_versions + team_activity_log (для team-промптов) в одном ответе.
// Версии — всегда; activity — только если activity service подключён
// И промпт принадлежит команде (для личных промптов team_activity_log пустой).
func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	// Access check + подгрузка промпта одним вызовом.
	prompt, err := h.svc.GetByID(r.Context(), id, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	// Fetch version-снапшоты. 100 последних — типичный предел истории одного промпта.
	versions, _, err := h.svc.ListVersions(r.Context(), id, userID, 1, 100)
	if err != nil {
		respondError(w, err)
		return
	}
	versionResps := NewVersionListResponse(versions)
	actorsPartial := h.enrichVersionsWithActors(r.Context(), versions, versionResps)

	// Activity events (только для team-промптов и если activity подключён).
	// H1: repo/timeout fail больше не молчаливый — логируем Warn и возвращаем
	// флаг activity_partial, чтобы UI мог отличить «событий нет» от «данные
	// временно недоступны».
	activityItems := []any{}
	activityPartial := false
	if h.activity != nil && prompt.TeamID != nil {
		events, err := h.activity.GetPromptHistory(r.Context(), id, 100)
		if err != nil {
			slog.WarnContext(r.Context(), "prompt.history.activity_failed",
				"err", err, "prompt_id", id, "team_id", *prompt.TeamID)
			activityPartial = true
		} else {
			for _, e := range events {
				activityItems = append(activityItems, map[string]any{
					"id":           e.ID,
					"actor_id":     e.ActorID,
					"actor_email":  e.ActorEmail,
					"actor_name":   e.ActorName,
					"event_type":   e.EventType,
					"target_type":  e.TargetType,
					"target_id":    e.TargetID,
					"target_label": e.TargetLabel,
					"metadata":     e.Metadata,
					"created_at":   e.CreatedAt,
				})
			}
		}
	}

	utils.WriteOK(w, map[string]any{
		"versions":         versionResps,
		"activity":         activityItems,
		"actors_partial":   actorsPartial,
		"activity_partial": activityPartial,
	})
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
