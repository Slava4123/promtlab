package starter

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	starteruc "promptvault/internal/usecases/starter"
)

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, starteruc.ErrUnknownTemplate):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, starteruc.ErrAlreadyCompleted):
		httperr.Respond(w, httperr.Conflict(err.Error()))
	case errors.Is(err, starteruc.ErrUserNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
