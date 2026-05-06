package supervisor

import (
	"testing"
)

func makeConfigs(names ...string) []ProcessConfig {
	out := make([]ProcessConfig, len(names))
	for i, n := range names {
		out[i] = ProcessConfig{Name: n, Command: "true"}
	}
	return out
}

func TestProcessFilter_Exact(t *testing.T) {
	cfgs := makeConfigs("api", "worker", "scheduler")
	f := NewExactFilter("worker")
	result := f.Apply(cfgs)
	if len(result) != 1 || result[0].Name != "worker" {
		t.Fatalf("expected [worker], got %v", result)
	}
}

func TestProcessFilter_ExactNoMatch(t *testing.T) {
	cfgs := makeConfigs("api", "worker")
	f := NewExactFilter("missing")
	result := f.Apply(cfgs)
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestProcessFilter_Prefix(t *testing.T) {
	cfgs := makeConfigs("worker-a", "worker-b", "api", "worker-c")
	f := NewPrefixFilter("worker")
	result := f.Apply(cfgs)
	if len(result) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(result))
	}
	for _, r := range result {
		if r.Name == "api" {
			t.Fatalf("api should not be included")
		}
	}
}

func TestProcessFilter_All(t *testing.T) {
	cfgs := makeConfigs("a", "b", "c")
	f := NewAllFilter()
	result := f.Apply(cfgs)
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
}

func TestProcessFilter_Match(t *testing.T) {
	cfg := ProcessConfig{Name: "api-server", Command: "./api"}

	if !NewExactFilter("api-server").Match(cfg) {
		t.Error("exact filter should match")
	}
	if NewExactFilter("api").Match(cfg) {
		t.Error("exact filter should not match partial name")
	}
	if !NewPrefixFilter("api").Match(cfg) {
		t.Error("prefix filter should match")
	}
	if !NewAllFilter().Match(cfg) {
		t.Error("all filter should always match")
	}
}

func TestProcessFilter_ApplyNilSafe(t *testing.T) {
	f := NewAllFilter()
	result := f.Apply(nil)
	if result != nil && len(result) != 0 {
		t.Fatalf("expected nil or empty, got %v", result)
	}
}
