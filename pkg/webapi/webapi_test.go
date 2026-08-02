package webapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func seedStore(t *testing.T) *store.MemStore {
	t.Helper()
	ctx := context.Background()
	ms := store.NewMemStore()

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
	return ms
}

func TestBuildExecutionResponse(t *testing.T) {
	ctx := context.Background()
	ms := seedStore(t)
	svc := &service.Service{Store: ms}

	resp, err := BuildExecutionResponse(ctx, svc)
	if err != nil {
		t.Fatal(err)
	}

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

	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

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

	resp, err := BuildMaturityResponse(ctx, svc)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Models) != 1 {
		t.Errorf("expected 1 model, got %d", len(resp.Models))
	}

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

	resp, err := BuildSpecsResponse(ctx, svc)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Workflows) != 1 {
		t.Errorf("expected 1 workflow, got %d", len(resp.Workflows))
	}

	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}
	if len(data) == 0 {
		t.Error("empty JSON output")
	}
}

func TestBuildSpendResponseNoData(t *testing.T) {
	ctx := context.Background()
	ms := seedStore(t)
	svc := &service.Service{Store: ms}

	// Point at an empty temp dir: no events directory exists.
	resp, err := BuildSpendResponse(ctx, svc, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if resp.HasData {
		t.Error("expected hasData=false without an events directory")
	}
	if resp.DataNote == "" {
		t.Error("expected a data note explaining missing token data")
	}
}

func TestRegisterRoutesContract(t *testing.T) {
	ms := seedStore(t)
	connect := func() (*service.Service, func(), error) {
		return &service.Service{Store: ms}, func() {}, nil
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, connect, t.TempDir())
	srv := httptest.NewServer(mux)
	defer srv.Close()

	for _, path := range []string{"/api/execution", "/api/spend", "/api/maturity", "/api/specs"} {
		res, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		if res.StatusCode != http.StatusOK {
			t.Errorf("GET %s: expected 200, got %d", path, res.StatusCode)
		}
		if ct := res.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("GET %s: expected application/json, got %q", path, ct)
		}
		if origin := res.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("GET %s: expected CORS wildcard, got %q", path, origin)
		}
		var decoded map[string]any
		if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
			t.Errorf("GET %s: invalid JSON: %v", path, err)
		}
		if err := res.Body.Close(); err != nil {
			t.Fatalf("close body for %s: %v", path, err)
		}
	}
}
