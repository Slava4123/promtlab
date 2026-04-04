package errors

import (
	stderrors "errors"
	"log/slog"
	"net/http"

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
