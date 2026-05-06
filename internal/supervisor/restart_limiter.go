package supervisor

import (
	"sync"
	"time"
)

// RestartLimiter tracks restart attempts and enforces backoff and max-restart limits.
type RestartLimiter struct {
	mu          sync.Mutex
	maxRestarts int
	window      time.Duration
	backoff     time.Duration
	maxBackoff  time.Duration
	attempts    []time.Time
}

// NewRestartLimiter creates a RestartLimiter with the given constraints.
// maxRestarts <= 0 means unlimited. window is the rolling time window for counting restarts.
func NewRestartLimiter(maxRestarts int, window, backoff, maxBackoff time.Duration) *RestartLimiter {
	return &RestartLimiter{
		maxRestarts: maxRestarts,
		window:      window,
		backoff:     backoff,
		maxBackoff:  maxBackoff,
	}
}

// Allow reports whether another restart is permitted right now.
func (r *RestartLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.maxRestarts <= 0 {
		return true
	}
	r.prune()
	return len(r.attempts) < r.maxRestarts
}

// Record registers a restart attempt at the current time.
func (r *RestartLimiter) Record() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.attempts = append(r.attempts, time.Now())
}

// Backoff returns the current backoff duration, capped at maxBackoff.
// Each call doubles the backoff up to the cap.
func (r *RestartLimiter) Backoff() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.backoff <= 0 {
		return 0
	}
	n := len(r.attempts)
	if n == 0 {
		return r.backoff
	}
	delay := r.backoff
	for i := 1; i < n; i++ {
		delay *= 2
		if r.maxBackoff > 0 && delay > r.maxBackoff {
			return r.maxBackoff
		}
	}
	if r.maxBackoff > 0 && delay > r.maxBackoff {
		return r.maxBackoff
	}
	return delay
}

// Reset clears all recorded attempts.
func (r *RestartLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.attempts = r.attempts[:0]
}

// prune removes attempts outside the rolling window. Must be called with lock held.
func (r *RestartLimiter) prune() {
	if r.window <= 0 {
		return
	}
	cutoff := time.Now().Add(-r.window)
	i := 0
	for i < len(r.attempts) && r.attempts[i].Before(cutoff) {
		i++
	}
	r.attempts = r.attempts[i:]
}
