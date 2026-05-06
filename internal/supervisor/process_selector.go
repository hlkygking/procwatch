package supervisor

import "fmt"

// ProcessSelector resolves a named or wildcard target string into a
// ProcessFilter and applies it to a Config, returning the matching
// ProcessConfig entries or an error if nothing matched.
type ProcessSelector struct {
	cfg Config
}

// NewProcessSelector creates a selector backed by the given config.
func NewProcessSelector(cfg Config) *ProcessSelector {
	return &ProcessSelector{cfg: cfg}
}

// Select resolves target into matching ProcessConfig entries.
// A target of "*" or "all" selects every process.
// A target ending in "*" is treated as a prefix match (e.g. "worker*").
// Otherwise an exact match is performed.
func (s *ProcessSelector) Select(target string) ([]ProcessConfig, error) {
	var f ProcessFilter

	switch {
	case target == "*" || target == "all":
		f = NewAllFilter()
	case len(target) > 1 && target[len(target)-1] == '*':
		f = NewPrefixFilter(target[:len(target)-1])
	default:
		f = NewExactFilter(target)
	}

	matches := f.Apply(s.cfg.Processes)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no processes matched target %q", target)
	}
	return matches, nil
}

// MustSelect is like Select but panics on error; useful in tests.
func (s *ProcessSelector) MustSelect(target string) []ProcessConfig {
	result, err := s.Select(target)
	if err != nil {
		panic(err)
	}
	return result
}
