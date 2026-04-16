package auth

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// ctxKey — приватный тип для ключа контекста (M-7 OAuth referredBy).
type ctxKey int

const ctxReferredByKey ctxKey = 1

// WithReferredBy кладёт реферальный код в контекст. Вызывается из OAuth-callback
// handler перед s.CompleteCallback — сервис прочитает его при создании нового юзера.
func WithReferredBy(ctx context.Context, code string) context.Context {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxReferredByKey, code)
}

func referredByFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(ctxReferredByKey).(string)
	return v
}

// GenerateReferralCode — 8-символьный код из Crockford-Base32 (без I/L/O/U/0/1
// для читаемости на телефонах). 8 символов × 5 бит = 40 бит энтропии —
// на 2^40 кодов коллизия становится вероятной после ~2^20 юзеров, что с запасом
// для MVP. При INSERT с UNIQUE constraint на collision ретраим в вызывающем коде.
func GenerateReferralCode() (string, error) {
	var raw [5]byte // 5 байт × 8 бит = 40 бит → 8 base32 символов
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	// std base32 с padding, берём первые 8 символов и убираем неоднозначные.
	s := base32.StdEncoding.EncodeToString(raw[:])
	s = strings.ToUpper(s)
	// Маппинг стандартного base32 → Crockford-стиль (упрощённый): заменим
	// неоднозначные символы, чтобы юзеры не путали I/1, O/0.
	repl := strings.NewReplacer("I", "J", "O", "P", "L", "M", "U", "V", "0", "2", "1", "3")
	s = repl.Replace(s)
	if len(s) < 8 {
		return s, nil
	}
	return s[:8], nil
}

// ReferralInfo — DTO для GET /api/auth/referral (M-7).
type ReferralInfo struct {
	Code          string `json:"code"`
	InvitedCount  int64  `json:"invited_count"`
	ReferredBy    string `json:"referred_by,omitempty"`
	RewardGranted bool   `json:"reward_granted"`
}

// GetReferralInfo возвращает код юзера + счётчик приглашённых.
func (s *Service) GetReferralInfo(ctx context.Context, userID uint) (*ReferralInfo, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	count, err := s.users.CountReferredBy(ctx, user.ReferralCode)
	if err != nil {
		return nil, err
	}
	return &ReferralInfo{
		Code:          user.ReferralCode,
		InvitedCount:  count,
		ReferredBy:    user.ReferredBy,
		RewardGranted: user.ReferralRewardedAt != nil,
	}, nil
}

// createUserWithReferralCode генерит уникальный ReferralCode и создаёт юзера.
// 3 попытки на случай UNIQUE collision (40 бит энтропии — практически нереально
// до миллионов юзеров, но перестраховка).
// Распознавание collision error'а через SQLSTATE недоступно без прямого доступа
// к pgconn; используем substring match в сообщении — GORM оборачивает ошибку
// так, что "referral_code" всегда есть в тексте при дубликате этой колонки.
// Standalone-функция (не метод), чтобы шарить между Service и OAuthService.
func createUserWithReferralCode(ctx context.Context, users repo.UserRepository, user *models.User) error {
	const maxAttempts = 3
	for i := 0; i < maxAttempts; i++ {
		code, err := GenerateReferralCode()
		if err != nil {
			return fmt.Errorf("generate referral code: %w", err)
		}
		user.ReferralCode = code
		err = users.Create(ctx, user)
		if err == nil {
			return nil
		}
		if strings.Contains(strings.ToLower(err.Error()), "referral_code") {
			continue
		}
		return err
	}
	return errors.New("referral code generation: collisions exceeded")
}
