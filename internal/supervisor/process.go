package supervisor

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// RestartPolicy defines how a process should be restarted.
type RestartPolicy string

const (
	RestartAlways    RestartPolicy = "always"
	RestartOnFailure RestartPolicy = "on-failure"
	RestartNever     RestartPolicy = "never"
)

// ProcessConfig holds configuration for a managed process.
type ProcessConfig struct {
	Name          string
	Command       string
	Args          []string
	RestartPolicy RestartPolicy
	MaxRestarts   int
	BackoffDelay  time.Duration
}

// Process represents a supervised process.
type Process struct {
	cfg         ProcessConfig
	cmd         *exec.Cmd
	restarts    int
	mu          sync.Mutex
	logger      Logger
}

// NewProcess creates a new supervised Process.
func NewProcess(cfg ProcessConfig, logger Logger) *Process {
	if cfg.BackoffDelay == 0 {
		cfg.BackoffDelay = 2 * time.Second
	}
	return &Process{cfg: cfg, logger: logger}
}

// Run starts the process and supervises it according to the restart policy.
func (p *Process) Run(ctx context.Context) error {
	for {
		p.mu.Lock()
		p.cmd = exec.CommandContext(ctx, p.cfg.Command, p.cfg.Args...)
		p.mu.Unlock()

		p.logger.Info("starting process", map[string]interface{}{
			"name":    p.cfg.Name,
			"command": p.cfg.Command,
			"args":    p.cfg.Args,
		})

		err := p.cmd.Run()
		exitCode := 0
		if p.cmd.ProcessState != nil {
			exitCode = p.cmd.ProcessState.ExitCode()
		}

		p.logger.Info("process exited", map[string]interface{}{
			"name":      p.cfg.Name,
			"exit_code": exitCode,
			"error":     fmt.Sprintf("%v", err),
		})

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if !p.shouldRestart(err) {
			return err
		}

		p.mu.Lock()
		p.restarts++
		restarts := p.restarts
		p.mu.Unlock()

		p.logger.Info("scheduling restart", map[string]interface{}{
			"name":     p.cfg.Name,
			"restarts": restarts,
			"delay":    p.cfg.BackoffDelay.String(),
		})

		select {
		case <-time.After(p.cfg.BackoffDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Process) shouldRestart(err error) bool {
	if p.cfg.MaxRestarts > 0 && p.restarts >= p.cfg.MaxRestarts {
		return false
	}
	switch p.cfg.RestartPolicy {
	case RestartAlways:
		return true
	case RestartOnFailure:
		return err != nil
	default:
		return false
	}
}
