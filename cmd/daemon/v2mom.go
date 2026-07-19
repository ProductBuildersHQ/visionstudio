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
	"github.com/ProductBuildersHQ/visionstudio/pkg/config"
)

// handleListV2MOMs lists all V2MOMs in a project
func (s *Server) handleListV2MOMs(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.ListV2MOMsResponse{V2MOMs: []api.V2MOMSummary{}})
		return
	}

	v2moms := s.discoverV2MOMs(tracked.Path)
	s.writeJSON(w, http.StatusOK, api.ListV2MOMsResponse{V2MOMs: v2moms})
}

// discoverV2MOMs finds all V2MOM files in a project
//
//nolint:dupl // Similar structure to discoverCapabilities but different types
func (s *Server) discoverV2MOMs(projectPath string) []api.V2MOMSummary {
	v2moms := make([]api.V2MOMSummary, 0)

	v2momDir := filepath.Join(projectPath, "v2mom")
	if _, err := os.Stat(v2momDir); os.IsNotExist(err) { //nolint:gosec // G703: Path from tracked project config
		return v2moms
	}

	_ = filepath.WalkDir(v2momDir, func(path string, d os.DirEntry, err error) error { //nolint:gosec // G703: Path from tracked project config
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		// Load V2MOM to get metadata
		data, err := os.ReadFile(path) //nolint:gosec // G122: Path from WalkDir in trusted project directory
		if err != nil {
			return nil
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil
		}

		// Extract ID and name from metadata
		id := strings.TrimSuffix(d.Name(), ".json")
		id = strings.TrimSuffix(id, ".v2mom")
		name := id
		var parentID string

		if meta, ok := raw["metadata"].(map[string]any); ok {
			if n, ok := meta["name"].(string); ok {
				name = n
			}
			if metaID, ok := meta["id"].(string); ok {
				id = metaID
			}
			if pid, ok := meta["parentId"].(string); ok {
				parentID = pid
			}
		}

		relPath, _ := filepath.Rel(projectPath, path)

		v2moms = append(v2moms, api.V2MOMSummary{
			ID:       id,
			Name:     name,
			ParentID: parentID,
			Path:     relPath,
		})

		return nil
	})

	return v2moms
}

// handleGetV2MOM returns a specific V2MOM
func (s *Server) handleGetV2MOM(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	v2momID := chi.URLParam(r, "v2momId")

	// Validate v2momID for path traversal
	if err := osutil.ValidateNoTraversal(v2momID); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.GetV2MOMResponse{
			Error: "Invalid V2MOM ID",
		})
		return
	}

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetV2MOMResponse{
			Error: "Project not found",
		})
		return
	}

	// Discover V2MOMs and find the path for the requested ID
	v2moms := s.discoverV2MOMs(tracked.Path)
	var v2momPath string
	for _, v := range v2moms {
		if v.ID == v2momID {
			v2momPath = v.Path
			break
		}
	}

	if v2momPath == "" {
		// Fall back to loading by ID pattern
		v2mom, err := s.loadV2MOM(tracked.Path, v2momID)
		if err != nil {
			s.writeJSON(w, http.StatusNotFound, api.GetV2MOMResponse{
				Error: "V2MOM not found: " + v2momID,
			})
			return
		}
		s.writeJSON(w, http.StatusOK, api.GetV2MOMResponse{V2MOM: v2mom})
		return
	}

	// Load using discovered path
	v2mom, err := s.loadV2MOMFromPath(tracked.Path, v2momPath)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetV2MOMResponse{
			Error: "V2MOM not found: " + v2momID,
		})
		return
	}

	s.writeJSON(w, http.StatusOK, api.GetV2MOMResponse{V2MOM: v2mom})
}

