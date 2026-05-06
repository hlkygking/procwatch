package supervisor

// ProcessEventLogger listens on a ProcessEventBus and writes structured
// log entries for each event using the project Logger.
type ProcessEventLogger struct {
	bus    *ProcessEventBus
	logger *Logger
}

// NewProcessEventLogger creates a ProcessEventLogger that will consume
// events from bus and emit structured log lines via logger.
func NewProcessEventLogger(bus *ProcessEventBus, logger *Logger) *ProcessEventLogger {
	return &ProcessEventLogger{bus: bus, logger: logger}
}

// Start subscribes to the bus and logs events until ch is closed or the
// done channel is closed. It is intended to run in its own goroutine.
func (l *ProcessEventLogger) Start(done <-chan struct{}) {
	ch := l.bus.Subscribe(32)
	go func() {
		for {
			select {
			case e, ok := <-ch:
				if !ok {
					return
				}
				l.log(e)
			case <-done:
				return
			}
		}
	}()
}

// log emits a single structured log entry for the given event.
func (l *ProcessEventLogger) log(e ProcessEvent) {
	fields := map[string]interface{}{
		"process": e.ProcessName,
		"event":   string(e.Type),
	}
	if e.ExitCode != 0 {
		fields["exit_code"] = e.ExitCode
	}
	if e.Message != "" {
		fields["message"] = e.Message
	}

	switch e.Type {
	case EventFailed:
		l.logger.WithFields(fields).Error("process event")
	case EventStopped:
		if e.ExitCode != 0 {
			l.logger.WithFields(fields).Warn("process event")
		} else {
			l.logger.WithFields(fields).Info("process event")
		}
	default:
		l.logger.WithFields(fields).Info("process event")
	}
}
