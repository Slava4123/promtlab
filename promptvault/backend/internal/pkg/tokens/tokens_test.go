package tokens

import (
	"strings"
	"testing"
)

func TestNew_HasPrefix(t *testing.T) {
	raw, hash, err := New(PrefixAccessToken)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if !strings.HasPrefix(raw, PrefixAccessToken) {
		t.Fatalf("raw = %q, want prefix %q", raw, PrefixAccessToken)
	}
	if len(hash) != 64 {
		t.Fatalf("hash len = %d, want 64 (SHA256 hex)", len(hash))
	}
}

func TestNew_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := range 1000 {
		raw, _, err := New(PrefixAccessToken)
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if seen[raw] {
			t.Fatalf("duplicate token at iteration %d: %q", i, raw)
		}
		seen[raw] = true
	}
}

func TestHash_Deterministic(t *testing.T) {
	if Hash("x") != Hash("x") {
		t.Fatal("Hash must be deterministic")
	}
	if Hash("x") == Hash("y") {
		t.Fatal("Hash must differ for different inputs")
	}
}
