package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// ProcessCheckpointReporter renders checkpoint logs as table or JSON.
type ProcessCheckpointReporter struct {
	log    *ProcessCheckpointLog
	writer io.Writer
}

// NewProcessCheckpointReporter creates a reporter for the given log.
func NewProcessCheckpointReporter(log *ProcessCheckpointLog, w io.Writer) *ProcessCheckpointReporter {
	if w == nil {
		w = os.Stdout
	}
	return &ProcessCheckpointReporter{log: log, writer: w}
}

// PrintTable writes a human-readable table of checkpoints.
func (r *ProcessCheckpointReporter) PrintTable() {
	entries := r.log.All()
	tw := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PROCESS\tKIND\tTIMESTAMP")
	for _, e := range entries {
		fmt.Fprintf(tw, "%s\t%s\t%s\n",
			e.Process,
			string(e.Kind),
			e.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		)
	}
	tw.Flush()
}

// PrintJSON writes all checkpoints as a JSON array.
func (r *ProcessCheckpointReporter) PrintJSON() error {
	entries := r.log.All()
	enc := json.NewEncoder(r.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

// PrintSummary writes a concise per-process checkpoint count summary.
func (r *ProcessCheckpointReporter) PrintSummary() {
	entries := r.log.All()
	counts := make(map[string]int)
	for _, e := range entries {
		counts[e.Process]++
	}
	tw := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PROCESS\tCHECKPOINTS")
	for process, count := range counts {
		fmt.Fprintf(tw, "%s\t%d\n", process, count)
	}
	tw.Flush()
}
