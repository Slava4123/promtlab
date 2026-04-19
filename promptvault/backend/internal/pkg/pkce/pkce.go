// Package pkce реализует server-side проверку PKCE challenge/verifier для
// OAuth 2.1 Authorization Code Grant (RFC 7636).
//
// Клиент генерирует code_verifier (43-128 байт URL-safe base64) и отправляет
// code_challenge = BASE64URL(SHA256(verifier)) в authorize-запросе.
// На token-exchange клиент присылает verifier, мы пересчитываем challenge
// и сравниваем constant-time с тем, что был в authorize.
//
// MCP spec 2025-06-18 §Authorization Code Protection: PKCE S256 MUST.
package pkce

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
)

const (
	MethodS256  = "S256"
	MethodPlain = "plain" // RFC 7636 допускает, но мы принимаем только S256

	minVerifierLen = 43
	maxVerifierLen = 128
)

var (
	ErrUnsupportedMethod = errors.New("pkce: unsupported code_challenge_method, only S256 allowed")
	ErrVerifierLength    = errors.New("pkce: code_verifier length must be 43-128 characters")
	ErrMismatch          = errors.New("pkce: code_verifier does not match code_challenge")
)

// ComputeS256 возвращает BASE64URL(SHA256(verifier)) без padding — формат RFC 7636.
func ComputeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// Verify проверяет, что клиент прислал valid verifier для ранее сохранённого challenge.
// Возвращает ErrMismatch при несовпадении, constant-time сравнение.
func Verify(method, challenge, verifier string) error {
	if method != MethodS256 {
		return ErrUnsupportedMethod
	}
	if len(verifier) < minVerifierLen || len(verifier) > maxVerifierLen {
		return ErrVerifierLength
	}
	expected := ComputeS256(verifier)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(challenge)) != 1 {
		return ErrMismatch
	}
	return nil
}
