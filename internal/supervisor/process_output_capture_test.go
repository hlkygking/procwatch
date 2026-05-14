package supervisor

import (
	"fmt"
	"testing"
)

func TestProcessOutputCapture_StdoutCapture(t *testing.T) {
	c := NewProcessOutputCapture(10)
	w := c.StdoutWriter()

	fmt.Fprintln(w, "hello stdout")
	fmt.Fprintln(w, "second line")

	lines := c.Stdout()
	if len(lines) != 2 {
		t.Fatalf("expected 2 stdout lines, got %d", len(lines))
	}
	if lines[0] != "hello stdout" {
		t.Errorf("unexpected line[0]: %q", lines[0])
	}
	if lines[1] != "second line" {
		t.Errorf("unexpected line[1]: %q", lines[1])
	}
}

func TestProcessOutputCapture_StderrCapture(t *testing.T) {
	c := NewProcessOutputCapture(10)
	w := c.StderrWriter()

	fmt.Fprintln(w, "error occurred")

	lines := c.Stderr()
	if len(lines) != 1 {
		t.Fatalf("expected 1 stderr line, got %d", len(lines))
	}
	if lines[0] != "error occurred" {
		t.Errorf("unexpected stderr line: %q", lines[0])
	}
}

func TestProcessOutputCapture_StdoutAndStderrAreIndependent(t *testing.T) {
	c := NewProcessOutputCapture(10)

	fmt.Fprintln(c.StdoutWriter(), "out line")
	fmt.Fprintln(c.StderrWriter(), "err line")

	if len(c.Stdout()) != 1 {
		t.Errorf("expected 1 stdout line, got %d", len(c.Stdout()))
	}
	if len(c.Stderr()) != 1 {
		t.Errorf("expected 1 stderr line, got %d", len(c.Stderr()))
	}
}

func TestProcessOutputCapture_Reset(t *testing.T) {
	c := NewProcessOutputCapture(10)

	fmt.Fprintln(c.StdoutWriter(), "before reset")
	fmt.Fprintln(c.StderrWriter(), "before reset err")

	c.Reset()

	if len(c.Stdout()) != 0 {
		t.Errorf("expected stdout to be empty after reset")
	}
	if len(c.Stderr()) != 0 {
		t.Errorf("expected stderr to be empty after reset")
	}
}

func TestProcessOutputCapture_DefaultSize(t *testing.T) {
	c := NewProcessOutputCapture(0)
	if c.stdout == nil || c.stderr == nil {
		t.Fatal("expected non-nil ring buffers with default size")
	}
}

func TestProcessOutputCapture_OverflowKeepsLatest(t *testing.T) {
	c := NewProcessOutputCapture(3)
	w := c.StdoutWriter()

	for i := 0; i < 5; i++ {
		fmt.Fprintf(w, "line %d\n", i)
	}

	lines := c.Stdout()
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (ring buffer size), got %d", len(lines))
	}
	if lines[len(lines)-1] != "line 4" {
		t.Errorf("expected last line to be 'line 4', got %q", lines[len(lines)-1])
	}
}

func TestProcessOutputCapture_ResetAllowsReuse(t *testing.T) {
	c := NewProcessOutputCapture(10)

	fmt.Fprintln(c.StdoutWriter(), "before reset")
	c.Reset()

	// Write new lines after reset and verify only new content is present.
	fmt.Fprintln(c.StdoutWriter(), "after reset")

	lines := c.Stdout()
	if len(lines) != 1 {
		t.Fatalf("expected 1 stdout line after reset and reuse, got %d", len(lines))
	}
	if lines[0] != "after reset" {
		t.Errorf("expected 'after reset', got %q", lines[0])
	}
}
