package supervisor

// LoadPrioritiesFromConfig reads priority fields from a slice of ProcessConfig
// entries and populates the given ProcessPriorityStore. Unknown or empty
// priority strings default to PriorityNormal and do not produce an error.
func LoadPrioritiesFromConfig(configs []ProcessConfig, store *ProcessPriorityStore) []error {
	var errs []error
	for _, cfg := range configs {
		raw := ""
		if v, ok := cfg.Env["PROCWATCH_PRIORITY"]; ok {
			raw = v
		}
		p, err := ParsePriority(raw)
		if err != nil {
			errs = append(errs, err)
			p = PriorityNormal
		}
		store.Set(cfg.Name, p)
	}
	return errs
}

// OrderedByPriority returns process configs sorted by their assigned priority
// (highest first), preserving original order for equal priorities.
func OrderedByPriority(configs []ProcessConfig, store *ProcessPriorityStore) []ProcessConfig {
	type indexed struct {
		cfg ProcessConfig
		pri Priority
		idx int
	}
	items := make([]indexed, len(configs))
	for i, c := range configs {
		p, ok := store.Get(c.Name)
		if !ok {
			p = PriorityNormal
		}
		items[i] = indexed{cfg: c, pri: p, idx: i}
	}
	// stable insertion sort by priority descending, then original index ascending
	for i := 1; i < len(items); i++ {
		for j := i; j > 0; j-- {
			a, b := items[j-1], items[j]
			if b.pri > a.pri || (b.pri == a.pri && b.idx < a.idx) {
				items[j-1], items[j] = items[j], items[j-1]
			} else {
				break
			}
		}
	}
	out := make([]ProcessConfig, len(items))
	for i, it := range items {
		out[i] = it.cfg
	}
	return out
}
