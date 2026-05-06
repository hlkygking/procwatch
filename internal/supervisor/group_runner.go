package supervisor

import (
	"context"
	"sync"
)

// GroupRunner manages concurrent execution of multiple supervised processes.
type GroupRunner struct {
	processes []*Process
	logger    *Logger
	health    *HealthTracker
	wg        sync.WaitGroup
}

// NewGroupRunner creates a GroupRunner from a slice of ProcessConfig entries.
func NewGroupRunner(configs []ProcessConfig, logger *Logger, health *HealthTracker) (*GroupRunner, error) {
	g := &GroupRunner{
		logger: logger,
		health: health,
	}
	for _, cfg := range configs {
		p, err := NewProcess(cfg, logger, health)
		if err != nil {
			return nil, err
		}
		g.processes = append(g.processes, p)
	}
	return g, nil
}

// Run starts all processes concurrently and blocks until all have exited or
// the context is cancelled.
func (g *GroupRunner) Run(ctx context.Context) {
	for _, p := range g.processes {
		g.wg.Add(1)
		go func(proc *Process) {
			defer g.wg.Done()
			g.logger.Info("starting process", map[string]interface{}{
				"process": proc.Config.Name,
			})
			proc.Run(ctx)
			g.logger.Info("process exited", map[string]interface{}{
				"process": proc.Config.Name,
			})
		}(p)
	}
	g.wg.Wait()
}

// Len returns the number of processes managed by this runner.
func (g *GroupRunner) Len() int {
	return len(g.processes)
}
