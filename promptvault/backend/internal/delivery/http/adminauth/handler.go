package adminauth

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	adminauthuc "promptvault/internal/usecases/adminauth"
)

type Handler struct {
	svc      *adminauthuc.Service
	validate *validator.Validate
}

func NewHandler(svc *adminauthuc.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// POST /api/admin/totp/enroll
// Генерирует новый TOTP secret + 10 backup codes. Требует role='admin'.
// BackupCodes возвращаются ОДИН РАЗ — юзер обязан их сохранить.
func (h *Handler) Enroll(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	result, err := h.svc.Enroll(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}
	utils.WriteCreated(w, NewEnrollResponse(result))
}

// POST /api/admin/totp/verify-enrollment
// Подтверждает enrollment первым кодом из Authenticator. После этого
// юзер должен будет вводить TOTP при каждом login.
func (h *Handler) ConfirmEnrollment(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[ConfirmEnrollmentRequest](r, h.validate)
	if err != nil {
		// Raw validator error содержит технический текст
		// ("Key: 'ConfirmEnrollmentRequest.Code' Error:...") — маппим в
		// человеческое сообщение. 422 вместо 400 чтобы client.ts не делал
		// auto-refresh retry (консистентно с respondError для бизнес-401).
		httperr.Respond(w, &httperr.AppError{
			Code:    http.StatusUnprocessableEntity,
			Message: "введите 6-значный код из Authenticator",
		})
		return
	}

	if err := h.svc.ConfirmEnrollment(r.Context(), userID, req.Code); err != nil {
		respondError(w, err)
		return
	}
	utils.WriteOK(w, ConfirmEnrollmentResponse{Confirmed: true})
}

// POST /api/admin/totp/backup-codes/regenerate
// Инвалидирует все существующие backup codes и генерирует 10 новых.
// Старые коды больше не принимаются. Юзер должен сохранить новые.
func (h *Handler) RegenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	codes, err := h.svc.RegenerateBackupCodes(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}
	utils.WriteOK(w, RegenerateBackupCodesResponse{BackupCodes: codes})
}

// GET /api/admin/totp/status
// Возвращает флаги enrolled/confirmed для фронта. Используется в Admin UI
// чтобы показать «настроить TOTP» если юзер впервые заходит как admin.
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	confirmed, err := h.svc.IsConfirmed(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}
	utils.WriteOK(w, StatusResponse{Enrolled: true, Confirmed: confirmed})
}
