package supervisor

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// SignalPolicy defines which OS signal to send to a process on shutdown or restart.
type SignalPolicy struct {
	ShutdownSignal os.Signal
	RestartSignal  os.Signal
}

// DefaultSignalPolicy returns a policy using SIGTERM for shutdown and SIGHUP for restart.
func DefaultSignalPolicy() SignalPolicy {
	return SignalPolicy{
		ShutdownSignal: os.Interrupt, // SIGINT / SIGTERM equivalent on all platforms
		RestartSignal:  os.Interrupt,
	}
}

// ProcessSignalPolicyStore stores per-process signal policies.
type ProcessSignalPolicyStore struct {
	mu       sync.RWMutex
	policies map[string]SignalPolicy
}

// NewProcessSignalPolicyStore creates an empty store.
func NewProcessSignalPolicyStore() *ProcessSignalPolicyStore {
	return &ProcessSignalPolicyStore{
		policies: make(map[string]SignalPolicy),
	}
}

// Set assigns a signal policy to a named process.
func (s *ProcessSignalPolicyStore) Set(name string, policy SignalPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[name] = policy
}

// Get returns the signal policy for a process, or the default if none is set.
func (s *ProcessSignalPolicyStore) Get(name string) SignalPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.policies[name]; ok {
		return p
	}
	return DefaultSignalPolicy()
}

// ParseSignalName converts a string like "SIGTERM" or "SIGHUP" to an os.Signal.
// Only SIGINT and SIGTERM are supported cross-platform in the standard library.
func ParseSignalName(name string) (os.Signal, error) {
	switch strings.ToUpper(strings.TrimPrefix(strings.ToUpper(name), "SIG")) {
	case "INT", "SIGINT":
		return os.Interrupt, nil
	case "TERM", "SIGTERM":
		return os.Interrupt, nil // mapped to os.Interrupt for portability
	default:
		return nil, fmt.Errorf("unsupported signal: %q (supported: SIGINT, SIGTERM)", name)
	}
}

// All returns a snapshot of all stored process signal policies.
func (s *ProcessSignalPolicyStore) All() map[string]SignalPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]SignalPolicy, len(s.policies))
	for k, v := range s.policies {
		out[k] = v
	}
	return out
}
