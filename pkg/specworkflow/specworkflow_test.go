package specworkflow

import (
	"testing"
)

func TestDefaultLoaderListWorkflows(t *testing.T) {
	loader := DefaultLoader()
	workflows, err := loader.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(workflows) == 0 {
		t.Fatal("expected at least one workflow")
	}

	// Check that pbhq-lite is available
	found := false
	for _, w := range workflows {
		if w.ID == "pbhq-lite" {
			found = true
			if w.Name == "" {
				t.Error("pbhq-lite has empty name")
			}
			if len(w.SpecsRequired) == 0 {
				t.Error("pbhq-lite has no required specs")
			}
			t.Logf("pbhq-lite: required=%v, optional=%v", w.SpecsRequired, w.SpecsOptional)
			break
		}
	}
	if !found {
		t.Error("pbhq-lite workflow not found")
	}
}

func TestLoadPBHQLite(t *testing.T) {
	loader := DefaultLoader()
	loaded, err := loader.Load("pbhq-lite")
	if err != nil {
		t.Fatalf("Load(pbhq-lite) failed: %v", err)
	}
	if loaded.Workflow == nil {
		t.Fatal("loaded.Workflow is nil")
	}
	if loaded.Workflow.Name != "pbhq-lite" {
		t.Errorf("expected name pbhq-lite, got %s", loaded.Workflow.Name)
	}

	// Check rubrics loaded
	if len(loaded.Rubrics) == 0 {
		t.Error("no rubrics loaded")
	}
	if _, ok := loaded.Rubrics["prd"]; !ok {
		t.Error("prd rubric not loaded")
	}
}

func TestGetRubric(t *testing.T) {
	loader := DefaultLoader()
	rubric, err := loader.GetRubric("pbhq-lite", "prd")
	if err != nil {
		t.Fatalf("GetRubric() failed: %v", err)
	}
	if rubric == nil {
		t.Fatal("rubric is nil")
	}
	if rubric.Name == "" {
		t.Error("rubric has no name")
	}
	if len(rubric.Categories) == 0 {
		t.Error("rubric has no categories")
	}
	t.Logf("prd rubric: %s with %d categories", rubric.Name, len(rubric.Categories))
}

func TestGetSynthesisGuidance(t *testing.T) {
	loader := DefaultLoader()
	sources, guidance, err := loader.GetSynthesisGuidance("pbhq-lite", "trd")
	if err != nil {
		t.Fatalf("GetSynthesisGuidance() failed: %v", err)
	}
	if len(sources) == 0 {
		t.Error("no sources for trd synthesis")
	}
	if guidance == "" {
		t.Error("no guidance for trd synthesis")
	}
	t.Logf("trd synthesis: sources=%v, guidance=%q", sources, guidance)
}
