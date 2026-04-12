package adminauth

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	adminauthuc "promptvault/internal/usecases/adminauth"
	repo "promptvault/internal/interface/repository"
)

// respondError маппит доменные ошибки usecases/adminauth в HTTP статусы.
//
// ВАЖНО: business-validation ошибки (ErrInvalidCode, ErrTOTPNotEnrolled)
// возвращаются как 422 Unprocessable Entity, НЕ 401. Причина: фронтенд
// client.ts автоматически ретритит любой 401 на protected endpoint через
// ensureFreshToken — это нужно только для истёкшего JWT. Бизнес-валидация
// (неверный TOTP код) не является auth failure и не должна триггерить
// token refresh. См. BUG #1 из QA отчёта.
func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, adminauthuc.ErrNotAdmin):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, adminauthuc.ErrInvalidCode),
		errors.Is(err, adminauthuc.ErrTOTPNotEnrolled):
		httperr.Respond(w, &httperr.AppError{
			Code:    http.StatusUnprocessableEntity,
			Message: err.Error(),
		})
	case errors.Is(err, adminauthuc.ErrTOTPAlreadyConfirmed):
		httperr.Respond(w, httperr.Conflict(err.Error()))
	case errors.Is(err, adminauthuc.ErrGenerateFailed):
		httperr.Respond(w, httperr.Internal(err))
	case errors.Is(err, repo.ErrNotFound):
		httperr.Respond(w, httperr.NotFound("не найдено"))
	default:
		httperr.Respond(w, httperr.Internal(err))
	}
}
