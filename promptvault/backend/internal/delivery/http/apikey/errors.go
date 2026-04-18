package apikey

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	apikeyuc "promptvault/internal/usecases/apikey"
)

// errTeamAccessDenied возвращается handler'ом когда userID не является членом team_id.
var errTeamAccessDenied = errors.New("Вы не являетесь участником команды")

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apikeyuc.ErrNameEmpty),
		errors.Is(err, apikeyuc.ErrNameTooLong),
		errors.Is(err, apikeyuc.ErrInvalidExpires),
		errors.Is(err, apikeyuc.ErrInvalidToolName):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, apikeyuc.ErrMaxKeysReached):
		httperr.Respond(w, httperr.Conflict(err.Error()))
	case errors.Is(err, apikeyuc.ErrKeyNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, errTeamAccessDenied):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
