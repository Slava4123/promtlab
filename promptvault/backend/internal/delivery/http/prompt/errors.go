package prompt

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	promptuc "promptvault/internal/usecases/prompt"
	quotauc "promptvault/internal/usecases/quota"
)

func respondError(w http.ResponseWriter, err error) {
	var qe *quotauc.QuotaExceededError
	if errors.As(err, &qe) {
		httperr.RespondQuotaError(w, qe.QuotaType, qe.Used, qe.Limit, qe.PlanID, qe.Message)
		return
	}

	switch {
	case errors.Is(err, promptuc.ErrNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, promptuc.ErrVersionNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, promptuc.ErrForbidden):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, promptuc.ErrViewerReadOnly):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, promptuc.ErrPinForbidden):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, promptuc.ErrWorkspaceMismatch):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
