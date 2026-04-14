package auth

import (
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	adminauthuc "promptvault/internal/usecases/adminauth"
	authuc "promptvault/internal/usecases/auth"
	changeloguc "promptvault/internal/usecases/changelog"
)

type Handler struct {
	auth          *authuc.Service
	adminauth     *adminauthuc.Service
	changelog     *changeloguc.Service
	validate      *validator.Validate
	secureCookies bool
}

// NewHandler — adminauthSvc может быть nil для тестов без admin flow;
// в production всегда передаётся из app.New.
func NewHandler(authSvc *authuc.Service, adminauthSvc *adminauthuc.Service, changelogSvc *changeloguc.Service, secureCookies bool) *Handler {
	return &Handler{
		auth:          authSvc,
		adminauth:     adminauthSvc,
		changelog:     changelogSvc,
		validate:      validator.New(),
		secureCookies: secureCookies,
	}
}

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   h.secureCookies,
		// Lax (не Strict) чтобы cookie отправлялась при возврате с OAuth-провайдера
		// (GitHub/Google/Yandex) и при F5 на защищённых страницах. Strict блокировал
		// cookie при top-level cross-site navigation, что ломало OAuth login.
		// От CSRF защищает то, что POST-запросы на protected endpoints требуют
		// Authorization header с access-token (не cookie).
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	h.setRefreshCookie(w, "", -1)
}

// POST /api/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeAndValidate[RegisterRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if req.Username != "" {
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, req.Username); !matched {
			httperr.Respond(w, httperr.BadRequest("Никнейм может содержать только латинские буквы, цифры и _"))
			return
		}
	}

	user, err := h.auth.Register(r.Context(), req.Email, req.Password, utils.SanitizeString(req.Name), strings.TrimSpace(req.Username))
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteCreated(w, map[string]any{
		"email":   user.Email,
		"message": "Код подтверждения отправлен на email",
	})
}

// POST /api/auth/verify-email
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeAndValidate[VerifyEmailRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	user, tokens, err := h.auth.VerifyEmail(r.Context(), req.Email, req.Code)
	if err != nil {
		respondError(w, err)
		return
	}

	h.setRefreshCookie(w, tokens.RefreshToken, 7*24*3600)
	utils.WriteOK(w, NewAuthResponse(*user, tokens))
}

// POST /api/auth/resend-code
func (h *Handler) ResendCode(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeAndValidate[ResendCodeRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if err := h.auth.ResendCode(r.Context(), req.Email); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, map[string]string{"message": "Код отправлен"})
}

// POST /api/auth/login
// Flow:
//  1. Проверка credentials через AuthenticatePassword (не issue tokens).
//  2. Если user — обычный, issue tokens и вернуть как раньше.
//  3. Если user — admin с confirmed TOTP: вернуть pre_auth_token + totp_required=true.
//     НЕ issue полный JWT — клиент должен пройти /api/auth/verify-totp.
//  4. Если user — admin без TOTP (первый заход как admin): issue tokens +
//     totp_enrollment_required=true (клиент покажет /admin/totp enroll wizard).
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeAndValidate[LoginRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	user, err := h.auth.AuthenticatePassword(r.Context(), req.Email, req.Password)
	if err != nil {
		respondError(w, err)
		return
	}

	// Admin flow: проверить confirmed TOTP и либо требовать verify, либо
	// пропустить в enrollment wizard. Если adminauth service не wired —
	// degrade gracefully: admin логинится как обычный user (с предупреждением в логах).
	if user.IsAdmin() && h.adminauth != nil {
		confirmed, err := h.adminauth.IsConfirmed(r.Context(), user.ID)
		if err != nil {
			slog.Error("auth.login.totp_status_failed", "user_id", user.ID, "error", err)
			// Fail-safe: не разрешаем admin login без явного TOTP-check.
			httperr.Respond(w, httperr.Internal(err))
			return
		}
		if confirmed {
			// Confirmed TOTP — требуется verify step.
			preAuth, err := h.auth.IssuePreAuthToken(user.ID)
			if err != nil {
				httperr.Respond(w, httperr.Internal(err))
				return
			}
			utils.WriteOK(w, AdminLoginStepResponse{
				TOTPRequired: true,
				PreAuthToken: preAuth,
				User:         NewUserResponse(*user),
			})
			return
		}
		// Не confirmed — первый заход как admin, должен enroll.
		// Issue полные tokens + hint для UI.
		tokens, err := h.auth.IssueTokens(r.Context(), user)
		if err != nil {
			httperr.Respond(w, httperr.Internal(err))
			return
		}
		h.setRefreshCookie(w, tokens.RefreshToken, 7*24*3600)
		utils.WriteOK(w, AdminLoginStepResponse{
			TOTPEnrollmentRequired: true,
			AccessToken:            tokens.AccessToken,
			ExpiresIn:              tokens.ExpiresIn,
			User:                   NewUserResponse(*user),
		})
		return
	}

	// Обычный user — стандартный flow как раньше.
	tokens, err := h.auth.IssueTokens(r.Context(), user)
	if err != nil {
		httperr.Respond(w, httperr.Internal(err))
		return
	}
	h.setRefreshCookie(w, tokens.RefreshToken, 7*24*3600)
	utils.WriteOK(w, NewAuthResponse(*user, tokens))
}

