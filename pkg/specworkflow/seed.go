// Package specworkflow provides spec workflow management and built-in workflow definitions.
package specworkflow

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ProductBuildersHQ/visionstudio/pkg/initiative"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// BuiltInWorkflows returns the default spec workflow definitions.
func BuiltInWorkflows() []*store.SpecWorkflow {
	return []*store.SpecWorkflow{
		{
			ID:            "quick-fix",
			Name:          "Quick Fix",
			Description:   "Minimal workflow for small fixes and maintenance tasks",
			SpecsRequired: []string{"ROADMAP.md"},
			SpecsOptional: []string{},
			InitTypes:     []string{initiative.TypeMaintenance},
		},
		{
			ID:            "pbhq-lite",
			Name:          "PBHQ Lite",
			Description:   "Lightweight workflow for refactors and migrations",
			SpecsRequired: []string{"PLAN.md", "ROADMAP.md"},
			SpecsOptional: []string{"PRD.md", "TRD.md"},
			InitTypes:     []string{initiative.TypeRefactor, initiative.TypeMigration},
		},
		{
			ID:            "pbhq-standard",
			Name:          "PBHQ Standard",
			Description:   "Standard workflow for features and compliance initiatives",
			SpecsRequired: []string{"PRD.md", "TRD.md", "PLAN.md", "ROADMAP.md"},
			SpecsOptional: []string{},
			InitTypes:     []string{initiative.TypeFeature, initiative.TypeCompliance},
		},
		{
			ID:            "aws-working-backwards",
			Name:          "AWS Working Backwards",
			Description:   "Full PR/FAQ process for major product initiatives",
			SpecsRequired: []string{"PRFAQ.md", "PRD.md", "TRD.md", "PLAN.md", "ROADMAP.md"},
			SpecsOptional: []string{},
			InitTypes:     []string{}, // opt-in only
		},
	}
}

// DefaultWorkflowForType returns the default workflow ID for an initiative type.
func DefaultWorkflowForType(initType string) string {
	switch initType {
	case initiative.TypeMaintenance:
		return "quick-fix"
	case initiative.TypeRefactor, initiative.TypeMigration:
		return "pbhq-lite"
	case initiative.TypeFeature, initiative.TypeCompliance:
		return "pbhq-standard"
	default:
		return "pbhq-standard"
	}
}

// SpecDir returns the canonical spec directory for an initiative,
// relative to its home repository root: docs/specs/initiatives/{INIT-ID}/
func SpecDir(initiativeID string) string {
	return filepath.Join("docs", "specs", "initiatives", initiativeID)
}

// Resolve returns the workflow that applies to an initiative: its explicit
// WorkflowID override if set, otherwise the default for its type. It checks
// the store first, falling back to the built-in definitions so resolution
// works even before `workflow seed` has been run.
func Resolve(ctx context.Context, s store.SpecWorkflowStore, init *store.Initiative) (*store.SpecWorkflow, error) {
	id := init.WorkflowID
	if id == "" {
		id = DefaultWorkflowForType(init.InitType)
	}

	wf, err := s.GetSpecWorkflow(ctx, id)
	if err == nil {
		return wf, nil
	}

	for _, builtin := range BuiltInWorkflows() {
		if builtin.ID == id {
			return builtin, nil
		}
	}
	return nil, fmt.Errorf("workflow %q not found (store: %w)", id, err)
}

// SeedBuiltIn creates the built-in workflows if they don't exist.
// Returns the number of workflows created.
func SeedBuiltIn(ctx context.Context, s store.SpecWorkflowStore) (int, error) {
	existing, err := s.ListSpecWorkflows(ctx)
	if err != nil {
		return 0, fmt.Errorf("list workflows: %w", err)
	}
	existingIDs := make(map[string]bool, len(existing))
	for _, wf := range existing {
		existingIDs[wf.ID] = true
	}

	var created int
	for _, wf := range BuiltInWorkflows() {
		if existingIDs[wf.ID] {
			continue
		}
		if err := s.CreateSpecWorkflow(ctx, wf); err != nil {
			return created, fmt.Errorf("create %s: %w", wf.ID, err)
		}
		created++
	}
	return created, nil
}
