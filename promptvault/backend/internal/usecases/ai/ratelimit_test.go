package ai

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllow_WithinLimit(t *testing.T) {
	rl := newUserLimiter(5)

	for i := 0; i < 3; i++ {
		assert.True(t, rl.Allow(1), "request %d should be allowed", i+1)
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	rl := newUserLimiter(3)

	for i := 0; i < 3; i++ {
		assert.True(t, rl.Allow(1), "request %d should be allowed", i+1)
	}

	assert.False(t, rl.Allow(1), "4th request should be blocked")
}

func TestAllow_UserIsolation(t *testing.T) {
	rl := newUserLimiter(2)

	assert.True(t, rl.Allow(1))
	assert.True(t, rl.Allow(1))
	assert.False(t, rl.Allow(1), "user 1 should be blocked after reaching limit")

	assert.True(t, rl.Allow(2), "user 2 should still be allowed")
}

func TestAllow_ZeroLimit(t *testing.T) {
	rl := newUserLimiter(0)

	for i := 0; i < 10; i++ {
		assert.True(t, rl.Allow(1), "request %d should be allowed with zero limit", i+1)
	}
}

func TestAllow_Concurrent(t *testing.T) {
	rl := newUserLimiter(50)

	var allowed atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow(1) {
				allowed.Add(1)
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(50), allowed.Load(), "exactly 50 of 100 concurrent requests should be allowed")
}

func TestAllow_ExactlyAtLimit(t *testing.T) {
	rl := newUserLimiter(3)

	assert.True(t, rl.Allow(1), "1st should be allowed")
	assert.True(t, rl.Allow(1), "2nd should be allowed")
	assert.True(t, rl.Allow(1), "3rd should be allowed (exactly at limit)")
	assert.False(t, rl.Allow(1), "4th should be blocked (over limit)")
}
