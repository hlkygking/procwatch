package supervisor

import "strings"

// FilterMode controls how process names are matched.
type FilterMode int

const (
	FilterExact  FilterMode = iota
	FilterPrefix            // match names with a given prefix
	FilterAll               // match all processes
)

// ProcessFilter selects a subset of ProcessConfig entries.
type ProcessFilter struct {
	mode  FilterMode
	value string
}

// NewExactFilter returns a filter that matches a single process by name.
func NewExactFilter(name string) ProcessFilter {
	return ProcessFilter{mode: FilterExact, value: name}
}

// NewPrefixFilter returns a filter that matches processes whose name starts with prefix.
func NewPrefixFilter(prefix string) ProcessFilter {
	return ProcessFilter{mode: FilterPrefix, value: prefix}
}

// NewAllFilter returns a filter that matches every process.
func NewAllFilter() ProcessFilter {
	return ProcessFilter{mode: FilterAll}
}

// Match reports whether the given ProcessConfig is selected by this filter.
func (f ProcessFilter) Match(cfg ProcessConfig) bool {
	switch f.mode {
	case FilterAll:
		return true
	case FilterExact:
		return cfg.Name == f.value
	case FilterPrefix:
		return strings.HasPrefix(cfg.Name, f.value)
	}
	return false
}

// Apply returns only the configs from the provided slice that match the filter.
func (f ProcessFilter) Apply(configs []ProcessConfig) []ProcessConfig {
	var out []ProcessConfig
	for _, c := range configs {
		if f.Match(c) {
			out = append(out, c)
		}
	}
	return out
}
