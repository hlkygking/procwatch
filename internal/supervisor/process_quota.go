package supervisor

import (
	"fmt"
	"sync"
	"time"
)

// ProcessQuota defines resource usage limits for a process.
type ProcessQuota struct {
	MaxRestarts   int           // 0 = unlimited
	MaxUptime     time.Duration // 0 = unlimited
	MaxCrashRate  float64       // crashes per minute, 0 = unlimited
}

// ProcessQuotaViolation describes why a quota was exceeded.
type ProcessQuotaViolation struct {
	Process   string
	Reason    string
	At        time.Time
}

func (v ProcessQuotaViolation) String() string {
	return fmt.Sprintf("quota violation [%s]: %s at %s", v.Process, v.Reason, v.At.Format(time.RFC3339))
}

// ProcessQuotaStore tracks quotas and violations per process.
type ProcessQuotaStore struct {
	mu         sync.Mutex
	quotas     map[string]ProcessQuota
	violations []ProcessQuotaViolation
}

func NewProcessQuotaStore() *ProcessQuotaStore {
	return &ProcessQuotaStore{
		quotas: make(map[string]ProcessQuota),
	}
}

func (s *ProcessQuotaStore) Set(process string, q ProcessQuota) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.quotas[process] = q
}

func (s *ProcessQuotaStore) Get(process string) (ProcessQuota, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	q, ok := s.quotas[process]
	return q, ok
}

// CheckRestarts returns a violation if the restart count exceeds the quota.
func (s *ProcessQuotaStore) CheckRestarts(process string, restarts int) *ProcessQuotaViolation {
	s.mu.Lock()
	defer s.mu.Unlock()
	q, ok := s.quotas[process]
	if !ok || q.MaxRestarts == 0 {
		return nil
	}
	if restarts > q.MaxRestarts {
		v := ProcessQuotaViolation{
			Process: process,
			Reason:  fmt.Sprintf("restart count %d exceeds max %d", restarts, q.MaxRestarts),
			At:      time.Now(),
		}
		s.violations = append(s.violations, v)
		return &v
	}
	return nil
}

// CheckUptime returns a violation if the uptime exceeds the quota.
func (s *ProcessQuotaStore) CheckUptime(process string, uptime time.Duration) *ProcessQuotaViolation {
	s.mu.Lock()
	defer s.mu.Unlock()
	q, ok := s.quotas[process]
	if !ok || q.MaxUptime == 0 {
		return nil
	}
	if uptime > q.MaxUptime {
		v := ProcessQuotaViolation{
			Process: process,
			Reason:  fmt.Sprintf("uptime %s exceeds max %s", uptime.Round(time.Second), q.MaxUptime),
			At:      time.Now(),
		}
		s.violations = append(s.violations, v)
		return &v
	}
	return nil
}

// Violations returns all recorded quota violations.
func (s *ProcessQuotaStore) Violations() []ProcessQuotaViolation {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ProcessQuotaViolation, len(s.violations))
	copy(out, s.violations)
	return out
}

// ViolationsFor returns violations for a specific process.
func (s *ProcessQuotaStore) ViolationsFor(process string) []ProcessQuotaViolation {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []ProcessQuotaViolation
	for _, v := range s.violations {
		if v.Process == process {
			out = append(out, v)
		}
	}
	return out
}
