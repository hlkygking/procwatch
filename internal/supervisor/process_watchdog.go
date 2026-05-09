package supervisor

import (
	"context"
	"sync"
	"time"
)

// WatchdogConfig holds configuration for a process watchdog.
type WatchdogConfig struct {
	ProcessName string
	Interval    time.Duration
	MaxMissed   int
	PingFn      func(ctx context.Context) bool
	OnDead      func(name string, missed int)
}

// ProcessWatchdog monitors a process by periodically calling a ping function
// and triggering a callback when the process is considered unresponsive.
type ProcessWatchdog struct {
	cfg    WatchdogConfig
	mu     sync.Mutex
	missed int
	running bool
}

// NewProcessWatchdog creates a new ProcessWatchdog with the given config.
func NewProcessWatchdog(cfg WatchdogConfig) *ProcessWatchdog {
	if cfg.Interval <= 0 {
		cfg.Interval = 5 * time.Second
	}
	if cfg.MaxMissed <= 0 {
		cfg.MaxMissed = 3
	}
	return &ProcessWatchdog{cfg: cfg}
}

// Start begins the watchdog loop, blocking until ctx is cancelled.
func (w *ProcessWatchdog) Start(ctx context.Context) {
	w.mu.Lock()
	w.running = true
	w.mu.Unlock()

	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.mu.Lock()
			w.running = false
			w.mu.Unlock()
			return
		case <-ticker.C:
			w.check(ctx)
		}
	}
}

// Reset clears the missed-ping counter, signalling the process is alive.
func (w *ProcessWatchdog) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.missed = 0
}

// MissedCount returns the current number of consecutive missed pings.
func (w *ProcessWatchdog) MissedCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.missed
}

func (w *ProcessWatchdog) check(ctx context.Context) {
	alive := w.cfg.PingFn(ctx)
	w.mu.Lock()
	defer w.mu.Unlock()
	if alive {
		w.missed = 0
		return
	}
	w.missed++
	if w.missed >= w.cfg.MaxMissed && w.cfg.OnDead != nil {
		w.cfg.OnDead(w.cfg.ProcessName, w.missed)
	}
}
