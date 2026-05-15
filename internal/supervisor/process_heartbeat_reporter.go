package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
)

// ProcessHeartbeatReporter renders heartbeat status for all tracked processes.
type ProcessHeartbeatReporter struct {
	store  *ProcessHeartbeatStore
	writer io.Writer
}

// NewProcessHeartbeatReporter creates a reporter backed by the given store.
// If writer is nil, os.Stdout is used.
func NewProcessHeartbeatReporter(store *ProcessHeartbeatStore, writer io.Writer) *ProcessHeartbeatReporter {
	if writer == nil {
		writer = os.Stdout
	}
	return &ProcessHeartbeatReporter{store: store, writer: writer}
}

// PrintTable writes a human-readable table of heartbeat statuses.
func (r *ProcessHeartbeatReporter) PrintTable() {
	statuses := r.store.All()
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Process < statuses[j].Process
	})
	tw := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PROCESS\tALIVE\tMISSED\tLAST BEAT")
	for _, s := range statuses {
		alive := "yes"
		if !s.Alive {
			alive = "no"
		}
		last := "-"
		if !s.LastBeat.IsZero() {
			last = s.LastBeat.Format("15:04:05")
		}
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n", s.Process, alive, s.Missed, last)
	}
	tw.Flush()
}

// PrintJSON writes heartbeat statuses as a JSON array.
func (r *ProcessHeartbeatReporter) PrintJSON() error {
	statuses := r.store.All()
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Process < statuses[j].Process
	})
	enc := json.NewEncoder(r.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(statuses)
}
