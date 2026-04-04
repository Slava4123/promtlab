package utils

import (
	"html"
	"strings"
)

// SanitizeString escapes HTML entities and trims whitespace.
func SanitizeString(s string) string {
	return html.EscapeString(strings.TrimSpace(s))
}

// SanitizeStringPtr escapes HTML entities for a pointer string.
func SanitizeStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := SanitizeString(*s)
	return &v
}

// ValidateURL checks that a URL starts with https:// (blocks javascript:, data:, http://, etc.)
func ValidateURL(u string) bool {
	u = strings.TrimSpace(u)
	return u == "" || strings.HasPrefix(u, "https://")
}
