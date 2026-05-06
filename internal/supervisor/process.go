package supervisor

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// Process supervises a single child process according to its config and restart policy.
type Process struct {
	cfg     ProcessConfig
	policy  RestartPolicy
	limiter *RestartLimiter
	logger  *ProcessLogger
	health  *HealthTracker
	mu      sync.Mutex
}

// ProcessConfig holds the runtime configuration for a supervised process.
type ProcessConfig struct {
	Name    string
	Command string
	Args    []string
	Env     []string
}

// NewProcess constructs a Process ready to be Run.
func NewProcess(cfg ProcessConfig, policy RestartPolicy, logger *ProcessLogger, health *HealthTracker) *Process {
	limiter := NewRestartLimiter(
		10,
		5*time.Minute,
		500*time.Millisecond,
		30*time.Second,
	)
	return &Process{
		cfg:    cfg,
		policy: policy,
		limiter: limiter,
		logger: logger,
		health: health,
	}
}

// Run starts the process and supervises it until ctx is cancelled or the policy
// says not to restart.
func (p *Process) Run(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		cmd := exec.CommandContext(ctx, p.cfg.Command, p.cfg.Args...)
		if len(p.cfg.Env) > 0 {
			cmd.Env = p.cfg.Env
		}
		if p.logger != nil {
			cmd.Stdout = p.logger.Stdout()
			cmd.Stderr = p.logger.Stderr()
		}

		if p.health != nil {
			p.health.RecordStart(p.cfg.Name)
		}

		p.log("starting process")
		if err := cmd.Start(); err != nil {
			p.log(fmt.Sprintf("failed to start: %v", err))
			return err
		}

		waitErr := cmd.Wait()
		exitCode := 0
		if waitErr != nil {
			if exitErr, ok := waitErr.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}

		if p.health != nil {
			p.health.RecordExit(p.cfg.Name, exitCode)
		}
		p.log(fmt.Sprintf("exited with code %d", exitCode))

		if !p.policy.ShouldRestart(exitCode) {
			p.log("policy says do not restart")
			return waitErr
		}

		if !p.limiter.Allow() {
			p.log("restart limit reached, giving up")
			if p.health != nil {
				p.health.SetStatus(p.cfg.Name, StatusFailed)
			}
			return fmt.Errorf("process %q exceeded restart limit", p.cfg.Name)
		}

		delay := p.limiter.Backoff()
		p.limiter.Record()
		p.log(fmt.Sprintf("restarting in %v", delay))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}

func (p *Process) log(msg string) {
	if p.logger != nil {
		p.logger.Info(msg)
	}
}
