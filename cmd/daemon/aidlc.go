package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/grokify/mogo/os/osutil"

	"github.com/ProductBuildersHQ/visionspec/pkg/aidlc"
	"github.com/ProductBuildersHQ/visionstudio/pkg/api"
	"github.com/ProductBuildersHQ/visionstudio/pkg/config"
)

// handleGetAIDLCState returns the AIDLC state for a project
func (s *Server) handleGetAIDLCState(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetAIDLCStateResponse{
			Error: "Project not found: " + projectName,
		})
		return
	}

	// Look for aidlc-state.md in the project
	statePath := filepath.Join(tracked.Path, "aidlc-docs", "aidlc-state.md")

	state, err := aidlc.ParseStateFile(statePath)
	if err != nil {
		// Return empty state if file doesn't exist
		s.writeJSON(w, http.StatusOK, api.GetAIDLCStateResponse{
			State: api.AIDLCState{
				CurrentPhase:    api.AIDLCPhaseInception,
				CompletedDocs:   []string{},
				PendingDocs:     []string{},
				InProgressDocs:  []string{},
				PhaseProgress:   make(map[api.AIDLCPhase]float64),
				OverallProgress: 0.0,
				LastUpdated:     time.Now().Format(time.RFC3339),
			},
		})
		return
	}

	s.writeJSON(w, http.StatusOK, api.GetAIDLCStateResponse{
		State: convertAIDLCState(state),
	})
}

// handleGetAIDLCWorkflow returns the AIDLC workflow DAG for a project
func (s *Server) handleGetAIDLCWorkflow(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetAIDLCWorkflowResponse{
			Error: "Project not found: " + projectName,
		})
		return
	}

	// Create default workflow
	workflow := aidlc.DefaultWorkflow()

	// Try to update from state file if it exists
	statePath := filepath.Join(tracked.Path, "aidlc-docs", "aidlc-state.md")
	if state, err := aidlc.ParseStateFile(statePath); err == nil {
		workflow.UpdateFromState(state)
	}

	// Convert to API type
	apiWorkflow := convertAIDLCWorkflow(workflow)

	s.writeJSON(w, http.StatusOK, api.GetAIDLCWorkflowResponse{
		Workflow: apiWorkflow,
		Mermaid:  workflow.ToMermaid(),
	})
}

// handleListAIDLCDocuments lists all AIDLC documents in a project
func (s *Server) handleListAIDLCDocuments(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.ListAIDLCDocumentsResponse{
			Error: "Project not found: " + projectName,
		})
		return
	}

	// Scan for AIDLC documents
	aidlcDocsDir := filepath.Join(tracked.Path, "aidlc-docs")
	visionSpecDir := filepath.Join(tracked.Path, ".visionspec")

	engine := aidlc.NewSyncEngine(visionSpecDir, aidlcDocsDir)

	ctx := r.Context()
	diff, err := engine.DiffState(ctx)
	if err != nil {
		s.logger.Error("Failed to scan AIDLC documents", "error", err)
	}

	// Collect documents from both sources
	documents := make([]api.AIDLCDocument, 0)

	// Get all document types and their status
	for _, docType := range aidlc.AllDocTypes() {
		doc := api.AIDLCDocument{
			Type:      string(docType),
			Phase:     api.AIDLCPhase(docType.Phase()),
			Title:     docType.DisplayName(),
			Status:    api.AIDLCDocStatusPending,
			UpdatedAt: time.Now().Format(time.RFC3339),
		}

		// Check if document exists in either directory
		aidlcPath := filepath.Join(aidlcDocsDir, string(docType.Phase()), docType.Filename())
		vsPath := filepath.Join(visionSpecDir, docType.Filename())

		if exists, _ := osutil.Exists(aidlcPath); exists {
			doc.Path = aidlcPath
			doc.Status = api.AIDLCDocStatusDraft
		} else if exists, _ := osutil.Exists(vsPath); exists {
			doc.Path = vsPath
			doc.Status = api.AIDLCDocStatusDraft
		}

		// Check if there are pending sync actions
		if diff != nil {
			for _, action := range diff.Actions {
				if action.DocType == docType {
					doc.Status = api.AIDLCDocStatusInProgress
					break
				}
			}
		}

		documents = append(documents, doc)
	}

	s.writeJSON(w, http.StatusOK, api.ListAIDLCDocumentsResponse{
		Documents: documents,
	})
}

