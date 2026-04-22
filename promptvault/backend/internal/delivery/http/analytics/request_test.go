package analytics

import (
	"testing"

	analyticsuc "promptvault/internal/usecases/analytics"
)

func TestParseRange(t *testing.T) {
	tests := map[string]analyticsuc.RangeID{
		"7d":        analyticsuc.Range7d,
		"30d":       analyticsuc.Range30d,
		"90d":       analyticsuc.Range90d,
		"365d":      analyticsuc.Range365d,
		"":          analyticsuc.Range7d, // default
		"unknown":   analyticsuc.Range7d, // safe fallback
		"1y":        analyticsuc.Range7d, // safe fallback
		"365 days":  analyticsuc.Range7d, // spaces — not handled
	}
	for in, want := range tests {
		if got := parseRange(in); got != want {
			t.Errorf("parseRange(%q) = %q, want %q", in, got, want)
		}
	}
}
