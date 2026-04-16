package auth

type RegisterRequest struct {
	Email      string `json:"email" validate:"required,email,max=255"`
	Password   string `json:"password" validate:"required,min=8,max=128"`
	Name       string `json:"name" validate:"required,min=1,max=100"`
	Username   string `json:"username" validate:"omitempty,min=3,max=30"`
	ReferredBy string `json:"referred_by" validate:"omitempty,alphanum,len=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

type ResendCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type SetPasswordRequest struct {
	Code     string `json:"code" validate:"required,len=6"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Code        string `json:"code" validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

type UpdateProfileRequest struct {
	Name      string  `json:"name" validate:"required,min=1,max=100"`
	Username  *string `json:"username" validate:"omitempty,min=3,max=30"`
	AvatarURL string  `json:"avatar_url" validate:"omitempty,url,max=500"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

// VerifyTOTPRequest — POST /api/auth/verify-totp.
// Code — либо 6-значный TOTP из Authenticator, либо backup code формата xxxxx-xxxxx.
type VerifyTOTPRequest struct {
	PreAuthToken string `json:"pre_auth_token" validate:"required"`
	Code         string `json:"code" validate:"required,min=6,max=20"`
}
