package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

// ProcessSnapshotReporter prints snapshot data in table or JSON format.
type ProcessSnapshotReporter struct {
	store  *ProcessSnapshotStore
	writer io.Writer
}

// NewProcessSnapshotReporter creates a reporter backed by the given store.
func NewProcessSnapshotReporter(store *ProcessSnapshotStore, w io.Writer) *ProcessSnapshotReporter {
	if w == nil {
		w = os.Stdout
	}
	return &ProcessSnapshotReporter{store: store, writer: w}
}

// PrintTable writes a human-readable table of the latest snapshot per process.
func (r *ProcessSnapshotReporter) PrintTable(names []string) {
	tw := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PROCESS\tSTATUS\tRESTARTS\tUPTIME\tTAKEN_AT")
	for _, name := range names {
		snap, ok := r.store.Latest(name)
		if !ok {
			continue
		}
		fmt.Fprintf(tw, "%s\t%s\t%d\t%.1fs\t%s\n",
			snap.Name, snap.Status, snap.Restarts,
			snap.Uptime, snap.TakenAt.Format(time.RFC3339))
	}
	tw.Flush()
}

// PrintJSON writes all snapshots for the given process names as a JSON array.
func (r *ProcessSnapshotReporter) PrintJSON(names []string) error {
	var out []ProcessSnapshot
	for _, name := range names {
		out = append(out, r.store.ForProcess(name)...)
	}
	enc := json.NewEncoder(r.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