// loadV2MOM loads a V2MOM by ID
func (s *Server) loadV2MOM(projectPath, v2momID string) (api.V2MOM, error) {
	v2momDir := filepath.Join(projectPath, "v2mom")

	// Try different file patterns
	patterns := []string{
		filepath.Join(v2momDir, v2momID+".json"),
		filepath.Join(v2momDir, v2momID+".v2mom.json"),
		filepath.Join(v2momDir, "teams", v2momID+".json"),
		filepath.Join(v2momDir, "teams", v2momID+".v2mom.json"),
	}

	var data []byte
	var err error
	for _, pattern := range patterns {
		data, err = os.ReadFile(pattern) //nolint:gosec // G703: Path from tracked project config
		if err == nil {
			break
		}
	}

	if err != nil {
		return api.V2MOM{}, err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return api.V2MOM{}, err
	}

	return s.parseV2MOM(raw), nil
}

// loadV2MOMFromPath loads a V2MOM from a relative path
func (s *Server) loadV2MOMFromPath(projectPath, relPath string) (api.V2MOM, error) {
	fullPath := filepath.Join(projectPath, relPath)

	data, err := os.ReadFile(fullPath) //nolint:gosec // G703: Path from tracked project config
	if err != nil {
		return api.V2MOM{}, err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return api.V2MOM{}, err
	}

	return s.parseV2MOM(raw), nil
}

// parseV2MOM converts raw JSON to V2MOM struct
func (s *Server) parseV2MOM(raw map[string]any) api.V2MOM {
	v2mom := api.V2MOM{}

	if meta, ok := raw["metadata"].(map[string]any); ok {
		v2mom.Metadata = meta
	}

	if vision, ok := raw["vision"].(string); ok {
		v2mom.Vision = vision
	}

	if values, ok := raw["values"].([]any); ok {
		v2mom.Values = toMapSlice(values)
	}

	if methods, ok := raw["methods"].([]any); ok {
		v2mom.Methods = toMapSlice(methods)
	}

	if obstacles, ok := raw["obstacles"].([]any); ok {
		v2mom.Obstacles = toMapSlice(obstacles)
	}

	if measures, ok := raw["measures"].([]any); ok {
		v2mom.Measures = toMapSlice(measures)
	}

	return v2mom
}

// toMapSlice converts []any to []map[string]any
func toMapSlice(items []any) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			result = append(result, m)
		}
	}
	return result
}

// handleGetV2MOMCascade returns hierarchical V2MOMs
func (s *Server) handleGetV2MOMCascade(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetV2MOMCascadeResponse{
			Error: "Project not found",
		})
		return
	}

	// Discover all V2MOMs
	summaries := s.discoverV2MOMs(tracked.Path)

	if len(summaries) == 0 {
		s.writeJSON(w, http.StatusOK, api.GetV2MOMCascadeResponse{
			Cascade: api.V2MOMCascade{},
		})
		return
	}

	// Find root (no parentId) and children
	var rootSummary *api.V2MOMSummary
	childSummaries := make([]api.V2MOMSummary, 0)

	for i, summary := range summaries {
		if summary.ParentID == "" {
			rootSummary = &summaries[i]
		} else {
			childSummaries = append(childSummaries, summary)
		}
	}

	// If no explicit root, use first one
	if rootSummary == nil && len(summaries) > 0 {
		rootSummary = &summaries[0]
		childSummaries = summaries[1:]
	}

	cascade := api.V2MOMCascade{}

	// Load root V2MOM using its path
	if rootSummary != nil {
		root, err := s.loadV2MOMFromPath(tracked.Path, rootSummary.Path)
		if err == nil {
			cascade.Root = root
		}
	}

	// Load children using their paths
	for _, childSummary := range childSummaries {
		child, err := s.loadV2MOMFromPath(tracked.Path, childSummary.Path)
		if err == nil {
			cascade.Children = append(cascade.Children, child)
		}
	}

	s.writeJSON(w, http.StatusOK, api.GetV2MOMCascadeResponse{Cascade: cascade})
}
