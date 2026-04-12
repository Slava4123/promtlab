package apikey

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	apikeyuc "promptvault/internal/usecases/apikey"
)

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apikeyuc.ErrNameEmpty):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, apikeyuc.ErrNameTooLong):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, apikeyuc.ErrMaxKeysReached):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, apikeyuc.ErrKeyNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
