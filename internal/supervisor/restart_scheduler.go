package supervisor

import (
	"context"
	"time"
)

// RestartScheduler coordinates restart timing using a Throttle and RestartLimiter.
type RestartScheduler struct {
	throttle *Throttle
	limiter  *RestartLimiter
	logger   *Logger
	name     string
}

// NewRestartScheduler creates a RestartScheduler for a named process.
func NewRestartScheduler(name string, limiter *RestartLimiter, policy ThrottlePolicy, logger *Logger) *RestartScheduler {
	return &RestartScheduler{
		name:     name,
		throttle: NewThrottle(policy),
		limiter:  limiter,
		logger:   logger,
	}
}

// Schedule blocks for the appropriate backoff duration and reports whether
// a restart is permitted. Returns false if the context is cancelled or the
// limiter denies the restart.
func (rs *RestartScheduler) Schedule(ctx context.Context, exitCode int) bool {
	if !rs.limiter.Allow() {
		rs.logger.Info("restart limit reached", map[string]any{
			"process": rs.name,
		})
		return false
	}

	delay := rs.throttle.Next()
	rs.logger.Info("scheduling restart", map[string]any{
		"process":   rs.name,
		"delay":     delay.String(),
		"exit_code": exitCode,
	})

	select {
	case <-time.After(delay):
		return true
	case <-ctx.Done():
		return false
	}
}

// RecordSuccess resets the throttle backoff after a successful long-running period.
func (rs *RestartScheduler) RecordSuccess() {
	rs.throttle.Reset()
	rs.limiter.Reset()
}
