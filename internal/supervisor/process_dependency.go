package supervisor

import (
	"errors"
	"fmt"
)

// ProcessDependency represents a directed dependency between two processes.
type ProcessDependency struct {
	From string // process that depends on
	To   string // process that must start first
}

// DependencyGraph holds a set of process dependencies and supports
// topological ordering to determine safe startup sequence.
type DependencyGraph struct {
	edges map[string][]string // from -> list of deps
	nodes map[string]struct{}
}

// NewDependencyGraph creates an empty DependencyGraph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		edges: make(map[string][]string),
		nodes: make(map[string]struct{}),
	}
}

// Add registers a dependency: 'from' depends on 'to'.
func (g *DependencyGraph) Add(from, to string) {
	g.nodes[from] = struct{}{}
	g.nodes[to] = struct{}{}
	g.edges[from] = append(g.edges[from], to)
}

// Order returns process names in topological order (dependencies first).
// Returns an error if a cycle is detected.
func (g *DependencyGraph) Order() ([]string, error) {
	visited := make(map[string]int) // 0=unvisited, 1=visiting, 2=done
	var result []string

	var visit func(node string) error
	visit = func(node string) error {
		switch visited[node] {
		case 1:
			return fmt.Errorf("cycle detected at process %q", node)
		case 2:
			return nil
		}
		visited[node] = 1
		for _, dep := range g.edges[node] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visited[node] = 2
		result = append(result, node)
		return nil
	}

	for node := range g.nodes {
		if err := visit(node); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Validate checks that all referenced process names exist in the provided set.
func (g *DependencyGraph) Validate(known map[string]struct{}) error {
	for node := range g.nodes {
		if _, ok := known[node]; !ok {
			return fmt.Errorf("unknown process %q referenced in dependencies", node)
		}
	}
	return nil
}

// ErrCycle is returned when a dependency cycle is detected.
var ErrCycle = errors.New("dependency cycle detected")
