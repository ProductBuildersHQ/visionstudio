package specworkflow_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/initiative"
	"github.com/ProductBuildersHQ/visionstudio/pkg/specworkflow"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestBuiltInWorkflows(t *testing.T) {
	workflows := specworkflow.BuiltInWorkflows()
	if len(workflows) != 4 {
		t.Fatalf("expected 4 built-in workflows, got %d", len(workflows))
	}
	wantIDs := map[string]bool{
		"quick-fix":             false,
		"pbhq-lite":             false,
		"pbhq-standard":         false,
		"aws-working-backwards": false,
	}
	for _, wf := range workflows {
		if _, ok := wantIDs[wf.ID]; !ok {
			t.Errorf("unexpected workflow ID %q", wf.ID)
		}
		wantIDs[wf.ID] = true
		if len(wf.SpecsRequired) == 0 {
			t.Errorf("workflow %q has no required specs", wf.ID)
		}
	}
	for id, found := range wantIDs {
		if !found {
			t.Errorf("expected built-in workflow %q not found", id)
		}
	}
}

func TestDefaultWorkflowForType(t *testing.T) {
	tests := []struct {
		initType string
		want     string
	}{
		{initiative.TypeMaintenance, "quick-fix"},
		{initiative.TypeRefactor, "pbhq-lite"},
		{initiative.TypeMigration, "pbhq-lite"},
		{initiative.TypeFeature, "pbhq-standard"},
		{initiative.TypeCompliance, "pbhq-standard"},
		{"", "pbhq-standard"},
	}
	for _, tt := range tests {
		if got := specworkflow.DefaultWorkflowForType(tt.initType); got != tt.want {
			t.Errorf("DefaultWorkflowForType(%q) = %q, want %q", tt.initType, got, tt.want)
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

func TestSeedBuiltIn(t *testing.T) {
	ctx := context.Background()
	s := store.NewMemStore()

	created, err := specworkflow.SeedBuiltIn(ctx, s)
	if err != nil {
		t.Fatalf("SeedBuiltIn: %v", err)
	}
	if created != 4 {
		t.Errorf("expected 4 created, got %d", created)
	}

	workflows, err := s.ListSpecWorkflows(ctx)
	if err != nil {
		t.Fatalf("ListSpecWorkflows: %v", err)
	}
	if len(workflows) != 4 {
		t.Errorf("expected 4 stored workflows, got %d", len(workflows))
	}

	// Re-seeding is idempotent.
	created, err = specworkflow.SeedBuiltIn(ctx, s)
	if err != nil {
		t.Fatalf("SeedBuiltIn (second run): %v", err)
	}
	if created != 0 {
		t.Errorf("expected 0 created on re-seed, got %d", created)
	}
}

func TestResolve(t *testing.T) {
	ctx := context.Background()
	s := store.NewMemStore()

	t.Run("falls back to built-in when store empty", func(t *testing.T) {
		init := &store.Initiative{ID: "INIT-A", InitType: initiative.TypeFeature}
		wf, err := specworkflow.Resolve(ctx, s, init)
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if wf.ID != "pbhq-standard" {
			t.Errorf("expected pbhq-standard, got %q", wf.ID)
		}
	})

	t.Run("uses explicit WorkflowID override", func(t *testing.T) {
		init := &store.Initiative{ID: "INIT-B", InitType: initiative.TypeMaintenance, WorkflowID: "aws-working-backwards"}
		wf, err := specworkflow.Resolve(ctx, s, init)
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if wf.ID != "aws-working-backwards" {
			t.Errorf("expected aws-working-backwards, got %q", wf.ID)
		}
	})

	t.Run("prefers stored workflow over built-in default", func(t *testing.T) {
		custom := &store.SpecWorkflow{
			ID:            "pbhq-standard",
			Name:          "Custom Standard",
			SpecsRequired: []string{"CUSTOM.md"},
		}
		if err := s.CreateSpecWorkflow(ctx, custom); err != nil {
			t.Fatalf("CreateSpecWorkflow: %v", err)
		}
		init := &store.Initiative{ID: "INIT-C", InitType: initiative.TypeFeature}
		wf, err := specworkflow.Resolve(ctx, s, init)
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if wf.Name != "Custom Standard" {
			t.Errorf("expected stored override to win, got %q", wf.Name)
		}
	})

	t.Run("errors on unknown workflow ID", func(t *testing.T) {
		init := &store.Initiative{ID: "INIT-D", InitType: initiative.TypeFeature, WorkflowID: "does-not-exist"}
		if _, err := specworkflow.Resolve(ctx, s, init); err == nil {
			t.Error("expected error for unknown workflow ID")
		}
	})
}
