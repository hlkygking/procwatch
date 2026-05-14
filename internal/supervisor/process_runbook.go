package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// RunbookEntry records a named operational note or runbook step associated with a process.
type RunbookEntry struct {
	Process   string    `json:"process"`
	Step      string    `json:"step"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// ProcessRunbook stores runbook entries per process.
type ProcessRunbook struct {
	mu      sync.RWMutex
	entries []RunbookEntry
}

// NewProcessRunbook creates an empty ProcessRunbook.
func NewProcessRunbook() *ProcessRunbook {
	return &ProcessRunbook{}
}

// Add appends a runbook entry for the given process.
func (r *ProcessRunbook) Add(process, step, note string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, RunbookEntry{
		Process:   process,
		Step:      step,
		Note:      note,
		CreatedAt: time.Now(),
	})
}

// ForProcess returns all runbook entries for the given process name.
func (r *ProcessRunbook) ForProcess(process string) []RunbookEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []RunbookEntry
	for _, e := range r.entries {
		if e.Process == process {
			out = append(out, e)
		}
	}
	return out
}

// All returns all runbook entries.
func (r *ProcessRunbook) All() []RunbookEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]RunbookEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

// PrintTable writes a human-readable table of runbook entries to w.
func (r *ProcessRunbook) PrintTable(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	entries := r.All()
	if len(entries) == 0 {
		fmt.Fprintln(w, "no runbook entries")
		return
	}
	fmt.Fprintf(w, "%-20s %-20s %s\n", "PROCESS", "STEP", "NOTE")
	for _, e := range entries {
		fmt.Fprintf(w, "%-20s %-20s %s\n", e.Process, e.Step, e.Note)
	}
}

// PrintJSON writes runbook entries as JSON lines to w.
func (r *ProcessRunbook) PrintJSON(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	for _, e := range r.All() {
		b, _ := json.Marshal(e)
		fmt.Fprintln(w, string(b))
	}
}
