package supervisor

// ProcessAuditLogSubscriber wires a ProcessEventBus to a ProcessAuditLog,
// translating process events into structured audit entries automatically.
type ProcessAuditLogSubscriber struct {
	bus *ProcessEventBus
	log *ProcessAuditLog
}

// NewProcessAuditLogSubscriber creates a subscriber and registers it with the
// provided event bus. Call Stop to deregister.
func NewProcessAuditLogSubscriber(bus *ProcessEventBus, log *ProcessAuditLog) *ProcessAuditLogSubscriber {
	s := &ProcessAuditLogSubscriber{bus: bus, log: log}
	bus.Subscribe(s.handleEvent)
	return s
}

// handleEvent translates a ProcessEvent into an AuditEntry and records it.
func (s *ProcessAuditLogSubscriber) handleEvent(evt ProcessEvent) {
	entry := AuditEntry{
		Timestamp: evt.Timestamp,
		Process:   evt.ProcessName,
		Message:   evt.Message,
	}

	switch evt.Kind {
	case EventKindStarted:
		entry.Kind = AuditEventStart
	case EventKindStopped:
		entry.Kind = AuditEventStop
		if evt.ExitCode != 0 {
			code := evt.ExitCode
			entry.ExitCode = &code
		}
	case EventKindRestarting:
		entry.Kind = AuditEventRestart
		entry.RestartNum = evt.RestartNum
	case EventKindKilled:
		entry.Kind = AuditEventKill
	case EventKindSkipped:
		entry.Kind = AuditEventSkip
	default:
		// Unknown event kinds are silently ignored.
		return
	}

	s.log.Record(entry)
}
