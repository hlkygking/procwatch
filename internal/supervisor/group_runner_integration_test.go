package supervisor

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestGroupRunner_HealthTrackedForAll(t *testing.T) {
	configs := []ProcessConfig{
		{Name: "proc-a", Command: "echo", Args: []string{"a"}, Restart: "never"},
		{Name: "proc-b", Command: "echo", Args: []string{"b"}, Restart: "never"},
	}
	gr, health := makeGroupRunner(t, configs)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	gr.Run(ctx)

	all := health.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 health entries, got %d", len(all))
	}
}

func TestGroupRunner_LogsProcessNames(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	health := NewHealthTracker()
	configs := []ProcessConfig{
		{Name: "named-proc", Command: "echo", Args: []string{"ok"}, Restart: "never"},
	}
	gr, err := NewGroupRunner(configs, logger, health)
	if err != nil {
		t.Fatalf("NewGroupRunner: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	gr.Run(ctx)

	output := buf.String()
	if !strings.Contains(output, "named-proc") {
		t.Errorf("expected log output to contain process name, got: %s", output)
	}
}
