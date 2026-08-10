package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestBuildExecutionResponse(t *testing.T) {
	ctx := context.Background()
	ms := store.NewMemStore()

	// Create test data
	prog := &store.Program{ID: "prog-1", Name: "Test Program"}
	if err := ms.CreateProgram(ctx, prog); err != nil {
		t.Fatal(err)
	}

	init := &store.Initiative{
		ID:        "INIT-TEST-001",
		Title:     "Test Initiative",
		Status:    "in_progress",
		ProgramID: "prog-1",
	}
	if err := ms.CreateInitiative(ctx, init); err != nil {
		t.Fatal(err)
	}

	phase := &store.Phase{
		ID:             "phase-1",
		InitiativeID:   "INIT-TEST-001",
		Title:          "Phase 1",
		SequenceNumber: 1,
	}
	if err := ms.CreatePhase(ctx, phase); err != nil {
		t.Fatal(err)
	}

	rmi := &store.RoadmapItem{
		ID:             "RMI-TEST-001",
		InitiativeID:   "INIT-TEST-001",
		PhaseID:        "phase-1",
		Title:          "Test RMI",
		Status:         "completed",
		SequenceNumber: 1,
	}
	if err := ms.CreateRMI(ctx, rmi); err != nil {
		t.Fatal(err)
	}

	svc := &service.Service{Store: ms}

	resp, err := buildExecutionResponse(ctx, svc)
	if err != nil {
		t.Fatal(err)
	}

	// Verify structure
	if len(resp.Programs) != 1 {
		t.Errorf("expected 1 program, got %d", len(resp.Programs))
	}
	if len(resp.Initiatives) != 1 {
		t.Errorf("expected 1 initiative, got %d", len(resp.Initiatives))
	}
	if len(resp.Phases) != 1 {
		t.Errorf("expected 1 phase, got %d", len(resp.Phases))
	}
	if len(resp.RMIs) != 1 {
		t.Errorf("expected 1 RMI, got %d", len(resp.RMIs))
	}

	// Verify JSON serialization works
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	// Verify key fields in JSON
	var decoded ExecutionResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.Initiatives[0].Progress != 1.0 {
		t.Errorf("expected progress 1.0, got %f", decoded.Initiatives[0].Progress)
	}
	if decoded.Initiatives[0].ProgramName != "Test Program" {
		t.Errorf("expected program name 'Test Program', got %q", decoded.Initiatives[0].ProgramName)
	}
}

func TestBuildMaturityResponse(t *testing.T) {
	ctx := context.Background()
	ms := store.NewMemStore()

	model := &store.CapabilityModel{
		ID:       "test-model",
		Name:     "Test Model",
		MaxLevel: 5,
	}
	if err := ms.CreateCapabilityModel(ctx, model); err != nil {
		t.Fatal(err)
	}

	svc := &service.Service{Store: ms}

	resp, err := buildMaturityResponse(ctx, svc)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Models) != 1 {
		t.Errorf("expected 1 model, got %d", len(resp.Models))
	}

	// Verify JSON serialization
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}
	if len(data) == 0 {
		t.Error("empty JSON output")
	}
}

func TestBuildSpecsResponse(t *testing.T) {
	ctx := context.Background()
	ms := store.NewMemStore()

	wf := &store.SpecWorkflow{
		ID:            "pbhq-standard",
		Name:          "Standard Workflow",
		SpecsRequired: []string{"PRD.md", "TRD.md"},
	}
	if err := ms.CreateSpecWorkflow(ctx, wf); err != nil {
		t.Fatal(err)
	}

	svc := &service.Service{Store: ms}

	resp, err := buildSpecsResponse(ctx, svc)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Workflows) != 1 {
		t.Errorf("expected 1 workflow, got %d", len(resp.Workflows))
	}

	// Verify JSON serialization
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}
	if len(data) == 0 {
		t.Error("empty JSON output")
	}
}

func TestIsValidInitiativeID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"valid ID", "INIT-VISIONSTUDIO-001", true},
		{"empty", "", false},
		{"dot", ".", false},
		{"dot-dot", "..", false},
		{"path traversal", "../../etc/passwd", false},
		{"forward slash", "foo/bar", false},
		{"backslash", "foo\\bar", false},
		{"leading traversal no slash suffix", "..secret", true}, // no separator, so not traversal
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidInitiativeID(tt.id); got != tt.want {
				t.Errorf("isValidInitiativeID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestWithinDir(t *testing.T) {
	tests := []struct {
		name string
		root string
		path string
		want bool
	}{
		{"exact root", "/a/b", "/a/b", true},
		{"direct child", "/a/b", "/a/b/c", true},
		{"nested child", "/a/b", "/a/b/c/d.md", true},
		{"sibling that shares a prefix", "/a/b", "/a/bc", false},
		{"escapes via ..", "/a/b", "/a/b/../../etc/passwd", false},
		{"outright outside", "/a/b", "/etc/passwd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := withinDir(tt.root, tt.path); got != tt.want {
				t.Errorf("withinDir(%q, %q) = %v, want %v", tt.root, tt.path, got, tt.want)
			}
		})
	}
}
