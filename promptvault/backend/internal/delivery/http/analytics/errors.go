package analytics

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	analyticsuc "promptvault/internal/usecases/analytics"
)

// respondError — маппинг analytics domain errors → HTTP.
// 403 для ErrForbidden (не член команды / нет доступа к промпту),
// 404 для ErrNotFound (промпт не существует),
// 402 для ErrMaxRequired/ErrProRequired (tier gate — H5),
// 500 — default.
func respondError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, analyticsuc.ErrForbidden):
		httperr.Respond(w, httperr.Forbidden("Нет доступа"))
	case errors.Is(err, analyticsuc.ErrNotFound):
		httperr.Respond(w, httperr.NotFound("Не найдено"))
	case errors.Is(err, analyticsuc.ErrMaxRequired):
		// Refresh insights endpoint — Max-only (rate-limited 1/час).
		respondTierRequired(w, "insights", "max")
	case errors.Is(err, analyticsuc.ErrProRequired):
		// Pricing iteration v3 (Task 8): ErrProRequired поднимается из двух мест:
		//   - GetInsightsGated (insights teaser — Free → 402, Pro/Max → данные),
		//   - ExportGate (CSV/XLSX export — Free → 402, Pro/Max → данные).
		// Generic "premium_feature" label корректно описывает оба контекста.
		// plan="pro" нужен фронту для upgrade prompt (CTA на /pricing) —
		// lowercase для match с frontend planConfig lookup (PlanBadge).
		respondTierRequired(w, "premium_feature", "pro")
	default:
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
	}
}

// respondTierRequired — 402 Payment Required для фич выше текущего тарифа.
// План юзера проверяется в service — handler не знает current plan,
// передаёт только required plan для UI upgrade prompt.
func respondTierRequired(w http.ResponseWriter, feature, requiredPlan string) {
	httperr.RespondQuotaError(w, feature, 0, 0, requiredPlan,
		"Фича доступна на тарифе "+requiredPlan+". Обновите план на /pricing.")
}
