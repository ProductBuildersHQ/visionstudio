package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/ProductBuildersHQ/visionstudio/pkg/api"
	"github.com/ProductBuildersHQ/visionstudio/pkg/config"
)

// handleGetRoadmap returns the roadmap for a project
func (s *Server) handleGetRoadmap(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetRoadmapResponse{
			Error: "Project not found",
		})
		return
	}

	roadmap, err := s.loadRoadmap(tracked.Path)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetRoadmapResponse{
			Error: "Roadmap not found",
		})
		return
	}

	s.writeJSON(w, http.StatusOK, api.GetRoadmapResponse{Roadmap: roadmap})
}

// loadRoadmap loads the roadmap from a project
func (s *Server) loadRoadmap(projectPath string) (api.Roadmap, error) {
	roadmapDir := filepath.Join(projectPath, "roadmap")

	// Find first roadmap file
	var roadmapFile string

	entries, err := os.ReadDir(roadmapDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				roadmapFile = filepath.Join(roadmapDir, entry.Name())
				break
			}
		}
	}

	// Fallback: check for roadmap.json in project root
	if roadmapFile == "" {
		roadmapFile = filepath.Join(projectPath, "roadmap.json")
	}

	data, err := os.ReadFile(roadmapFile) //nolint:gosec // G703: Path from tracked project config
	if err != nil {
		return api.Roadmap{}, err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return api.Roadmap{}, err
	}

	return s.parseRoadmap(raw), nil
}

// parseRoadmap converts raw JSON to Roadmap struct
func (s *Server) parseRoadmap(raw map[string]any) api.Roadmap {
	roadmap := api.Roadmap{}

	if meta, ok := raw["metadata"].(map[string]any); ok {
		roadmap.Metadata = meta
	}

	if items, ok := raw["items"].([]any); ok {
		roadmap.Items = make([]api.RoadmapItem, 0, len(items))
		for _, item := range items {
			if m, ok := item.(map[string]any); ok {
				roadmap.Items = append(roadmap.Items, s.parseRoadmapItem(m))
			}
		}
	}

	return roadmap
}

// parseRoadmapItem converts a raw item to RoadmapItem struct
func (s *Server) parseRoadmapItem(raw map[string]any) api.RoadmapItem {
	item := api.RoadmapItem{}

	if v, ok := raw["id"].(string); ok {
		item.ID = v
	}
	// Try "title" first, fall back to "name"
	if v, ok := raw["title"].(string); ok {
		item.Title = v
	} else if v, ok := raw["name"].(string); ok {
		item.Title = v
	}
	if v, ok := raw["description"].(string); ok {
		item.Description = v
	}
	if v, ok := raw["status"].(string); ok {
		item.Status = v
	}
	if v, ok := raw["priority"].(string); ok {
		item.Priority = v
	}
	if v, ok := raw["quarter"].(string); ok {
		item.Quarter = v
	}
	// Handle effort as string or object (use tshirt_size if object)
	if v, ok := raw["effort"].(string); ok {
		item.Effort = v
	} else if effortObj, ok := raw["effort"].(map[string]any); ok {
		if size, ok := effortObj["tshirt_size"].(string); ok {
			item.Effort = size
		}
	}
	if v, ok := raw["rice"].(map[string]any); ok {
		item.RICE = v
	}
	if v, ok := raw["capability_refs"].([]any); ok {
		item.CapabilityRefs = toStringSlice(v)
	}
	if v, ok := raw["goal_refs"].([]any); ok {
		item.GoalRefs = toStringSlice(v)
	}

	return item
}

// toStringSlice converts []any to []string
func toStringSlice(items []any) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
