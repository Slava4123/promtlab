package analytics

import (
	"fmt"
	"mime"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	analyticsuc "promptvault/internal/usecases/analytics"
)

// Handler — HTTP делегат analytics.Service.
// Endpoints (Phase 14 B.4):
//   - GET /api/analytics/personal?range=7d|30d|90d|365d
//   - GET /api/analytics/teams/{id}?range=...
//   - GET /api/analytics/prompts/{id}
//   - GET /api/analytics/insights (Max only — 402 для остальных)
//   - GET /api/analytics/export?format=csv&scope=personal|team&team_id=&range=
//
// H5: plan-check (Max/Pro gate) вынесен в service (GetInsightsGated, ExportGate).
// Handler только маппит доменные ошибки в HTTP.
type Handler struct {
	svc *analyticsuc.Service
}

func NewHandler(svc *analyticsuc.Service) *Handler {
	return &Handler{svc: svc}
}

// Personal — личный dashboard. Service сам clamp'ает range по тарифу.
func (h *Handler) Personal(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	rng := parseRange(r.URL.Query().Get("range"))

	dash, err := h.svc.GetPersonalDashboard(r.Context(), userID, rng)
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteOK(w, dash)
}

// Team — team dashboard. Membership проверяется внутри svc.GetTeamDashboard.
func (h *Handler) Team(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	teamIDStr := chi.URLParam(r, "id")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный team_id"))
		return
	}
	rng := parseRange(r.URL.Query().Get("range"))

	dash, err := h.svc.GetTeamDashboard(r.Context(), userID, uint(teamID), rng)
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteOK(w, dash)
}

// Prompt — per-prompt analytics. Access check внутри svc.
func (h *Handler) Prompt(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	promptIDStr := chi.URLParam(r, "id")
	promptID, err := strconv.ParseUint(promptIDStr, 10, 32)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный prompt ID"))
		return
	}

	data, err := h.svc.GetPromptAnalytics(r.Context(), uint(promptID), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteOK(w, data)
}

// Insights — Max only. Free/Pro → 402 с upgrade_url.
func (h *Handler) Insights(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	insights, err := h.svc.GetInsightsGated(r.Context(), userID, nil)
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteOK(w, map[string]any{"items": toInsightResponses(insights)})
}

// Export — CSV streaming для usage-данных. Free → 402 (export = Pro+).
// Current MVP: один sheet с colon-сепарированными строками по дням.
// Scope: personal или team (по query param).
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	q := r.URL.Query()

	format := q.Get("format")
	if format == "" {
		format = "csv"
	}
	if format != "csv" {
		httperr.Respond(w, httperr.BadRequest("Поддерживается только format=csv"))
		return
	}

	if err := h.svc.ExportGate(r.Context(), userID); err != nil {
		respondError(w, r, err)
		return
	}

	scope := q.Get("scope")
	if scope == "" {
		scope = "personal"
	}
	rng := parseRange(q.Get("range"))

	var filename string
	var points []repo.UsagePoint

	switch scope {
	case "personal":
		dash, err := h.svc.GetPersonalDashboard(r.Context(), userID, rng)
		if err != nil {
			respondError(w, r, err)
			return
		}
		points = dash.UsagePerDay
		filename = fmt.Sprintf("analytics-personal-%s.csv", dash.Range)
	case "team":
		teamIDStr := q.Get("team_id")
		teamID64, perr := strconv.ParseUint(teamIDStr, 10, 32)
		if perr != nil {
			httperr.Respond(w, httperr.BadRequest("Для scope=team нужен team_id"))
			return
		}
		dash, err := h.svc.GetTeamDashboard(r.Context(), userID, uint(teamID64), rng)
		if err != nil {
			respondError(w, r, err)
			return
		}
		points = dash.UsagePerDay
		filename = fmt.Sprintf("analytics-team-%d-%s.csv", teamID64, dash.Range)
	default:
		httperr.Respond(w, httperr.BadRequest("scope должен быть personal или team"))
		return
	}

	// Stream CSV: простой header "date,uses".
	// M7: Content-Disposition через mime.FormatMediaType — корректное
	// квотирование/экранирование filename, защита от Response Splitting.
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	w.WriteHeader(http.StatusOK)

	if _, err := fmt.Fprintln(w, "date,uses"); err != nil {
		return
	}
	for _, p := range points {
		if _, err := fmt.Fprintf(w, "%s,%d\n", p.Day.Format("2006-01-02"), p.Count); err != nil {
			return
		}
	}
}
