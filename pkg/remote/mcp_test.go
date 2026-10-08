package remote_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ProductBuildersHQ/visionstudio/pkg/mcpserver"
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
