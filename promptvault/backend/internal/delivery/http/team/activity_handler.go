package team

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	activityuc "promptvault/internal/usecases/activity"
	teamuc "promptvault/internal/usecases/team"
)

// ActivityHandler — endpoint GET /api/teams/{slug}/activity (Phase 14).
//
// Принимает query params: event_type, actor_id, target_type, target_id,
// from, to, page, page_size. Пагинация — offset (consistency с остальными
// HTTP endpoints; MCP использует cursor).
//
// GDPR (Q1): viewer-ам возвращаем имя без email; owner/editor видят полный actor_email.
type ActivityHandler struct {
	teams    *teamuc.Service
	activity *activityuc.Service
}

func NewActivityHandler(teams *teamuc.Service, activity *activityuc.Service) *ActivityHandler {
	return &ActivityHandler{teams: teams, activity: activity}
}

// ActivityItemResponse — DTO одного события в feed.
type ActivityItemResponse struct {
	ID          uint      `json:"id"`
	ActorID     *uint     `json:"actor_id,omitempty"`
	ActorEmail  string    `json:"actor_email,omitempty"`
	ActorName   string    `json:"actor_name,omitempty"`
	EventType   string    `json:"event_type"`
	TargetType  string    `json:"target_type"`
	TargetID    *uint     `json:"target_id,omitempty"`
	TargetLabel string    `json:"target_label,omitempty"`
	Metadata    any       `json:"metadata,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	activityMaxPage     = 1000 // H2: cap страниц, защита от int-overflow в Limit
	activityMaxPageSize = 200
	activityDefaultSize = 50
)

// List handles GET /api/teams/{slug}/activity
func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	// GetBySlug также проверяет membership (роль viewer минимум) и возвращает
	// список членов команды — из него достаём роль запрашивающего для GDPR mask.
	team, members, err := h.teams.GetBySlug(r.Context(), slug, userID)
	if err != nil {
		respondActivityError(w, r, err)
		return
	}

	q := r.URL.Query()
	filter := repo.TeamActivityFilter{
		TeamID:     team.ID,
		EventType:  q.Get("event_type"),
		TargetType: q.Get("target_type"),
	}

	// M4: невалидные query params → BadRequest, а не silent ignore.
	if s := q.Get("actor_id"); s != "" {
		v, perr := strconv.ParseUint(s, 10, 32)
		if perr != nil {
			httperr.Respond(w, httperr.BadRequest("неверный формат actor_id: ожидается целое число"))
			return
		}
		u := uint(v)
		filter.ActorID = &u
	}
	if s := q.Get("target_id"); s != "" {
		v, perr := strconv.ParseUint(s, 10, 32)
		if perr != nil {
			httperr.Respond(w, httperr.BadRequest("неверный формат target_id: ожидается целое число"))
			return
		}
		u := uint(v)
		filter.TargetID = &u
	}
	if s := q.Get("from"); s != "" {
		t, perr := time.Parse(time.RFC3339, s)
		if perr != nil {
			httperr.Respond(w, httperr.BadRequest("неверный формат from: ожидается RFC3339"))
			return
		}
		filter.FromTime = &t
	}
	if s := q.Get("to"); s != "" {
		t, perr := time.Parse(time.RFC3339, s)
		if perr != nil {
			httperr.Respond(w, httperr.BadRequest("неверный формат to: ожидается RFC3339"))
			return
		}
		filter.ToTime = &t
	}

	// Offset-based pagination для HTTP (consistency с остальными endpoints).
	// Преобразуем в cursor-based под капотом repo: skip через CursorBefore
	// неэффективен — просто читаем с лимитом и отбрасываем первые (page-1)*limit.
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if page < 1 {
		page = 1
	}
	if page > activityMaxPage { // H2: защита от int overflow при pageSize*page
		page = activityMaxPage
	}
	if pageSize < 1 || pageSize > activityMaxPageSize {
		pageSize = activityDefaultSize
	}
	// M1: sentinel-пагинация — запрашиваем на одну запись больше, чтобы
	// точно определить наличие следующей страницы, не полагаясь на точную
	// границу `== pageSize`.
	filter.Limit = pageSize*page + 1

	events, _, err := h.activity.ListByTeam(r.Context(), filter)
	if err != nil {
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
		return
	}
	// Отбрасываем первые (page-1)*pageSize (offset по факту).
	skip := (page - 1) * pageSize
	if skip >= len(events) {
		events = nil
	} else {
		events = events[skip:]
	}
	hasMore := len(events) > pageSize
	if hasMore {
		events = events[:pageSize]
	}

	items := toActivityItemResponses(events, requesterRole(members, userID))
	utils.WriteOK(w, map[string]any{
		"items":     items,
		"page":      page,
		"page_size": pageSize,
		"has_more":  hasMore,
	})
}

// requesterRole находит роль текущего userID среди членов команды.
// Возвращает пустую строку, если не найден (не должно случаться — checkAccess
// уже прошёл в GetBySlug, но подстраховываемся).
func requesterRole(members []models.TeamMember, userID uint) models.TeamRole {
	for _, m := range members {
		if m.UserID == userID {
			return m.Role
		}
	}
	return ""
}

// canSeeActorEmail возвращает true для ролей, которым показываем actor_email.
// GDPR (Q1): viewer-ам отдаём только actor_id + actor_name.
func canSeeActorEmail(role models.TeamRole) bool {
	return role == models.RoleOwner || role == models.RoleEditor
}

func toActivityItemResponses(events []models.TeamActivityLog, requester models.TeamRole) []ActivityItemResponse {
	showEmail := canSeeActorEmail(requester)
	items := make([]ActivityItemResponse, len(events))
	for i, e := range events {
		items[i] = ActivityItemResponse{
			ID:          e.ID,
			ActorID:     e.ActorID,
			ActorName:   e.ActorName,
			EventType:   e.EventType,
			TargetType:  e.TargetType,
			TargetID:    e.TargetID,
			TargetLabel: e.TargetLabel,
			CreatedAt:   e.CreatedAt,
		}
		if showEmail {
			items[i].ActorEmail = e.ActorEmail
		}
		if len(e.Metadata) > 0 {
			items[i].Metadata = e.Metadata
		}
	}
	return items
}

func respondActivityError(w http.ResponseWriter, r *http.Request, err error) {
	// GetBySlug возвращает teamuc.ErrNotFound / ErrForbidden / ErrNotOwner.
	switch {
	case errors.Is(err, teamuc.ErrNotFound):
		httperr.Respond(w, httperr.NotFound("Команда не найдена"))
	case errors.Is(err, teamuc.ErrForbidden), errors.Is(err, teamuc.ErrNotOwner):
		httperr.Respond(w, httperr.Forbidden("Нет доступа к команде"))
	default:
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
	}
}
