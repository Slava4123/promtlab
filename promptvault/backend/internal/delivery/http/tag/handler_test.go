// MN-7 — handler-level tests для tag-пакета.
// Полное покрытие Create/List/Delete требует mock'ов TagRepository (объёмно)
// — здесь pure helper parseTeamID и базовая HTTP integration.
package tag

import (
	"net/http/httptest"
	"testing"
)

func TestParseTeamID(t *testing.T) {
	cases := []struct {
		name string
		q    string
		want *uint
	}{
		{"empty — nil", "", nil},
		{"valid — 42", "?team_id=42", uintPtr(42)},
		{"non-numeric — nil", "?team_id=abc", nil},
		{"negative — nil", "?team_id=-1", nil},
		{"overflow uint32 — nil", "?team_id=99999999999999999999", nil},
		{"zero — &0", "?team_id=0", uintPtr(0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/api/tags"+tc.q, nil)
			got := parseTeamID(r)
			switch {
			case tc.want == nil && got != nil:
				t.Errorf("ожидался nil, got %d", *got)
			case tc.want != nil && got == nil:
				t.Errorf("ожидался %d, got nil", *tc.want)
			case tc.want != nil && got != nil && *tc.want != *got:
				t.Errorf("ожидался %d, got %d", *tc.want, *got)
			}
		})
	}
}

func uintPtr(v uint) *uint { return &v }
