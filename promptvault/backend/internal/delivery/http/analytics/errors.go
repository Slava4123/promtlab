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
		respondTierRequired(w, "insights", "Max")
	case errors.Is(err, analyticsuc.ErrProRequired):
		respondTierRequired(w, "export", "Pro")
	default:
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
	}
}

// respondTierRequired — 402 Payment Required для фич выше текущего тарифа
// (Smart Insights = Max only, Export CSV = Pro+). План юзера проверяется
// в service — handler не знает current plan, передаём только required.
func respondTierRequired(w http.ResponseWriter, feature, requiredPlan string) {
	httperr.RespondQuotaError(w, feature, 0, 0, "",
		"Фича доступна на тарифе "+requiredPlan+". Обновите план на /pricing.")
}
