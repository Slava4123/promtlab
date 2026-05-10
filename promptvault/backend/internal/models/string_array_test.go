// Tests for MJ-38 StringArray (drop-in for pq.StringArray).
package models

import (
	"reflect"
	"testing"
)

func TestStringArray_Value_Empty(t *testing.T) {
	got, err := StringArray{}.Value()
	if err != nil {
		t.Fatalf("Value() returned err: %v", err)
	}
	if got != "{}" {
		t.Errorf("empty array: got %q, want {}", got)
	}
}

func TestStringArray_Value_Nil(t *testing.T) {
	var a StringArray
	got, err := a.Value()
	if err != nil {
		t.Fatalf("Value() returned err: %v", err)
	}
	if got != nil {
		t.Errorf("nil array: got %v, want nil", got)
	}
}

func TestStringArray_Value_SimpleStrings(t *testing.T) {
	got, err := StringArray{"a", "b", "c"}.Value()
	if err != nil {
		t.Fatalf("Value() returned err: %v", err)
	}
	want := `{"a","b","c"}`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStringArray_Value_EscapesSpecialChars(t *testing.T) {
	got, err := StringArray{`with"quote`, `with\backslash`}.Value()
	if err != nil {
		t.Fatalf("Value() returned err: %v", err)
	}
	want := `{"with\"quote","with\\backslash"}`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStringArray_Scan_String(t *testing.T) {
	var a StringArray
	if err := a.Scan(`{"hello","world"}`); err != nil {
		t.Fatalf("Scan returned err: %v", err)
	}
	if !reflect.DeepEqual([]string(a), []string{"hello", "world"}) {
		t.Errorf("got %v, want [hello world]", a)
	}
}

func TestStringArray_Scan_Bytes(t *testing.T) {
	var a StringArray
	if err := a.Scan([]byte(`{"a","b"}`)); err != nil {
		t.Fatalf("Scan returned err: %v", err)
	}
	if !reflect.DeepEqual([]string(a), []string{"a", "b"}) {
		t.Errorf("got %v, want [a b]", a)
	}
}

func TestStringArray_Scan_Empty(t *testing.T) {
	var a StringArray
	if err := a.Scan(`{}`); err != nil {
		t.Fatalf("Scan returned err: %v", err)
	}
	if len(a) != 0 {
		t.Errorf("got %v, want empty", a)
	}
}

func TestStringArray_Scan_Nil(t *testing.T) {
	a := StringArray{"will", "be", "cleared"}
	if err := a.Scan(nil); err != nil {
		t.Fatalf("Scan returned err: %v", err)
	}
	if a != nil {
		t.Errorf("got %v, want nil", a)
	}
}

func TestStringArray_Scan_EscapedChars(t *testing.T) {
	var a StringArray
	if err := a.Scan(`{"with\"quote","with\\backslash"}`); err != nil {
		t.Fatalf("Scan returned err: %v", err)
	}
	want := []string{`with"quote`, `with\backslash`}
	if !reflect.DeepEqual([]string(a), want) {
		t.Errorf("got %v, want %v", []string(a), want)
	}
}

func TestStringArray_Scan_InvalidLiteral_Errors(t *testing.T) {
	var a StringArray
	if err := a.Scan(`not-an-array`); err == nil {
		t.Errorf("expected error on invalid literal, got nil")
	}
}

func TestStringArray_Scan_UnsupportedType_Errors(t *testing.T) {
	var a StringArray
	if err := a.Scan(42); err == nil {
		t.Errorf("expected error on int input, got nil")
	}
}

func TestStringArray_RoundTrip(t *testing.T) {
	// Value → Scan должны быть обратимы для типичных значений.
	original := StringArray{"alpha", "beta", "gamma", `with"quote`, `back\slash`}
	literal, err := original.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	var roundTripped StringArray
	if err := roundTripped.Scan(literal); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !reflect.DeepEqual([]string(original), []string(roundTripped)) {
		t.Errorf("round-trip:\noriginal: %v\nresult:   %v", []string(original), []string(roundTripped))
	}
}
