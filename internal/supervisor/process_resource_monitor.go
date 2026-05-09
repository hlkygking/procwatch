package supervisor

import (
	"os"
	"sync"
	"time"
)

// ResourceSample holds a single snapshot of resource usage for a process.
type ResourceSample struct {
	PID       int
	Name      string
	Timestamp time.Time
	RSSBytes  int64
	CPUPct    float64
}

// ProcessResourceMonitor polls resource usage for tracked processes at a
// configurable interval and retains the most recent sample per process.
type ProcessResourceMonitor struct {
	mu       sync.RWMutex
	samples  map[string]*ResourceSample
	interval time.Duration
	pids     map[string]int
	stop     chan struct{}
}

// NewProcessResourceMonitor creates a monitor that samples at the given interval.
func NewProcessResourceMonitor(interval time.Duration) *ProcessResourceMonitor {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &ProcessResourceMonitor{
		samples:  make(map[string]*ResourceSample),
		pids:     make(map[string]int),
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Track registers a named process PID for monitoring.
func (m *ProcessResourceMonitor) Track(name string, pid int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pids[name] = pid
}

// Untrack removes a process from monitoring.
func (m *ProcessResourceMonitor) Untrack(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pids, name)
}

// Sample returns the latest ResourceSample for a process, or nil if unavailable.
func (m *ProcessResourceMonitor) Sample(name string) *ResourceSample {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.samples[name]
	if s == nil {
		return nil
	}
	copy := *s
	return &copy
}

// Start begins the background polling loop.
func (m *ProcessResourceMonitor) Start() {
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.poll()
			case <-m.stop:
				return
			}
		}
	}()
}

// Stop halts the background polling loop.
func (m *ProcessResourceMonitor) Stop() {
	close(m.stop)
}

func (m *ProcessResourceMonitor) poll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, pid := range m.pids {
		rss := readRSS(pid)
		m.samples[name] = &ResourceSample{
			PID:       pid,
			Name:      name,
			Timestamp: time.Now(),
			RSSBytes:  rss,
			CPUPct:    0, // CPU% requires two samples; placeholder for now
		}
	}
}

// readRSS attempts to read the resident set size for a PID from /proc.
// Returns 0 if unavailable (non-Linux or process gone).
func readRSS(pid int) int64 {
	path := "/proc/" + itoa(pid) + "/statm"
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var pages int64
	for _, b := range data {
		if b == ' ' || b == '\n' {
			break
		}
		if b >= '0' && b <= '9' {
			pages = pages*10 + int64(b-'0')
		}
	}
	return pages * 4096
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
