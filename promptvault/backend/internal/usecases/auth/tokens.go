package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token helpers — generate/validate JWT пары. Q-13: выделено из auth.go
// чтобы оставить auth.go только для user-flow (register/login/verify/profile).

// ValidateAccessToken — удобная обёртка над ValidateToken для access-токенов.
func (s *Service) ValidateAccessToken(token string) (*Claims, error) {
	return s.ValidateToken(token, TokenTypeAccess)
}

// ValidateToken проверяет подпись, срок, тип. Возвращает конкретные ошибки:
// ErrExpiredToken — подпись верна, но expired; ErrInvalidToken — всё остальное
// (плохая подпись, не тот type, malformed). Handlers отличают их чтобы
// 401 expired → попробовать refresh, 401 invalid → forced re-login.
func (s *Service) ValidateToken(tokenStr string, expectedType TokenType) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid || claims.Type != expectedType {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *Service) generateTokenPair(userID uint, nonce string) (*TokenPair, error) {
	now := time.Now()

	accessToken, err := s.generateToken(userID, TokenTypeAccess, "", now, s.accessDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(userID, TokenTypeRefresh, nonce, now, s.refreshDuration)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessDuration.Seconds()),
	}, nil
}

func (s *Service) generateToken(userID uint, tokenType TokenType, nonce string, now time.Time, duration time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Type:   tokenType,
		Nonce:  nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}
