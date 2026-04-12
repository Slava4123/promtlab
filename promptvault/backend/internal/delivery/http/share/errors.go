package share

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	shareuc "promptvault/internal/usecases/share"
)

func respondError(w http.ResponseWriter, r *http.Request, err error) {
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
