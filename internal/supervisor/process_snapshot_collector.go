package supervisor

import (
	"context"
	"time"
)

// SnapshotSource provides the current state of a named process.
type SnapshotSource interface {
	Snapshots() []ProcessSnapshot
}

// ProcessSnapshotCollector periodically samples process state and records snapshots.
type ProcessSnapshotCollector struct {
	store    *ProcessSnapshotStore
	source   func() []ProcessSnapshot
	interval time.Duration
}

// NewProcessSnapshotCollector creates a collector that samples via source every interval.
func NewProcessSnapshotCollector(
	store *ProcessSnapshotStore,
	source func() []ProcessSnapshot,
	interval time.Duration,
) *ProcessSnapshotCollector {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &ProcessSnapshotCollector{
		store:    store,
		source:   source,
		interval: interval,
	}
}

// Run starts periodic collection until ctx is cancelled.
func (c *ProcessSnapshotCollector) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

// collect samples the source once and stores all returned snapshots.
func (c *ProcessSnapshotCollector) collect() {
	for _, snap := range c.source() {
		snap.TakenAt = time.Now()
		c.store.Record(snap)
	}
}
