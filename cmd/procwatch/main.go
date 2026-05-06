package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/yourorg/procwatch/internal/supervisor"
)

const defaultConfigPath = "procwatch.json"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "procwatch: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		configPath  string
		showStatus  bool
		filterName  string
		filterPrefix string
		logJSON     bool
	)

	flag.StringVar(&configPath, "config", defaultConfigPath, "path to procwatch JSON config file")
	flag.BoolVar(&showStatus, "status", false, "print process status snapshot and exit")
	flag.StringVar(&filterName, "name", "", "run only the process with this exact name")
	flag.StringVar(&filterPrefix, "prefix", "", "run only processes whose names start with this prefix")
	flag.BoolVar(&logJSON, "json", false, "emit structured JSON log output")
	flag.Parse()

	cfg, err := supervisor.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Build structured logger writing to stdout.
	logger := supervisor.NewLogger(os.Stdout)
	if logJSON {
		logger = supervisor.NewLogger(os.Stdout)
	}

	// Resolve which processes to run based on filter flags.
	var filter supervisor.ProcessFilter
	switch {
	case filterName != "":
		filter = supervisor.NewExactFilter(filterName)
	case filterPrefix != "":
		filter = supervisor.NewPrefixFilter(filterPrefix)
	default:
		filter = supervisor.NewAllFilter()
	}

	selector := supervisor.NewProcessSelector(cfg.Processes, filter)
	selected, err := selector.Select()
	if err != nil {
		return fmt.Errorf("selecting processes: %w", err)
	}
	if len(selected) == 0 {
		return fmt.Errorf("no processes matched the given filter")
	}

	// Shared health tracker and event bus used across all subsystems.
	healthTracker := supervisor.NewHealthTracker()
	eventBus := supervisor.NewProcessEventBus()

	// Wire up event-driven logging for process lifecycle events.
	eventLogger := supervisor.NewProcessEventLogger(eventBus, logger)
	_ = eventLogger // subscribed as side-effect

	// Build the group runner from the selected process configs.
	runner := supervisor.NewGroupRunner(selected, logger, healthTracker, eventBus)

	// If --status was requested, print a snapshot of current health and exit.
	// At startup there is nothing running yet, but this validates config + wiring.
	if showStatus {
		reporter := supervisor.NewStatusReporter(healthTracker)
		reporter.PrintTable(os.Stdout)
		return nil
	}

	// Set up OS signal handling so SIGINT/SIGTERM trigger graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigHandler := supervisor.NewSignalHandler(logger)
	sigCtx := sigHandler.Start(ctx)

	logger.Info("procwatch starting", map[string]interface{}{
		"config":    configPath,
		"processes": len(selected),
	})

	// Run all selected processes under the signal-aware context.
	runner.Run(sigCtx)

	logger.Info("procwatch stopped", nil)
	return nil
}
