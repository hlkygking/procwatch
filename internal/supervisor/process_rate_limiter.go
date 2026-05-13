package supervisor

import (
	"sync"
	"time"
)

// ProcessRateLimiter tracks per-process event rates and enforces a maximum
// number of events within a sliding time window.
type ProcessRateLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	maxRate int
	events  map[string][]time.Time
}

// NewProcessRateLimiter creates a ProcessRateLimiter with the given sliding
// window duration and maximum allowed events per window.
func NewProcessRateLimiter(window time.Duration, maxRate int) *ProcessRateLimiter {
	return &ProcessRateLimiter{
		window:  window,
		maxRate: maxRate,
		events:  make(map[string][]time.Time),
	}
}

// Allow records an event for the named process and reports whether it is
// within the allowed rate. Events older than the window are pruned on each
// call.
func (r *ProcessRateLimiter) Allow(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	prev := r.events[name]
	filtered := prev[:0]
	for _, t := range prev {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= r.maxRate {
		r.events[name] = filtered
		return false
	}

	r.events[name] = append(filtered, now)
	return true
}

// Count returns the number of events recorded within the current window for
// the named process.
func (r *ProcessRateLimiter) Count(name string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-r.window)
	count := 0
	for _, t := range r.events[name] {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// Reset clears all recorded events for the named process.
func (r *ProcessRateLimiter) Reset(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.events, name)
}
