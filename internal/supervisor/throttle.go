package supervisor

import (
	"sync"
	"time"
)

// ThrottlePolicy defines how restart throttling behaves.
type ThrottlePolicy struct {
	MinDelay time.Duration
	MaxDelay time.Duration
	Factor   float64
}

// DefaultThrottlePolicy returns sensible defaults for restart throttling.
func DefaultThrottlePolicy() ThrottlePolicy {
	return ThrottlePolicy{
		MinDelay: 100 * time.Millisecond,
		MaxDelay: 30 * time.Second,
		Factor:   2.0,
	}
}

// Throttle controls restart delay with exponential backoff.
type Throttle struct {
	mu      sync.Mutex
	policy  ThrottlePolicy
	current time.Duration
	resets  int
}

// NewThrottle creates a Throttle with the given policy.
func NewThrottle(policy ThrottlePolicy) *Throttle {
	return &Throttle{
		policy:  policy,
		current: policy.MinDelay,
	}
}

// Next returns the current delay and advances the backoff state.
func (t *Throttle) Next() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()

	delay := t.current
	next := time.Duration(float64(t.current) * t.policy.Factor)
	if next > t.policy.MaxDelay {
		next = t.policy.MaxDelay
	}
	t.current = next
	return delay
}

// Reset resets the backoff to the minimum delay.
func (t *Throttle) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.current = t.policy.MinDelay
	t.resets++
}

// Resets returns how many times the throttle has been reset.
func (t *Throttle) Resets() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.resets
}

// Current returns the current delay without advancing state.
func (t *Throttle) Current() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.current
}
