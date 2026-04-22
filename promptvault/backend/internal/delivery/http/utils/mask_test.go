package utils

import "testing"

func TestMaskEmail(t *testing.T) {
	tests := map[string]string{
		"alice@example.com":    "a***@example.com",
		"a@b.c":                "a***@b.c",
		"":                     "",
		"no-at-sign":           "",
		"@nouser.com":          "",
		"very.long.name@x.io":  "v***@x.io",
	}
	for in, want := range tests {
		got := MaskEmail(in)
		if got != want {
			t.Errorf("MaskEmail(%q) = %q, want %q", in, got, want)
		}
	}
}
