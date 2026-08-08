package ir

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/grokify/gogit/scanner"
)

// GoLeverageAdapter builds a LeverageGraph from Go module scan results.
type GoLeverageAdapter struct {
	// InternalPrefixes are module path prefixes considered "internal".
	InternalPrefixes []string

	// Scope describes what was scanned.
	Scope string
}

// DefaultGoAdapter returns an adapter configured for grokify/plexusone/ProductBuildersHQ.
func DefaultGoAdapter() *GoLeverageAdapter {
	return &GoLeverageAdapter{
		InternalPrefixes: []string{
			"github.com/grokify/",
			"github.com/plexusone/",
			"github.com/ProductBuildersHQ/",
		},
		Scope: "platform-orgs",
	}
}

// BuildGraph constructs a LeverageGraph from gogit scan results.
func (a *GoLeverageAdapter) BuildGraph(results []scanner.RepoResult) *LeverageGraph {
	graph := &LeverageGraph{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Ecosystem:   "go",
		Scope:       a.Scope,
		Modules:     []LeverageModule{},
		Edges:       []LeverageEdge{},
	}

	// Track modules and their dependents
	moduleSet := make(map[string]*LeverageModule)
	dependentCount := make(map[string]int)
	dependencyCount := make(map[string]int)

	// First pass: collect all internal modules
	for _, r := range results {
		if !r.HasGoMod || r.ModuleName == "" {
			continue
		}

		mod := &LeverageModule{
			ID:   r.ModuleName,
			Name: filepath.Base(r.ModuleName),
			Path: r.ModuleName,
			Kind: a.classifyModule(r.ModuleName),
			Org:  a.extractOrg(r.ModuleName),
			Stats: ModuleStats{
				DirectDependencies: len(r.DirectDependencies),
			},
		}
		moduleSet[r.ModuleName] = mod
	}

	// Track direct deps separately for consumer stats
	directDepsSet := make(map[string]map[string]bool) // module -> set of direct deps

	// Second pass: build edges and count dependents
	for _, r := range results {
		if !r.HasGoMod || r.ModuleName == "" {
			continue
		}

		// Track direct deps for this module
		directDepsSet[r.ModuleName] = make(map[string]bool)
		for _, dep := range r.DirectDependencies {
			directDepsSet[r.ModuleName][dep] = true
		}

		// Process all dependencies (direct + indirect)
		for _, dep := range r.Dependencies {
			isDirect := directDepsSet[r.ModuleName][dep]
			kind := "indirect"
			if isDirect {
				kind = "direct"
			}

			// Create edge
			edge := LeverageEdge{
				From: r.ModuleName,
				To:   dep,
				Kind: kind,
			}
			graph.Edges = append(graph.Edges, edge)

			// Count dependents for the target
			dependentCount[dep]++
			dependencyCount[r.ModuleName]++

			// Ensure external deps are in module set
			if _, exists := moduleSet[dep]; !exists {
				moduleSet[dep] = &LeverageModule{
					ID:   dep,
					Name: filepath.Base(dep),
					Path: dep,
					Kind: a.classifyModule(dep),
					Org:  a.extractOrg(dep),
				}
			}
		}
	}

	// Update stats and collect modules
	for id, mod := range moduleSet {
		mod.Stats.DirectDependents = dependentCount[id]
		if mod.Stats.DirectDependencies == 0 {
			mod.Stats.DirectDependencies = dependencyCount[id]
		}

		// Compute leverage score (dependents / max(dependencies, 1))
		deps := mod.Stats.DirectDependencies
		if deps == 0 {
			deps = 1
		}
		mod.Stats.LeverageScore = float64(mod.Stats.DirectDependents) / float64(deps)

		graph.Modules = append(graph.Modules, *mod)
	}

	// Sort modules by path for consistent output
	sort.Slice(graph.Modules, func(i, j int) bool {
		return graph.Modules[i].Path < graph.Modules[j].Path
	})

	// Compute summary
	graph.Summary = a.computeSummary(graph)

	return graph
}

// classifyModule returns "internal", "external", or "stdlib".
func (a *GoLeverageAdapter) classifyModule(path string) string {
	// stdlib has no dots in first segment
	if !strings.Contains(strings.Split(path, "/")[0], ".") {
		return "stdlib"
	}

	for _, prefix := range a.InternalPrefixes {
		if strings.HasPrefix(path, prefix) {
			return "internal"
		}
	}
	return "external"
}

