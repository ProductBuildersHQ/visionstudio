// Package export writes JSONL snapshots of all PRISM Control tables
// plus a manifest into a target directory.
package export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// Manifest records metadata about an export run.
type Manifest struct {
	ExportedAt time.Time      `json:"exported_at"`
	Counts     map[string]int `json:"counts"`
}

// Result summarizes what was exported.
type Result struct {
	Dir      string
	Manifest Manifest
}

// Run exports all tables from the store as JSONL files into dir.
func Run(ctx context.Context, s store.Store, dir string) (*Result, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create export dir: %w", err)
	}

	counts := make(map[string]int)

	initiatives, err := s.ListInitiatives(ctx)
	if err != nil {
		return nil, fmt.Errorf("list initiatives: %w", err)
	}
	if err := writeJSONL(filepath.Join(dir, "initiatives.jsonl"), initiatives); err != nil {
		return nil, err
	}
	counts["initiatives"] = len(initiatives)

	var allPhases []*store.Phase
	for _, init := range initiatives {
		phases, err := s.ListPhases(ctx, init.ID)
		if err != nil {
			return nil, fmt.Errorf("list phases for %s: %w", init.ID, err)
		}
		allPhases = append(allPhases, phases...)
	}
	if err := writeJSONL(filepath.Join(dir, "phases.jsonl"), allPhases); err != nil {
		return nil, err
	}
	counts["phases"] = len(allPhases)

	rmis, err := s.ListAllRMIs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list RMIs: %w", err)
	}
	if err := writeJSONL(filepath.Join(dir, "roadmap_items.jsonl"), rmis); err != nil {
		return nil, err
	}
	counts["roadmap_items"] = len(rmis)

	deps, err := s.ListAllDependencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list RMI dependencies: %w", err)
	}
	if err := writeJSONL(filepath.Join(dir, "rmi_dependencies.jsonl"), deps); err != nil {
		return nil, err
	}
	counts["rmi_dependencies"] = len(deps)

	assignments, err := s.ListAllAssignments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list assignments: %w", err)
	}
	if err := writeJSONL(filepath.Join(dir, "assignments.jsonl"), assignments); err != nil {
		return nil, err
	}
	counts["assignments"] = len(assignments)

	evidence, err := s.ListAllEvidence(ctx)
	if err != nil {
		return nil, fmt.Errorf("list evidence: %w", err)
	}
	if err := writeJSONL(filepath.Join(dir, "evidence.jsonl"), evidence); err != nil {
		return nil, err
	}
	counts["evidence"] = len(evidence)

	repos, err := s.ListRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}
	if err := writeJSONL(filepath.Join(dir, "repositories.jsonl"), repos); err != nil {
		return nil, err
	}
	counts["repositories"] = len(repos)

	repoDeps, err := s.ListAllRepoDependencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repo dependencies: %w", err)
	}
	if err := writeJSONL(filepath.Join(dir, "repository_dependencies.jsonl"), repoDeps); err != nil {
		return nil, err
	}
	counts["repository_dependencies"] = len(repoDeps)

	manifest := Manifest{
		ExportedAt: time.Now(),
		Counts:     counts,
	}

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), manifestBytes, 0o600); err != nil {
		return nil, fmt.Errorf("write manifest: %w", err)
	}

	return &Result{Dir: dir, Manifest: manifest}, nil
}

func writeJSONL[T any](path string, items []T) (retErr error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", filepath.Base(path), err)
	}
	defer func() {
		if err := f.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("close %s: %w", filepath.Base(path), err)
		}
	}()

	enc := json.NewEncoder(f)
	for _, item := range items {
		if err := enc.Encode(item); err != nil {
			return fmt.Errorf("encode to %s: %w", filepath.Base(path), err)
		}
	}
	return nil
}
