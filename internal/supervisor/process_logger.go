package supervisor

import (
	"bufio"
	"io"
)

// ProcessLogger tees stdout/stderr lines from a child process into both a
// structured Logger and an in-memory RingBuffer for later inspection.
type ProcessLogger struct {
	logger *Logger
	buffer *RingBuffer
}

// NewProcessLogger creates a ProcessLogger for the named process.
// bufSize controls the RingBuffer capacity (0 → default 100).
func NewProcessLogger(w io.Writer, processName string, bufSize int) *ProcessLogger {
	return &ProcessLogger{
		logger: NewLogger(w, processName),
		buffer: NewRingBuffer(bufSize),
	}
}

// StreamLines reads lines from r, writing each to the logger at the given level
// and to the ring buffer. Blocks until r is closed.
func (pl *ProcessLogger) StreamLines(r io.Reader, level LogLevel) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		_, _ = pl.buffer.Write([]byte(line))
		pl.logger.log(level, line, nil)
	}
}

// RecentLines returns the most recent log lines captured from the process.
func (pl *ProcessLogger) RecentLines() []string {
	return pl.buffer.Lines()
}

// Logger returns the underlying structured Logger.
func (pl *ProcessLogger) Logger() *Logger {
	return pl.logger
}