// extractOrg extracts the org from a module path.
func (a *GoLeverageAdapter) extractOrg(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// computeSummary calculates rollup statistics.
func (a *GoLeverageAdapter) computeSummary(graph *LeverageGraph) LeverageSummary {
	summary := LeverageSummary{
		TotalModules: len(graph.Modules),
		TotalEdges:   len(graph.Edges),
	}

	// Count by kind
	internalModules := make(map[string]bool)
	for _, m := range graph.Modules {
		switch m.Kind {
		case "internal":
			summary.InternalModules++
			internalModules[m.ID] = true
		case "external":
			summary.ExternalModules++
		}
	}

	// Count internal edges and find orphans
	orphanSet := make(map[string]bool)
	for id := range internalModules {
		orphanSet[id] = true
	}

	for _, e := range graph.Edges {
		fromInternal := internalModules[e.From]
		toInternal := internalModules[e.To]
		if fromInternal && toInternal {
			summary.InternalEdges++
		}
		// Remove from orphan set if something depends on it
		if toInternal {
			delete(orphanSet, e.To)
		}
	}

	if summary.TotalEdges > 0 {
		summary.InternalRatio = float64(summary.InternalEdges) / float64(summary.TotalEdges) * 100
	}

	// Collect orphans (internal modules with no dependents)
	for id := range orphanSet {
		summary.Orphans = append(summary.Orphans, id)
	}
	sort.Strings(summary.Orphans)

	// Find top leveraged internal modules (most depended upon)
	type modScore struct {
		id    string
		count int
		score float64
	}
	var leveragedScores []modScore
	for _, m := range graph.Modules {
		if m.Kind == "internal" && m.Stats.DirectDependents > 0 {
			leveragedScores = append(leveragedScores, modScore{m.ID, m.Stats.DirectDependents, m.Stats.LeverageScore})
		}
	}
	sort.Slice(leveragedScores, func(i, j int) bool {
		return leveragedScores[i].count > leveragedScores[j].count
	})

	// Top 10 leveraged
	limit := 10
	if len(leveragedScores) < limit {
		limit = len(leveragedScores)
	}
	for i := 0; i < limit; i++ {
		summary.TopLeveraged = append(summary.TopLeveraged, ModuleRank{
			ModuleID:   leveragedScores[i].id,
			Dependents: leveragedScores[i].count,
			Score:      leveragedScores[i].score,
		})
	}

	// Find top consumers (modules that use the most internal dependencies)
	type consumerStats struct {
		total    int
		direct   int
		indirect int
	}
	consumerMap := make(map[string]*consumerStats)

	for _, e := range graph.Edges {
		if internalModules[e.To] {
			if consumerMap[e.From] == nil {
				consumerMap[e.From] = &consumerStats{}
			}
			consumerMap[e.From].total++
			if e.Kind == "direct" {
				consumerMap[e.From].direct++
			} else {
				consumerMap[e.From].indirect++
			}
		}
	}

	type consumerScore struct {
		id       string
		total    int
		direct   int
		indirect int
	}
	var consumerScores []consumerScore
	for modID, stats := range consumerMap {
		if internalModules[modID] && stats.total > 0 {
			consumerScores = append(consumerScores, consumerScore{modID, stats.total, stats.direct, stats.indirect})
		}
	}
	sort.Slice(consumerScores, func(i, j int) bool {
		return consumerScores[i].total > consumerScores[j].total
	})

	// Top 10 consumers
	limit = 10
	if len(consumerScores) < limit {
		limit = len(consumerScores)
	}
	for i := 0; i < limit; i++ {
		summary.TopConsumers = append(summary.TopConsumers, ModuleRank{
			ModuleID:   consumerScores[i].id,
			Dependents: consumerScores[i].total,
			Direct:     consumerScores[i].direct,
			Indirect:   consumerScores[i].indirect,
		})
	}

	return summary
}

// ScanAndBuild scans directories and builds a graph in one call.
func (a *GoLeverageAdapter) ScanAndBuild(dirs []string) (*LeverageGraph, error) {
	var allResults []scanner.RepoResult

	for _, dir := range dirs {
		results, err := scanner.ScanDirectoryWithProgress(dir, nil, scanner.ScanOptions{})
		if err != nil {
			continue
		}
		allResults = append(allResults, results...)
	}

	return a.BuildGraph(allResults), nil
}
