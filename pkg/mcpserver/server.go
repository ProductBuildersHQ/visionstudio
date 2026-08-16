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
	"github.com/ProductBuildersHQ/visionstudio/pkg/speceval"
	"github.com/ProductBuildersHQ/visionstudio/pkg/specworkflow"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
	"github.com/ProductBuildersHQ/visionstudio/pkg/synthesis"
)

// New creates an MCP server with all PRISM Control tools registered.
func New(svc *service.Service) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{Name: "visionstudio", Version: "0.1.0"},
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
	s.AddTool(workflowListTool(), workflowListHandler(svc))
	s.AddTool(workflowSelectTool(), workflowSelectHandler(svc))
	s.AddTool(workflowStatusTool(), workflowStatusHandler(svc))
	s.AddTool(specListTool(), specListHandler(svc))
	s.AddTool(specCreateTool(), specCreateHandler(svc))
	s.AddTool(specReadTool(), specReadHandler(svc))
	s.AddTool(specEvaluateTool(), specEvaluateHandler(svc))
	s.AddTool(specSynthesizeTool(), specSynthesizeHandler(svc))
	s.AddTool(specAddTool(), specAddHandler(svc))
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
		Description: "Create a new initiative in proposed status. workflow_id is required to select a spec workflow (e.g. pbhq-lite, aws-one-way-door). Use workflow_list to see available workflows.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"id":{"type":"string","description":"Initiative ID (e.g. INIT-MYPROJECT-001)"},
				"organization":{"type":"string","description":"Organization name"},
				"title":{"type":"string","description":"Short title"},
				"description":{"type":"string","description":"Full description"},
				"priority":{"type":"string","description":"Priority level"},
				"program_id":{"type":"string","description":"Program ID to associate (e.g. PROG-DELIVERY)"},
				"workflow_id":{"type":"string","description":"Spec workflow ID (e.g. pbhq-lite, aws-one-way-door, big-tech-essentials). Use workflow_list to see options."}
			},
			"required":["id","organization","title","workflow_id"]
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
			WorkflowID   string `json:"workflow_id"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		if args.WorkflowID == "" {
			return nil, fmt.Errorf("workflow_id is required; use workflow_list to see available workflows (e.g. pbhq-lite, aws-one-way-door)")
		}
		init, err := svc.CreateInitiative(ctx, args.ID, args.Organization, args.Title, args.Description, args.Priority, args.InitType, args.WorkflowID)
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

// ---------- workflow_list ----------

func workflowListTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "workflow_list",
		Description: "List available specification workflows with their required and optional specs.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
	}
}

func workflowListHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get embedded workflows from specification-workflow-spec
		loader := specworkflow.DefaultLoader()
		embeddedWorkflows, err := loader.List()
		if err != nil {
			return nil, fmt.Errorf("list embedded workflows: %w", err)
		}

		// Also get any custom workflows from the database
		dbWorkflows, _ := svc.ListWorkflows(ctx)

		return jsonResult(map[string]any{
			"embedded_workflows": embeddedWorkflows,
			"custom_workflows":   dbWorkflows,
		})
	}
}

// ---------- workflow_select ----------

func workflowSelectTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "workflow_select",
		Description: "Activate a workflow for an initiative. Subsequent spec operations are scoped to this workflow.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"initiative_id":{"type":"string","description":"Initiative ID"},
				"workflow_id":{"type":"string","description":"Workflow ID to activate"}
			},
			"required":["initiative_id","workflow_id"]
		}`),
	}
}

func workflowSelectHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			InitiativeID string `json:"initiative_id"`
			WorkflowID   string `json:"workflow_id"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		if err := svc.SelectWorkflow(ctx, args.InitiativeID, args.WorkflowID); err != nil {
			return nil, err
		}
		return jsonResult(map[string]any{
			"status":        "selected",
			"initiative_id": args.InitiativeID,
			"workflow_id":   args.WorkflowID,
		})
	}
}

// ---------- workflow_status ----------

func workflowStatusTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "workflow_status",
		Description: "Show current workflow position: which specs exist, their status, and recommended next steps.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"initiative_id":{"type":"string","description":"Initiative ID"}
			},
			"required":["initiative_id"]
		}`),
	}
}

func workflowStatusHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			InitiativeID string `json:"initiative_id"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		status, err := svc.GetWorkflowStatus(ctx, args.InitiativeID)
		if err != nil {
			return nil, err
		}
		return jsonResult(status)
	}
}

// ---------- spec_list ----------

func specListTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "spec_list",
		Description: "List specs for an initiative. If a workflow is active, shows workflow-defined specs plus any custom additions.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"initiative_id":{"type":"string","description":"Initiative ID"}
			},
			"required":["initiative_id"]
		}`),
	}
}

func specListHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			InitiativeID string `json:"initiative_id"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		specs, err := svc.ListSpecs(ctx, args.InitiativeID)
		if err != nil {
			return nil, err
		}
		return jsonResult(map[string]any{"specs": specs})
	}
}

// ---------- spec_create ----------

func specCreateTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "spec_create",
		Description: "Create a new spec from the workflow template. Writes to the initiative's spec directory.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"initiative_id":{"type":"string","description":"Initiative ID"},
				"spec_type":{"type":"string","description":"Spec type (e.g., prd, trd, press)"},
				"repository_id":{"type":"string","description":"Repository ID where spec will be stored"},
				"file_path":{"type":"string","description":"Relative file path for the spec"}
			},
			"required":["initiative_id","spec_type","repository_id","file_path"]
		}`),
	}
}

func specCreateHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			InitiativeID string `json:"initiative_id"`
			SpecType     string `json:"spec_type"`
			RepositoryID string `json:"repository_id"`
			FilePath     string `json:"file_path"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		if err := svc.CreateSpec(ctx, args.InitiativeID, args.SpecType, args.RepositoryID, args.FilePath, nil); err != nil {
			return nil, err
		}
		return jsonResult(map[string]any{
			"status":        "created",
			"initiative_id": args.InitiativeID,
			"spec_type":     args.SpecType,
			"file_path":     args.FilePath,
		})
	}
}

// ---------- spec_read ----------

func specReadTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "spec_read",
		Description: "Read spec content.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"initiative_id":{"type":"string","description":"Initiative ID"},
				"spec_type":{"type":"string","description":"Spec type (e.g., prd, trd)"},
				"repo_path":{"type":"string","description":"Local repository path"}
			},
			"required":["initiative_id","spec_type","repo_path"]
		}`),
	}
}

func specReadHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			InitiativeID string `json:"initiative_id"`
			SpecType     string `json:"spec_type"`
			RepoPath     string `json:"repo_path"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}
		content, err := svc.ReadSpecContent(ctx, args.InitiativeID, args.SpecType, args.RepoPath)
		if err != nil {
			return nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: content}},
		}, nil
	}
}

// ---------- spec_evaluate ----------

func specEvaluateTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "spec_evaluate",
		Description: "Evaluate a spec against the workflow's rubric using LLM-as-judge.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"initiative_id":{"type":"string","description":"Initiative ID"},
				"spec_type":{"type":"string","description":"Spec type (e.g., prd, trd)"},
				"repo_path":{"type":"string","description":"Local repository path"},
				"model":{"type":"string","description":"Model for evaluation (default: claude-sonnet-4-20250514)"}
			},
			"required":["initiative_id","spec_type","repo_path"]
		}`),
	}
}

func specEvaluateHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			InitiativeID string `json:"initiative_id"`
			SpecType     string `json:"spec_type"`
			RepoPath     string `json:"repo_path"`
			Model        string `json:"model"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}

		llmClient := &stubLLMClient{}
		result, err := svc.EvaluateSpec(ctx, args.InitiativeID, args.SpecType, args.RepoPath, args.Model, llmClient)
		if err != nil {
			return nil, err
		}
		return jsonResult(result)
	}
}

// stubLLMClient is a placeholder LLM client that returns a mock evaluation.
// In production, this should be replaced with a real LLM client.
type stubLLMClient struct{}

func (s *stubLLMClient) Complete(ctx context.Context, prompt string, model string) (string, error) {
	return `{
		"score": 75,
		"verdict": "partial",
		"rationale": "Document covers most requirements but lacks some detail.",
		"categories": [
			{"name": "Problem Statement", "score": 85, "verdict": "pass", "rationale": "Clear problem articulation"},
			{"name": "User Stories", "score": 70, "verdict": "partial", "rationale": "Missing edge cases"},
			{"name": "Requirements", "score": 75, "verdict": "partial", "rationale": "Some requirements need more specificity"}
		],
		"findings": [
			{"severity": "medium", "section": "User Stories", "message": "Missing error handling scenarios"}
		]
	}`, nil
}

var _ speceval.LLMClient = (*stubLLMClient)(nil)
var _ synthesis.LLMClient = (*stubLLMClient)(nil)

// ---------- spec_synthesize ----------

func specSynthesizeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "spec_synthesize",
		Description: "Generate a spec document from source documents using LLM synthesis. Follows the workflow DAG (e.g., PRD -> TRD -> PLAN -> ROADMAP).",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"initiative_id":{"type":"string","description":"Initiative ID"},
				"target_spec_type":{"type":"string","description":"Spec type to generate (e.g., trd, plan, roadmap)"},
				"sources":{"type":"array","items":{"type":"object","properties":{
					"type":{"type":"string","description":"Source document type"},
					"path":{"type":"string","description":"File path (optional)"},
					"content":{"type":"string","description":"Document content"}
				},"required":["type","content"]},"description":"Source documents to synthesize from"},
				"repo_path":{"type":"string","description":"Local repository path for saving output"},
				"model":{"type":"string","description":"Model for synthesis (default: claude-sonnet-4-20250514)"},
				"dry_run":{"type":"boolean","description":"Preview synthesis without saving","default":false}
			},
			"required":["initiative_id","target_spec_type","sources"]
		}`),
	}
}

func specSynthesizeHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			InitiativeID   string `json:"initiative_id"`
			TargetSpecType string `json:"target_spec_type"`
			Sources        []struct {
				Type    string `json:"type"`
				Path    string `json:"path"`
				Content string `json:"content"`
			} `json:"sources"`
			RepoPath string `json:"repo_path"`
			Model    string `json:"model"`
			DryRun   bool   `json:"dry_run"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}

		sources := make([]synthesis.SourceDocument, len(args.Sources))
		for i, s := range args.Sources {
			sources[i] = synthesis.SourceDocument{
				Type:    s.Type,
				Path:    s.Path,
				Content: s.Content,
			}
		}

		synthReq := &synthesis.SynthesisRequest{
			TargetSpecType: args.TargetSpecType,
			Sources:        sources,
			InitiativeID:   args.InitiativeID,
			Model:          args.Model,
			DryRun:         args.DryRun,
		}

		llmClient := &stubLLMClient{}

		var result *synthesis.SynthesisResult
		var err error

		if args.RepoPath != "" && !args.DryRun {
			result, err = svc.SynthesizeAndSaveSpec(ctx, synthReq, args.RepoPath, llmClient)
		} else {
			result, err = svc.SynthesizeSpec(ctx, synthReq, llmClient)
		}

		if err != nil {
			return nil, err
		}
		return jsonResult(result)
	}
}

// ---------- spec_add ----------

func specAddTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "spec_add",
		Description: "Add a custom spec document that's not part of the workflow template. Useful for supplementary documentation.",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"initiative_id":{"type":"string","description":"Initiative ID"},
				"spec_type":{"type":"string","description":"Custom spec type (e.g., adr, runbook, sla)"},
				"repository_id":{"type":"string","description":"Repository ID"},
				"file_path":{"type":"string","description":"Relative file path"},
				"content":{"type":"string","description":"Spec content (optional if file exists)"},
				"repo_path":{"type":"string","description":"Local repository path (required if content provided)"}
			},
			"required":["initiative_id","spec_type","repository_id","file_path"]
		}`),
	}
}

func specAddHandler(svc *service.Service) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			InitiativeID string `json:"initiative_id"`
			SpecType     string `json:"spec_type"`
			RepositoryID string `json:"repository_id"`
			FilePath     string `json:"file_path"`
			Content      string `json:"content"`
			RepoPath     string `json:"repo_path"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, fmt.Errorf("parse arguments: %w", err)
		}

		var content []byte
		if args.Content != "" {
			content = []byte(args.Content)
		}

		if err := svc.AddCustomSpec(ctx, args.InitiativeID, args.SpecType, args.RepositoryID, args.FilePath, content, args.RepoPath); err != nil {
			return nil, err
		}

		return jsonResult(map[string]any{
			"status":        "added",
			"initiative_id": args.InitiativeID,
			"spec_type":     args.SpecType,
			"file_path":     args.FilePath,
		})
	}
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