// POST /api/auth/verify-totp
// Обменивает pre_auth_token + code на полный JWT pair. code может быть либо
// 6-значным TOTP кодом из Authenticator, либо recovery backup code.
func (h *Handler) VerifyTOTP(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeAndValidate[VerifyTOTPRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if h.adminauth == nil {
		httperr.Respond(w, httperr.Internal(nil))
		return
	}

	userID, err := h.auth.ValidatePreAuthToken(req.PreAuthToken)
	if err != nil {
		httperr.Respond(w, httperr.Unauthorized("pre_auth_token недействителен или истёк"))
		return
	}

	result, err := h.adminauth.Verify(r.Context(), userID, req.Code)
	if err != nil {
		// Маппим adminauth errors в HTTP через отдельный adminauth handler, но
		// здесь локальный respondError подходит — проверим только ErrInvalidCode.
		switch err {
		case adminauthuc.ErrInvalidCode:
			httperr.Respond(w, httperr.Unauthorized("неверный код"))
		case adminauthuc.ErrTOTPNotEnrolled:
			httperr.Respond(w, httperr.Conflict("TOTP не настроен"))
		default:
			httperr.Respond(w, httperr.Internal(err))
		}
		return
	}

	user, err := h.auth.Me(r.Context(), userID)
	if err != nil {
		httperr.Respond(w, httperr.Internal(err))
		return
	}

	tokens, err := h.auth.IssueTokens(r.Context(), user)
	if err != nil {
		httperr.Respond(w, httperr.Internal(err))
		return
	}

	h.setRefreshCookie(w, tokens.RefreshToken, 7*24*3600)
	utils.WriteOK(w, VerifyTOTPResponse{
		AccessToken:          tokens.AccessToken,
		ExpiresIn:            tokens.ExpiresIn,
		User:                 NewUserResponse(*user),
		UsedBackupCode:       result.UsedBackupCode,
		RemainingBackupCodes: result.RemainingBackupCodes,
	})
}

// POST /api/auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Читаем refresh_token из cookie (primary) или из body (fallback)
	var refreshToken string
	if c, err := r.Cookie("refresh_token"); err == nil && c.Value != "" {
		refreshToken = c.Value
	} else {
		var req RefreshRequest
		if err := utils.DecodeJSON(r, &req); err != nil {
			httperr.Respond(w, httperr.BadRequest(err.Error()))
			return
		}
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		httperr.Respond(w, httperr.Unauthorized("refresh token не предоставлен"))
		return
	}

	_, tokens, err := h.auth.Refresh(r.Context(), refreshToken)
	if err != nil {
		h.clearRefreshCookie(w)
		respondError(w, err)
		return
	}

	h.setRefreshCookie(w, tokens.RefreshToken, 7*24*3600)
	utils.WriteOK(w, AuthTokens{AccessToken: tokens.AccessToken, ExpiresIn: tokens.ExpiresIn})
}

// GET /api/auth/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	user, err := h.auth.Me(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	resp := NewUserResponse(*user)
	if h.changelog != nil {
		if unread, err := h.changelog.HasUnread(r.Context(), userID); err == nil {
			resp.HasUnreadChangelog = unread
		}
	}
	utils.WriteOK(w, resp)
}

// POST /api/auth/set-password/initiate — отправить код на email
func (h *Handler) InitiateSetPassword(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	if err := h.auth.InitiateSetPassword(r.Context(), userID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, map[string]string{"message": "Код отправлен на email"})
}

// POST /api/auth/set-password/confirm — проверить код и установить пароль
func (h *Handler) ConfirmSetPassword(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[SetPasswordRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if err := h.auth.ConfirmSetPassword(r.Context(), userID, req.Code, req.Password); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, map[string]string{"message": "Пароль установлен"})
}

// POST /api/auth/forgot-password — отправить код сброса (публичный)
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeAndValidate[ForgotPasswordRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if err := h.auth.ForgotPassword(r.Context(), req.Email); err != nil {
		respondError(w, err)
		return
	}

	// Всегда отвечаем одинаково — не раскрываем существование аккаунта
	utils.WriteOK(w, map[string]string{"message": "Если аккаунт существует, код отправлен на email"})
}

// POST /api/auth/reset-password — проверить код и установить новый пароль (публичный)
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeAndValidate[ResetPasswordRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if err := h.auth.ResetPassword(r.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, map[string]string{"message": "Пароль изменён"})
}

// PUT /api/auth/profile
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[UpdateProfileRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if req.AvatarURL != "" && !utils.ValidateURL(req.AvatarURL) {
		httperr.Respond(w, httperr.BadRequest("Некорректный URL аватара"))
		return
	}

	if req.Username != nil && *req.Username != "" {
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, *req.Username); !matched {
			httperr.Respond(w, httperr.BadRequest("Никнейм может содержать только латинские буквы, цифры и _"))
			return
		}
	}

	if req.Username != nil {
		trimmed := strings.TrimSpace(*req.Username)
		req.Username = &trimmed
	}

	user, err := h.auth.UpdateProfile(r.Context(), userID, utils.SanitizeString(req.Name), req.AvatarURL, req.Username)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, NewUserResponse(*user))
}

// PUT /api/auth/password
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[ChangePasswordRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if err := h.auth.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, map[string]string{"message": "Пароль изменён"})
}

// GET /api/auth/linked-accounts
func (h *Handler) LinkedAccounts(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	accounts, err := h.auth.GetLinkedAccounts(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, accounts)
}

// DELETE /api/auth/unlink/{provider}
func (h *Handler) UnlinkProvider(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		httperr.Respond(w, httperr.BadRequest("Укажите провайдер"))
		return
	}

	if err := h.auth.UnlinkProvider(r.Context(), userID, provider); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, map[string]string{"message": "Провайдер отвязан"})
}

// POST /api/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	if userID > 0 {
		if err := h.auth.InvalidateTokens(r.Context(), userID); err != nil {
			slog.Error("failed to invalidate tokens on logout", "user_id", userID, "error", err)
		}
	}
	h.clearRefreshCookie(w)
	utils.WriteOK(w, map[string]string{"message": "Выход выполнен"})
}

