package team

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	quotauc "promptvault/internal/usecases/quota"
	teamuc "promptvault/internal/usecases/team"
)

// UsageHandler — endpoint GET /api/teams/{slug}/usage. Возвращает usage
// всех team-pool ресурсов команды против её лимитов (Pack T, миграция 000070).
//
// Access: любой member команды (viewer+) — лимит/usage не секретная информация
// для участников. teams.GetBySlug делает membership-check и выкинет 403/404
// если юзер не в команде.
type UsageHandler struct {
	teams  *teamuc.Service
	quotas *quotauc.Service
}

func NewUsageHandler(teams *teamuc.Service, quotas *quotauc.Service) *UsageHandler {
	return &UsageHandler{teams: teams, quotas: quotas}
}

// GET /api/teams/{slug}/usage
func (h *UsageHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	team, _, err := h.teams.GetBySlug(r.Context(), slug, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	summary, err := h.quotas.GetTeamUsageSummary(r.Context(), team)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, summary)
}
