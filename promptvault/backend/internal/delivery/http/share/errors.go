package share

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	quotauc "promptvault/internal/usecases/quota"
	shareuc "promptvault/internal/usecases/share"
)

func respondError(w http.ResponseWriter, r *http.Request, err error) {
	var qe *quotauc.QuotaExceededError
	if errors.As(err, &qe) {
		httperr.RespondQuotaError(w, qe.QuotaType, qe.Used, qe.Limit, qe.PlanID, qe.Message)
		return
	}

	switch {
	case errors.Is(err, shareuc.ErrNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, shareuc.ErrPromptNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, shareuc.ErrForbidden):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, shareuc.ErrViewerReadOnly):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	default:
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
	}
}
