package auth

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	authuc "promptvault/internal/usecases/auth"
)

func respondError(w http.ResponseWriter, err error) {
	var emailTaken *authuc.EmailTakenError

	switch {
	case errors.As(err, &emailTaken):
		httperr.Respond(w, httperr.Conflict(err.Error()))
	case errors.Is(err, authuc.ErrInvalidCredentials):
		httperr.Respond(w, httperr.Unauthorized(err.Error()))
	case errors.Is(err, authuc.ErrInvalidToken), errors.Is(err, authuc.ErrExpiredToken):
		httperr.Respond(w, httperr.Unauthorized(err.Error()))
	case errors.Is(err, authuc.ErrUserNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, authuc.ErrInvalidCode), errors.Is(err, authuc.ErrExpiredCode), errors.Is(err, authuc.ErrTooManyAttempts):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, authuc.ErrEmailNotVerified):
		httperr.Respond(w, httperr.New(http.StatusForbidden, err.Error(), nil))
	case errors.Is(err, authuc.ErrCannotUnlinkLast):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, authuc.ErrWrongPassword):
		httperr.Respond(w, httperr.Unauthorized(err.Error()))
	case errors.Is(err, authuc.ErrNoPassword):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, authuc.ErrPasswordAlreadySet):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, authuc.ErrOAuthNotConfigured):
		httperr.Respond(w, httperr.New(http.StatusNotImplemented, err.Error(), nil))
	case errors.Is(err, authuc.ErrOAuthExchangeFailed), errors.Is(err, authuc.ErrOAuthProfileFailed):
		httperr.Respond(w, httperr.Unauthorized(err.Error()))
	case errors.Is(err, authuc.ErrOAuthStateMismatch):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, authuc.ErrProviderLinkedToOther), errors.Is(err, authuc.ErrProviderAlreadyLinked), errors.Is(err, authuc.ErrUsernameTaken):
		httperr.Respond(w, httperr.Conflict(err.Error()))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
