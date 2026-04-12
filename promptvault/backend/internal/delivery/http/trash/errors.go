package trash

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	trashuc "promptvault/internal/usecases/trash"
)

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, trashuc.ErrNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, trashuc.ErrForbidden):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, trashuc.ErrViewerReadOnly):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, trashuc.ErrInvalidType):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
