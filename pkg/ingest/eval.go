package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/plexusone/structured-evaluation/rubric"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// EvalResult is the outcome of ingesting evaluation files.
type EvalResult struct {
	InitiativeID string
	FilesFound   int
	Imported     int
	Errors       []string
}

// EvalFile represents the JSON structure of an evaluation file.
type EvalFile struct {
	ID           string          `json:"id"`
	InitiativeID string          `json:"initiative_id"`
	SpecPath     string          `json:"spec_path"`
	SpecType     string          `json:"spec_type"`
	RubricID     string          `json:"rubric_id"`
	Score        float64         `json:"score"`
	Verdict      string          `json:"verdict"`
	Model        string          `json:"model"`
	EvaluatedAt  time.Time       `json:"evaluated_at"`
	Rationale    string          `json:"rationale"`
	Categories   []EvalCategory  `json:"categories,omitempty"`
	Findings     []EvalFinding   `json:"findings,omitempty"`
}

// EvalCategory is a rubric category evaluation result.
type EvalCategory struct {
	Name      string  `json:"name"`
	Score     float64 `json:"score"`
	Verdict   string  `json:"verdict"`
	Weight    float64 `json:"weight"`
	Rationale string  `json:"rationale"`
}

// EvalFinding is an issue found during evaluation.
type EvalFinding struct {
	Severity string `json:"severity"`
	Section  string `json:"section"`
	Message  string `json:"message"`
}

// Evals ingests all *.eval.json files from an initiative's evaluations directory.
func Evals(ctx context.Context, svc *service.Service, initiativeID, repoPath string) (*EvalResult, error) {
	result := &EvalResult{
		InitiativeID: initiativeID,
	}

	evalDir := filepath.Join(repoPath, "docs", "specs", "initiatives", initiativeID, "evaluations")
	entries, err := os.ReadDir(evalDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("read evaluations dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".eval.json") {
			continue
		}
		result.FilesFound++

		filePath := filepath.Join(evalDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", entry.Name(), err))
			continue
		}

		judgeResult := parseEvalFileData(data, initiativeID, entry.Name())
		if judgeResult == nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: parse error", entry.Name()))
			continue
		}

		if err := svc.Store.CreateJudgeResult(ctx, judgeResult); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				result.Errors = append(result.Errors, fmt.Sprintf("%s: store error: %v", entry.Name(), err))
				continue
			}
		}
		result.Imported++
	}

	return result, nil
}

// EvalsFromRepo scans a repository for evaluation files in all initiatives.
func EvalsFromRepo(ctx context.Context, svc *service.Service, repoPath string) ([]*EvalResult, error) {
	initiativesDir := filepath.Join(repoPath, "docs", "specs", "initiatives")
	entries, err := os.ReadDir(initiativesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read initiatives dir: %w", err)
	}

	var results []*EvalResult
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "INIT-") {
			continue
		}

		result, err := Evals(ctx, svc, entry.Name(), repoPath)
		if err != nil {
			return nil, fmt.Errorf("ingest %s: %w", entry.Name(), err)
		}
		if result.FilesFound > 0 {
			results = append(results, result)
		}
	}

	return results, nil
}

// parseEvalFileData parses an eval file supporting both legacy and structured-evaluation formats.
func parseEvalFileData(data []byte, initiativeID, filename string) *store.JudgeResult {
	// Try structured-evaluation rubric.Rubric format first
	var report rubric.Rubric
	if err := json.Unmarshal(data, &report); err == nil && report.ReviewType != "" {
		// Valid rubric report
		specType := strings.TrimSuffix(filename, ".eval.json")
		return &store.JudgeResult{
			ID:           fmt.Sprintf("eval-%s-%s", initiativeID, specType),
			InitiativeID: initiativeID,
			SpecPath:     report.Metadata.Document,
			SpecType:     specType,
			RubricID:     report.RubricID,
			EvaluatedAt:  report.Metadata.GeneratedAt,
			Report:       &report,
		}
	}

	// Fall back to legacy format
	var legacy EvalFile
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil
	}

	// Convert legacy to rubric.Rubric
	report = rubric.Rubric{
		Metadata: rubric.ReportMetadata{
			Document:    legacy.SpecPath,
			GeneratedAt: legacy.EvaluatedAt,
		},
		ReviewType: legacy.SpecType,
		RubricID:   legacy.RubricID,
		IntScore:   rubric.IntegerScore(int(legacy.Score) / 20), // Convert 0-100 to 1-5
		Pass:       legacy.Verdict == "pass",
		Summary:    legacy.Rationale,
		Categories: make([]rubric.CategoryResult, 0, len(legacy.Categories)),
		Findings:   make([]rubric.Finding, 0, len(legacy.Findings)),
	}
	if legacy.Model != "" {
		report.Judge = &rubric.JudgeMetadata{Model: legacy.Model}
	}

	for _, cat := range legacy.Categories {
		catResult := rubric.CategoryResult{
			Category:  cat.Name,
			IntScore:  rubric.IntegerScore(int(cat.Score) / 20), // Convert 0-100 to 1-5
			Reasoning: cat.Rationale,
		}
		catResult.Score = catResult.IntScore.ToCategorical()
		report.Categories = append(report.Categories, catResult)
	}

	for i, f := range legacy.Findings {
		finding := rubric.Finding{
			ID:          fmt.Sprintf("finding-%d", i+1),
			Category:    f.Section,
			Severity:    rubric.Severity(f.Severity),
			Title:       f.Message,
			Description: f.Message,
		}
		report.Findings = append(report.Findings, finding)
	}

	return &store.JudgeResult{
		ID:           legacy.ID,
		InitiativeID: legacy.InitiativeID,
		SpecPath:     legacy.SpecPath,
		SpecType:     legacy.SpecType,
		RubricID:     legacy.RubricID,
		EvaluatedAt:  legacy.EvaluatedAt,
		Report:       &report,
	}
}
