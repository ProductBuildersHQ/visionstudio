// Package ir defines the intermediate representation types for VisionStudio.
// This file defines the LeverageGraph IR for dependency-based leverage analysis.
// The IR is language-agnostic; adapters (Go, TypeScript, Python, etc.) populate it.
package ir

// LeverageGraph represents a dependency graph for leverage analysis.
// It captures which modules depend on which, enabling visualization of
// reuse patterns and opportunities for consolidation.
type LeverageGraph struct {
	// GeneratedAt is the ISO 8601 timestamp when this graph was generated.
	GeneratedAt string `json:"generatedAt"`

	// Ecosystem identifies the language/package ecosystem (e.g., "go", "npm", "pypi").
	Ecosystem string `json:"ecosystem"`

	// Scope describes what was scanned (e.g., "github.com/grokify", "monorepo:platform").
	Scope string `json:"scope"`

	// Modules are all modules/packages discovered in the scan.
	Modules []LeverageModule `json:"modules"`

	// Edges represent dependency relationships.
	Edges []LeverageEdge `json:"edges"`

	// Summary provides rollup statistics.
	Summary LeverageSummary `json:"summary"`
}

// LeverageModule represents a single module/package in the graph.
type LeverageModule struct {
	// ID is a unique identifier (typically the module path).
	ID string `json:"id"`

	// Name is a human-readable short name.
	Name string `json:"name"`

	// Path is the full module path (e.g., "github.com/grokify/mogo").
	Path string `json:"path"`

	// Kind categorizes the module: "internal", "external", "stdlib".
	Kind string `json:"kind"`

	// Org is the organization/namespace (e.g., "grokify", "plexusone").
	Org string `json:"org,omitempty"`

	// Version is the current version if known.
	Version string `json:"version,omitempty"`

	// Stats contains usage statistics.
	Stats ModuleStats `json:"stats"`

	// Tags for filtering/grouping (e.g., ["platform", "utility", "framework"]).
	Tags []string `json:"tags,omitempty"`
}

// ModuleStats captures usage metrics for a module.
type ModuleStats struct {
	// DirectDependents is count of modules that directly import this one.
	DirectDependents int `json:"directDependents"`

	// TransitiveDependents is count of modules that transitively depend on this.
	TransitiveDependents int `json:"transitiveDependents,omitempty"`

	// DirectDependencies is count of modules this one directly imports.
	DirectDependencies int `json:"directDependencies"`

	// Depth is the longest path from a root to this module.
	Depth int `json:"depth,omitempty"`

	// LeverageScore is a computed score (dependents / dependencies ratio).
	LeverageScore float64 `json:"leverageScore,omitempty"`
}

// LeverageEdge represents a dependency relationship.
type LeverageEdge struct {
	// From is the module ID that has the dependency.
	From string `json:"from"`

	// To is the module ID being depended upon.
	To string `json:"to"`

	// Kind is "direct" or "indirect".
	Kind string `json:"kind"`

	// Version is the required version constraint if known.
	Version string `json:"version,omitempty"`
}

// LeverageSummary provides rollup statistics for the graph.
type LeverageSummary struct {
	// TotalModules is the count of all modules.
	TotalModules int `json:"totalModules"`

	// InternalModules is count of modules marked as internal.
	InternalModules int `json:"internalModules"`

	// ExternalModules is count of external dependencies.
	ExternalModules int `json:"externalModules"`

	// TotalEdges is the count of all edges.
	TotalEdges int `json:"totalEdges"`

	// InternalEdges is count of edges between internal modules.
	InternalEdges int `json:"internalEdges"`

	// InternalRatio is internalEdges / totalEdges as a percentage.
	InternalRatio float64 `json:"internalRatio"`

	// TopLeveraged lists the most-depended-upon internal modules.
	TopLeveraged []ModuleRank `json:"topLeveraged,omitempty"`

	// TopConsumers lists modules that consume the most internal dependencies.
	TopConsumers []ModuleRank `json:"topConsumers,omitempty"`

	// Orphans lists internal modules with zero dependents (reuse opportunities).
	Orphans []string `json:"orphans,omitempty"`

	// Clusters groups tightly-coupled modules for potential consolidation.
	Clusters []ModuleCluster `json:"clusters,omitempty"`
}

// ModuleRank is a module with its rank/score.
type ModuleRank struct {
	ModuleID   string  `json:"moduleId"`
	Dependents int     `json:"dependents"`
	Direct     int     `json:"direct,omitempty"`
	Indirect   int     `json:"indirect,omitempty"`
	Score      float64 `json:"score,omitempty"`
}

// ModuleCluster groups related modules.
type ModuleCluster struct {
	ID      string   `json:"id"`
	Name    string   `json:"name,omitempty"`
	Modules []string `json:"modules"`
	Reason  string   `json:"reason,omitempty"`
}