// handleGetAIDLCDocument returns a specific AIDLC document
func (s *Server) handleGetAIDLCDocument(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	docID := chi.URLParam(r, "docId")

	// Validate path traversal
	if err := osutil.ValidateNoTraversal(docID); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.GetAIDLCDocumentResponse{
			Error: "Invalid document ID",
		})
		return
	}

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetAIDLCDocumentResponse{
			Error: "Project not found: " + projectName,
		})
		return
	}

	docType := aidlc.DocType(docID)
	if !docType.IsValid() {
		s.writeJSON(w, http.StatusBadRequest, api.GetAIDLCDocumentResponse{
			Error: "Invalid document type: " + docID,
		})
		return
	}

	// Try to find the document in aidlc-docs or .visionspec
	aidlcDocsDir := filepath.Join(tracked.Path, "aidlc-docs")
	visionSpecDir := filepath.Join(tracked.Path, ".visionspec")

	var docPath string
	aidlcPath := filepath.Join(aidlcDocsDir, string(docType.Phase()), docType.Filename())
	vsPath := filepath.Join(visionSpecDir, docType.Filename())

	if exists, _ := osutil.Exists(aidlcPath); exists {
		docPath = aidlcPath
	} else if exists, _ := osutil.Exists(vsPath); exists {
		docPath = vsPath
	}

	doc := api.AIDLCDocument{
		Type:      string(docType),
		Phase:     api.AIDLCPhase(docType.Phase()),
		Path:      docPath,
		Title:     docType.DisplayName(),
		Status:    api.AIDLCDocStatusPending,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	if docPath != "" {
		// Read content
		content, err := osutil.ReadFileSecure(docPath)
		if err == nil {
			doc.Content = string(content)
			doc.Status = api.AIDLCDocStatusDraft
		}
	}

	s.writeJSON(w, http.StatusOK, api.GetAIDLCDocumentResponse{
		Document: doc,
	})
}

// handleGetAIDLCSyncDiff returns the sync diff between visionspec and aidlc directories
func (s *Server) handleGetAIDLCSyncDiff(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetAIDLCSyncDiffResponse{
			Error: "Project not found: " + projectName,
		})
		return
	}

	aidlcDocsDir := filepath.Join(tracked.Path, "aidlc-docs")
	visionSpecDir := filepath.Join(tracked.Path, ".visionspec")

	engine := aidlc.NewSyncEngine(visionSpecDir, aidlcDocsDir)

	ctx := r.Context()
	diff, err := engine.DiffState(ctx)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.GetAIDLCSyncDiffResponse{
			Error: "Failed to compute sync diff: " + err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, api.GetAIDLCSyncDiffResponse{
		Diff: convertAIDLCSyncDiff(diff),
	})
}

// handleAIDLCSync triggers a sync operation
func (s *Server) handleAIDLCSync(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.AIDLCSyncResponse{
			Error: "Project not found: " + projectName,
		})
		return
	}

	var req api.AIDLCSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to bidirectional sync
		req.Direction = "bidirectional"
	}

	aidlcDocsDir := filepath.Join(tracked.Path, "aidlc-docs")
	visionSpecDir := filepath.Join(tracked.Path, ".visionspec")

	engine := aidlc.NewSyncEngine(visionSpecDir, aidlcDocsDir)
	engine.DryRun = req.DryRun

	ctx := r.Context()
	var result *aidlc.SyncResult

	switch req.Direction {
	case "to_aidlc":
		result, err = engine.ExportToAIDLC(ctx)
	case "from_aidlc":
		result, err = engine.ImportFromAIDLC(ctx)
	default: // bidirectional
		result, err = engine.Sync(ctx)
	}

	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.AIDLCSyncResponse{
			Error: "Sync failed: " + err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, api.AIDLCSyncResponse{
		Result: convertAIDLCSyncResult(result),
	})
}

// Conversion helpers

func convertAIDLCState(state *aidlc.State) api.AIDLCState {
	completedDocs := make([]string, len(state.CompletedDocs))
	for i, d := range state.CompletedDocs {
		completedDocs[i] = string(d)
	}

	pendingDocs := make([]string, len(state.PendingDocs))
	for i, d := range state.PendingDocs {
		pendingDocs[i] = string(d)
	}

	inProgressDocs := make([]string, len(state.InProgressDocs))
	for i, d := range state.InProgressDocs {
		inProgressDocs[i] = string(d)
	}

	phaseProgress := make(map[api.AIDLCPhase]float64)
	for p, progress := range state.PhaseProgress {
		phaseProgress[api.AIDLCPhase(p)] = progress
	}

	documentScores := make(map[string]*api.AIDLCQualityScore)
	for dt, score := range state.DocumentScores {
		if score != nil {
			documentScores[string(dt)] = convertAIDLCQualityScore(score)
		}
	}

	return api.AIDLCState{
		CurrentPhase:    api.AIDLCPhase(state.CurrentPhase),
		CurrentDocument: string(state.CurrentDocument),
		CompletedDocs:   completedDocs,
		PendingDocs:     pendingDocs,
		InProgressDocs:  inProgressDocs,
		DocumentScores:  documentScores,
		PhaseProgress:   phaseProgress,
		OverallProgress: state.OverallProgress(),
		LastUpdated:     state.LastUpdated.Format(time.RFC3339),
	}
}

