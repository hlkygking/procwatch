package supervisor

import "sync"

// RingBuffer is a fixed-capacity, thread-safe circular buffer of log lines.
type RingBuffer struct {
	mu   sync.Mutex
	buf  []string
	size int
	head int
	count int
}

// NewRingBuffer creates a RingBuffer with the given capacity.
func NewRingBuffer(size int) *RingBuffer {
	if size <= 0 {
		size = 100
	}
	return &RingBuffer{
		buf:  make([]string, size),
		size: size,
	}
}

// Write appends a line to the buffer, overwriting the oldest entry when full.
func (r *RingBuffer) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.head] = string(p)
	r.head = (r.head + 1) % r.size
	if r.count < r.size {
		r.count++
	}
	return len(p), nil
}

// Lines returns all stored lines in insertion order (oldest first).
func (r *RingBuffer) Lines() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]string, r.count)
	start := (r.head - r.count + r.size) % r.size
	for i := 0; i < r.count; i++ {
		result[i] = r.buf[(start+i)%r.size]
	}
	return result
}

// Len returns the number of stored entries.
func (r *RingBuffer) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.count
}

// Reset clears all entries.
func (r *RingBuffer) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.head = 0
	r.count = 0
}
