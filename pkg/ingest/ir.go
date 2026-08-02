package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/ir"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// IRResult is the outcome of ingesting an IR file.
type IRResult struct {
	RepoID         string
	FilePath       string
	DevXReports    int
	PRISMRoadmaps  int
	PRISMGoals     int
	PRISMDocuments int
	SpecDocuments  int
	ExecutionItems int
	MaturityItems  int
	Err            error
}

// IR ingests a single *.ir.json file into the database.
func IR(ctx context.Context, svc *service.Service, repoID, filePath string) (*IRResult, error) {
	result := &IRResult{
		RepoID:   repoID,
		FilePath: filePath,
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read IR file: %w", err)
	}

	var snapshot ir.RepoSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("parse IR file: %w", err)
	}

	now := time.Now()

	if snapshot.DevX != nil {
		for _, pr := range snapshot.DevX.PeriodReports {
			if pr == nil || pr.Report == nil {
				continue
			}
			report := &store.DevXPeriodReport{
				ID:            fmt.Sprintf("%s-%s-%s", repoID, pr.Report.Subject.PersonID, pr.Label),
				Organization:  snapshot.Org,
				RepositoryID:  repoID,
				PersonID:      pr.Report.Subject.PersonID,
				PeriodType:    string(pr.Type),
				PeriodLabel:   pr.Label,
				PeriodStart:   pr.Report.Period.Start,
				PeriodEnd:     pr.Report.Period.End,
				CoverageScore: pr.Report.Quality.CoverageScore,
				CreatedAt:     now,
			}
			if pr.Report.Metrics.Combined != nil {
				metricsMap := make(map[string]any)
				for k, v := range pr.Report.Metrics.Combined {
					metricsMap[k] = v.Value
				}
				report.Metrics = metricsMap
			}
			if pr.Report.Metrics.ByModel != nil {
				byModelMap := make(map[string]any)
				for model, metrics := range pr.Report.Metrics.ByModel {
					modelMetrics := make(map[string]float64)
					for k, v := range metrics {
						modelMetrics[k] = v.Value
					}
					byModelMap[model] = modelMetrics
				}
				report.ByModel = byModelMap
			}
			if err := svc.CreateDevXPeriodReport(ctx, report); err != nil {
				if !strings.Contains(err.Error(), "already exists") {
					return nil, fmt.Errorf("create devx period report: %w", err)
				}
			} else {
				result.DevXReports++
			}
		}
	}

	if snapshot.Roadmap != nil {
		if snapshot.Roadmap.Roadmap != nil {
			roadmap := &store.PRISMRoadmap{
				ID:           fmt.Sprintf("%s-roadmap", repoID),
				Organization: snapshot.Org,
				RepositoryID: repoID,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if snapshot.Roadmap.Roadmap.Phases != nil {
				phases := make([]any, len(snapshot.Roadmap.Roadmap.Phases))
				for i, p := range snapshot.Roadmap.Roadmap.Phases {
					phases[i] = p
				}
				roadmap.Phases = phases
			}
			if err := svc.CreatePRISMRoadmap(ctx, roadmap); err != nil {
				if !strings.Contains(err.Error(), "already exists") {
					return nil, fmt.Errorf("create prism roadmap: %w", err)
				}
			} else {
				result.PRISMRoadmaps++
			}
		}

		if snapshot.Roadmap.Goals != nil {
			goal := &store.PRISMGoal{
				ID:           fmt.Sprintf("%s-goals", repoID),
				Organization: snapshot.Org,
				RepositoryID: repoID,
				GoalType:     "goals",
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			doc := make(map[string]any)
			docJSON, _ := json.Marshal(snapshot.Roadmap.Goals)
			if err := json.Unmarshal(docJSON, &doc); err == nil {
				goal.Document = doc
			}
			if err := svc.CreatePRISMGoal(ctx, goal); err != nil {
				if !strings.Contains(err.Error(), "already exists") {
					return nil, fmt.Errorf("create prism goal: %w", err)
				}
			} else {
				result.PRISMGoals++
			}
		}
	}

	if snapshot.Maturity != nil && snapshot.Maturity.PRISMDocument != nil {
		prismDoc := snapshot.Maturity.PRISMDocument
		doc := &store.PRISMDocument{
			ID:           fmt.Sprintf("%s-prism", repoID),
			Organization: snapshot.Org,
			RepositoryID: repoID,
			Name:         prismDoc.Metadata.Name,
			Description:  prismDoc.Metadata.Description,
			Version:      prismDoc.Metadata.Version,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if len(prismDoc.Domains) > 0 {
			domains := make([]any, len(prismDoc.Domains))
			for i, d := range prismDoc.Domains {
				domains[i] = d
			}
			doc.Domains = domains
		}
		if len(prismDoc.Layers) > 0 {
			layers := make([]any, len(prismDoc.Layers))
			for i, l := range prismDoc.Layers {
				layers[i] = l
			}
			doc.Layers = layers
		}
		if len(prismDoc.Metrics) > 0 {
			metrics := make([]any, len(prismDoc.Metrics))
			for i, m := range prismDoc.Metrics {
				metrics[i] = m
			}
			doc.Metrics = metrics
		}
		if prismDoc.Maturity != nil {
			maturityJSON, _ := json.Marshal(prismDoc.Maturity)
			var maturityMap map[string]any
			if err := json.Unmarshal(maturityJSON, &maturityMap); err == nil {
				doc.Maturity = maturityMap
			}
		}
		if prismDoc.SLIState != nil {
			sliJSON, _ := json.Marshal(prismDoc.SLIState)
			var sliMap map[string]any
			if err := json.Unmarshal(sliJSON, &sliMap); err == nil {
				doc.SLIState = sliMap
			}
		}
		if prismDoc.MaturityState != nil {
			msJSON, _ := json.Marshal(prismDoc.MaturityState)
			var msMap map[string]any
			if err := json.Unmarshal(msJSON, &msMap); err == nil {
				doc.MaturityState = msMap
			}
		}
		if err := svc.CreatePRISMDocument(ctx, doc); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return nil, fmt.Errorf("create prism document: %w", err)
			}
		} else {
			result.PRISMDocuments++
		}
	}

	return result, nil
}

// IRFromRepo scans a repository for *.ir.json files and ingests them.
func IRFromRepo(ctx context.Context, svc *service.Service, repoID string) ([]*IRResult, error) {
	repo, err := svc.GetRepository(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("get repository: %w", err)
	}
	if repo.LocalPath == "" {
		return nil, fmt.Errorf("repository %s has no local_path", repoID)
	}

	var results []*IRResult
	err = filepath.Walk(repo.LocalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".ir.json") {
			result, err := IR(ctx, svc, repoID, path)
			if err != nil {
				result = &IRResult{RepoID: repoID, FilePath: path, Err: err}
			}
			results = append(results, result)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk repository: %w", err)
	}
	return results, nil
}
