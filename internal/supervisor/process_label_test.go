package supervisor

import (
	"testing"
)

func TestLabelSet_SetAndGet(t *testing.T) {
	ls := NewLabelSet("worker")
	ls.Set("env", "prod")
	v, ok := ls.Get("env")
	if !ok || v != "prod" {
		t.Fatalf("expected env=prod, got %q ok=%v", v, ok)
	}
}

func TestLabelSet_GetMissing(t *testing.T) {
	ls := NewLabelSet("worker")
	_, ok := ls.Get("missing")
	if ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestLabelSet_Keys_Sorted(t *testing.T) {
	ls := NewLabelSet("worker")
	ls.Set("z", "1")
	ls.Set("a", "2")
	ls.Set("m", "3")
	keys := ls.Keys()
	if keys[0] != "a" || keys[1] != "m" || keys[2] != "z" {
		t.Fatalf("unexpected key order: %v", keys)
	}
}

func TestLabelSet_Matches_AllPresent(t *testing.T) {
	ls := NewLabelSet("svc")
	ls.Set("env", "prod")
	ls.Set("region", "us-east")
	if !ls.Matches(map[string]string{"env": "prod", "region": "us-east"}) {
		t.Fatal("expected match")
	}
}

func TestLabelSet_Matches_PartialMismatch(t *testing.T) {
	ls := NewLabelSet("svc")
	ls.Set("env", "prod")
	if ls.Matches(map[string]string{"env": "staging"}) {
		t.Fatal("expected no match")
	}
}

func TestLabelSet_Clone_IsIndependent(t *testing.T) {
	ls := NewLabelSet("svc")
	ls.Set("k", "v")
	clone := ls.Clone()
	clone.Set("k", "changed")
	if v, _ := ls.Get("k"); v != "v" {
		t.Fatal("original label set was mutated by clone")
	}
}

func TestProcessLabelStore_GetOrCreate_Idempotent(t *testing.T) {
	s := NewProcessLabelStore()
	a := s.GetOrCreate("svc")
	b := s.GetOrCreate("svc")
	if a != b {
		t.Fatal("expected same pointer for same process")
	}
}

func TestProcessLabelStore_SelectByLabels(t *testing.T) {
	s := NewProcessLabelStore()
	ls1 := s.GetOrCreate("web")
	ls1.Set("env", "prod")
	ls2 := s.GetOrCreate("worker")
	ls2.Set("env", "staging")

	results := s.SelectByLabels(map[string]string{"env": "prod"})
	if len(results) != 1 || results[0].Process != "web" {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestProcessLabelStore_All(t *testing.T) {
	s := NewProcessLabelStore()
	s.GetOrCreate("a")
	s.GetOrCreate("b")
	if len(s.All()) != 2 {
		t.Fatal("expected 2 label sets")
	}
}
