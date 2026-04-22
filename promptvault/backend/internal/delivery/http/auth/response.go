package auth

import (
	"cmp"
	"time"

	"promptvault/internal/models"
	authuc "promptvault/internal/usecases/auth"
)

// AuthTokens содержит только access token для JSON response.
// Refresh token доставляется через HttpOnly cookie, не в JSON.
type AuthTokens struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type AuthResponse struct {
	User   UserResponse `json:"user"`
	Tokens AuthTokens   `json:"tokens"`
}

type UserResponse struct {
	ID                    uint       `json:"id"`
	Email                 string     `json:"email"`
	Name                  string     `json:"name"`
	Username              string     `json:"username,omitempty"`
	AvatarURL             string     `json:"avatar_url,omitempty"`
	EmailVerified         bool       `json:"email_verified"`
	HasPassword           bool       `json:"has_password"`
	PlanID                string     `json:"plan_id"`
	Role                  string     `json:"role"`
	Status                string     `json:"status"`
	OnboardingCompletedAt *time.Time `json:"onboarding_completed_at,omitempty"`
	HasUnreadChangelog    bool       `json:"has_unread_changelog"`
}

func NewUserResponse(u models.User) UserResponse {
	return UserResponse{
		ID:                    u.ID,
		Email:                 u.Email,
		Name:                  u.Name,
		Username:              u.Username,
		AvatarURL:             u.AvatarURL,
		EmailVerified:         u.EmailVerified,
		HasPassword:           u.HasPassword(),
		PlanID:                cmp.Or(u.PlanID, "free"),
		Role:                  string(u.Role),
		Status:                string(u.Status),
		OnboardingCompletedAt: u.OnboardingCompletedAt,
	}
}

func NewAuthResponse(u models.User, tokens *authuc.TokenPair) AuthResponse {
	return AuthResponse{
		User: NewUserResponse(u),
		Tokens: AuthTokens{
			AccessToken: tokens.AccessToken,
			ExpiresIn:   tokens.ExpiresIn,
		},
	}
}

// AdminLoginStepResponse — вариативный response для POST /api/auth/login
// когда юзер — admin. Три состояния:
//  1. TOTPRequired=true: у admin confirmed TOTP → фронт показывает TOTP input
//     и POST'ит в /api/auth/verify-totp с PreAuthToken. AccessToken НЕ отдан.
//  2. TOTPEnrollmentRequired=true: admin впервые логинится → у него ещё нет
//     TOTP enrollment, фронт ведёт на /admin/totp wizard. AccessToken ОТДАН
//     (юзер залогинен, но должен настроить TOTP перед admin action'ами).
//  3. Оба false — не используется для admin flow (тогда возвращается AuthResponse).
type AdminLoginStepResponse struct {
	// Для "требуется TOTP verification":
	TOTPRequired bool   `json:"totp_required,omitempty"`
	PreAuthToken string `json:"pre_auth_token,omitempty"`

	// Для "первый заход как admin, нужен enrollment":
	TOTPEnrollmentRequired bool   `json:"totp_enrollment_required,omitempty"`
	AccessToken            string `json:"access_token,omitempty"`
	ExpiresIn              int64  `json:"expires_in,omitempty"`

	User UserResponse `json:"user"`
}

// VerifyTOTPResponse — результат POST /api/auth/verify-totp. Аналогичен
// обычному AuthResponse, плюс информация о использованном backup code.
// UsedBackupCode=true — UI показывает баннер «осталось N backup кодов».
type VerifyTOTPResponse struct {
	AccessToken          string       `json:"access_token"`
	ExpiresIn            int64        `json:"expires_in"`
	User                 UserResponse `json:"user"`
	UsedBackupCode       bool         `json:"used_backup_code"`
	RemainingBackupCodes int          `json:"remaining_backup_codes"`
}
