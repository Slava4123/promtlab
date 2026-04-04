package auth

import authuc "promptvault/internal/usecases/auth"

type TokenValidator interface {
	ValidateAccessToken(token string) (*authuc.Claims, error)
}
