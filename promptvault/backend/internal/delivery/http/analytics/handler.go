package analytics

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	analyticsuc "promptvault/internal/usecases/analytics"
)

// addBreadcrumb кладёт запись в Sentry-hub для диагностики ошибок Max-юзеров.
// Намеренно не пишем PII (email, prompt_id допустим — уже не PII).
func addBreadcrumb(r *http.Request, category, message string, data map[string]any) {
	hub := sentry.GetHubFromContext(r.Context())
	if hub == nil {
		return
	}
	hub.Scope().AddBreadcrumb(&sentry.Breadcrumb{
		Category: category,
		Message:  message,
		Level:    sentry.LevelInfo,
		Data:     data,
	}, 10)
}

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
// Поддерживает drill-down через ?tag_id=:id и ?collection_id=:id (задача #9).
func (h *Handler) Personal(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	rng := parseRange(r.URL.Query().Get("range"))
	tagID := parseOptionalUint(r.URL.Query().Get("tag_id"))
	collectionID := parseOptionalUint(r.URL.Query().Get("collection_id"))

	dash, err := h.svc.GetPersonalDashboardFiltered(r.Context(), userID, rng, tagID, collectionID)
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteOK(w, dash)
}

// parseOptionalUint — возвращает nil если параметр пуст или неконвертируется.
// Используется для опциональных query-фильтров (tag_id, collection_id).
func parseOptionalUint(raw string) *uint {
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return nil
	}
	u := uint(v)
	return &u
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

// RefreshInsights — POST-endpoint для форсированного пересчёта инсайтов
// (обычно считаются ежесуточно cron'ом). Max-only + rate-limit 1/час
// (middleware в app.go). Возвращает свежий список после пересчёта.
func (h *Handler) RefreshInsights(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	addBreadcrumb(r, "analytics", "insights.refresh.trigger", map[string]any{
		"user_id": userID,
	})

	insights, err := h.svc.RefreshInsightsGated(r.Context(), userID, nil)
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteOK(w, map[string]any{"items": toInsightResponses(insights)})
}

// Export — CSV или XLSX экспорт. Free → 402 (export = Pro+).
// Scope: personal или team (по query param).
// Format: csv (один sheet — только usage_per_day) или xlsx (4 sheets).
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	q := r.URL.Query()

	format := q.Get("format")
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "xlsx" {
		httperr.Respond(w, httperr.BadRequest("Поддерживается format=csv или format=xlsx"))
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

	addBreadcrumb(r, "analytics", "export.trigger", map[string]any{
		"user_id": userID,
		"format":  format,
		"scope":   scope,
		"range":   string(rng),
	})

	switch scope {
	case "personal":
		dash, err := h.svc.GetPersonalDashboard(r.Context(), userID, rng)
		if err != nil {
			respondError(w, r, err)
			return
		}
		base := fmt.Sprintf("analytics-personal-%s", dash.Range)
		if format == "xlsx" {
			writePersonalXLSX(w, base, dash)
		} else {
			writeUsageCSV(w, base, dash.UsagePerDay)
		}
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
		base := fmt.Sprintf("analytics-team-%d-%s", teamID64, dash.Range)
		if format == "xlsx" {
			writeTeamXLSX(w, base, dash)
		} else {
			writeUsageCSV(w, base, dash.UsagePerDay)
		}
	default:
		httperr.Respond(w, httperr.BadRequest("scope должен быть personal или team"))
	}
}
