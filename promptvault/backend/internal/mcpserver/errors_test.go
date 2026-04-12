package mcpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"

	colluc "promptvault/internal/usecases/collection"
	promptuc "promptvault/internal/usecases/prompt"
	shareuc "promptvault/internal/usecases/share"
	taguc "promptvault/internal/usecases/tag"
)

func TestMapDomainError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		// prompts
		{"prompt not found", promptuc.ErrNotFound, "prompt not found"},
		{"prompt forbidden", promptuc.ErrForbidden, "access denied"},
		{"prompt viewer read-only", promptuc.ErrViewerReadOnly, "read-only access"},
		{"prompt version not found", promptuc.ErrVersionNotFound, "version not found"},
		{"prompt workspace mismatch", promptuc.ErrWorkspaceMismatch, "collections and tags must belong to the same workspace"},
		{"prompt pin forbidden", promptuc.ErrPinForbidden, "pin forbidden for viewers"},

		// collections
		{"collection not found", colluc.ErrNotFound, "collection not found"},
		{"collection forbidden", colluc.ErrForbidden, "access denied"},
		{"collection viewer read-only", colluc.ErrViewerReadOnly, "read-only access"},
		{"collection invalid color", colluc.ErrInvalidColor, "invalid color: use #RRGGBB format"},
		{"collection invalid icon", colluc.ErrInvalidIcon, "invalid icon"},

		// tags
		{"tag not found", taguc.ErrNotFound, "tag not found"},
		{"tag forbidden", taguc.ErrForbidden, "access denied"},
		{"tag viewer read-only", taguc.ErrViewerReadOnly, "read-only access"},
		{"tag name empty", taguc.ErrNameEmpty, "tag name is required"},

		// shares
		{"share not found", shareuc.ErrNotFound, "share link not found"},
		{"share prompt not found", shareuc.ErrPromptNotFound, "prompt not found"},
		{"share forbidden", shareuc.ErrForbidden, "access denied"},
		{"share viewer read-only", shareuc.ErrViewerReadOnly, "read-only access"},

		// unknown
		{"unknown error", assert.AnError, "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapDomainError(tt.err)
			assert.EqualError(t, got, tt.wantMsg)
		})
	}
}
