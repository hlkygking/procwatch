package supervisor

// ProcessTagFilter filters process configs by tag membership.
// Tags are arbitrary string labels attached to ProcessConfig entries.
type ProcessTagFilter struct {
	tag string
}

// NewTagFilter returns a filter that matches processes containing the given tag.
func NewTagFilter(tag string) *ProcessTagFilter {
	return &ProcessTagFilter{tag: tag}
}

// Match returns true if the config's Tags slice contains the filter tag.
func (f *ProcessTagFilter) Match(cfg ProcessConfig) bool {
	for _, t := range cfg.Tags {
		if t == f.tag {
			return true
		}
	}
	return false
}

// FilterByTag returns only those configs whose Tags include the given tag.
func FilterByTag(configs []ProcessConfig, tag string) []ProcessConfig {
	f := NewTagFilter(tag)
	var out []ProcessConfig
	for _, c := range configs {
		if f.Match(c) {
			out = append(out, c)
		}
	}
	return out
}

// FilterByAllTags returns configs that contain every tag in the provided set.
func FilterByAllTags(configs []ProcessConfig, tags []string) []ProcessConfig {
	var out []ProcessConfig
	for _, c := range configs {
		if hasAllTags(c.Tags, tags) {
			out = append(out, c)
		}
	}
	return out
}

func hasAllTags(have []string, want []string) bool {
	set := make(map[string]struct{}, len(have))
	for _, t := range have {
		set[t] = struct{}{}
	}
	for _, t := range want {
		if _, ok := set[t]; !ok {
			return false
		}
	}
	return true
}
