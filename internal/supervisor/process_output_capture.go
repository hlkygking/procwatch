package supervisor

import (
	"io"
	"sync"
)

// ProcessOutputCapture captures stdout/stderr output from a process
// into a fixed-size ring buffer per stream, safe for concurrent use.
type ProcessOutputCapture struct {
	mu     sync.RWMutex
	stdout *RingBuffer
	stderr *RingBuffer
}

// NewProcessOutputCapture creates a new output capture with the given buffer size.
// If size <= 0, the default ring buffer size is used.
func NewProcessOutputCapture(size int) *ProcessOutputCapture {
	var stdout, stderr *RingBuffer
	if size > 0 {
		stdout = NewRingBuffer(size)
		stderr = NewRingBuffer(size)
	} else {
		stdout = NewRingBuffer(0)
		stderr = NewRingBuffer(0)
	}
	return &ProcessOutputCapture{
		stdout: stdout,
		stderr: stderr,
	}
}

// StdoutWriter returns an io.Writer that appends lines to the stdout buffer.
func (c *ProcessOutputCapture) StdoutWriter() io.Writer {
	return &captureWriter{capture: c, target: c.stdout}
}

// StderrWriter returns an io.Writer that appends lines to the stderr buffer.
func (c *ProcessOutputCapture) StderrWriter() io.Writer {
	return &captureWriter{capture: c, target: c.stderr}
}

// Stdout returns all captured stdout lines.
func (c *ProcessOutputCapture) Stdout() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stdout.Entries()
}

// Stderr returns all captured stderr lines.
func (c *ProcessOutputCapture) Stderr() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stderr.Entries()
}

// Reset clears both stdout and stderr buffers.
func (c *ProcessOutputCapture) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stdout.Reset()
	c.stderr.Reset()
}

// captureWriter implements io.Writer, writing lines into a RingBuffer.
type captureWriter struct {
	capture *ProcessOutputCapture
	target  *RingBuffer
}

func (w *captureWriter) Write(p []byte) (int, error) {
	w.capture.mu.Lock()
	defer w.capture.mu.Unlock()
	line := string(p)
	// Trim trailing newline for cleaner storage
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	if len(line) > 0 {
		w.target.Write(line)
	}
	return len(p), nil
}
