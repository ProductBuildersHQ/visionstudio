package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/apitypes"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/specworkflow"
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
	svc := &service.Service{Store: ms}

	resp, err := buildSpecsResponse(ctx, svc)
	if err != nil {
		t.Fatal(err)
	}

	// Workflows come from the visionspec catalog, not the DB.
	wantCount := len(specworkflow.DefaultLoader().Available())
	if len(resp.Workflows) != wantCount {
		t.Errorf("expected %d catalog workflows, got %d", wantCount, len(resp.Workflows))
	}
	byID := map[string]apitypes.SpecWorkflow{}
	for _, w := range resp.Workflows {
		byID[w.ID] = w
	}
	for _, id := range []string{"pbhq-lite", "quick-fix", "aws-one-way-door", "aws-two-way-door"} {
		if _, ok := byID[id]; !ok {
			t.Errorf("catalog workflow %q missing from response", id)
		}
	}
	if aws := byID["aws-two-way-door"]; len(aws.Sequence) == 0 || len(aws.Phases) == 0 {
		t.Errorf("aws-two-way-door should include sequence and phases, got sequence=%v phases=%v",
			aws.Sequence, aws.Phases)
	}
	if got := byID["pbhq-lite"].SpecsRequired; len(got) != 4 || got[0] != "PRD.md" {
		t.Errorf("pbhq-lite required = %v, want [PRD.md TRD.md PLAN.md ROADMAP.md]", got)
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
		{"dots not allowed by allowlist", "..secret", false},
		{"space not allowed by allowlist", "foo bar", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidInitiativeID(tt.id); got != tt.want {
				t.Errorf("isValidInitiativeID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestReadSpecFilesFlatStructure(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "PRD.md"), []byte("# PRD"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "evaluations"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "evaluations", "prd.eval.json"), []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}

	files, err := readSpecFiles(dir, "INIT-TEST-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Content != "# PRD" {
		t.Errorf("unexpected content: %q", files[0].Content)
	}
	if files[0].EvalJSON != "{}" {
		t.Errorf("expected eval JSON to be picked up, got %q", files[0].EvalJSON)
	}
}

func TestReadSpecFilesVisionSpecStructure(t *testing.T) {
	dir := t.TempDir()
	sourceDir := filepath.Join(dir, "source")
	evalDir := filepath.Join(dir, "eval")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(evalDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "prd.md"), []byte("# PRD"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(evalDir, "prd.json"), []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}

	files, err := readSpecFiles(dir, "INIT-TEST-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Content != "# PRD" {
		t.Errorf("unexpected content: %q", files[0].Content)
	}
	if files[0].EvalJSON != "{}" {
		t.Errorf("expected eval JSON to be picked up, got %q", files[0].EvalJSON)
	}
}

func TestHandleCreateInitiative(t *testing.T) {
	newReq := func(body string) *http.Request {
		return httptest.NewRequest(http.MethodPost, "/api/initiatives", strings.NewReader(body))
	}
	newSvc := func() *service.Service {
		return &service.Service{Store: store.NewMemStore()}
	}

	t.Run("creates with valid workflow", func(t *testing.T) {
		svc := newSvc()
		resp, status, err := handleCreateInitiative(newReq(`{
			"id": "INIT-TEST-001",
			"title": "Test Initiative",
			"initType": "feature",
			"workflowId": "aws-two-way-door",
			"description": "desc"
		}`), svc)
		if err != nil {
			t.Fatalf("handleCreateInitiative: %v (status %d)", err, status)
		}
		if resp.ID != "INIT-TEST-001" || resp.Status != "proposed" {
			t.Errorf("resp = %+v, want ID=INIT-TEST-001 status=proposed", resp)
		}
		init, err := svc.Store.GetInitiative(context.Background(), "INIT-TEST-001")
		if err != nil {
			t.Fatalf("GetInitiative: %v", err)
		}
		if init.WorkflowID != "aws-two-way-door" {
			t.Errorf("WorkflowID = %q, want aws-two-way-door", init.WorkflowID)
		}
	})

	t.Run("rejects unknown workflow", func(t *testing.T) {
		_, status, err := handleCreateInitiative(newReq(`{
			"id": "INIT-TEST-002", "title": "T", "workflowId": "no-such-workflow"
		}`), newSvc())
		if err == nil || status != http.StatusBadRequest {
			t.Errorf("expected 400 for unknown workflow, got status=%d err=%v", status, err)
		}
	})

	t.Run("rejects missing fields and bad IDs", func(t *testing.T) {
		for _, body := range []string{
			`{"title": "T", "workflowId": "pbhq-lite"}`,
			`{"id": "INIT-X-001", "workflowId": "pbhq-lite"}`,
			`{"id": "INIT-X-001", "title": "T"}`,
			`{"id": "../evil", "title": "T", "workflowId": "pbhq-lite"}`,
		} {
			if _, status, err := handleCreateInitiative(newReq(body), newSvc()); err == nil || status != http.StatusBadRequest {
				t.Errorf("body %s: expected 400, got status=%d err=%v", body, status, err)
			}
		}
	})

	t.Run("conflict on duplicate", func(t *testing.T) {
		svc := newSvc()
		body := `{"id": "INIT-DUP-001", "title": "T", "workflowId": "pbhq-lite"}`
		if _, _, err := handleCreateInitiative(newReq(body), svc); err != nil {
			t.Fatalf("first create: %v", err)
		}
		if _, status, err := handleCreateInitiative(newReq(body), svc); err == nil || status != http.StatusConflict {
			t.Errorf("expected 409 on duplicate, got status=%d err=%v", status, err)
		}
	})

	t.Run("sets program when provided", func(t *testing.T) {
		svc := newSvc()
		_, _, err := handleCreateInitiative(newReq(`{
			"id": "INIT-PROG-001", "title": "T", "workflowId": "pbhq-lite", "programId": "PROG-X"
		}`), svc)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		init, err := svc.Store.GetInitiative(context.Background(), "INIT-PROG-001")
		if err != nil {
			t.Fatal(err)
		}
		if init.ProgramID != "PROG-X" {
			t.Errorf("ProgramID = %q, want PROG-X", init.ProgramID)
		}
	})
}

func TestBuildWorkflowSpecDetail(t *testing.T) {
	t.Run("template and rubric for aws faq", func(t *testing.T) {
		d, err := buildWorkflowSpecDetail("aws-one-way-door", "FAQ")
		if err != nil {
			t.Fatal(err)
		}
		if d.Template == "" {
			t.Error("expected FAQ template content")
		}
		if !strings.Contains(d.RubricJSON, "disconfirmation_rigor") {
			t.Error("expected FAQ rubric JSON with LP category disconfirmation_rigor")
		}
		if d.SpecType != "FAQ" {
			t.Errorf("SpecType = %q, want FAQ", d.SpecType)
		}
	})

	t.Run("rubric only when no template (pbhq-lite prd)", func(t *testing.T) {
		d, err := buildWorkflowSpecDetail("pbhq-lite", "prd")
		if err != nil {
			t.Fatal(err)
		}
		if d.RubricJSON == "" {
			t.Error("expected pbhq-lite prd rubric")
		}
	})

	t.Run("unknown workflow errors", func(t *testing.T) {
		if _, err := buildWorkflowSpecDetail("no-such-workflow", "prd"); err == nil {
			t.Error("expected error for unknown workflow")
		}
	})
}
