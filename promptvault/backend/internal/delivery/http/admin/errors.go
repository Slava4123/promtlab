package admin

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	adminauthuc "promptvault/internal/usecases/adminauth"
	adminuc "promptvault/internal/usecases/admin"
)

// respondError маппит доменные ошибки admin.Service/adminauth.Service → HTTP.
func respondError(w http.ResponseWriter, err error) {
	// Если это уже httperr.AppError (например, из verifyTOTP 400 BadRequest) —
	// пропускаем как есть, без wrapping в Internal().
	var appErr *httperr.AppError
	if errors.As(err, &appErr) {
		httperr.Respond(w, appErr)
		return
	}

	switch {
	case errors.Is(err, adminuc.ErrUserNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, adminuc.ErrBadgeNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, adminuc.ErrBadgeAlreadyUnlocked):
		httperr.Respond(w, httperr.Conflict(err.Error()))
	case errors.Is(err, adminuc.ErrCannotFreezeSelf),
		errors.Is(err, adminuc.ErrCannotRevokeSelfRole):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, adminuc.ErrInvalidStatus):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, adminuc.ErrTierNotImplemented):
		httperr.Respond(w, &httperr.AppError{
			Code:    http.StatusNotImplemented,
			Message: err.Error(),
		})
	case errors.Is(err, adminuc.ErrEmailNotConfigured):
		httperr.Respond(w, &httperr.AppError{
			Code:    http.StatusServiceUnavailable,
			Message: err.Error(),
		})
	case errors.Is(err, adminauthuc.ErrInvalidCode),
		errors.Is(err, adminauthuc.ErrTOTPNotEnrolled):
		// Sudo mode validation failure — возвращаем 422 Unprocessable Entity,
		// НЕ 401. Причина: фронтенд client.ts автоматически ретритит 401 через
		// ensureFreshToken (для истёкшего JWT) — для бизнес-валидации TOTP это
		// приводит к двойному запросу. См. BUG #1 из QA отчёта.
		httperr.Respond(w, &httperr.AppError{
			Code:    http.StatusUnprocessableEntity,
			Message: err.Error(),
		})
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
