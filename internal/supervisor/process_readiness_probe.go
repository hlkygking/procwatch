package supervisor

import (
	"context"
	"net"
	"net/http"
	"time"
)

// ProbeKind identifies the type of readiness check.
type ProbeKind string

const (
	ProbeHTTP ProbeKind = "http"
	ProbeTCP  ProbeKind = "tcp"
)

// ReadinessProbeConfig holds configuration for a single probe.
type ReadinessProbeConfig struct {
	Kind            ProbeKind     `json:"kind"`
	Target          string        `json:"target"`           // URL for HTTP, host:port for TCP
	Timeout         time.Duration `json:"timeout"`          // per-attempt timeout
	Interval        time.Duration `json:"interval"`         // time between attempts
	SuccessThreshold int          `json:"success_threshold"` // consecutive successes needed
	FailureThreshold int          `json:"failure_threshold"` // consecutive failures before unhealthy
}

// ReadinessResult is the outcome of a single probe attempt.
type ReadinessResult struct {
	Process string
	Ready   bool
	Err     error
}

// ProcessReadinessProbe runs readiness checks for a named process.
type ProcessReadinessProbe struct {
	cfg     ReadinessProbeConfig
	process string
	client  *http.Client
}

// NewProcessReadinessProbe creates a probe for the given process and config.
func NewProcessReadinessProbe(process string, cfg ReadinessProbeConfig) *ProcessReadinessProbe {
	if cfg.Timeout == 0 {
		cfg.Timeout = 2 * time.Second
	}
	if cfg.Interval == 0 {
		cfg.Interval = 5 * time.Second
	}
	if cfg.SuccessThreshold == 0 {
		cfg.SuccessThreshold = 1
	}
	if cfg.FailureThreshold == 0 {
		cfg.FailureThreshold = 3
	}
	return &ProcessReadinessProbe{
		cfg:     cfg,
		process: process,
		client:  &http.Client{Timeout: cfg.Timeout},
	}
}

// Probe performs a single readiness check and returns the result.
func (p *ProcessReadinessProbe) Probe(ctx context.Context) ReadinessResult {
	var err error
	switch p.cfg.Kind {
	case ProbeHTTP:
		err = p.probeHTTP(ctx)
	case ProbeTCP:
		err = p.probeTCP(ctx)
	default:
		err = p.probeHTTP(ctx)
	}
	return ReadinessResult{Process: p.process, Ready: err == nil, Err: err}
}

func (p *ProcessReadinessProbe) probeHTTP(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.Target, nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return &probeError{status: resp.StatusCode}
}

func (p *ProcessReadinessProbe) probeTCP(ctx context.Context) error {
	d := &net.Dialer{Timeout: p.cfg.Timeout}
	conn, err := d.DialContext(ctx, "tcp", p.cfg.Target)
	if err != nil {
		return err
	}
	return conn.Close()
}

type probeError struct{ status int }

func (e *probeError) Error() string {
	return "unexpected status: " + http.StatusText(e.status)
}
