package specworkflow_test

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/initiative"
	"github.com/ProductBuildersHQ/visionstudio/pkg/specworkflow"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestDefaultWorkflowForType(t *testing.T) {
	tests := []struct {
		initType string
		want     string
	}{
		{initiative.TypeMaintenance, "quick-fix"},
		{initiative.TypeRefactor, "quick-fix"},
		{initiative.TypeMigration, "quick-fix"},
		{initiative.TypeFeature, "pbhq-lite"},
		{initiative.TypeCompliance, "pbhq-lite"},
		{"", "pbhq-lite"},
	}
	for _, tt := range tests {
		if got := specworkflow.DefaultWorkflowForType(tt.initType); got != tt.want {
			t.Errorf("DefaultWorkflowForType(%q) = %q, want %q", tt.initType, got, tt.want)
		}
	}
}

func TestSpecFileName(t *testing.T) {
	tests := []struct {
		specType string
		want     string
	}{
		{"prd", "PRD.md"},
		{"opportunity-spec", "OPPORTUNITY-SPEC.md"},
		{"narrative-6p", "NARRATIVE-6P.md"},
	}
	for _, tt := range tests {
		if got := specworkflow.SpecFileName(tt.specType); got != tt.want {
			t.Errorf("SpecFileName(%q) = %q, want %q", tt.specType, got, tt.want)
		}
	}
}

func TestSpecDir(t *testing.T) {
	got := specworkflow.SpecDir("INIT-FOO-001")
	want := filepath.Join("docs", "specs", "initiatives", "INIT-FOO-001")
	if got != want {
		t.Errorf("SpecDir() = %q, want %q", got, want)
	}
}

func TestResolve(t *testing.T) {
	loader := specworkflow.DefaultLoader()

	t.Run("type default when no WorkflowID", func(t *testing.T) {
		init := &store.Initiative{ID: "INIT-A", InitType: initiative.TypeFeature}
		wf, err := specworkflow.Resolve(loader, init)
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if wf.ID != "pbhq-lite" {
			t.Errorf("expected pbhq-lite, got %q", wf.ID)
		}
		want := []string{"PRD.md", "TRD.md", "PLAN.md", "ROADMAP.md"}
		if !slices.Equal(wf.SpecsRequired, want) {
			t.Errorf("pbhq-lite required = %v, want %v", wf.SpecsRequired, want)
		}
	})

	t.Run("explicit WorkflowID override", func(t *testing.T) {
		init := &store.Initiative{ID: "INIT-B", InitType: initiative.TypeMaintenance, WorkflowID: "aws-two-way-door"}
		wf, err := specworkflow.Resolve(loader, init)
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if wf.ID != "aws-two-way-door" {
			t.Errorf("expected aws-two-way-door, got %q", wf.ID)
		}
		// Required set follows the upstream profile (execution-sequence order).
		// The PR/FAQ lead; OPPORTUNITY-SPEC is optional post-FAQ deepening.
		want := []string{
			"PRESS.md", "FAQ.md", "PRD.md",
			"UXD.md", "TRD.md", "TPD.md",
		}
		if !slices.Equal(wf.SpecsRequired, want) {
			t.Errorf("aws-two-way-door required = %v, want %v", wf.SpecsRequired, want)
		}
	})

	t.Run("quick-fix requires only ROADMAP", func(t *testing.T) {
		init := &store.Initiative{ID: "INIT-C", InitType: initiative.TypeMaintenance}
		wf, err := specworkflow.Resolve(loader, init)
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if !slices.Equal(wf.SpecsRequired, []string{"ROADMAP.md"}) {
			t.Errorf("quick-fix required = %v, want [ROADMAP.md]", wf.SpecsRequired)
		}
	})

	t.Run("errors on unknown workflow ID", func(t *testing.T) {
		init := &store.Initiative{ID: "INIT-D", InitType: initiative.TypeFeature, WorkflowID: "does-not-exist"}
		if _, err := specworkflow.Resolve(loader, init); err == nil {
			t.Error("expected error for unknown workflow ID")
		}
	})

	t.Run("retired IDs are not resolvable", func(t *testing.T) {
		init := &store.Initiative{ID: "INIT-E", InitType: initiative.TypeFeature, WorkflowID: "pbhq-standard"}
		if _, err := specworkflow.Resolve(loader, init); err == nil {
			t.Error("expected error for retired workflow ID pbhq-standard")
		}
	})
}

