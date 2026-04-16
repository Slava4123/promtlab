package ai

import "promptvault/internal/middleware/ratelimit"

// userLimiter is a type alias for the generic rate limiter keyed by user ID.
type userLimiter = ratelimit.Limiter[uint]

// newUserLimiter creates a new per-user rate limiter.
func newUserLimiter(rpm int) *userLimiter {
	return ratelimit.NewLimiter[uint](rpm, ratelimit.UintHash)
}
