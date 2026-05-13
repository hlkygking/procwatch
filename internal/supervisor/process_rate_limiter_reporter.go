package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"time"
)

// ProcessRateLimiterReporter periodically logs rate-limit statistics for all
// tracked processes using the shared ProcessRateLimiter.
type ProcessRateLimiterReporter struct {
	mu       sync.Mutex
	limiter  *ProcessRateLimiter
	logger   *Logger
	interval time.Duration
	names    []string
	stop     chan struct{}
	w        io.Writer
}

// NewProcessRateLimiterReporter creates a reporter that emits JSON rate-limit
// snapshots at the given interval. If w is nil, os.Stdout is used.
func NewProcessRateLimiterReporter(limiter *ProcessRateLimiter, logger *Logger, interval time.Duration, w io.Writer) *ProcessRateLimiterReporter {
	if w == nil {
		w = os.Stdout
	}
	return &ProcessRateLimiterReporter{
		limiter:  limiter,
		logger:   logger,
		interval: interval,
		stop:     make(chan struct{}),
		w:        w,
	}
}

// Track registers a process name to be included in reports.
func (r *ProcessRateLimiterReporter) Track(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, n := range r.names {
		if n == name {
			return
		}
	}
	r.names = append(r.names, name)
	sort.Strings(r.names)
}

// Start begins the background reporting loop.
func (r *ProcessRateLimiterReporter) Start() {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				r.report()
			case <-r.stop:
				return
			}
		}
	}()
}

// Stop halts the background reporting loop.
func (r *ProcessRateLimiterReporter) Stop() {
	close(r.stop)
}

func (r *ProcessRateLimiterReporter) report() {
	r.mu.Lock()
	names := make([]string, len(r.names))
	copy(names, r.names)
	r.mu.Unlock()

	type entry struct {
		Process string `json:"process"`
		Count   int    `json:"event_count"`
	}

	var entries []entry
	for _, name := range names {
		entries = append(entries, entry{Process: name, Count: r.limiter.Count(name)})
	}

	b, err := json.Marshal(map[string]interface{}{
		"kind":    "rate_limit_report",
		"time":    time.Now().UTC().Format(time.RFC3339),
		"entries": entries,
	})
	if err != nil {
		return
	}
	fmt.Fprintln(r.w, string(b))
}