func TestSyncFromCatalog(t *testing.T) {
	ctx := context.Background()
	loader := specworkflow.DefaultLoader()
	s := store.NewMemStore()

	// Seed pre-consolidation state: retired rows plus initiatives on them.
	for _, wf := range []*store.SpecWorkflow{
		{ID: "pbhq-standard", Name: "PBHQ Standard", SpecsRequired: []string{"PRD.md", "TRD.md", "PLAN.md", "ROADMAP.md"}},
		{ID: "aws-working-backwards", Name: "AWS Working Backwards", SpecsRequired: []string{"PRFAQ.md"}},
		{ID: "pbhq-lite", Name: "PBHQ Lite (old local)", SpecsRequired: []string{"PLAN.md", "ROADMAP.md"}},
	} {
		if err := s.CreateSpecWorkflow(ctx, wf); err != nil {
			t.Fatalf("seed workflow %s: %v", wf.ID, err)
		}
	}
	for _, init := range []*store.Initiative{
		{ID: "INIT-STD-001", InitType: initiative.TypeFeature, WorkflowID: "pbhq-standard"},
		{ID: "INIT-LITE-001", InitType: initiative.TypeRefactor, WorkflowID: "pbhq-lite"},
		{ID: "INIT-AWS-001", InitType: initiative.TypeFeature, WorkflowID: "aws-product"},
		{ID: "INIT-NULL-001", InitType: initiative.TypeFeature},
	} {
		if err := s.CreateInitiative(ctx, init); err != nil {
			t.Fatalf("seed initiative %s: %v", init.ID, err)
		}
	}

	res, err := specworkflow.SyncFromCatalog(ctx, s, loader)
	if err != nil {
		t.Fatalf("SyncFromCatalog: %v", err)
	}

	// Every catalog workflow now has a row; pbhq-lite was updated, not duplicated.
	rows, err := s.ListSpecWorkflows(ctx)
	if err != nil {
		t.Fatalf("ListSpecWorkflows: %v", err)
	}
	byID := map[string]*store.SpecWorkflow{}
	for _, r := range rows {
		byID[r.ID] = r
	}
	if len(loader.Available()) != len(rows) {
		t.Errorf("rows = %d, want %d (catalog size)", len(rows), len(loader.Available()))
	}
	if got := byID["pbhq-lite"]; got == nil {
		t.Fatal("pbhq-lite row missing after sync")
	} else if !slices.Equal(got.SpecsRequired, []string{"PRD.md", "TRD.md", "PLAN.md", "ROADMAP.md"}) {
		t.Errorf("pbhq-lite row not updated to canonical definition: %v", got.SpecsRequired)
	}

	// Retired rows: pbhq-standard remapped away then deleted; aws-working-backwards deleted.
	if _, ok := byID["pbhq-standard"]; ok {
		t.Error("pbhq-standard row should have been deleted")
	}
	if _, ok := byID["aws-working-backwards"]; ok {
		t.Error("aws-working-backwards row should have been deleted")
	}
	if got := res.Remapped["INIT-STD-001"]; got != "pbhq-lite" {
		t.Errorf("INIT-STD-001 remap = %q, want pbhq-lite", got)
	}
	remapped, err := s.GetInitiative(ctx, "INIT-STD-001")
	if err != nil {
		t.Fatalf("GetInitiative: %v", err)
	}
	if remapped.WorkflowID != "pbhq-lite" {
		t.Errorf("INIT-STD-001 WorkflowID = %q, want pbhq-lite", remapped.WorkflowID)
	}

	// The door rename remap: aws-product → aws-one-way-door. (Guards against
	// the remap keys themselves being renamed by a careless bulk edit — the
	// keys must stay the OLD ids or the remap is a no-op.)
	if got := res.Remapped["INIT-AWS-001"]; got != "aws-one-way-door" {
		t.Errorf("INIT-AWS-001 remap = %q, want aws-one-way-door", got)
	}

	// Idempotent: second run creates nothing, remaps nothing, deletes nothing.
	res2, err := specworkflow.SyncFromCatalog(ctx, s, loader)
	if err != nil {
		t.Fatalf("SyncFromCatalog (second run): %v", err)
	}
	if res2.Created != 0 || len(res2.Remapped) != 0 || len(res2.Deleted) != 0 {
		t.Errorf("second sync not idempotent: created=%d remapped=%d deleted=%d",
			res2.Created, len(res2.Remapped), len(res2.Deleted))
	}
}
