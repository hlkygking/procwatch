package supervisor

import (
	"context"
	"testing"
	"time"
)

func makeSnapshot(name, status string, restarts int) ProcessSnapshot {
	return ProcessSnapshot{Name: name, Status: status, Restarts: restarts}
}

func TestProcessSnapshotCollector_CollectsOnTick(t *testing.T) {
	store := NewProcessSnapshotStore()
	called := 0
	source := func() []ProcessSnapshot {
		called++
		return []ProcessSnapshot{makeSnapshot("api", "running", 0)}
	}

	collector := NewProcessSnapshotCollector(store, source, 20*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Millisecond)
	defer cancel()

	collector.Run(ctx)

	if called < 2 {
		t.Errorf("expected at least 2 collections, got %d", called)
	}
	snaps := store.ForProcess("api")
	if len(snaps) < 2 {
		t.Errorf("expected at least 2 snapshots stored, got %d", len(snaps))
	}
}

func TestProcessSnapshotCollector_StopsOnContextCancel(t *testing.T) {
	store := NewProcessSnapshotStore()
	called := 0
	source := func() []ProcessSnapshot {
		called++
		return []ProcessSnapshot{makeSnapshot("svc", "running", 0)}
	}

	collector := NewProcessSnapshotCollector(store, source, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		collector.Run(ctx)
		close(done)
	}()

	time.Sleep(35 * time.Millisecond)
	cancel()
	select {
	case <-done:
		// ok
	case <-time.After(200 * time.Millisecond):
		t.Error("collector did not stop after context cancel")
	}
}

func TestProcessSnapshotCollector_DefaultInterval(t *testing.T) {
	store := NewProcessSnapshotStore()
	source := func() []ProcessSnapshot { return nil }
	collector := NewProcessSnapshotCollector(store, source, 0)
	if collector.interval != 30*time.Second {
		t.Errorf("expected default interval 30s, got %v", collector.interval)
	}
}

func TestProcessSnapshotCollector_MultipleProcesses(t *testing.T) {
	store := NewProcessSnapshotStore()
	source := func() []ProcessSnapshot {
		return []ProcessSnapshot{
			makeSnapshot("web", "running", 1),
			makeSnapshot("db", "running", 0),
		}
	}

	collector := NewProcessSnapshotCollector(store, source, 15*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	collector.Run(ctx)

	for _, name := range []string{"web", "db"} {
		if snaps := store.ForProcess(name); len(snaps) == 0 {
			t.Errorf("expected snapshots for %s, got none", name)
		}
	}
}
