package tag

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	taguc "promptvault/internal/usecases/tag"
)

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taguc.ErrNameEmpty):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, taguc.ErrNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, taguc.ErrForbidden):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, taguc.ErrViewerReadOnly):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
