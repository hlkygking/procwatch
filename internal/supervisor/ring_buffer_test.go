package supervisor

import (
	"fmt"
	"testing"
)

func TestRingBuffer_BasicWrite(t *testing.T) {
	rb := NewRingBuffer(5)
	_, _ = rb.Write([]byte("line1"))
	_, _ = rb.Write([]byte("line2"))

	if rb.Len() != 2 {
		t.Fatalf("expected len 2, got %d", rb.Len())
	}
	lines := rb.Lines()
	if lines[0] != "line1" || lines[1] != "line2" {
		t.Errorf("unexpected lines: %v", lines)
	}
}

func TestRingBuffer_Overflow(t *testing.T) {
	rb := NewRingBuffer(3)
	for i := 1; i <= 5; i++ {
		_, _ = rb.Write([]byte(fmt.Sprintf("line%d", i)))
	}
	if rb.Len() != 3 {
		t.Fatalf("expected len 3, got %d", rb.Len())
	}
	lines := rb.Lines()
	// Oldest entries should have been overwritten; last 3 remain.
	if lines[0] != "line3" || lines[2] != "line5" {
		t.Errorf("unexpected overflow lines: %v", lines)
	}
}

func TestRingBuffer_Reset(t *testing.T) {
	rb := NewRingBuffer(4)
	_, _ = rb.Write([]byte("a"))
	rb.Reset()
	if rb.Len() != 0 {
		t.Errorf("expected len 0 after reset, got %d", rb.Len())
	}
	if len(rb.Lines()) != 0 {
		t.Error("expected empty lines after reset")
	}
}

func TestRingBuffer_DefaultSize(t *testing.T) {
	rb := NewRingBuffer(0)
	if rb.size != 100 {
		t.Errorf("expected default size 100, got %d", rb.size)
	}
}

func TestRingBuffer_SingleEntry(t *testing.T) {
	rb := NewRingBuffer(1)
	_, _ = rb.Write([]byte("only"))
	_, _ = rb.Write([]byte("latest"))
	lines := rb.Lines()
	if len(lines) != 1 || lines[0] != "latest" {
		t.Errorf("expected [latest], got %v", lines)
	}
}
