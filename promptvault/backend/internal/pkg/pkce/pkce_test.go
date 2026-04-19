package pkce

import (
	"errors"
	"strings"
	"testing"
)

func TestComputeS256_RFC7636Vector(t *testing.T) {
	// Пример из RFC 7636 Appendix B.
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if got := ComputeS256(verifier); got != want {
		t.Fatalf("ComputeS256 = %q, want %q", got, want)
	}
}

func TestVerify_OK(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := ComputeS256(verifier)
	if err := Verify(MethodS256, challenge, verifier); err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
}

func TestVerify_Mismatch(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := ComputeS256("different-verifier-value-must-be-at-least-43-chars")
	err := Verify(MethodS256, challenge, verifier)
	if !errors.Is(err, ErrMismatch) {
		t.Fatalf("expected ErrMismatch, got %v", err)
	}
}

func TestVerify_UnsupportedMethod(t *testing.T) {
	err := Verify(MethodPlain, "x", "y")
	if !errors.Is(err, ErrUnsupportedMethod) {
		t.Fatalf("expected ErrUnsupportedMethod, got %v", err)
	}
}

func TestVerify_ShortVerifier(t *testing.T) {
	err := Verify(MethodS256, "challenge", "short")
	if !errors.Is(err, ErrVerifierLength) {
		t.Fatalf("expected ErrVerifierLength, got %v", err)
	}
}

func TestVerify_LongVerifier(t *testing.T) {
	err := Verify(MethodS256, "challenge", strings.Repeat("a", maxVerifierLen+1))
	if !errors.Is(err, ErrVerifierLength) {
		t.Fatalf("expected ErrVerifierLength, got %v", err)
	}
}
