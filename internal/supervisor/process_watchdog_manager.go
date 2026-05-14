package supervisor

import (
	"context"
	"fmt"
	"sync"
)

// ProcessWatchdogManager tracks and manages watchdogs for multiple processes.
type ProcessWatchdogManager struct {
	mu       sync.Mutex
	dogs     map[string]*ProcessWatchdog
	cancels  map[string]context.CancelFunc
	logger   *Logger
}

// NewProcessWatchdogManager creates a new manager using the given logger.
func NewProcessWatchdogManager(logger *Logger) *ProcessWatchdogManager {
	return &ProcessWatchdogManager{
		dogs:    make(map[string]*ProcessWatchdog),
		cancels: make(map[string]context.CancelFunc),
		logger:  logger,
	}
}

// Register adds and starts a watchdog for the named process.
// If a watchdog already exists for that name, it is stopped first.
func (m *ProcessWatchdogManager) Register(ctx context.Context, cfg WatchdogConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cancel, ok := m.cancels[cfg.ProcessName]; ok {
		cancel()
	}

	watchCtx, cancel := context.WithCancel(ctx)
	wdg := NewProcessWatchdog(cfg)
	m.dogs[cfg.ProcessName] = wdg
	m.cancels[cfg.ProcessName] = cancel

	go func() {
		m.logger.Info(fmt.Sprintf("watchdog started for %s", cfg.ProcessName))
		wdg.Start(watchCtx)
		m.logger.Info(fmt.Sprintf("watchdog stopped for %s", cfg.ProcessName))
	}()
}

// Unregister stops and removes the watchdog for the named process.
func (m *ProcessWatchdogManager) Unregister(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cancel, ok := m.cancels[name]; ok {
		cancel()
		delete(m.cancels, name)
		delete(m.dogs, name)
	}
}

// Reset resets the missed-ping counter for the named process, if tracked.
func (m *ProcessWatchdogManager) Reset(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if wdg, ok := m.dogs[name]; ok {
		wdg.Reset()
	}
}

// Names returns the list of currently tracked process names.
func (m *ProcessWatchdogManager) Names() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	names := make([]string, 0, len(m.dogs))
	for name := range m.dogs {
		names = append(names, name)
	}
	return names
}

// Has returns true if a watchdog is currently registered for the given process name.
func (m *ProcessWatchdogManager) Has(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.dogs[name]
	return ok
}
