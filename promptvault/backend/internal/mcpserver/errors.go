package mcpserver

import (
	"errors"
	"fmt"

	colluc "promptvault/internal/usecases/collection"
	promptuc "promptvault/internal/usecases/prompt"
	shareuc "promptvault/internal/usecases/share"
	taguc "promptvault/internal/usecases/tag"
)

func mapDomainError(err error) error {
	switch {
	// prompts
	case errors.Is(err, promptuc.ErrNotFound):
		return fmt.Errorf("prompt not found")
	case errors.Is(err, promptuc.ErrForbidden):
		return fmt.Errorf("access denied")
	case errors.Is(err, promptuc.ErrViewerReadOnly):
		return fmt.Errorf("read-only access")
	case errors.Is(err, promptuc.ErrVersionNotFound):
		return fmt.Errorf("version not found")
	case errors.Is(err, promptuc.ErrWorkspaceMismatch):
		return fmt.Errorf("collections and tags must belong to the same workspace")
	case errors.Is(err, promptuc.ErrPinForbidden):
		return fmt.Errorf("pin forbidden for viewers")

	// collections
	case errors.Is(err, colluc.ErrNotFound):
		return fmt.Errorf("collection not found")
	case errors.Is(err, colluc.ErrForbidden):
		return fmt.Errorf("access denied")
	case errors.Is(err, colluc.ErrViewerReadOnly):
		return fmt.Errorf("read-only access")
	case errors.Is(err, colluc.ErrInvalidColor):
		return fmt.Errorf("invalid color: use #RRGGBB format")
	case errors.Is(err, colluc.ErrInvalidIcon):
		return fmt.Errorf("invalid icon")

	// tags
	case errors.Is(err, taguc.ErrNotFound):
		return fmt.Errorf("tag not found")
	case errors.Is(err, taguc.ErrForbidden):
		return fmt.Errorf("access denied")
	case errors.Is(err, taguc.ErrViewerReadOnly):
		return fmt.Errorf("read-only access")
	case errors.Is(err, taguc.ErrNameEmpty):
		return fmt.Errorf("tag name is required")

	// shares
	case errors.Is(err, shareuc.ErrNotFound):
		return fmt.Errorf("share link not found")
	case errors.Is(err, shareuc.ErrPromptNotFound):
		return fmt.Errorf("prompt not found")
	case errors.Is(err, shareuc.ErrForbidden):
		return fmt.Errorf("access denied")
	case errors.Is(err, shareuc.ErrViewerReadOnly):
		return fmt.Errorf("read-only access")
	}
	return fmt.Errorf("internal server error")
}
