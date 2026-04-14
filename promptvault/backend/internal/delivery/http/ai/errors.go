package ai

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	aiuc "promptvault/internal/usecases/ai"
	quotauc "promptvault/internal/usecases/quota"
)

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, aiuc.ErrRateLimited):
		httperr.Respond(w, httperr.TooManyRequests(err.Error()))
	case errors.Is(err, aiuc.ErrModelNotFound) || errors.Is(err, aiuc.ErrEmptyContent):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, aiuc.ErrAPIKeyMissing):
		httperr.Respond(w, httperr.New(http.StatusServiceUnavailable, err.Error(), err))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}

func respondQuotaError(w http.ResponseWriter, err error) {
	var qe *quotauc.QuotaExceededError
	if errors.As(err, &qe) {
		httperr.RespondQuotaError(w, qe.QuotaType, qe.Used, qe.Limit, qe.PlanID, qe.Message)
		return
	}
	httperr.Respond(w, httperr.Internal(err))
}
