package supervisor

import (
	"fmt"
	"time"
)

// RetryBehavior defines what action to take after a process exits.
type RetryBehavior int

const (
	RetryNever   RetryBehavior = iota // Never retry regardless of exit code
	RetryAlways                        // Always retry regardless of exit code
	RetryOnError                       // Retry only on non-zero exit code
)

// ProcessRetryPolicy defines per-process retry configuration.
type ProcessRetryPolicy struct {
	Behavior    RetryBehavior
	MaxAttempts int           // 0 means unlimited
	Delay       time.Duration // base delay before retry
	MaxDelay    time.Duration // cap on delay after backoff
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() ProcessRetryPolicy {
	return ProcessRetryPolicy{
		Behavior:    RetryOnError,
		MaxAttempts: 5,
		Delay:       500 * time.Millisecond,
		MaxDelay:    30 * time.Second,
	}
}

// ShouldRetry returns true if another attempt is permitted given the
// exit code and number of attempts already made.
func (p ProcessRetryPolicy) ShouldRetry(exitCode, attempts int) bool {
	if p.MaxAttempts > 0 && attempts >= p.MaxAttempts {
		return false
	}
	switch p.Behavior {
	case RetryNever:
		return false
	case RetryAlways:
		return true
	case RetryOnError:
		return exitCode != 0
	}
	return false
}

// NextDelay returns the delay before the nth retry attempt using
// exponential backoff capped at MaxDelay.
func (p ProcessRetryPolicy) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return p.Delay
	}
	delay := p.Delay
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay > p.MaxDelay {
			return p.MaxDelay
		}
	}
	return delay
}

// String returns a human-readable description of the policy.
func (p ProcessRetryPolicy) String() string {
	behavior := "on-error"
	switch p.Behavior {
	case RetryNever:
		behavior = "never"
	case RetryAlways:
		behavior = "always"
	}
	max := fmt.Sprintf("%d", p.MaxAttempts)
	if p.MaxAttempts == 0 {
		max = "unlimited"
	}
	return fmt.Sprintf("RetryPolicy{behavior=%s, maxAttempts=%s, delay=%s, maxDelay=%s}",
		behavior, max, p.Delay, p.MaxDelay)
}
