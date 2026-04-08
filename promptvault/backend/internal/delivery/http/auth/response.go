package auth

import (
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
	DefaultModel          string     `json:"default_model"`
	OnboardingCompletedAt *time.Time `json:"onboarding_completed_at,omitempty"`
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
		DefaultModel:          u.DefaultModel,
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
