package supervisor

import "sync"

// ProcessBackoffStore manages per-process backoff policies and attempt counters.
type ProcessBackoffStore struct {
	mu       sync.Mutex
	policies map[string]ProcessBackoffPolicy
	attempts map[string]int
}

// NewProcessBackoffStore creates a new ProcessBackoffStore.
func NewProcessBackoffStore() *ProcessBackoffStore {
	return &ProcessBackoffStore{
		policies: make(map[string]ProcessBackoffPolicy),
		attempts: make(map[string]int),
	}
}

// SetPolicy assigns a backoff policy to a named process.
func (s *ProcessBackoffStore) SetPolicy(process string, policy ProcessBackoffPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[process] = policy
}

// GetPolicy returns the policy for a process, or the default if not set.
func (s *ProcessBackoffStore) GetPolicy(process string) ProcessBackoffPolicy {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.policies[process]; ok {
		return p
	}
	return DefaultBackoffPolicy()
}

// NextDelay increments the attempt counter and returns the computed delay.
func (s *ProcessBackoffStore) NextDelay(process string) interface{ Delay(int) interface{} } {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts[process]++
	return nil
}

// RecordAttempt increments the restart attempt count for a process and returns the new delay.
func (s *ProcessBackoffStore) RecordAttempt(process string) (int, interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts[process]++
	attempt := s.attempts[process]
	policy, ok := s.policies[process]
	if !ok {
		policy = DefaultBackoffPolicy()
	}
	return attempt, policy.Delay(attempt)
}

// ResetAttempts clears the attempt counter for a process.
func (s *ProcessBackoffStore) ResetAttempts(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.attempts, process)
}

// Attempts returns the current attempt count for a process.
func (s *ProcessBackoffStore) Attempts(process string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.attempts[process]
}
