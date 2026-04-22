package analytics

import "errors"

var (
	ErrForbidden    = errors.New("analytics: forbidden")
	ErrNotFound     = errors.New("analytics: not found")
	ErrMaxRequired  = errors.New("analytics: feature requires Max plan")
	ErrProRequired  = errors.New("analytics: feature requires Pro or Max plan")
)
