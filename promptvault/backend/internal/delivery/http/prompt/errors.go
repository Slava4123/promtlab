package prompt

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	promptuc "promptvault/internal/usecases/prompt"
)

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, promptuc.ErrNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, promptuc.ErrVersionNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, promptuc.ErrForbidden):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, promptuc.ErrViewerReadOnly):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
