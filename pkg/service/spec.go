package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/speceval"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
	"github.com/ProductBuildersHQ/visionstudio/pkg/synthesis"
)

// SpecStatus represents the evaluation status of a spec document.
type SpecStatus struct {
	Type        string `json:"type"`
	Status      string `json:"status"`
	EvalScore   *int   `json:"eval_score,omitempty"`
	EvalVerdict string `json:"eval_verdict,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
	Required    bool   `json:"required"`
}

// WorkflowStatus summarizes an initiative's workflow progress.
type WorkflowStatus struct {
	InitiativeID string       `json:"initiative_id"`
	WorkflowID   string       `json:"workflow_id,omitempty"`
	WorkflowName string       `json:"workflow_name,omitempty"`
	Specs        []SpecStatus `json:"specs"`
	GatesPassed  bool         `json:"gates_passed"`
	Blockers     []string     `json:"blockers,omitempty"`
	NextSteps    []string     `json:"next_steps,omitempty"`
}

// ListWorkflows returns all available spec workflows.
func (s *Service) ListWorkflows(ctx context.Context) ([]*store.SpecWorkflow, error) {
	return s.Store.ListSpecWorkflows(ctx)
}

// GetWorkflow returns a single spec workflow by ID.
func (s *Service) GetWorkflow(ctx context.Context, id string) (*store.SpecWorkflow, error) {
	return s.Store.GetSpecWorkflow(ctx, id)
}

// SelectWorkflow activates a workflow for an initiative.
func (s *Service) SelectWorkflow(ctx context.Context, initiativeID, workflowID string) error {
	_, err := s.Store.GetInitiative(ctx, initiativeID)
	if err != nil {
		return fmt.Errorf("initiative not found: %w", err)
	}
	_, err = s.Store.GetSpecWorkflow(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}
	return s.Store.SelectWorkflowForInitiative(ctx, initiativeID, workflowID)
}

// GetWorkflowStatus returns the workflow status for an initiative.
func (s *Service) GetWorkflowStatus(ctx context.Context, initiativeID string) (*WorkflowStatus, error) {
	init, err := s.Store.GetInitiative(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("initiative not found: %w", err)
	}

	result := &WorkflowStatus{
		InitiativeID: initiativeID,
	}

	iw, err := s.Store.GetWorkflowForInitiative(ctx, initiativeID)
	if err != nil {
		return nil, err
	}

	var workflow *store.SpecWorkflow
	if iw != nil {
		result.WorkflowID = iw.WorkflowID
		workflow, _ = s.Store.GetSpecWorkflow(ctx, iw.WorkflowID)
		if workflow != nil {
			result.WorkflowName = workflow.Name
		}
	}

	docs, err := s.Store.ListSpecDocumentsByInitiative(ctx, initiativeID)
	if err != nil {
		return nil, err
	}
	docMap := make(map[string]*store.SpecDocument)
	for _, d := range docs {
		docMap[d.SpecType] = d
	}

	if workflow != nil {
		for _, st := range workflow.SpecsRequired {
			status := SpecStatus{Type: st, Status: "missing", Required: true}
			if doc, ok := docMap[st]; ok {
				status.Status = doc.Status
				status.EvalScore = doc.EvalScore
				status.EvalVerdict = doc.EvalVerdict
				status.FilePath = doc.FilePath
			}
			result.Specs = append(result.Specs, status)
		}
		for _, st := range workflow.SpecsOptional {
			status := SpecStatus{Type: st, Status: "missing", Required: false}
			if doc, ok := docMap[st]; ok {
				status.Status = doc.Status
				status.EvalScore = doc.EvalScore
				status.EvalVerdict = doc.EvalVerdict
				status.FilePath = doc.FilePath
			}
			result.Specs = append(result.Specs, status)
		}
	} else {
		for st, fp := range init.Specs {
			status := SpecStatus{Type: st, Status: "draft", FilePath: fp}
			if doc, ok := docMap[st]; ok {
				status.Status = doc.Status
				status.EvalScore = doc.EvalScore
				status.EvalVerdict = doc.EvalVerdict
			}
			result.Specs = append(result.Specs, status)
		}
	}

	result.NextSteps = s.computeNextSteps(result)
	result.GatesPassed, result.Blockers = s.computeGates(result)
	return result, nil
}

func (s *Service) computeNextSteps(ws *WorkflowStatus) []string {
	var steps []string
	for _, sp := range ws.Specs {
		if sp.Required && sp.Status == "missing" {
			steps = append(steps, fmt.Sprintf("Create %s spec", sp.Type))
		} else if sp.Status == "draft" {
			steps = append(steps, fmt.Sprintf("Evaluate %s spec", sp.Type))
		} else if sp.EvalScore != nil && *sp.EvalScore < 85 {
			steps = append(steps, fmt.Sprintf("Improve %s spec (score %d < 85)", sp.Type, *sp.EvalScore))
		}
	}
	return steps
}

func (s *Service) computeGates(ws *WorkflowStatus) (bool, []string) {
	var blockers []string
	for _, sp := range ws.Specs {
		if sp.Required && sp.Status == "missing" {
			blockers = append(blockers, fmt.Sprintf("Missing required spec: %s", sp.Type))
		} else if sp.Required && sp.EvalVerdict == "fail" {
			blockers = append(blockers, fmt.Sprintf("Spec %s failed evaluation", sp.Type))
		} else if sp.Required && sp.EvalScore != nil && *sp.EvalScore < 85 {
			blockers = append(blockers, fmt.Sprintf("Spec %s below threshold (score %d < 85)", sp.Type, *sp.EvalScore))
		}
	}
	return len(blockers) == 0, blockers
}

// ListSpecs returns all spec documents for an initiative.
func (s *Service) ListSpecs(ctx context.Context, initiativeID string) ([]*store.SpecDocument, error) {
	return s.Store.ListSpecDocumentsByInitiative(ctx, initiativeID)
}

// GetSpec returns a single spec document.
func (s *Service) GetSpec(ctx context.Context, id string) (*store.SpecDocument, error) {
	return s.Store.GetSpecDocument(ctx, id)
}

// ReadSpecContent reads the content of a spec file from the filesystem.
func (s *Service) ReadSpecContent(ctx context.Context, initiativeID, specType, repoPath string) (string, error) {
	docs, err := s.Store.ListSpecDocumentsByInitiative(ctx, initiativeID)
	if err != nil {
		return "", err
	}
	for _, doc := range docs {
		if doc.SpecType == specType {
			fullPath := filepath.Join(repoPath, doc.FilePath)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return "", fmt.Errorf("read spec file: %w", err)
			}
			return string(content), nil
		}
	}
	return "", fmt.Errorf("spec %s not found for initiative %s", specType, initiativeID)
}

// CreateSpec creates a new spec document entry and optional file.
func (s *Service) CreateSpec(ctx context.Context, initiativeID, specType, repoID, filePath string, content []byte) error {
	now := time.Now()
	hash := sha256.Sum256(content)
	doc := &store.SpecDocument{
		ID:           fmt.Sprintf("%s/%s", initiativeID, specType),
		InitiativeID: initiativeID,
		RepositoryID: repoID,
		SpecType:     specType,
		FilePath:     filePath,
		Status:       "draft",
		ContentHash:  hex.EncodeToString(hash[:]),
		SyncedAt:     now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	iw, _ := s.Store.GetWorkflowForInitiative(ctx, initiativeID)
	if iw != nil {
		doc.WorkflowID = iw.WorkflowID
	}

	return s.Store.CreateSpecDocument(ctx, doc)
}

// UpdateSpecEvaluation updates a spec's evaluation results.
func (s *Service) UpdateSpecEvaluation(ctx context.Context, id string, score int, verdict string) error {
	doc, err := s.Store.GetSpecDocument(ctx, id)
	if err != nil {
		return err
	}
	doc.EvalScore = &score
	doc.EvalVerdict = verdict
	doc.Status = "evaluated"
	doc.UpdatedAt = time.Now()
	return s.Store.UpdateSpecDocument(ctx, doc)
}

// EvaluateSpec runs LLM-as-judge evaluation on a spec document.
func (s *Service) EvaluateSpec(ctx context.Context, initiativeID, specType, repoPath, model string, llmClient speceval.LLMClient) (*speceval.EvaluationResult, error) {
	content, err := s.ReadSpecContent(ctx, initiativeID, specType, repoPath)
	if err != nil {
		return nil, fmt.Errorf("read spec content: %w", err)
	}

	evaluator := speceval.NewEvaluator(llmClient)
	for st, r := range speceval.DefaultRubrics() {
		evaluator.RegisterRubric(st, r)
	}

	result, err := evaluator.Evaluate(ctx, specType, content, model)
	if err != nil {
		return nil, fmt.Errorf("evaluate spec: %w", err)
	}

	docID := fmt.Sprintf("%s/%s", initiativeID, specType)
	if err := s.UpdateSpecEvaluation(ctx, docID, result.Score, result.Verdict); err != nil {
		return nil, fmt.Errorf("persist evaluation: %w", err)
	}

	return result, nil
}

// CheckGates verifies if all workflow gates are satisfied for an initiative.
func (s *Service) CheckGates(ctx context.Context, initiativeID string) (bool, []string, error) {
	status, err := s.GetWorkflowStatus(ctx, initiativeID)
	if err != nil {
		return false, nil, err
	}

	var blockers []string
	for _, sp := range status.Specs {
		if sp.Required && sp.Status == "missing" {
			blockers = append(blockers, fmt.Sprintf("Missing required spec: %s", sp.Type))
		} else if sp.Required && sp.EvalScore != nil && *sp.EvalScore < 85 {
			blockers = append(blockers, fmt.Sprintf("Spec %s below threshold (score %d < 85)", sp.Type, *sp.EvalScore))
		} else if sp.Required && sp.EvalVerdict == "fail" {
			blockers = append(blockers, fmt.Sprintf("Spec %s failed evaluation", sp.Type))
		}
	}

	return len(blockers) == 0, blockers, nil
}

// SynthesizeSpec generates a spec document from source documents using LLM.
func (s *Service) SynthesizeSpec(ctx context.Context, req *synthesis.SynthesisRequest, llmClient synthesis.LLMClient) (*synthesis.SynthesisResult, error) {
	executor := synthesis.NewExecutor(llmClient)
	return executor.Synthesize(ctx, req)
}

// SynthesizeAndSaveSpec generates a spec and saves it to both filesystem and database.
func (s *Service) SynthesizeAndSaveSpec(ctx context.Context, req *synthesis.SynthesisRequest, repoPath string, llmClient synthesis.LLMClient) (*synthesis.SynthesisResult, error) {
	result, err := s.SynthesizeSpec(ctx, req, llmClient)
	if err != nil {
		return nil, err
	}

	if req.DryRun {
		return result, nil
	}

	specDir := filepath.Join(repoPath, "docs", "specs", "initiatives", req.InitiativeID)
	if err := os.MkdirAll(specDir, 0755); err != nil {
		return nil, fmt.Errorf("create spec directory: %w", err)
	}

	filename := fmt.Sprintf("%s.md", req.TargetSpecType)
	filePath := filepath.Join(specDir, filename)
	if err := os.WriteFile(filePath, []byte(result.Content), 0644); err != nil {
		return nil, fmt.Errorf("write spec file: %w", err)
	}

	relPath := filepath.Join("docs", "specs", "initiatives", req.InitiativeID, filename)
	init, err := s.Store.GetInitiative(ctx, req.InitiativeID)
	if err != nil {
		return nil, fmt.Errorf("get initiative: %w", err)
	}

	if err := s.CreateSpec(ctx, req.InitiativeID, req.TargetSpecType, init.HomeRepo, relPath, []byte(result.Content)); err != nil {
		return nil, fmt.Errorf("create spec record: %w", err)
	}

	return result, nil
}

// AddCustomSpec adds a custom spec document that's not part of the workflow template.
func (s *Service) AddCustomSpec(ctx context.Context, initiativeID, specType, repoID, filePath string, content []byte, repoPath string) error {
	if repoPath != "" && content != nil {
		fullPath := filepath.Join(repoPath, filePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
	}

	return s.CreateSpec(ctx, initiativeID, specType, repoID, filePath, content)
}
