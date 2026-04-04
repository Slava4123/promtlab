package collection

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	colluc "promptvault/internal/usecases/collection"
)

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, colluc.ErrNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, colluc.ErrForbidden):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, colluc.ErrViewerReadOnly):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
