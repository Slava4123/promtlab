// MN-1 — pure-function tests for user package: maskEmail.
// Service.Search/GetByID требуют UserRepository mock — отдельный PR.
package user

import "testing"

func TestMaskEmail(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"alice@example.com", "a***@example.com"},
		{"bob@gmail.com", "b***@gmail.com"},
		{"x@y.z", "x***@y.z"},
		{"@example.com", "***"},     // empty local part
		{"noatsign", "***"},          // no @
		{"", "***"},                  // empty
		{"first.last@domain.tld", "f***@domain.tld"},
		{"u@a", "u***@a"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := maskEmail(tc.input)
			if got != tc.want {
				t.Errorf("maskEmail(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
