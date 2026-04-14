package errors

import (
	stderrors "errors"
	"log/slog"
	"net/http"

	"github.com/getsentry/sentry-go"

	"promptvault/internal/delivery/http/utils"
)

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"error"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func New(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func BadRequest(msg string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: msg}
}

func Forbidden(msg string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: msg}
}

func NotFound(msg string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg}
}

func Conflict(msg string) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg}
}

func TooManyRequests(msg string) *AppError {
	return &AppError{Code: http.StatusTooManyRequests, Message: msg}
}

func Internal(err error) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: "internal error", Err: err}
}

// RespondQuotaError пишет enriched 402 ответ для QuotaExceededError.
// Используется всеми delivery-пакетами при маппинге quota errors.
func RespondQuotaError(w http.ResponseWriter, quotaType string, used, limit int, planID, message string) {
	utils.WriteJSON(w, http.StatusPaymentRequired, map[string]any{
		"error":       message,
		"quota_type":  quotaType,
		"used":        used,
		"limit":       limit,
		"plan":        planID,
		"upgrade_url": "/pricing",
	})
}

func Respond(w http.ResponseWriter, err error) {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		if appErr.Err != nil {
			slog.Error("request error", "error", appErr.Err, "message", appErr.Message)
		}
		utils.WriteJSON(w, appErr.Code, map[string]string{"error": appErr.Message})
		return
	}

	slog.Error("unhandled error", "error", err)
	utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}

// RespondWithRequest — вариант Respond с доступом к *http.Request для
// отправки ошибок в Sentry через per-request Hub. Используется там, где
// важна атрибуция к user.id (protected endpoints с Sentry UserContext middleware).
//
// Захватываются только:
//   - AppError с wrapped Err (5xx по своей природе — internal failures);
//   - unhandled errors (не AppError — definite bug).
//
// 4xx AppError без Err не захватываются — это обычно user errors (validation,
// not found), не имеет смысла алертить.
func RespondWithRequest(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		if appErr.Err != nil {
			slog.Error("request error", "error", appErr.Err, "message", appErr.Message)
			if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
				hub.CaptureException(appErr.Err)
			}
		}
		utils.WriteJSON(w, appErr.Code, map[string]string{"error": appErr.Message})
		return
	}

	slog.Error("unhandled error", "error", err)
	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		hub.CaptureException(err)
	}
	utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}
