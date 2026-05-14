package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
)

// ProcessCooldownReporter renders cooldown state for all tracked processes.
type ProcessCooldownReporter struct {
	cooldown *ProcessCooldown
	names    []string
	writer   io.Writer
}

// NewProcessCooldownReporter creates a reporter for the given cooldown tracker
// and process name list. If writer is nil, os.Stdout is used.
func NewProcessCooldownReporter(cd *ProcessCooldown, names []string, w io.Writer) *ProcessCooldownReporter {
	if w == nil {
		w = os.Stdout
	}
	return &ProcessCooldownReporter{cooldown: cd, names: names, writer: w}
}

type cooldownEntry struct {
	Process string `json:"process"`
	Cooled  bool   `json:"cooled"`
}

func (r *ProcessCooldownReporter) entries() []cooldownEntry {
	sorted := make([]string, len(r.names))
	copy(sorted, r.names)
	sort.Strings(sorted)

	out := make([]cooldownEntry, 0, len(sorted))
	for _, name := range sorted {
		out = append(out, cooldownEntry{
			Process: name,
			Cooled:  r.cooldown.IsCooled(name),
		})
	}
	return out
}

// PrintTable writes a human-readable table of cooldown states.
func (r *ProcessCooldownReporter) PrintTable() {
	fmt.Fprintf(r.writer, "%-30s %s\n", "PROCESS", "COOLED")
	for _, e := range r.entries() {
		fmt.Fprintf(r.writer, "%-30s %v\n", e.Process, e.Cooled)
	}
}

// PrintJSON writes a JSON array of cooldown states.
func (r *ProcessCooldownReporter) PrintJSON() error {
	enc := json.NewEncoder(r.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(r.entries())
}
