package supervisor

import (
	"testing"
)

func TestDependencyGraph_EmptyOrder(t *testing.T) {
	g := NewDependencyGraph()
	order, err := g.Order()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 0 {
		t.Errorf("expected empty order, got %v", order)
	}
}

func TestDependencyGraph_LinearOrder(t *testing.T) {
	g := NewDependencyGraph()
	g.Add("web", "db")
	g.Add("worker", "db")

	order, err := g.Order()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pos := make(map[string]int, len(order))
	for i, name := range order {
		pos[name] = i
	}

	if pos["db"] >= pos["web"] {
		t.Errorf("expected db before web in order %v", order)
	}
	if pos["db"] >= pos["worker"] {
		t.Errorf("expected db before worker in order %v", order)
	}
}

func TestDependencyGraph_CycleDetected(t *testing.T) {
	g := NewDependencyGraph()
	g.Add("a", "b")
	g.Add("b", "c")
	g.Add("c", "a")

	_, err := g.Order()
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

func TestDependencyGraph_Validate_AllKnown(t *testing.T) {
	g := NewDependencyGraph()
	g.Add("web", "db")

	known := map[string]struct{}{
		"web": {},
		"db":  {},
	}
	if err := g.Validate(known); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestDependencyGraph_Validate_UnknownProcess(t *testing.T) {
	g := NewDependencyGraph()
	g.Add("web", "db")

	known := map[string]struct{}{
		"web": {},
		// db is missing
	}
	if err := g.Validate(known); err == nil {
		t.Error("expected validation error for unknown process, got nil")
	}
}

func TestDependencyGraph_NoDuplicateNodes(t *testing.T) {
	g := NewDependencyGraph()
	g.Add("web", "db")
	g.Add("web", "cache")

	order, err := g.Order()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	seen := make(map[string]int)
	for _, name := range order {
		seen[name]++
	}
	for name, count := range seen {
		if count > 1 {
			t.Errorf("process %q appears %d times in order", name, count)
		}
	}
}
