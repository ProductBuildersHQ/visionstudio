package remote_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ProductBuildersHQ/visionstudio/pkg/mcpserver"
	"github.com/ProductBuildersHQ/visionstudio/pkg/remote/remotetest"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// TestMCPToolsAgainstRemote runs the unmodified MCP server over a remote
// store and drives it end to end through an MCP client: tool calls go
// MCP → service → remote store → HTTP → (fake) cloud API.
func TestMCPToolsAgainstRemote(t *testing.T) {
	fake := newFake(t)
	ctx := context.Background()

	server := mcpserver.New(service.New(newStore(t, fake)))
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

	// initiative_create → POST /initiatives
	text := callTool(t, cs, "initiative_create", map[string]any{
		"id": "INIT-ACME-001", "organization": "acme", "title": "Launch", "workflow_id": "pbhq-lite",
	})
	var created store.Initiative
	if err := json.Unmarshal([]byte(text), &created); err != nil {
		t.Fatalf("decode initiative_create: %v (%s)", err, text)
	}
	if created.ID != "INIT-ACME-001" || created.Status != "proposed" {
		t.Fatalf("initiative_create = %+v", created)
	}

	// Phases have no MCP tool; create one through the remote store.
	if _, err := service.New(newStore(t, fake)).CreatePhase(ctx, "INIT-ACME-001/phase-1", "INIT-ACME-001", 1, "Foundation", ""); err != nil {
		t.Fatal(err)
	}
	if err := fake.Service.Store.CreateRepository(ctx, &store.Repository{ID: "github.com/acme/app", Organization: "acme", RepositoryName: "app", Status: "active"}); err != nil {
		t.Fatal(err)
	}

	// rmi_create → POST /rmis
	text = callTool(t, cs, "rmi_create", map[string]any{
		"id": "RMI-ACME-001", "repository_id": "github.com/acme/app", "initiative_id": "INIT-ACME-001",
		"phase_id": "INIT-ACME-001/phase-1", "title": "Build it", "item_type": "capability",
	})
	if !strings.Contains(text, `"RMI-ACME-001"`) {
		t.Fatalf("rmi_create = %s", text)
	}

	// Both writes landed in the tenant's (server-side) store.
	if _, err := fake.Service.GetRMI(ctx, "RMI-ACME-001"); err != nil {
		t.Fatalf("server-side GetRMI: %v", err)
	}

	// initiative_list → GET /initiatives
	text = callTool(t, cs, "initiative_list", map[string]any{})
	var inits []store.Initiative
	if err := json.Unmarshal([]byte(text), &inits); err != nil {
		t.Fatalf("decode initiative_list: %v (%s)", err, text)
	}
	if len(inits) != 1 || inits[0].ID != "INIT-ACME-001" {
		t.Fatalf("initiative_list = %+v", inits)
	}

	// Every HTTP call carried the bearer credential.
	for _, r := range fake.Requests() {
		if r.Header.Get("Authorization") != "Bearer "+testToken {
			t.Fatalf("%s %s missing bearer token", r.Method, r.Path)
		}
	}
}

func callTool(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) string {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool %s: %v", name, err)
	}
	text := toolText(t, res)
	if res.IsError {
		t.Fatalf("CallTool %s returned error: %s", name, text)
	}
	return text
}

func toolText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatal("empty tool result")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content = %T, want TextContent", res.Content[0])
	}
	return tc.Text
}

// newMCPSession connects an MCP client to an unmodified MCP server running
// over a remote store against fake.
func newMCPSession(t *testing.T, fake *remotetest.Server) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	server := mcpserver.New(service.New(newStore(t, fake)))
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
	return cs
}

