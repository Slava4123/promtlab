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
	// PlanID — текущий тариф юзера на момент выпуска access-токена.
	// Phase 15: позволяет analytics-handlers читать plan без users.GetByID
	// на каждый запрос (M9). Может устареть в течение access TTL (15 мин)
	// если admin сменил тариф — приемлемо для retention clamp.
	// Старые JWT (выпущенные до Phase 15) не имеют поля → fallback на DB.
	PlanID string `json:"plan_id,omitempty"`
	jwt.RegisteredClaims
}
