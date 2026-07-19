package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/grokify/mogo/os/osutil"

	"github.com/ProductBuildersHQ/visionstudio/pkg/api"
)

// getSamplesDir returns the path to the samples directory
func (s *Server) getSamplesDir() string {
	// Try to find samples directory relative to executable
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}

	// Go up from bin/ to repo root, then to samples/
	repoRoot := filepath.Dir(filepath.Dir(execPath))
	samplesDir := filepath.Join(repoRoot, "samples")

	if _, err := os.Stat(samplesDir); err == nil {
		return samplesDir
	}

	// Fallback: check if we're running from repo root (e.g., go run)
	cwd, _ := os.Getwd()
	samplesDir = filepath.Join(cwd, "samples")
	if _, err := os.Stat(samplesDir); err == nil {
		return samplesDir
	}

	return ""
}

// handleListSamples lists all available sample projects
func (s *Server) handleListSamples(w http.ResponseWriter, r *http.Request) {
	samplesDir := s.getSamplesDir()
	if samplesDir == "" {
		s.writeJSON(w, http.StatusOK, api.ListSamplesResponse{Samples: []api.SampleSummary{}})
		return
	}

	entries, err := os.ReadDir(samplesDir)
	if err != nil {
		s.logger.Error("Failed to read samples directory", "error", err)
		s.writeJSON(w, http.StatusOK, api.ListSamplesResponse{Samples: []api.SampleSummary{}})
		return
	}

	samples := make([]api.SampleSummary, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		samplePath := filepath.Join(samplesDir, entry.Name())
		summary := s.loadSampleSummary(entry.Name(), samplePath)
		if summary.ID != "" {
			samples = append(samples, summary)
		}
	}

	s.writeJSON(w, http.StatusOK, api.ListSamplesResponse{Samples: samples})
}

// loadSampleSummary loads summary info for a sample
func (s *Server) loadSampleSummary(id, path string) api.SampleSummary {
	summary := api.SampleSummary{
		ID:         id,
		Path:       path,
		FileCounts: make(map[string]int),
	}

	// Try to load project.json for metadata
	projectPath := filepath.Join(path, "project.json")
	if data, err := os.ReadFile(projectPath); err == nil { //nolint:gosec // G703: Path from embedded samples directory
		var projectJSON map[string]any
		if json.Unmarshal(data, &projectJSON) == nil {
			if meta, ok := projectJSON["metadata"].(map[string]any); ok {
				if name, ok := meta["title"].(string); ok {
					summary.Name = name
				}
				if desc, ok := meta["description"].(string); ok {
					summary.Description = desc
				}
			}
		}
	}

	// Fallback name from directory
	if summary.Name == "" {
		summary.Name = strings.Title(strings.ReplaceAll(id, "-", " "))
	}

	// Determine complexity based on directory name or structure
	if id == "simple" {
		summary.Complexity = "simple"
	} else {
		summary.Complexity = "enterprise"
	}

	// Count files in each category
	summary.FileCounts["v2mom"] = countJSONFiles(filepath.Join(path, "v2mom"))
	summary.FileCounts["capability"] = countJSONFiles(filepath.Join(path, "capability"))
	summary.FileCounts["maturity"] = countJSONFiles(filepath.Join(path, "maturity"))
	summary.FileCounts["roadmap"] = countJSONFiles(filepath.Join(path, "roadmap"))

	return summary
}

// countJSONFiles counts JSON files in a directory (recursively)
func countJSONFiles(dir string) int {
	count := 0
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error { //nolint:gosec // G703: Path from embedded samples directory
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
			count++
		}
		return nil
	})
	return count
}

// handleGetSample returns detailed information about a specific sample
func (s *Server) handleGetSample(w http.ResponseWriter, r *http.Request) {
	sampleID := chi.URLParam(r, "sampleId")

	// Validate sampleID for path traversal
	if err := osutil.ValidateNoTraversal(sampleID); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.GetSampleResponse{
			Error: "Invalid sample ID",
		})
		return
	}

	samplesDir := s.getSamplesDir()
	if samplesDir == "" {
		s.writeJSON(w, http.StatusNotFound, api.GetSampleResponse{
			Error: "Samples directory not found",
		})
		return
	}

	samplePath := filepath.Join(samplesDir, sampleID)
	if _, err := os.Stat(samplePath); os.IsNotExist(err) { //nolint:gosec // G703: Path from embedded samples directory
		s.writeJSON(w, http.StatusNotFound, api.GetSampleResponse{
			Error: "Sample not found: " + sampleID,
		})
		return
	}

	// Load summary
	summary := s.loadSampleSummary(sampleID, samplePath)

	detail := api.SampleDetail{
		SampleSummary: summary,
	}

	// Load project.json
	projectPath := filepath.Join(samplePath, "project.json")
	if data, err := os.ReadFile(projectPath); err == nil { //nolint:gosec // G703: Path from embedded samples directory
		var projectJSON map[string]any
		if json.Unmarshal(data, &projectJSON) == nil {
			detail.ProjectJSON = projectJSON
		}
	}

	// Load README.md
	readmePath := filepath.Join(samplePath, "README.md")
	if data, err := os.ReadFile(readmePath); err == nil { //nolint:gosec // G703: Path from embedded samples directory
		detail.README = string(data)
	}

	s.writeJSON(w, http.StatusOK, api.GetSampleResponse{Sample: detail})
}
