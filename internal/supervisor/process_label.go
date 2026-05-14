package supervisor

import "sort"

// ProcessLabel represents a key-value metadata label attached to a process config.
type ProcessLabel struct {
	Key   string
	Value string
}

// LabelSet holds a collection of labels for a named process.
type LabelSet struct {
	Process string
	Labels  map[string]string
}

// NewLabelSet creates an empty LabelSet for the given process.
func NewLabelSet(process string) *LabelSet {
	return &LabelSet{
		Process: process,
		Labels:  make(map[string]string),
	}
}

// Set adds or overwrites a label key with the given value.
func (ls *LabelSet) Set(key, value string) {
	ls.Labels[key] = value
}

// Get returns the value for a label key and whether it was present.
func (ls *LabelSet) Get(key string) (string, bool) {
	v, ok := ls.Labels[key]
	return v, ok
}

// Keys returns all label keys in sorted order.
func (ls *LabelSet) Keys() []string {
	keys := make([]string, 0, len(ls.Labels))
	for k := range ls.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Matches returns true if all provided key-value pairs exist in the label set.
func (ls *LabelSet) Matches(selector map[string]string) bool {
	for k, v := range selector {
		if ls.Labels[k] != v {
			return false
		}
	}
	return true
}

// Clone returns a deep copy of the LabelSet.
func (ls *LabelSet) Clone() *LabelSet {
	copy := NewLabelSet(ls.Process)
	for k, v := range ls.Labels {
		copy.Labels[k] = v
	}
	return copy
}
