package supervisor

import (
	"testing"
)

func makeTaggedConfigs() []ProcessConfig {
	return []ProcessConfig{
		{Name: "web", Command: "serve", Tags: []string{"http", "frontend"}},
		{Name: "api", Command: "api", Tags: []string{"http", "backend"}},
		{Name: "worker", Command: "work", Tags: []string{"backend", "queue"}},
		{Name: "cron", Command: "cron", Tags: []string{}},
	}
}

func TestTagFilter_Match(t *testing.T) {
	f := NewTagFilter("http")
	cfg := ProcessConfig{Name: "web", Command: "serve", Tags: []string{"http", "frontend"}}
	if !f.Match(cfg) {
		t.Error("expected match for tag 'http'")
	}
}

func TestTagFilter_NoMatch(t *testing.T) {
	f := NewTagFilter("database")
	cfg := ProcessConfig{Name: "web", Command: "serve", Tags: []string{"http", "frontend"}}
	if f.Match(cfg) {
		t.Error("expected no match for tag 'database'")
	}
}

func TestTagFilter_EmptyTags(t *testing.T) {
	f := NewTagFilter("http")
	cfg := ProcessConfig{Name: "cron", Command: "cron", Tags: []string{}}
	if f.Match(cfg) {
		t.Error("expected no match for config with no tags")
	}
}

func TestFilterByTag_ReturnsMatching(t *testing.T) {
	cfgs := makeTaggedConfigs()
	result := FilterByTag(cfgs, "http")
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	names := map[string]bool{}
	for _, c := range result {
		names[c.Name] = true
	}
	if !names["web"] || !names["api"] {
		t.Error("expected 'web' and 'api' in results")
	}
}

func TestFilterByTag_NoMatch(t *testing.T) {
	cfgs := makeTaggedConfigs()
	result := FilterByTag(cfgs, "nonexistent")
	if len(result) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result))
	}
}

func TestFilterByAllTags_AllPresent(t *testing.T) {
	cfgs := makeTaggedConfigs()
	result := FilterByAllTags(cfgs, []string{"http", "backend"})
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Name != "api" {
		t.Errorf("expected 'api', got '%s'", result[0].Name)
	}
}

func TestFilterByAllTags_EmptyTagList(t *testing.T) {
	cfgs := makeTaggedConfigs()
	result := FilterByAllTags(cfgs, []string{})
	if len(result) != len(cfgs) {
		t.Errorf("expected all %d configs, got %d", len(cfgs), len(result))
	}
}

func TestFilterByAllTags_PartialMatch(t *testing.T) {
	cfgs := makeTaggedConfigs()
	result := FilterByAllTags(cfgs, []string{"http", "queue"})
	if len(result) != 0 {
		t.Errorf("expected 0 results for non-overlapping tags, got %d", len(result))
	}
}
