package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// ProcessLabelReporter prints label information for processes.
type ProcessLabelReporter struct {
	store  *ProcessLabelStore
	writer io.Writer
}

// NewProcessLabelReporter creates a reporter backed by the given store.
// If writer is nil, os.Stdout is used.
func NewProcessLabelReporter(store *ProcessLabelStore, writer io.Writer) *ProcessLabelReporter {
	if writer == nil {
		writer = os.Stdout
	}
	return &ProcessLabelReporter{store: store, writer: writer}
}

// PrintTable writes a human-readable table of process labels to the writer.
func (r *ProcessLabelReporter) PrintTable() {
	sets := r.store.All()
	if len(sets) == 0 {
		fmt.Fprintln(r.writer, "no labels defined")
		return
	}
	for _, ls := range sets {
		pairs := make([]string, 0, len(ls.Labels))
		for _, k := range ls.Keys() {
			pairs = append(pairs, fmt.Sprintf("%s=%s", k, ls.Labels[k]))
		}
		fmt.Fprintf(r.writer, "%-20s %s\n", ls.Process, strings.Join(pairs, "  "))
	}
}

// PrintJSON writes all label sets as a JSON array to the writer.
func (r *ProcessLabelReporter) PrintJSON() error {
	type entry struct {
		Process string            `json:"process"`
		Labels  map[string]string `json:"labels"`
	}
	sets := r.store.All()
	entries := make([]entry, 0, len(sets))
	for _, ls := range sets {
		entries = append(entries, entry{Process: ls.Process, Labels: ls.Labels})
	}
	enc := json.NewEncoder(r.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
