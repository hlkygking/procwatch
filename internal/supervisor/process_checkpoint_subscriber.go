package supervisor

// ProcessCheckpointSubscriber listens to process events and records
// checkpoints automatically based on event kind.
type ProcessCheckpointSubscriber struct {
	log *ProcessCheckpointLog
}

// NewProcessCheckpointSubscriber creates a subscriber that writes to the given log.
func NewProcessCheckpointSubscriber(log *ProcessCheckpointLog) *ProcessCheckpointSubscriber {
	return &ProcessCheckpointSubscriber{log: log}
}

// Subscribe registers the subscriber on the given event bus.
func (s *ProcessCheckpointSubscriber) Subscribe(bus *ProcessEventBus) {
	bus.Subscribe(func(e ProcessEvent) {
		var kind CheckpointKind
		switch e.Kind {
		case EventStarted:
			kind = CheckpointStarted
		case EventStopped:
			kind = CheckpointStopped
		case EventFailed:
			kind = CheckpointFailed
		case EventRestarted:
			kind = CheckpointRestored
		default:
			return
		}
		meta := map[string]string{}
		if e.Message != "" {
			meta["message"] = e.Message
		}
		s.log.Record(e.Process, kind, meta)
	})
}
