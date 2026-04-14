package team

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	quotauc "promptvault/internal/usecases/quota"
	teamuc "promptvault/internal/usecases/team"
)

func respondError(w http.ResponseWriter, err error) {
	var qe *quotauc.QuotaExceededError
	if errors.As(err, &qe) {
		httperr.RespondQuotaError(w, qe.QuotaType, qe.Used, qe.Limit, qe.PlanID, qe.Message)
		return
	}

	switch {
	case errors.Is(err, teamuc.ErrNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, teamuc.ErrForbidden):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, teamuc.ErrNotOwner):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, teamuc.ErrUserNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, teamuc.ErrAlreadyMember):
		httperr.Respond(w, httperr.Conflict(err.Error()))
	case errors.Is(err, teamuc.ErrAlreadyInvited):
		httperr.Respond(w, httperr.Conflict(err.Error()))
	case errors.Is(err, teamuc.ErrInvitationNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, teamuc.ErrCannotInviteSelf):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, teamuc.ErrCannotRemoveOwner):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, teamuc.ErrCannotChangeOwnerRole):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, teamuc.ErrInvalidRole):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
