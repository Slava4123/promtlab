package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestOAuthHandler(secret string) *OAuthHandler {
	return &OAuthHandler{jwtSecret: secret}
}

func TestSignAndVerifyLinkCookie_RoundTrip(t *testing.T) {
	h := newTestOAuthHandler("test-secret-32-chars-long!!!!!!!!")

	cookie := h.signLinkCookie(42)
	uid, ok := h.verifyLinkCookie(cookie)

	require.True(t, ok, "verifyLinkCookie should return true for a freshly signed cookie")
	assert.Equal(t, uint(42), uid)
}

func TestVerifyLinkCookie_TamperedSignature(t *testing.T) {
	h := newTestOAuthHandler("test-secret-32-chars-long!!!!!!!!")

	cookie := h.signLinkCookie(42)

	// Tamper with the signature: flip last character
	tampered := cookie[:len(cookie)-1] + "X"

	uid, ok := h.verifyLinkCookie(tampered)
	assert.False(t, ok, "verifyLinkCookie should return false for a tampered signature")
	assert.Equal(t, uint(0), uid)
}

func TestVerifyLinkCookie_Expired(t *testing.T) {
	h := newTestOAuthHandler("test-secret-32-chars-long!!!!!!!!")

	// Manually construct a cookie with a timestamp older than 300 seconds
	userID := uint(42)
	oldTimestamp := time.Now().Unix() - 301 // 301 seconds ago, exceeds the 5-min window
	data := fmt.Sprintf("%d:%d", userID, oldTimestamp)

	mac := hmac.New(sha256.New, []byte(h.jwtSecret))
	mac.Write([]byte(data))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	cookie := data + ":" + sig

	uid, ok := h.verifyLinkCookie(cookie)
	assert.False(t, ok, "verifyLinkCookie should return false for an expired cookie")
	assert.Equal(t, uint(0), uid)
}

func TestVerifyLinkCookie_MalformedFormat(t *testing.T) {
	h := newTestOAuthHandler("test-secret-32-chars-long!!!!!!!!")

	tests := []struct {
		name  string
		value string
	}{
		{"empty string", ""},
		{"single part", "42"},
		{"two parts", "42:1234567890"},
		{"four parts", "42:1234567890:sig:extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid, ok := h.verifyLinkCookie(tt.value)
			assert.False(t, ok, "verifyLinkCookie should return false for malformed input")
			assert.Equal(t, uint(0), uid)
		})
	}
}

func TestVerifyLinkCookie_DifferentSecret(t *testing.T) {
	signer := newTestOAuthHandler("secret-A-32-chars-long!!!!!!!!!!")
	verifier := newTestOAuthHandler("secret-B-32-chars-long!!!!!!!!!!")

	cookie := signer.signLinkCookie(42)

	uid, ok := verifier.verifyLinkCookie(cookie)
	assert.False(t, ok, "verifyLinkCookie should return false when secrets differ")
	assert.Equal(t, uint(0), uid)
}
