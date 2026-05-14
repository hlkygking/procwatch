package supervisor

import (
	"sync"
	"time"
)

// CooldownPolicy defines how long a process must remain stable before its
// restart counter is considered eligible for a reset.
type CooldownPolicy struct {
	StableDuration time.Duration // how long a process must run without crashing
}

// DefaultCooldownPolicy returns a sensible default cooldown policy.
func DefaultCooldownPolicy() CooldownPolicy {
	return CooldownPolicy{
		StableDuration: 30 * time.Second,
	}
}

// ProcessCooldown tracks per-process cooldown state, recording when a process
// last started and whether it has satisfied the stable duration threshold.
type ProcessCooldown struct {
	mu      sync.Mutex
	policy  CooldownPolicy
	starts  map[string]time.Time
	cooled  map[string]bool
	nowFunc func() time.Time
}

// NewProcessCooldown creates a new ProcessCooldown with the given policy.
func NewProcessCooldown(policy CooldownPolicy) *ProcessCooldown {
	return &ProcessCooldown{
		policy:  policy,
		starts:  make(map[string]time.Time),
		cooled:  make(map[string]bool),
		nowFunc: time.Now,
	}
}

// RecordStart marks the process as having just started, resetting its cooldown state.
func (c *ProcessCooldown) RecordStart(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.starts[name] = c.nowFunc()
	c.cooled[name] = false
}

// RecordStable checks whether the process has been running long enough to be
// considered stable. If so, it marks the process as cooled down and returns true.
func (c *ProcessCooldown) RecordStable(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	start, ok := c.starts[name]
	if !ok {
		return false
	}
	if c.nowFunc().Sub(start) >= c.policy.StableDuration {
		c.cooled[name] = true
		return true
	}
	return false
}

// IsCooled returns true if the process has satisfied the stable duration threshold
// since its last recorded start.
func (c *ProcessCooldown) IsCooled(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cooled[name]
}

// Reset clears all cooldown state for a process.
func (c *ProcessCooldown) Reset(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.starts, name)
	delete(c.cooled, name)
}
