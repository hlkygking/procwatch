package supervisor

import (
	"testing"
)

func TestProcessStateStore_GetOrCreate(t *testing.T) {
	store := NewProcessStateStore()
	ps := store.GetOrCreate("alpha")
	if ps == nil {
		t.Fatal("expected non-nil ProcessState")
	}
	if store.Len() != 1 {
		t.Errorf("expected len 1, got %d", store.Len())
	}
}

func TestProcessStateStore_GetOrCreate_Idempotent(t *testing.T) {
	store := NewProcessStateStore()
	ps1 := store.GetOrCreate("alpha")
	ps2 := store.GetOrCreate("alpha")
	if ps1 != ps2 {
		t.Error("expected same pointer for same name")
	}
	if store.Len() != 1 {
		t.Errorf("expected len 1 after duplicate GetOrCreate, got %d", store.Len())
	}
}

func TestProcessStateStore_Get_Missing(t *testing.T) {
	store := NewProcessStateStore()
	_, ok := store.Get("missing")
	if ok {
		t.Error("expected not found for missing key")
	}
}

func TestProcessStateStore_Get_Existing(t *testing.T) {
	store := NewProcessStateStore()
	store.GetOrCreate("beta")
	ps, ok := store.Get("beta")
	if !ok {
		t.Fatal("expected to find 'beta'")
	}
	if ps == nil {
		t.Error("expected non-nil ProcessState")
	}
}

func TestProcessStateStore_Snapshots(t *testing.T) {
	store := NewProcessStateStore()
	store.GetOrCreate("svc-a").RecordStart(1)
	store.GetOrCreate("svc-b").RecordStart(2)

	snaps := store.Snapshots()
	if len(snaps) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(snaps))
	}
	names := map[string]bool{}
	for _, s := range snaps {
		names[s.Name] = true
	}
	if !names["svc-a"] || !names["svc-b"] {
		t.Errorf("expected both svc-a and svc-b in snapshots, got %v", names)
	}
}

func TestProcessStateStore_Len(t *testing.T) {
	store := NewProcessStateStore()
	if store.Len() != 0 {
		t.Error("expected empty store")
	}
	store.GetOrCreate("x")
	store.GetOrCreate("y")
	store.GetOrCreate("z")
	if store.Len() != 3 {
		t.Errorf("expected len 3, got %d", store.Len())
	}
}