func convertAIDLCQualityScore(score *aidlc.QualityScore) *api.AIDLCQualityScore {
	if score == nil {
		return nil
	}

	issues := make([]api.AIDLCIssue, len(score.Issues))
	for i, issue := range score.Issues {
		issues[i] = api.AIDLCIssue{
			Severity:   string(issue.Severity),
			Category:   issue.Category,
			Code:       issue.Code,
			Message:    issue.Message,
			Location:   issue.Location,
			Suggestion: issue.Suggestion,
		}
	}

	dimensions := make(map[string]float64)
	for id, dim := range score.Dimensions {
		dimensions[id] = dim.Score
	}

	return &api.AIDLCQualityScore{
		Rating:      api.AIDLCQualityRating(score.Rating),
		Score:       score.Score,
		Issues:      issues,
		Dimensions:  dimensions,
		EvaluatedAt: score.EvaluatedAt.Format(time.RFC3339),
	}
}

func convertAIDLCWorkflow(w *aidlc.Workflow) api.AIDLCWorkflow {
	phases := make([]api.AIDLCWorkflowPhase, len(w.Phases))
	for i, p := range w.Phases {
		phases[i] = api.AIDLCWorkflowPhase{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Order:       p.Order,
			NodeIDs:     p.NodeIDs,
		}
	}

	nodes := make(map[string]api.AIDLCWorkflowNode)
	for id, n := range w.Nodes {
		var score *api.AIDLCQualityScore
		if n.Score != nil {
			score = convertAIDLCQualityScore(n.Score)
		}
		nodes[id] = api.AIDLCWorkflowNode{
			ID:          n.ID,
			DocType:     string(n.DocType),
			Phase:       api.AIDLCPhase(n.Phase),
			Name:        n.Name,
			Description: n.Description,
			Status:      string(n.Status),
			Score:       score,
			DependsOn:   n.DependsOn,
			Blocks:      n.Blocks,
			Required:    n.Required,
			Automated:   n.Automated,
		}
	}

	edges := make([]api.AIDLCWorkflowEdge, len(w.Edges))
	for i, e := range w.Edges {
		edges[i] = api.AIDLCWorkflowEdge{
			From:  e.From,
			To:    e.To,
			Type:  string(e.Type),
			Label: e.Label,
		}
	}

	progress := w.Progress()

	return api.AIDLCWorkflow{
		Name:        w.Name,
		Description: w.Description,
		Phases:      phases,
		Nodes:       nodes,
		Edges:       edges,
		Progress: api.WorkflowProgress{
			Completed: progress.Completed,
			Total:     progress.Total,
			Percent:   progress.Percent,
		},
	}
}

func convertAIDLCSyncDiff(diff *aidlc.SyncDiff) api.AIDLCSyncDiff {
	actions := make([]api.AIDLCSyncAction, len(diff.Actions))
	for i, a := range diff.Actions {
		actions[i] = api.AIDLCSyncAction{
			Direction:  string(a.Direction),
			DocType:    string(a.DocType),
			SourcePath: a.SourcePath,
			DestPath:   a.DestPath,
			Action:     a.Action,
			Reason:     a.Reason,
		}
	}

	conflicts := make([]api.AIDLCSyncConflict, len(diff.Conflicts))
	for i, c := range diff.Conflicts {
		conflicts[i] = api.AIDLCSyncConflict{
			DocType:           string(c.DocType),
			VisionSpecPath:    c.VisionSpecPath,
			AIDLCPath:         c.AIDLCPath,
			VisionSpecModTime: c.VisionSpecModTime.Format(time.RFC3339),
			AIDLCModTime:      c.AIDLCModTime.Format(time.RFC3339),
			Reason:            c.Reason,
		}
	}

	return api.AIDLCSyncDiff{
		VisionSpecDir: diff.VisionSpecDir,
		AIDLCDocsDir:  diff.AIDLCDocsDir,
		Actions:       actions,
		Conflicts:     conflicts,
		ComputedAt:    diff.ComputedAt.Format(time.RFC3339),
	}
}