// TestMCPAgentWorkCycleAgainstRemote drives the agent work loop end to end
// in remote mode: work_ready → task_claim → task_update (status, evidence,
// handoff) → task_release → re-claim → task_update complete, with
// initiative_get reading the result — every store call crossing HTTP to the
// (fake) cloud API.
func TestMCPAgentWorkCycleAgainstRemote(t *testing.T) {
	fake := newFake(t)
	ctx := context.Background()
	seedWork(t, fake)
	for _, id := range []string{"RMI-ACME-001", "RMI-ACME-002"} {
		if _, err := fake.Service.UpdateRMIStatus(ctx, id, "ready"); err != nil {
			t.Fatal(err)
		}
	}
	// RMI-ACME-002 depends on RMI-ACME-001, so only 001 is ready to work.
	if err := fake.Service.CreateDependency(ctx, "RMI-ACME-002", "RMI-ACME-001", "requires"); err != nil {
		t.Fatal(err)
	}
	cs := newMCPSession(t, fake)

	text := callTool(t, cs, "work_ready", map[string]any{"initiative_id": "INIT-ACME-001"})
	if !strings.Contains(text, "RMI-ACME-001") || strings.Contains(text, "RMI-ACME-002") {
		t.Fatalf("work_ready = %s", text)
	}

	var claim struct {
		AssignmentID string `json:"assignment_id"`
		RMIID        string `json:"rmi_id"`
		TrailerLine  string `json:"trailer_line"`
	}
	text = callTool(t, cs, "task_claim", map[string]any{"rmi_id": "RMI-ACME-001", "worker": "session-1", "workspace": "/ws"})
	if err := json.Unmarshal([]byte(text), &claim); err != nil || claim.AssignmentID == "" || claim.TrailerLine == "" {
		t.Fatalf("task_claim = %s (%v)", text, err)
	}
	if srv, _ := fake.Service.GetRMI(ctx, "RMI-ACME-001"); srv.Status != "in_progress" {
		t.Fatalf("server-side status after claim = %q", srv.Status)
	}

	// A second session cannot claim the same RMI.
	msg := toolError(t, cs, "task_claim", map[string]any{"rmi_id": "RMI-ACME-001", "worker": "session-2"})
	if !strings.Contains(msg, "in_progress") {
		t.Fatalf("competing task_claim error = %q", msg)
	}

	text = callTool(t, cs, "task_update", map[string]any{
		"assignment_id": claim.AssignmentID,
		"evidence":      []map[string]any{{"type": "commit", "reference": "abc123"}, {"type": "pr", "reference": "https://example.com/pr/1"}},
		"handoff":       map[string]any{"completed": []string{"schema"}, "remaining": []string{"api"}, "next_action": "write handlers"},
	})
	if !strings.Contains(text, `"evidence_added": 2`) || !strings.Contains(text, `"handoff_updated": true`) {
		t.Fatalf("task_update = %s", text)
	}
	ev, err := fake.Service.Store.ListEvidenceByRMI(ctx, "RMI-ACME-001")
	if err != nil || len(ev) != 2 {
		t.Fatalf("server-side evidence = %+v, %v", ev, err)
	}

	text = callTool(t, cs, "task_release", map[string]any{"assignment_id": claim.AssignmentID, "handoff": map[string]any{"next_action": "resume api"}})
	if !strings.Contains(text, `"status": "released"`) {
		t.Fatalf("task_release = %s", text)
	}
	if srv, _ := fake.Service.GetRMI(ctx, "RMI-ACME-001"); srv.Status != "ready" {
		t.Fatalf("server-side status after release = %q", srv.Status)
	}

	text = callTool(t, cs, "task_claim", map[string]any{"rmi_id": "RMI-ACME-001", "worker": "session-2"})
	if err := json.Unmarshal([]byte(text), &claim); err != nil {
		t.Fatalf("re-claim = %s (%v)", text, err)
	}
	text = callTool(t, cs, "task_update", map[string]any{"assignment_id": claim.AssignmentID, "complete": true})
	if !strings.Contains(text, `"completed": true`) {
		t.Fatalf("task_update complete = %s", text)
	}
	if srv, _ := fake.Service.GetRMI(ctx, "RMI-ACME-001"); srv.Status != "completed" || srv.CompletedAt == nil {
		t.Fatalf("server-side RMI after complete = %+v", srv)
	}

	// The dependent RMI is now unblocked.
	text = callTool(t, cs, "work_ready", map[string]any{})
	if !strings.Contains(text, "RMI-ACME-002") {
		t.Fatalf("work_ready after completion = %s", text)
	}

	// task_update can also change status directly.
	text = callTool(t, cs, "task_update", map[string]any{"rmi_id": "RMI-ACME-002", "status": "blocked"})
	if !strings.Contains(text, `"rmi_status": "blocked"`) {
		t.Fatalf("task_update status = %s", text)
	}

	text = callTool(t, cs, "initiative_get", map[string]any{"id": "INIT-ACME-001"})
	for _, want := range []string{"INIT-ACME-001", "INIT-ACME-001/phase-1", "RMI-ACME-001", "completed"} {
		if !strings.Contains(text, want) {
			t.Fatalf("initiative_get missing %q: %s", want, text)
		}
	}
}

// toolError calls a tool that must fail and returns its error text.
func toolError(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) string {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	switch {
	case err != nil:
		return err.Error()
	case res.IsError:
		return toolText(t, res)
	}
	t.Fatalf("%s succeeded; want an error", name)
	return ""
}
