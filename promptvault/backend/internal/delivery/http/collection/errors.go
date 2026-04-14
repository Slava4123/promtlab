package collection

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	colluc "promptvault/internal/usecases/collection"
	quotauc "promptvault/internal/usecases/quota"
)

func respondError(w http.ResponseWriter, err error) {
	var qe *quotauc.QuotaExceededError
	if errors.As(err, &qe) {
		httperr.RespondQuotaError(w, qe.QuotaType, qe.Used, qe.Limit, qe.PlanID, qe.Message)
		return
	}

	switch {
	case errors.Is(err, colluc.ErrNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, colluc.ErrForbidden):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, colluc.ErrViewerReadOnly):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, colluc.ErrInvalidColor), errors.Is(err, colluc.ErrInvalidIcon):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
