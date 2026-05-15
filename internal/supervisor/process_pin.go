package supervisor

import (
	"fmt"
	"sync"
)

// PinMode controls how a process is pinned to a CPU or NUMA node.
type PinMode string

const (
	PinModeNone PinMode = "none"
	PinModeCPU  PinMode = "cpu"
	PinModeNUMA PinMode = "numa"
)

// ProcessPinPolicy describes the affinity pinning policy for a process.
type ProcessPinPolicy struct {
	Mode   PinMode
	Target int // CPU index or NUMA node index
}

// DefaultPinPolicy returns a policy that performs no pinning.
func DefaultPinPolicy() ProcessPinPolicy {
	return ProcessPinPolicy{Mode: PinModeNone, Target: 0}
}

// String returns a human-readable representation of the pin policy.
func (p ProcessPinPolicy) String() string {
	if p.Mode == PinModeNone {
		return "none"
	}
	return fmt.Sprintf("%s:%d", p.Mode, p.Target)
}

// ProcessPinStore stores per-process CPU/NUMA pin policies.
type ProcessPinStore struct {
	mu       sync.RWMutex
	policies map[string]ProcessPinPolicy
}

// NewProcessPinStore creates a new ProcessPinStore.
func NewProcessPinStore() *ProcessPinStore {
	return &ProcessPinStore{
		policies: make(map[string]ProcessPinPolicy),
	}
}

// Set assigns a pin policy to the named process.
func (s *ProcessPinStore) Set(process string, policy ProcessPinPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[process] = policy
}

// Get returns the pin policy for the named process, or the default if not set.
func (s *ProcessPinStore) Get(process string) ProcessPinPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.policies[process]; ok {
		return p
	}
	return DefaultPinPolicy()
}

// All returns a snapshot of all process pin policies.
func (s *ProcessPinStore) All() map[string]ProcessPinPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]ProcessPinPolicy, len(s.policies))
	for k, v := range s.policies {
		out[k] = v
	}
	return out
}

// Remove deletes the pin policy for the named process.
func (s *ProcessPinStore) Remove(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.policies, process)
}

// ParsePinMode converts a string to a PinMode, returning an error if unknown.
func ParsePinMode(s string) (PinMode, error) {
	switch PinMode(s) {
	case PinModeNone, PinModeCPU, PinModeNUMA:
		return PinMode(s), nil
	default:
		return PinModeNone, fmt.Errorf("unknown pin mode: %q", s)
	}
}
