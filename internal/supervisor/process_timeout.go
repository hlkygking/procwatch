package supervisor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ProcessTimeoutPolicy defines how long a process is allowed to run before
// it is considered hung and forcibly terminated.
type ProcessTimeoutPolicy struct {
	MaxRuntime time.Duration // 0 means no timeout
	GracePeriod time.Duration // time between SIGTERM and SIGKILL
}

// ProcessTimeoutStore tracks per-process timeout policies and active timers.
type ProcessTimeoutStore struct {
	mu       sync.Mutex
	policies map[string]ProcessTimeoutPolicy
	timers   map[string]*time.Timer
}

// NewProcessTimeoutStore creates an empty timeout store.
func NewProcessTimeoutStore() *ProcessTimeoutStore {
	return &ProcessTimeoutStore{
		policies: make(map[string]ProcessTimeoutPolicy),
		timers:   make(map[string]*time.Timer),
	}
}

// SetPolicy registers a timeout policy for the named process.
func (s *ProcessTimeoutStore) SetPolicy(name string, p ProcessTimeoutPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[name] = p
}

// GetPolicy returns the timeout policy for the named process and whether it exists.
func (s *ProcessTimeoutStore) GetPolicy(name string) (ProcessTimeoutPolicy, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.policies[name]
	return p, ok
}

// Arm starts a timeout timer for the named process. When the timer fires, the
// provided cancel function is called to terminate the process context. The
// timer is automatically disarmed when the context is already cancelled.
// Arm is a no-op if MaxRuntime is zero.
func (s *ProcessTimeoutStore) Arm(ctx context.Context, name string, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.policies[name]
	if !ok || p.MaxRuntime == 0 {
		return
	}

	// Cancel any existing timer for this process.
	if t, exists := s.timers[name]; exists {
		t.Stop()
	}

	t := time.AfterFunc(p.MaxRuntime, func() {
		select {
		case <-ctx.Done():
			// Already cancelled; nothing to do.
		default:
			cancel()
		}
	})
	s.timers[name] = t
}

// Disarm cancels any active timeout timer for the named process.
func (s *ProcessTimeoutStore) Disarm(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.timers[name]; ok {
		t.Stop()
		delete(s.timers, name)
	}
}

// String returns a human-readable description of the policy.
func (p ProcessTimeoutPolicy) String() string {
	if p.MaxRuntime == 0 {
		return "no timeout"
	}
	return fmt.Sprintf("max=%s grace=%s", p.MaxRuntime, p.GracePeriod)
}
