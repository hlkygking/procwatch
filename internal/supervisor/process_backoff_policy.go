package supervisor

import (
	"fmt"
	"time"
)

// BackoffStrategy defines how delay grows between restarts.
type BackoffStrategy string

const (
	BackoffFixed       BackoffStrategy = "fixed"
	BackoffLinear      BackoffStrategy = "linear"
	BackoffExponential BackoffStrategy = "exponential"
)

// ProcessBackoffPolicy defines per-process backoff configuration.
type ProcessBackoffPolicy struct {
	Strategy BackoffStrategy
	BaseDelay time.Duration
	MaxDelay  time.Duration
	Multiplier float64
}

// DefaultBackoffPolicy returns a sensible default exponential backoff policy.
func DefaultBackoffPolicy() ProcessBackoffPolicy {
	return ProcessBackoffPolicy{
		Strategy:   BackoffExponential,
		BaseDelay:  1 * time.Second,
		MaxDelay:   60 * time.Second,
		Multiplier: 2.0,
	}
}

// Delay computes the backoff delay for the given attempt number (1-based).
func (p ProcessBackoffPolicy) Delay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	var d time.Duration
	switch p.Strategy {
	case BackoffFixed:
		d = p.BaseDelay
	case BackoffLinear:
		d = p.BaseDelay * time.Duration(attempt)
	case BackoffExponential:
		d = p.BaseDelay
		for i := 1; i < attempt; i++ {
			d = time.Duration(float64(d) * p.Multiplier)
			if d >= p.MaxDelay {
				return p.MaxDelay
			}
		}
	default:
		d = p.BaseDelay
	}
	if p.MaxDelay > 0 && d > p.MaxDelay {
		return p.MaxDelay
	}
	return d
}

func (p ProcessBackoffPolicy) String() string {
	return fmt.Sprintf("strategy=%s base=%s max=%s multiplier=%.2f",
		p.Strategy, p.BaseDelay, p.MaxDelay, p.Multiplier)
}