func convertAIDLCSyncResult(result *aidlc.SyncResult) api.AIDLCSyncResult {
	return api.AIDLCSyncResult{
		Direction:   string(result.Direction),
		Created:     result.Created,
		Updated:     result.Updated,
		Skipped:     result.Skipped,
		Errors:      result.Errors,
		CompletedAt: result.CompletedAt.Format(time.RFC3339),
	}
}

// handleGetAIDLCPhaseRequirements returns the requirements for the current or target phase
func (s *Server) handleGetAIDLCPhaseRequirements(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetAIDLCPhaseRequirementsResponse{
			Error: "Project not found: " + projectName,
		})
		return
	}

	// Create default workflow and update from state
	workflow := aidlc.DefaultWorkflow()
	statePath := filepath.Join(tracked.Path, "aidlc-docs", "aidlc-state.md")
	if state, err := aidlc.ParseStateFile(statePath); err == nil {
		workflow.UpdateFromState(state)
	}

	// Get requirements for all phases
	allReqs := workflow.AllPhaseRequirements()

	phaseReqs := make([]api.AIDLCPhaseRequirements, len(allReqs))
	for i, req := range allReqs {
		requiredDocs := make([]string, len(req.RequiredDocs))
		for j, dt := range req.RequiredDocs {
			requiredDocs[j] = string(dt)
		}
		completedDocs := make([]string, len(req.CompletedDocs))
		for j, dt := range req.CompletedDocs {
			completedDocs[j] = string(dt)
		}
		pendingDocs := make([]string, len(req.PendingDocs))
		for j, dt := range req.PendingDocs {
			pendingDocs[j] = string(dt)
		}

		phaseReqs[i] = api.AIDLCPhaseRequirements{
			Phase:           api.AIDLCPhase(req.Phase),
			RequiredDocs:    requiredDocs,
			CompletedDocs:   completedDocs,
			PendingDocs:     pendingDocs,
			ProgressPercent: req.ProgressPercent,
			CanAdvance:      req.CanAdvance,
		}
	}

	currentPhase := workflow.CurrentPhase()

	s.writeJSON(w, http.StatusOK, api.GetAIDLCPhaseRequirementsResponse{
		CurrentPhase: api.AIDLCPhase(currentPhase),
		Requirements: phaseReqs,
	})
}

// handleAIDLCPhaseTransition attempts to transition to a new phase
func (s *Server) handleAIDLCPhaseTransition(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.AIDLCPhaseTransitionResponse{
			Error: "Project not found: " + projectName,
		})
		return
	}

	var req api.AIDLCPhaseTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.AIDLCPhaseTransitionResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	targetPhase := aidlc.Phase(req.TargetPhase)
	if !targetPhase.IsValid() {
		s.writeJSON(w, http.StatusBadRequest, api.AIDLCPhaseTransitionResponse{
			Error: "Invalid target phase: " + string(req.TargetPhase),
		})
		return
	}

	// Create workflow and update from state
	workflow := aidlc.DefaultWorkflow()
	statePath := filepath.Join(tracked.Path, "aidlc-docs", "aidlc-state.md")
	if state, err := aidlc.ParseStateFile(statePath); err == nil {
		workflow.UpdateFromState(state)
	}

	// Try to transition
	rules := aidlc.DefaultTransitionRules()
	result := workflow.CanTransitionTo(targetPhase, rules)

	// Convert blocking docs
	blockingDocs := make([]string, len(result.BlockingDocs))
	for i, dt := range result.BlockingDocs {
		blockingDocs[i] = string(dt)
	}

	s.writeJSON(w, http.StatusOK, api.AIDLCPhaseTransitionResponse{
		Success:        result.Success,
		FromPhase:      api.AIDLCPhase(result.FromPhase),
		ToPhase:        api.AIDLCPhase(result.ToPhase),
		BlockingDocs:   blockingDocs,
		BlockingIssues: result.BlockingIssues,
	})
}

