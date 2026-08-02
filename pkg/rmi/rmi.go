// Package rmi implements RMI domain logic: ID validation, status rules,
// dependency graph analysis, and the ready-work query.
package rmi

import (
	"fmt"
	"regexp"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// Status constants for roadmap items.
const (
	StatusPlanned    = "planned"
	StatusReady      = "ready"
	StatusInProgress = "in_progress"
	StatusBlocked    = "blocked"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
)

var rmiIDPattern = regexp.MustCompile(`^RMI-[A-Z0-9]+-\d{3}$`)

// ValidID reports whether id matches the RMI-<REPOSLUG>-<NNN> format.
func ValidID(id string) bool {
	return rmiIDPattern.MatchString(id)
}

// DependencyGraph holds directed edges between RMIs for analysis.
type DependencyGraph struct {
	edges map[string][]string // target -> sources (sources depend on target)
	deps  map[string][]string // source -> targets (source depends on these)
}

// NewDependencyGraph builds a graph from dependency records.
func NewDependencyGraph(deps []*store.RMIDependency) *DependencyGraph {
	g := &DependencyGraph{
		edges: make(map[string][]string),
		deps:  make(map[string][]string),
	}
	for _, d := range deps {
		if d.Relationship == "requires" {
			g.edges[d.TargetRMIID] = append(g.edges[d.TargetRMIID], d.SourceRMIID)
			g.deps[d.SourceRMIID] = append(g.deps[d.SourceRMIID], d.TargetRMIID)
		}
	}
	return g
}

// IsBlocked reports whether an RMI has any incomplete required dependencies.
func (g *DependencyGraph) IsBlocked(rmiID string, statuses map[string]string) bool {
	for _, dep := range g.deps[rmiID] {
		st, ok := statuses[dep]
		if !ok || (st != StatusCompleted && st != StatusCancelled) {
			return true
		}
	}
	return false
}

// DetectCycles returns any cycle found in the graph, or nil if acyclic.
func (g *DependencyGraph) DetectCycles() []string {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[string]int)
	var cycle []string

	var dfs func(string) bool
	dfs = func(node string) bool {
		color[node] = gray
		for _, dep := range g.deps[node] {
			switch color[dep] {
			case gray:
				cycle = append(cycle, dep, node)
				return true
			case white:
				if dfs(dep) {
					return true
				}
			}
		}
		color[node] = black
		return false
	}

	for node := range g.deps {
		if color[node] == white {
			if dfs(node) {
				return cycle
			}
		}
	}
	// also check nodes only present as targets
	for node := range g.edges {
		if color[node] == white {
			if dfs(node) {
				return cycle
			}
		}
	}
	return nil
}

// ReadyWork returns RMI IDs that are planned/ready, not blocked by
// incomplete dependencies, and have no active assignment.
func ReadyWork(rmis []*store.RoadmapItem, deps []*store.RMIDependency, activeAssignments map[string]bool) []string {
	statuses := make(map[string]string, len(rmis))
	for _, r := range rmis {
		statuses[r.ID] = r.Status
	}

	graph := NewDependencyGraph(deps)
	var ready []string
	for _, r := range rmis {
		if r.Status != StatusPlanned && r.Status != StatusReady {
			continue
		}
		if activeAssignments[r.ID] {
			continue
		}
		if graph.IsBlocked(r.ID, statuses) {
			continue
		}
		ready = append(ready, r.ID)
	}
	return ready
}

// ValidateID returns an error if the RMI ID format is invalid.
func ValidateID(id string) error {
	if !ValidID(id) {
		return fmt.Errorf("invalid RMI ID %q: must match RMI-<REPOSLUG>-<NNN>", id)
	}
	return nil
}
