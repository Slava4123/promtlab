// Package tokens генерирует и хэширует opaque-токены для OAuth access/refresh
// и authorization codes. Тот же паттерн, что используется для api_keys —
// клиенту отдаётся raw-значение с префиксом, в БД хранится SHA256.
package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// Prefixes для визуального разделения типов токенов в логах и UI.
const (
	PrefixAccessToken  = "pvoat_" // PromptVault OAuth Access Token
	PrefixRefreshToken = "pvort_" // PromptVault OAuth Refresh Token
	PrefixAuthCode     = "pvoac_" // PromptVault OAuth Auth Code
	PrefixClientID     = "pvoci_" // PromptVault OAuth Client ID
	PrefixClientSecret = "pvocs_" // PromptVault OAuth Client Secret
)

// randomBytes возвращает n криптостойких случайных байт.
func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("tokens: random read: %w", err)
	}
	return b, nil
}

// New возвращает пару (raw, hash). Raw отдаётся клиенту, hash хранится в БД.
// Формат raw: <prefix><32 байта base64url> → ~50 символов.
func New(prefix string) (raw, hash string, err error) {
	b, err := randomBytes(32)
	if err != nil {
		return "", "", err
	}
	raw = prefix + base64.RawURLEncoding.EncodeToString(b)
	hash = Hash(raw)
	return raw, hash, nil
}

// Hash — SHA256 hex, одинаковый для всех токенов в системе.
func Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
