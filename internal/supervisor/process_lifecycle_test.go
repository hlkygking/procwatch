package supervisor

import (
	"testing"
)

func TestProcessLifecycle_RecordAndAll(t *testing.T) {
	lc := NewProcessLifecycle("svc")
	lc.Record(LifecycleStarting, "about to start")
	lc.Record(LifecycleStarted, "running")

	all := lc.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 records, got %d", len(all))
	}
	if all[0].Event != LifecycleStarting {
		t.Errorf("expected first event %q, got %q", LifecycleStarting, all[0].Event)
	}
	if all[1].Event != LifecycleStarted {
		t.Errorf("expected second event %q, got %q", LifecycleStarted, all[1].Event)
	}
}

func TestProcessLifecycle_RecordExit_SetsCode(t *testing.T) {
	lc := NewProcessLifecycle("svc")
	lc.RecordExit(LifecycleFailed, 1, "non-zero exit")

	all := lc.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 record, got %d", len(all))
	}
	if all[0].ExitCode == nil || *all[0].ExitCode != 1 {
		t.Errorf("expected exit code 1, got %v", all[0].ExitCode)
	}
}

func TestProcessLifecycle_Last_ReturnsLatest(t *testing.T) {
	lc := NewProcessLifecycle("svc")
	if lc.Last() != nil {
		t.Fatal("expected nil for empty lifecycle")
	}
	lc.Record(LifecycleStarting, "")
	lc.Record(LifecycleStopped, "done")

	last := lc.Last()
	if last == nil || last.Event != LifecycleStopped {
		t.Errorf("expected last event %q, got %v", LifecycleStopped, last)
	}
}

func TestProcessLifecycle_CountOf(t *testing.T) {
	lc := NewProcessLifecycle("svc")
	lc.Record(LifecycleRestarting, "")
	lc.Record(LifecycleRestarting, "")
	lc.Record(LifecycleStarted, "")

	if got := lc.CountOf(LifecycleRestarting); got != 2 {
		t.Errorf("expected 2 restart events, got %d", got)
	}
	if got := lc.CountOf(LifecycleStarted); got != 1 {
		t.Errorf("expected 1 started event, got %d", got)
	}
}

func TestProcessLifecycleStore_GetOrCreate_Idempotent(t *testing.T) {
	store := NewProcessLifecycleStore()
	a := store.GetOrCreate("web")
	b := store.GetOrCreate("web")
	if a != b {
		t.Error("expected same lifecycle instance on repeated GetOrCreate")
	}
}

func TestProcessLifecycleStore_Get_Missing(t *testing.T) {
	store := NewProcessLifecycleStore()
	if store.Get("missing") != nil {
		t.Error("expected nil for unknown process")
	}
}

func TestProcessLifecycleStore_AllRecords(t *testing.T) {
	store := NewProcessLifecycleStore()
	store.GetOrCreate("web").Record(LifecycleStarted, "")
	store.GetOrCreate("worker").Record(LifecycleStarted, "")
	store.GetOrCreate("worker").Record(LifecycleStopped, "")

	all := store.AllRecords()
	if len(all) != 3 {
		t.Errorf("expected 3 total records, got %d", len(all))
	}
}

func TestProcessLifecycleStore_ProcessNames(t *testing.T) {
	store := NewProcessLifecycleStore()
	store.GetOrCreate("alpha")
	store.GetOrCreate("beta")

	names := store.ProcessNames()
	if len(names) != 2 {
		t.Errorf("expected 2 process names, got %d", len(names))
	}
}
