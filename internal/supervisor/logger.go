package supervisor

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// LogLevel represents the severity of a log entry.
type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelDebug LogLevel = "debug"
)

// LogEntry is a structured log record emitted by the supervisor.
type LogEntry struct {
	Timestamp string   `json:"timestamp"`
	Level     LogLevel `json:"level"`
	Process   string   `json:"process,omitempty"`
	Message   string   `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// Logger writes structured JSON log entries to an io.Writer.
type Logger struct {
	writer  io.Writer
	process string
}

// NewLogger creates a Logger that writes to w. If w is nil, os.Stdout is used.
func NewLogger(w io.Writer, process string) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{writer: w, process: process}
}

// log emits a single structured entry.
func (l *Logger) log(level LogLevel, msg string, fields map[string]any) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Process:   l.process,
		Message:   msg,
		Fields:    fields,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_, _ = l.writer.Write(append(data, '\n'))
}

func (l *Logger) Info(msg string, fields map[string]any)  { l.log(LogLevelInfo, msg, fields) }
func (l *Logger) Warn(msg string, fields map[string]any)  { l.log(LogLevelWarn, msg, fields) }
func (l *Logger) Error(msg string, fields map[string]any) { l.log(LogLevelError, msg, fields) }
func (l *Logger) Debug(msg string, fields map[string]any) { l.log(LogLevelDebug, msg, fields) }
