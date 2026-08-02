// Package mcpserver exposes PRISM Control as an MCP server.
// It registers the 9 core tools from TRD §7 and delegates to the
// shared service layer so protocol rules never diverge from the CLI.
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ProductBuildersHQ/visionstudio/pkg/contextbuild"
	"github.com/ProductBuildersHQ/visionstudio/pkg/report"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// New creates an MCP server with all PRISM Control tools registered.
func New(svc *service.Service) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{Name: "vistudio", Version: "0.1.0"},
		&mcp.ServerOptions{
			Instructions: "PRISM Control — Product Delivery Control Plane. " +
				"Browse initiatives, claim roadmap items, and update work status.",
		},
	)

	registerTools(s, svc)
	return s
}

// Run starts the MCP server on stdio, blocking until the client disconnects.
func Run(ctx context.Context, svc *service.Service) error {
	s := New(svc)
	return s.Run(ctx, &mcp.StdioTransport{})
}

func registerTools(s *mcp.Server, svc *service.Service) {
	s.AddTool(programListTool(), programListHandler(svc))
	s.AddTool(programCreateTool(), programCreateHandler(svc))
	s.AddTool(initiativeListTool(), initiativeListHandler(svc))
	s.AddTool(initiativeGetTool(), initiativeGetHandler(svc))
	s.AddTool(initiativeCreateTool(), initiativeCreateHandler(svc))
	s.AddTool(rmiCreateTool(), rmiCreateHandler(svc))
	s.AddTool(workReadyTool(), workReadyHandler(svc))
	s.AddTool(taskClaimTool(), taskClaimHandler(svc))
	s.AddTool(taskReleaseTool(), taskReleaseHandler(svc))
	s.AddTool(taskUpdateTool(), taskUpdateHandler(svc))
	s.AddTool(reportInitiativeTool(), reportInitiativeHandler(svc))
	s.AddTool(contextBuildTool(), contextBuildHandler(svc))
}

// ---------- program_list ----------

func programListTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "program_list",
		Description: "List all programs.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
	}
}

func programListHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		progs, err := svc.ListPrograms(ctx)
		if err != nil {
			return nil, err
		}
		return jsonResult(progs)
	}
}

// ---------- program_create ----------

func programCreateTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "program_create",
		Description: "Create a new program for grouping related initiatives.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"id":{"type":"string","description":"Program ID (e.g. PROG-DELIVERY)"},
				"name":{"type":"string","description":"Human-readable program name"},
				"organization":{"type":"string","description":"Organization name"},
				"description":{"type":"string","description":"Program description"}
			},
			"required":["id","name","organization"]
		}`),
	}
}

func programCreateHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Organization string `json:"organization"`
			Description  string `json:"description"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		prog, err := svc.CreateProgram(ctx, args.ID, args.Name, args.Organization, args.Description)
		if err != nil {
			return nil, err
		}
		return jsonResult(prog)
	}
}

// ---------- initiative_list ----------

func initiativeListTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "initiative_list",
		Description: "List all initiatives with their current status.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
	}
}

func initiativeListHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		inits, err := svc.ListInitiatives(ctx)
		if err != nil {
			return nil, err
		}
		return jsonResult(inits)
	}
}

// ---------- initiative_get ----------

func initiativeGetTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "initiative_get",
		Description: "Get initiative detail including phases with derived status and RMI breakdown.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"id":{"type":"string","description":"Initiative ID (e.g. INIT-PRISMCONTROL-001)"}
			},
			"required":["id"]
		}`),
	}
}

func initiativeGetHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		detail, err := svc.GetInitiativeDetail(ctx, args.ID)
		if err != nil {
			return nil, err
		}
		return jsonResult(detail)
	}
}

// ---------- initiative_create ----------

func initiativeCreateTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "initiative_create",
		Description: "Create a new initiative in proposed status.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"id":{"type":"string","description":"Initiative ID (e.g. INIT-MYPROJECT-001)"},
				"organization":{"type":"string","description":"Organization name"},
				"title":{"type":"string","description":"Short title"},
				"description":{"type":"string","description":"Full description"},
				"priority":{"type":"string","description":"Priority level"},
				"program_id":{"type":"string","description":"Program ID to associate (e.g. PROG-DELIVERY)"}
			},
			"required":["id","organization","title"]
		}`),
	}
}

func initiativeCreateHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			ID           string `json:"id"`
			Organization string `json:"organization"`
			Title        string `json:"title"`
			Description  string `json:"description"`
			Priority     string `json:"priority"`
			InitType     string `json:"init_type"`
			ProgramID    string `json:"program_id"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		init, err := svc.CreateInitiative(ctx, args.ID, args.Organization, args.Title, args.Description, args.Priority, args.InitType)
		if err != nil {
			return nil, err
		}
		if args.ProgramID != "" {
			init.ProgramID = args.ProgramID
			if err := svc.UpdateInitiative(ctx, init); err != nil {
				return nil, err
			}
		}
		return jsonResult(init)
	}
}

// ---------- rmi_create ----------

func rmiCreateTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "rmi_create",
		Description: "Create a new Roadmap Item (RMI) in proposed status.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"id":{"type":"string","description":"RMI ID (e.g. RMI-MYREPO-001)"},
				"repository_id":{"type":"string","description":"Repository ID (e.g. github.com/org/repo)"},
				"initiative_id":{"type":"string","description":"Parent initiative ID"},
				"phase_id":{"type":"string","description":"Phase ID within the initiative"},
				"title":{"type":"string","description":"Short title"},
				"description":{"type":"string","description":"Full description"},
				"item_type":{"type":"string","enum":["capability","task","spec","release"],"description":"Type of work"},
				"priority":{"type":"string","description":"Priority level"},
				"required":{"type":"boolean","description":"Whether this RMI is required for phase completion","default":true},
				"sequence_number":{"type":"integer","description":"Order within the phase"},
				"acceptance_criteria":{"type":"array","items":{"type":"string"},"description":"List of acceptance criteria"},
				"context_spec":{"type":"object","description":"Context assembly overrides","properties":{"extra_repos":{"type":"array","items":{"type":"string"},"description":"Additional repos to include"},"include_specs":{"type":"array","items":{"type":"string"},"description":"Spec files to include"},"exclude_specs":{"type":"array","items":{"type":"string"},"description":"Spec files to exclude"}}}
			},
			"required":["id","repository_id","title","item_type"]
		}`),
	}
}

func rmiCreateHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			ID                 string             `json:"id"`
			RepositoryID       string             `json:"repository_id"`
			InitiativeID       string             `json:"initiative_id"`
			PhaseID            string             `json:"phase_id"`
			Title              string             `json:"title"`
			Description        string             `json:"description"`
			ItemType           string             `json:"item_type"`
			Priority           string             `json:"priority"`
			Required           *bool              `json:"required"`
			SequenceNumber     int                `json:"sequence_number"`
			AcceptanceCriteria []string           `json:"acceptance_criteria"`
			ContextSpec        *store.ContextSpec `json:"context_spec"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		required := true
		if args.Required != nil {
			required = *args.Required
		}
		rmi, err := svc.CreateRMI(ctx,
			args.ID, args.RepositoryID, args.InitiativeID, args.PhaseID,
			args.Title, args.Description, args.ItemType, args.Priority,
			required, args.SequenceNumber, args.AcceptanceCriteria,
		)
		if err != nil {
			return nil, err
		}
		if args.ContextSpec != nil {
			rmi.ContextSpec = args.ContextSpec
			if err := svc.UpdateRMI(ctx, rmi); err != nil {
				return nil, err
			}
		}
		return jsonResult(rmi)
	}
}

// ---------- work_ready ----------

func workReadyTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "work_ready",
		Description: "List RMIs that are ready, unblocked by dependencies, and unclaimed. Use filters to narrow by initiative or repository.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"initiative_id":{"type":"string","description":"Filter by initiative ID"},
				"repository_id":{"type":"string","description":"Filter by repository ID"}
			}
		}`),
	}
}

func workReadyHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			InitiativeID string `json:"initiative_id"`
			RepositoryID string `json:"repository_id"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		ready, err := svc.WorkReady(ctx, service.WorkReadyFilters{
			InitiativeID: args.InitiativeID,
			RepoID:       args.RepositoryID,
		})
		if err != nil {
			return nil, err
		}
		return jsonResult(ready)
	}
}

// ---------- task_claim ----------

func taskClaimTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "task_claim",
		Description: "Claim an RMI for work, creating a lease-based assignment. Returns the git trailer line to use in commits.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"rmi_id":{"type":"string","description":"RMI ID to claim"},
				"worker":{"type":"string","description":"Worker/session identifier"},
				"workspace":{"type":"string","description":"Workspace path or identifier"},
				"lease_hours":{"type":"integer","description":"Lease duration in hours (default 4)","default":4}
			},
			"required":["rmi_id","worker"]
		}`),
	}
}

func taskClaimHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			RMIID      string `json:"rmi_id"`
			Worker     string `json:"worker"`
			Workspace  string `json:"workspace"`
			LeaseHours int    `json:"lease_hours"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		if args.LeaseHours == 0 {
			args.LeaseHours = 4
		}
		lease := time.Duration(args.LeaseHours) * time.Hour
		result, err := svc.ClaimRMI(ctx, args.RMIID, args.Worker, args.Workspace, lease)
		if err != nil {
			return nil, err
		}
		return jsonResult(map[string]any{
			"assignment_id": result.Assignment.ID,
			"rmi_id":        result.Assignment.RMIID,
			"worker":        result.Assignment.Worker,
			"lease_expires": result.Assignment.LeaseExpiresAt,
			"trailer_line":  result.TrailerLine,
		})
	}
}

// ---------- task_release ----------

func taskReleaseTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "task_release",
		Description: "Release a work claim. The RMI returns to ready status. Optionally include handoff notes for the next session.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"assignment_id":{"type":"string","description":"Assignment ID to release"},
				"handoff":{"type":"object","properties":{
					"completed":{"type":"array","items":{"type":"string"}},
					"remaining":{"type":"array","items":{"type":"string"}},
					"decisions":{"type":"array","items":{"type":"string"}},
					"next_action":{"type":"string"}
				},"description":"Handoff state for the next session"}
			},
			"required":["assignment_id"]
		}`),
	}
}

func taskReleaseHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			AssignmentID string         `json:"assignment_id"`
			Handoff      *store.Handoff `json:"handoff"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		a, err := svc.ReleaseWork(ctx, args.AssignmentID, args.Handoff)
		if err != nil {
			return nil, err
		}
		return jsonResult(map[string]any{
			"assignment_id": a.ID,
			"rmi_id":        a.RMIID,
			"status":        a.Status,
		})
	}
}

// ---------- task_update ----------

func taskUpdateTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "task_update",
		Description: "Update a task: change RMI status, add evidence, update handoff, or mark complete. Supports multiple operations in one call.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"assignment_id":{"type":"string","description":"Assignment ID (for handoff/complete operations)"},
				"rmi_id":{"type":"string","description":"RMI ID (for status change or evidence)"},
				"status":{"type":"string","enum":["proposed","planned","ready","in_progress","completed","blocked","cancelled"],"description":"New RMI status"},
				"complete":{"type":"boolean","description":"Mark the assignment as completed"},
				"evidence":{"type":"array","items":{"type":"object","properties":{
					"type":{"type":"string","enum":["commit","pr","release","changelog","test"]},
					"reference":{"type":"string"}
				},"required":["type","reference"]},"description":"Evidence to attach"},
				"handoff":{"type":"object","properties":{
					"completed":{"type":"array","items":{"type":"string"}},
					"remaining":{"type":"array","items":{"type":"string"}},
					"decisions":{"type":"array","items":{"type":"string"}},
					"next_action":{"type":"string"}
				},"description":"Handoff state update"}
			}
		}`),
	}
}

func taskUpdateHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			AssignmentID string `json:"assignment_id"`
			RMIID        string `json:"rmi_id"`
			Status       string `json:"status"`
			Complete     bool   `json:"complete"`
			Evidence     []struct {
				Type      string `json:"type"`
				Reference string `json:"reference"`
			} `json:"evidence"`
			Handoff *store.Handoff `json:"handoff"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}

		results := make(map[string]any)

		if args.Status != "" && args.RMIID != "" {
			rmi, err := svc.UpdateRMIStatus(ctx, args.RMIID, args.Status)
			if err != nil {
				return nil, fmt.Errorf("update status: %w", err)
			}
			results["rmi_status"] = rmi.Status
		}

		for _, ev := range args.Evidence {
			rmiID := args.RMIID
			if rmiID == "" && args.AssignmentID != "" {
				a, err := svc.GetAssignment(ctx, args.AssignmentID)
				if err != nil {
					return nil, fmt.Errorf("get assignment for evidence: %w", err)
				}
				rmiID = a.RMIID
			}
			if rmiID == "" {
				return nil, fmt.Errorf("rmi_id or assignment_id required to add evidence")
			}
			if _, err := svc.AddEvidence(ctx, rmiID, ev.Type, ev.Reference); err != nil {
				return nil, fmt.Errorf("add evidence: %w", err)
			}
		}
		if len(args.Evidence) > 0 {
			results["evidence_added"] = len(args.Evidence)
		}

		if args.Handoff != nil && args.AssignmentID != "" && !args.Complete {
			if _, err := svc.UpdateHandoff(ctx, args.AssignmentID, args.Handoff); err != nil {
				return nil, fmt.Errorf("update handoff: %w", err)
			}
			results["handoff_updated"] = true
		}

		if args.Complete && args.AssignmentID != "" {
			a, err := svc.CompleteWork(ctx, args.AssignmentID, args.Handoff)
			if err != nil {
				return nil, fmt.Errorf("complete: %w", err)
			}
			results["completed"] = true
			results["rmi_id"] = a.RMIID
		}

		if len(results) == 0 {
			results["message"] = "no operations performed — provide status, evidence, handoff, or complete"
		}

		return jsonResult(results)
	}
}

// ---------- report_initiative ----------

func reportInitiativeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "report_initiative",
		Description: "Generate an end-to-end initiative report: phases, RMI progress, assignments, and evidence summary.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"id":{"type":"string","description":"Initiative ID"}
			},
			"required":["id"]
		}`),
	}
}

func reportInitiativeHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}

		r, err := report.Generate(ctx, svc.Store, args.ID)
		if err != nil {
			return nil, err
		}
		return jsonResult(r)
	}
}

// ---------- context_build ----------

func contextBuildTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "context_build",
		Description: `Build a deterministic context package for a phase or RMI.

For a phase (e.g., INIT-FOO-001/phase-1): returns program/initiative context,
phase objectives and member RMIs, prerequisite phase handoffs, spec file
references with git revisions, and derived repository set.

For an RMI (e.g., RMI-REPO-001): includes all of the above plus current RMI
details with acceptance criteria and active assignment state.

Output is byte-identical across runs at the same Dolt/git revisions.`,
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"target_id":{"type":"string","description":"Phase ID (e.g. INIT-FOO-001/phase-1) or RMI ID (e.g. RMI-REPO-001)"},
				"format":{"type":"string","enum":["json","markdown"],"default":"json","description":"Output format"}
			},
			"required":["target_id"]
		}`),
	}
}

func contextBuildHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			TargetID string `json:"target_id"`
			Format   string `json:"format"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}

		if args.Format == "" {
			args.Format = "json"
		}

		builder := contextbuild.NewBuilder(svc.Store, "unknown")

		if isPhaseID(args.TargetID) {
			pkg, err := builder.BuildForPhase(ctx, args.TargetID)
			if err != nil {
				return nil, fmt.Errorf("build phase context: %w", err)
			}
			return renderContextResult(pkg, args.Format)
		}

		pkg, err := builder.BuildForRMI(ctx, args.TargetID)
		if err != nil {
			return nil, fmt.Errorf("build RMI context: %w", err)
		}
		return renderContextResult(pkg, args.Format)
	}
}

func isPhaseID(id string) bool {
	return strings.Contains(id, "/phase-")
}

func renderContextResult(pkg *contextbuild.ContextPackage, format string) (*mcp.CallToolResult, error) {
	var content string
	switch format {
	case "markdown":
		content = pkg.RenderMarkdown()
	default:
		data, err := pkg.RenderJSON()
		if err != nil {
			return nil, fmt.Errorf("render JSON: %w", err)
		}
		content = string(data)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: content}},
	}, nil
}

// ---------- helpers ----------

func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil
}
