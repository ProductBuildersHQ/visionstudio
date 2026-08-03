package mcpserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ProductBuildersHQ/visionstudio/pkg/report"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func setup(t *testing.T) (*mcp.ClientSession, *service.Service) {
	t.Helper()
	ctx := context.Background()
	ms := store.NewMemStore()
	svc := service.New(ms)
	server := New(svc)

	t1, t2 := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, t1, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0.0.1"}, nil)
	cs, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cs.Close()
		_ = ss.Wait()
	})
	return cs, svc
}

func callTool(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) string {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool %s: %v", name, err)
	}
	if len(res.Content) == 0 {
		t.Fatalf("CallTool %s: empty content", name)
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("CallTool %s: expected TextContent, got %T", name, res.Content[0])
	}
	return tc.Text
}

func seedInitiative(t *testing.T, svc *service.Service) {
	t.Helper()
	ctx := context.Background()
	now := time.Now()

	if err := svc.Store.CreateRepository(ctx, &store.Repository{
		ID: "github.com/test/repo", Organization: "test",
		RepositoryName: "repo", DefaultBranch: "main", Status: "active",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.CreateInitiative(ctx, "INIT-TEST-001", "test", "Test Initiative", "A test", "high", "", "pbhq-lite"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-TEST-001/phase-1", "INIT-TEST-001", 1, "Foundation", ""); err != nil {
		t.Fatal(err)
	}

	if err := svc.Store.CreateRMI(ctx, &store.RoadmapItem{
		ID: "RMI-TEST-001", RepositoryID: "github.com/test/repo",
		InitiativeID: "INIT-TEST-001", PhaseID: "INIT-TEST-001/phase-1",
		Title: "First item", ItemType: "capability", Status: "ready",
		Required: true, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := svc.Store.CreateRMI(ctx, &store.RoadmapItem{
		ID: "RMI-TEST-002", RepositoryID: "github.com/test/repo",
		InitiativeID: "INIT-TEST-001", PhaseID: "INIT-TEST-001/phase-1",
		Title: "Second item", ItemType: "task", Status: "planned",
		Required: true, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestInitiativeList(t *testing.T) {
	cs, svc := setup(t)
	seedInitiative(t, svc)

	text := callTool(t, cs, "initiative_list", map[string]any{})
	var inits []*store.Initiative
	if err := json.Unmarshal([]byte(text), &inits); err != nil {
		t.Fatal(err)
	}
	if len(inits) != 1 {
		t.Fatalf("expected 1 initiative, got %d", len(inits))
	}
	if inits[0].ID != "INIT-TEST-001" {
		t.Fatalf("expected INIT-TEST-001, got %s", inits[0].ID)
	}
}

func TestInitiativeGet(t *testing.T) {
	cs, svc := setup(t)
	seedInitiative(t, svc)

	text := callTool(t, cs, "initiative_get", map[string]any{"id": "INIT-TEST-001"})

	var detail struct {
		Initiative struct {
			ID string `json:"ID"`
		}
		Phases []struct {
			Status string
			RMIs   []struct {
				ID string
			}
		}
	}
	if err := json.Unmarshal([]byte(text), &detail); err != nil {
		t.Fatal(err)
	}
	if detail.Initiative.ID != "INIT-TEST-001" {
		t.Fatalf("expected INIT-TEST-001, got %s", detail.Initiative.ID)
	}
	if len(detail.Phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(detail.Phases))
	}
	if len(detail.Phases[0].RMIs) != 2 {
		t.Fatalf("expected 2 RMIs, got %d", len(detail.Phases[0].RMIs))
	}
}

func TestInitiativeCreate(t *testing.T) {
	cs, _ := setup(t)

	text := callTool(t, cs, "initiative_create", map[string]any{
		"id":           "INIT-NEW-001",
		"organization": "neworg",
		"title":        "New Initiative",
		"workflow_id":  "pbhq-lite",
	})

	var init store.Initiative
	if err := json.Unmarshal([]byte(text), &init); err != nil {
		t.Fatal(err)
	}
	if init.ID != "INIT-NEW-001" {
		t.Fatalf("expected INIT-NEW-001, got %s", init.ID)
	}
	if init.Status != "proposed" {
		t.Fatalf("expected proposed status, got %s", init.Status)
	}
}

func TestRMICreate(t *testing.T) {
	cs, svc := setup(t)
	seedInitiative(t, svc)

	text := callTool(t, cs, "rmi_create", map[string]any{
		"id":            "RMI-TEST-003",
		"repository_id": "github.com/test/repo",
		"initiative_id": "INIT-TEST-001",
		"phase_id":      "INIT-TEST-001/phase-1",
		"title":         "Third item",
		"item_type":     "capability",
		"required":      false,
	})

	var rmi store.RoadmapItem
	if err := json.Unmarshal([]byte(text), &rmi); err != nil {
		t.Fatal(err)
	}
	if rmi.ID != "RMI-TEST-003" {
		t.Fatalf("expected RMI-TEST-003, got %s", rmi.ID)
	}
	if rmi.Required {
		t.Fatal("expected required=false")
	}
}

func TestWorkReady(t *testing.T) {
	cs, svc := setup(t)
	seedInitiative(t, svc)

	text := callTool(t, cs, "work_ready", map[string]any{})
	var ready []*store.RoadmapItem
	if err := json.Unmarshal([]byte(text), &ready); err != nil {
		t.Fatal(err)
	}
	if len(ready) != 1 {
		t.Fatalf("expected 1 ready item, got %d", len(ready))
	}
	if ready[0].ID != "RMI-TEST-001" {
		t.Fatalf("expected RMI-TEST-001, got %s", ready[0].ID)
	}
}

func TestWorkReadyFilter(t *testing.T) {
	cs, svc := setup(t)
	seedInitiative(t, svc)

	text := callTool(t, cs, "work_ready", map[string]any{
		"initiative_id": "INIT-NONEXISTENT",
	})
	var ready []*store.RoadmapItem
	if err := json.Unmarshal([]byte(text), &ready); err != nil {
		t.Fatal(err)
	}
	if len(ready) != 0 {
		t.Fatalf("expected 0 ready items, got %d", len(ready))
	}
}

func TestTaskClaimAndRelease(t *testing.T) {
	cs, svc := setup(t)
	seedInitiative(t, svc)

	// Claim
	text := callTool(t, cs, "task_claim", map[string]any{
		"rmi_id": "RMI-TEST-001",
		"worker": "test-session-1",
	})
	var claim struct {
		AssignmentID string `json:"assignment_id"`
		TrailerLine  string `json:"trailer_line"`
	}
	if err := json.Unmarshal([]byte(text), &claim); err != nil {
		t.Fatal(err)
	}
	if claim.AssignmentID == "" {
		t.Fatal("expected non-empty assignment_id")
	}
	if claim.TrailerLine != "Refs: RMI-TEST-001" {
		t.Fatalf("expected trailer 'Refs: RMI-TEST-001', got %q", claim.TrailerLine)
	}

	// Verify RMI is now in_progress
	rmi, err := svc.GetRMI(context.Background(), "RMI-TEST-001")
	if err != nil {
		t.Fatal(err)
	}
	if rmi.Status != "in_progress" {
		t.Fatalf("expected in_progress, got %s", rmi.Status)
	}

	// Release
	text = callTool(t, cs, "task_release", map[string]any{
		"assignment_id": claim.AssignmentID,
		"handoff": map[string]any{
			"completed":   []string{"step1"},
			"remaining":   []string{"step2"},
			"next_action": "continue",
		},
	})
	var release struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(text), &release); err != nil {
		t.Fatal(err)
	}
	if release.Status != "released" {
		t.Fatalf("expected released status, got %s", release.Status)
	}

	// Verify RMI is back to ready
	rmi, err = svc.GetRMI(context.Background(), "RMI-TEST-001")
	if err != nil {
		t.Fatal(err)
	}
	if rmi.Status != "ready" {
		t.Fatalf("expected ready, got %s", rmi.Status)
	}
}

func TestTaskUpdateComplete(t *testing.T) {
	cs, svc := setup(t)
	seedInitiative(t, svc)

	// Claim first
	claimText := callTool(t, cs, "task_claim", map[string]any{
		"rmi_id": "RMI-TEST-001",
		"worker": "test-session-1",
	})
	var claim struct {
		AssignmentID string `json:"assignment_id"`
	}
	if err := json.Unmarshal([]byte(claimText), &claim); err != nil {
		t.Fatal(err)
	}

	// Add evidence
	callTool(t, cs, "task_update", map[string]any{
		"rmi_id": "RMI-TEST-001",
		"evidence": []map[string]any{
			{"type": "commit", "reference": "abc123"},
			{"type": "pr", "reference": "https://github.com/test/repo/pull/1"},
		},
	})

	// Complete
	text := callTool(t, cs, "task_update", map[string]any{
		"assignment_id": claim.AssignmentID,
		"complete":      true,
		"handoff": map[string]any{
			"completed":   []string{"all work"},
			"remaining":   []string{},
			"next_action": "none",
		},
	})
	var result struct {
		Completed bool   `json:"completed"`
		RMIID     string `json:"rmi_id"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Completed {
		t.Fatal("expected completed=true")
	}

	// Verify RMI is completed
	rmi, err := svc.GetRMI(context.Background(), "RMI-TEST-001")
	if err != nil {
		t.Fatal(err)
	}
	if rmi.Status != "completed" {
		t.Fatalf("expected completed, got %s", rmi.Status)
	}
}

func TestReportInitiative(t *testing.T) {
	cs, svc := setup(t)
	seedInitiative(t, svc)

	text := callTool(t, cs, "report_initiative", map[string]any{"id": "INIT-TEST-001"})

	var r report.Report
	if err := json.Unmarshal([]byte(text), &r); err != nil {
		t.Fatal(err)
	}
	if r.InitiativeID != "INIT-TEST-001" {
		t.Fatalf("expected INIT-TEST-001, got %s", r.InitiativeID)
	}
	if len(r.Phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(r.Phases))
	}
	if r.RMIs.Total != 2 {
		t.Fatalf("expected 2 total RMIs, got %d", r.RMIs.Total)
	}
	if r.RMIs.Ready != 1 {
		t.Fatalf("expected 1 ready, got %d", r.RMIs.Ready)
	}
}

func TestTaskUpdateStatusOnly(t *testing.T) {
	cs, svc := setup(t)
	seedInitiative(t, svc)

	text := callTool(t, cs, "task_update", map[string]any{
		"rmi_id": "RMI-TEST-002",
		"status": "ready",
	})
	var result struct {
		RMIStatus string `json:"rmi_status"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatal(err)
	}
	if result.RMIStatus != "ready" {
		t.Fatalf("expected ready, got %s", result.RMIStatus)
	}
}

func TestContextBuildPhase(t *testing.T) {
	cs, svc := setup(t)
	seedInitiative(t, svc)

	text := callTool(t, cs, "context_build", map[string]any{
		"target_id": "INIT-TEST-001/phase-1",
		"format":    "json",
	})

	var pkg struct {
		Version    string `json:"version"`
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
		Sections   struct {
			Phase struct {
				ID string `json:"id"`
			} `json:"phase"`
		} `json:"sections"`
	}
	if err := json.Unmarshal([]byte(text), &pkg); err != nil {
		t.Fatal(err)
	}
	if pkg.TargetType != "phase" {
		t.Fatalf("expected phase target_type, got %s", pkg.TargetType)
	}
	if pkg.TargetID != "INIT-TEST-001/phase-1" {
		t.Fatalf("expected INIT-TEST-001/phase-1, got %s", pkg.TargetID)
	}
	if pkg.Sections.Phase.ID != "INIT-TEST-001/phase-1" {
		t.Fatalf("expected phase ID INIT-TEST-001/phase-1, got %s", pkg.Sections.Phase.ID)
	}
}

func TestContextBuildRMI(t *testing.T) {
	cs, svc := setup(t)
	seedInitiative(t, svc)

	text := callTool(t, cs, "context_build", map[string]any{
		"target_id": "RMI-TEST-001",
	})

	var pkg struct {
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
		Sections   struct {
			CurrentRMI struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"current_rmi"`
		} `json:"sections"`
	}
	if err := json.Unmarshal([]byte(text), &pkg); err != nil {
		t.Fatal(err)
	}
	if pkg.TargetType != "rmi" {
		t.Fatalf("expected rmi target_type, got %s", pkg.TargetType)
	}
	if pkg.Sections.CurrentRMI.ID != "RMI-TEST-001" {
		t.Fatalf("expected RMI-TEST-001, got %s", pkg.Sections.CurrentRMI.ID)
	}
}

func TestContextBuildMarkdown(t *testing.T) {
	cs, svc := setup(t)
	seedInitiative(t, svc)

	text := callTool(t, cs, "context_build", map[string]any{
		"target_id": "RMI-TEST-001",
		"format":    "markdown",
	})

	if !strings.Contains(text, "# Context Package:") {
		t.Error("markdown output missing title")
	}
	if !strings.Contains(text, "RMI-TEST-001") {
		t.Error("markdown output missing RMI ID")
	}
}
