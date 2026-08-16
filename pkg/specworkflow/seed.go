// Package specworkflow provides spec workflow management backed by the
// specification-workflow-spec catalog (the single source of truth for all
// default workflow definitions).
package specworkflow

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ProductBuildersHQ/visionstudio/pkg/initiative"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// retiredRemap maps workflow IDs that no longer exist in the upstream catalog
// to their canonical replacement. Initiatives referencing a retired ID are
// remapped during SyncFromCatalog.
var retiredRemap = map[string]string{
	// pbhq-standard (all four docs required) is semantically identical to the
	// canonical upstream pbhq-lite definition.
	"pbhq-standard": "pbhq-lite",
	// The AWS Working Backwards profiles were renamed by Amazon's actual
	// selection criterion — decision reversibility — rather than the
	// product/feature taxonomy.
	"aws-product": "aws-one-way-door",
	"aws-feature": "aws-two-way-door",
}

// SpecFileName returns the canonical spec filename for an upstream spec-type
// ID (e.g. "prd" → "PRD.md", "opportunity-spec" → "OPPORTUNITY-SPEC.md").
func SpecFileName(specType string) string {
	return strings.ToUpper(specType) + ".md"
}

// DefaultWorkflowForType returns the default workflow ID for an initiative type.
func DefaultWorkflowForType(initType string) string {
	switch initType {
	case initiative.TypeMaintenance, initiative.TypeRefactor, initiative.TypeMigration:
		return "quick-fix"
	default:
		return "pbhq-lite"
	}
}

// DefaultInitTypes is the inverse of DefaultWorkflowForType, used to keep the
// DB index rows informative about which init types default to a workflow.
func DefaultInitTypes(workflowID string) []string {
	switch workflowID {
	case "quick-fix":
		return []string{initiative.TypeMaintenance, initiative.TypeRefactor, initiative.TypeMigration}
	case "pbhq-lite":
		return []string{initiative.TypeFeature, initiative.TypeCompliance}
	default:
		return nil // opt-in only
	}
}

// SpecDir returns the canonical spec directory for an initiative,
// relative to its home repository root: docs/specs/initiatives/{INIT-ID}/
func SpecDir(initiativeID string) string {
	return filepath.Join("docs", "specs", "initiatives", initiativeID)
}

// StoreWorkflow converts a loaded upstream workflow into the store
// representation used as a DB index row and by spec scaffolding. Required and
// optional specs are expressed as canonical filenames, ordered by the
// workflow's execution sequence where one is defined.
func StoreWorkflow(id string, lw *LoadedWorkflow) *store.SpecWorkflow {
	w := lw.Workflow

	orderSpecs := func(types []string) []string {
		ordered := orderBySequence(w, types)
		files := make([]string, len(ordered))
		for i, t := range ordered {
			files[i] = SpecFileName(t)
		}
		return files
	}

	var required, optional []string
	for specType, req := range w.SpecConfig {
		if req.Required {
			required = append(required, specType)
		} else {
			optional = append(optional, specType)
		}
	}

	return &store.SpecWorkflow{
		ID:            id,
		Name:          w.Name,
		Description:   w.Description,
		SpecsRequired: orderSpecs(required),
		SpecsOptional: orderSpecs(optional),
		InitTypes:     DefaultInitTypes(id),
	}
}

// Resolve returns the workflow that applies to an initiative: its explicit
// WorkflowID override if set, otherwise the default for its type. Definitions
// come from the upstream catalog via the loader — never from the store.
func Resolve(loader *Loader, init *store.Initiative) (*store.SpecWorkflow, error) {
	id := init.WorkflowID
	if id == "" {
		id = DefaultWorkflowForType(init.InitType)
	}
	lw, err := loader.Load(id)
	if err != nil {
		return nil, fmt.Errorf("workflow %q not found in catalog: %w", id, err)
	}
	return StoreWorkflow(id, lw), nil
}

// SyncResult reports what SyncFromCatalog changed.
type SyncResult struct {
	Created  int
	Updated  int
	Remapped map[string]string // initiativeID → new workflow ID
	Deleted  []string          // retired workflow rows removed
	Retained []string          // rows absent from catalog but still referenced
}

// SyncFromCatalog makes the store's spec_workflows table an index of the
// upstream catalog: it upserts a row per catalog workflow, remaps initiatives
// referencing retired IDs to their canonical replacement, and deletes rows
// that are absent from the catalog and unreferenced. Idempotent.
func SyncFromCatalog(ctx context.Context, s store.Store, loader *Loader) (*SyncResult, error) {
	res := &SyncResult{Remapped: map[string]string{}}

	existing, err := s.ListSpecWorkflows(ctx)
	if err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	existingByID := make(map[string]*store.SpecWorkflow, len(existing))
	for _, wf := range existing {
		existingByID[wf.ID] = wf
	}

	// Upsert every catalog workflow.
	catalogIDs := map[string]bool{}
	for _, id := range loader.Available() {
		lw, err := loader.Load(id)
		if err != nil {
			return nil, fmt.Errorf("load workflow %q: %w", id, err)
		}
		catalogIDs[id] = true
		row := StoreWorkflow(id, lw)
		if _, ok := existingByID[id]; ok {
			if err := s.UpdateSpecWorkflow(ctx, row); err != nil {
				return nil, fmt.Errorf("update workflow %q: %w", id, err)
			}
			res.Updated++
		} else {
			if err := s.CreateSpecWorkflow(ctx, row); err != nil {
				return nil, fmt.Errorf("create workflow %q: %w", id, err)
			}
			res.Created++
		}
	}

	// Remap initiatives referencing retired IDs.
	inits, err := s.ListInitiatives(ctx)
	if err != nil {
		return nil, fmt.Errorf("list initiatives: %w", err)
	}
	referenced := map[string]bool{}
	for _, init := range inits {
		if init.WorkflowID == "" {
			continue
		}
		if target, retired := retiredRemap[init.WorkflowID]; retired {
			init.WorkflowID = target
			if err := s.UpdateInitiative(ctx, init); err != nil {
				return nil, fmt.Errorf("remap initiative %s to workflow %q: %w", init.ID, target, err)
			}
			res.Remapped[init.ID] = target
			referenced[target] = true
			continue
		}
		referenced[init.WorkflowID] = true
	}

	// Delete rows absent from the catalog and unreferenced; retain (and
	// report) any absent-but-referenced rows rather than breaking them.
	for id := range existingByID {
		if catalogIDs[id] {
			continue
		}
		if referenced[id] {
			res.Retained = append(res.Retained, id)
			continue
		}
		if err := s.DeleteSpecWorkflow(ctx, id); err != nil {
			return nil, fmt.Errorf("delete retired workflow %q: %w", id, err)
		}
		res.Deleted = append(res.Deleted, id)
	}
	sort.Strings(res.Deleted)
	sort.Strings(res.Retained)

	return res, nil
}
