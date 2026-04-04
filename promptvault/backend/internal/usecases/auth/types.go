package auth

import "github.com/golang-jwt/jwt/v5"

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Claims struct {
	UserID uint      `json:"user_id"`
	Type   TokenType `json:"type"`
	Nonce  string    `json:"nonce,omitempty"`
	jwt.RegisteredClaims
}
