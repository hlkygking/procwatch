package supervisor

import (
	"testing"
	"time"
)

func TestProcessCheckpointLog_RecordAndAll(t *testing.T) {
	log := NewProcessCheckpointLog()
	log.Record("web", CheckpointStarted, nil)
	log.Record("worker", CheckpointReady, nil)

	all := log.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all[0].Process != "web" || all[0].Kind != CheckpointStarted {
		t.Errorf("unexpected first entry: %+v", all[0])
	}
}

func TestProcessCheckpointLog_TimestampSet(t *testing.T) {
	log := NewProcessCheckpointLog()
	before := time.Now()
	log.Record("svc", CheckpointReady, nil)
	after := time.Now()

	all := log.All()
	if all[0].Timestamp.Before(before) || all[0].Timestamp.After(after) {
		t.Errorf("timestamp out of range: %v", all[0].Timestamp)
	}
}

func TestProcessCheckpointLog_ForProcess(t *testing.T) {
	log := NewProcessCheckpointLog()
	log.Record("alpha", CheckpointStarted, nil)
	log.Record("beta", CheckpointStarted, nil)
	log.Record("alpha", CheckpointReady, nil)

	results := log.ForProcess("alpha")
	if len(results) != 2 {
		t.Fatalf("expected 2 for alpha, got %d", len(results))
	}
	for _, r := range results {
		if r.Process != "alpha" {
			t.Errorf("unexpected process in result: %s", r.Process)
		}
	}
}

func TestProcessCheckpointLog_ForProcess_NoMatch(t *testing.T) {
	log := NewProcessCheckpointLog()
	log.Record("alpha", CheckpointStarted, nil)

	results := log.ForProcess("ghost")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestProcessCheckpointLog_LastOf(t *testing.T) {
	log := NewProcessCheckpointLog()
	log.Record("svc", CheckpointStarted, nil)
	log.Record("svc", CheckpointReady, map[string]string{"port": "8080"})
	log.Record("svc", CheckpointReady, map[string]string{"port": "9090"})

	last := log.LastOf("svc", CheckpointReady)
	if last == nil {
		t.Fatal("expected a result")
	}
	if last.Meta["port"] != "9090" {
		t.Errorf("expected last ready port 9090, got %s", last.Meta["port"])
	}
}

func TestProcessCheckpointLog_LastOf_Missing(t *testing.T) {
	log := NewProcessCheckpointLog()
	log.Record("svc", CheckpointStarted, nil)

	result := log.LastOf("svc", CheckpointFailed)
	if result != nil {
		t.Errorf("expected nil, got %+v", result)
	}
}

func TestProcessCheckpointLog_MetaStored(t *testing.T) {
	log := NewProcessCheckpointLog()
	log.Record("svc", CheckpointRestored, map[string]string{"reason": "oom"})

	all := log.All()
	if all[0].Meta["reason"] != "oom" {
		t.Errorf("expected meta reason=oom, got %v", all[0].Meta)
	}
}