// handleListAIDLCTemplates returns all available AIDLC templates
func (s *Server) handleListAIDLCTemplates(w http.ResponseWriter, r *http.Request) {
	templates := aidlc.AllTemplates()

	apiTemplates := make([]api.AIDLCTemplate, 0, len(templates))
	for docType, tmpl := range templates {
		sections := make([]api.AIDLCTemplateSection, len(tmpl.Sections))
		for i, sec := range tmpl.Sections {
			sections[i] = api.AIDLCTemplateSection{
				ID:          sec.ID,
				Title:       sec.Title,
				Required:    sec.Required,
				Description: sec.Description,
			}
		}

		apiTemplates = append(apiTemplates, api.AIDLCTemplate{
			DocType:     string(docType),
			Name:        tmpl.Name,
			Description: tmpl.Description,
			Phase:       api.AIDLCPhase(docType.Phase()),
			Sections:    sections,
		})
	}

	s.writeJSON(w, http.StatusOK, api.ListAIDLCTemplatesResponse{
		Templates: apiTemplates,
	})
}

// handleGetAIDLCTemplate returns a specific template
func (s *Server) handleGetAIDLCTemplate(w http.ResponseWriter, r *http.Request) {
	docTypeStr := chi.URLParam(r, "docType")

	docType := aidlc.DocType(docTypeStr)
	if !docType.IsValid() {
		s.writeJSON(w, http.StatusBadRequest, api.GetAIDLCTemplateResponse{
			Error: "Invalid document type: " + docTypeStr,
		})
		return
	}

	tmpl, ok := aidlc.GetTemplate(docType)
	if !ok {
		s.writeJSON(w, http.StatusNotFound, api.GetAIDLCTemplateResponse{
			Error: "Template not found: " + docTypeStr,
		})
		return
	}

	sections := make([]api.AIDLCTemplateSection, len(tmpl.Sections))
	for i, sec := range tmpl.Sections {
		sections[i] = api.AIDLCTemplateSection{
			ID:          sec.ID,
			Title:       sec.Title,
			Required:    sec.Required,
			Description: sec.Description,
		}
	}

	s.writeJSON(w, http.StatusOK, api.GetAIDLCTemplateResponse{
		Template: api.AIDLCTemplate{
			DocType:     string(docType),
			Name:        tmpl.Name,
			Description: tmpl.Description,
			Phase:       api.AIDLCPhase(docType.Phase()),
			Sections:    sections,
			Content:     tmpl.Content,
		},
	})
}

// handleCreateAIDLCDocument creates a new document from a template
func (s *Server) handleCreateAIDLCDocument(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.CreateAIDLCDocumentResponse{
			Error: "Project not found: " + projectName,
		})
		return
	}

	var req api.CreateAIDLCDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.CreateAIDLCDocumentResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	docType := aidlc.DocType(req.DocType)
	if !docType.IsValid() {
		s.writeJSON(w, http.StatusBadRequest, api.CreateAIDLCDocumentResponse{
			Error: "Invalid document type: " + req.DocType,
		})
		return
	}

	// Render template
	data := aidlc.TemplateData{
		ProjectName: projectName,
		Title:       req.Title,
		Author:      req.Author,
		Date:        time.Now().Format("2006-01-02"),
		Version:     "1.0",
		Description: req.Description,
	}

	content, err := aidlc.RenderTemplate(docType, data)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.CreateAIDLCDocumentResponse{
			Error: "Failed to render template: " + err.Error(),
		})
		return
	}

	// Write to file
	aidlcDocsDir := filepath.Join(tracked.Path, "aidlc-docs")
	phaseDir := filepath.Join(aidlcDocsDir, string(docType.Phase()))
	docPath := filepath.Join(phaseDir, docType.Filename())

	// Create phase directory if needed
	if err := os.MkdirAll(phaseDir, 0755); err != nil { //nolint:gosec // G703: Path from tracked project config
		s.writeJSON(w, http.StatusInternalServerError, api.CreateAIDLCDocumentResponse{
			Error: "Failed to create directory: " + err.Error(),
		})
		return
	}

	// Check if file already exists
	if exists, _ := osutil.Exists(docPath); exists && !req.Overwrite {
		s.writeJSON(w, http.StatusConflict, api.CreateAIDLCDocumentResponse{
			Error: "Document already exists: " + docPath,
		})
		return
	}

	// Write file
	if err := osutil.WriteFileSecure(docPath, []byte(content), 0644); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.CreateAIDLCDocumentResponse{
			Error: "Failed to write file: " + err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, api.CreateAIDLCDocumentResponse{
		Document: api.AIDLCDocument{
			Type:      string(docType),
			Phase:     api.AIDLCPhase(docType.Phase()),
			Path:      docPath,
			Title:     req.Title,
			Content:   content,
			Status:    api.AIDLCDocStatusDraft,
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	})
}

// Ensure context is used to avoid unused import error
var _ = context.Background
