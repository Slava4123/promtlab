package feedback

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	feedbackuc "promptvault/internal/usecases/feedback"
)

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, feedbackuc.ErrInvalidType):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, feedbackuc.ErrMessageTooLong):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
