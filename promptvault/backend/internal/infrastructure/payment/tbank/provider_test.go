package tbank

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// computeRef вычисляет эталонную подпись для заданного набора (ключ-значение) + password.
// Алгоритм T-Bank: сортировка ключей → конкатенация значений → SHA-256(hex).
func computeRef(password string, kv map[string]string) string {
	kv = copyMap(kv)
	kv["Password"] = password
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	// Ручной bubble sort — никаких зависимостей.
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	concat := ""
	for _, k := range keys {
		concat += kv[k]
	}
	sum := sha256.Sum256([]byte(concat))
	return hex.EncodeToString(sum[:])
}

func copyMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func TestVerifyWebhookSignature(t *testing.T) {
	const password = "test-password"
	p := NewProvider(Config{Password: password})

	base := map[string]string{
		"TerminalKey": "1234567890DEMO",
		"OrderId":     "sub_42_abcdef012345",
		"Success":     "true",
		"Status":      "CONFIRMED",
		"PaymentId":   "9999999",
		"Amount":      "59900",
	}
	validToken := computeRef(password, base)

	tests := []struct {
		name   string
		params map[string]string
		token  string
		want   bool
	}{
		{
			name: "valid signature",
			params: mergeMap(base, map[string]string{
				"Token": validToken,
			}),
			token: validToken,
			want:  true,
		},
		{
			name: "signature is case-insensitive hex",
			params: mergeMap(base, map[string]string{
				"Token": validToken,
			}),
			token: upperCase(validToken),
			want:  true,
		},
		{
			name: "token tampered → invalid",
			params: mergeMap(base, map[string]string{
				"Token": validToken,
			}),
			token: "deadbeef" + validToken[8:],
			want:  false,
		},
		{
			name: "amount изменён → invalid",
			params: mergeMap(base, map[string]string{
				"Amount": "99900",
				"Token":  validToken,
			}),
			token: validToken,
			want:  false,
		},
		{
			name: "Receipt-объект исключается из подписи (подпись остаётся валидной)",
			params: mergeMap(base, map[string]string{
				"Token":   validToken,
				"Receipt": `{"Email":"a@b.ru","Items":[{"Name":"Pro"}]}`,
			}),
			token: validToken,
			want:  true,
		},
		{
			name: "DATA-объект исключается из подписи",
			params: mergeMap(base, map[string]string{
				"Token": validToken,
				"DATA":  `{"OS":"iOS"}`,
			}),
			token: validToken,
			want:  true,
		},
		{
			name: "missing required field (Amount) → подпись не сойдётся",
			params: map[string]string{
				"TerminalKey": base["TerminalKey"],
				"OrderId":     base["OrderId"],
				"Success":     base["Success"],
				"Status":      base["Status"],
				"PaymentId":   base["PaymentId"],
				"Token":       validToken,
			},
			token: validToken,
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := p.VerifyWebhookSignature(tc.params, tc.token)
			if got != tc.want {
				t.Fatalf("VerifyWebhookSignature = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGenerateToken_DeterministicOrdering(t *testing.T) {
	// Проверяем что порядок ключей в map не влияет на результат —
	// generateToken должен сортировать перед конкатенацией.
	p := NewProvider(Config{Password: "pwd"})

	params1 := map[string]string{
		"TerminalKey": "t1",
		"Amount":      "100",
		"OrderId":     "order1",
	}
	params2 := map[string]string{
		"OrderId":     "order1",
		"TerminalKey": "t1",
		"Amount":      "100",
	}

	if p.generateToken(params1) != p.generateToken(params2) {
		t.Fatalf("generateToken должен быть детерминированным по содержимому, не по порядку")
	}
}

func mergeMap(a, b map[string]string) map[string]string {
	out := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func upperCase(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'f' {
			b[i] = c - 32
		}
	}
	return string(b)
}
