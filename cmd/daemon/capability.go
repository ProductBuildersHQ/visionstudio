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

// handleListCapabilities lists all capability stacks in a project
func (s *Server) handleListCapabilities(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.ListCapabilitiesResponse{Capabilities: []api.CapabilitySummary{}})
		return
	}

	capabilities := s.discoverCapabilities(tracked.Path)
	s.writeJSON(w, http.StatusOK, api.ListCapabilitiesResponse{Capabilities: capabilities})
}

// discoverCapabilities finds all capability stack files in a project
//
//nolint:dupl // Similar structure to discoverV2MOMs but different types
func (s *Server) discoverCapabilities(projectPath string) []api.CapabilitySummary {
	capabilities := make([]api.CapabilitySummary, 0)

	capabilityDir := filepath.Join(projectPath, "capability")
	if _, err := os.Stat(capabilityDir); os.IsNotExist(err) { //nolint:gosec // G703: Path from tracked project config
		return capabilities
	}

	_ = filepath.WalkDir(capabilityDir, func(path string, d os.DirEntry, err error) error { //nolint:gosec // G703: Path from tracked project config
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		// Load capability to get metadata
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
		id = strings.TrimSuffix(id, ".capability")
		name := id
		var domain string

		if meta, ok := raw["metadata"].(map[string]any); ok {
			if n, ok := meta["name"].(string); ok {
				id = n
			}
			if title, ok := meta["title"].(string); ok {
				name = title
			}
			if d, ok := meta["domain"].(string); ok {
				domain = d
			}
		}

		relPath, _ := filepath.Rel(projectPath, path)

		capabilities = append(capabilities, api.CapabilitySummary{
			ID:     id,
			Name:   name,
			Domain: domain,
			Path:   relPath,
		})

		return nil
	})

	return capabilities
}

// handleGetCapability returns a specific capability stack
func (s *Server) handleGetCapability(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	capabilityID := chi.URLParam(r, "capabilityId")

	// Validate capabilityID for path traversal
	if err := osutil.ValidateNoTraversal(capabilityID); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.GetCapabilityResponse{
			Error: "Invalid capability ID",
		})
		return
	}

	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetCapabilityResponse{
			Error: "Project not found",
		})
		return
	}

	// Discover capabilities and find the path for the requested ID
	capabilities := s.discoverCapabilities(tracked.Path)
	var capPath string
	for _, cap := range capabilities {
		if cap.ID == capabilityID {
			capPath = cap.Path
			break
		}
	}

	if capPath == "" {
		// Fall back to loading by ID pattern
		capability, err := s.loadCapability(tracked.Path, capabilityID)
		if err != nil {
			s.writeJSON(w, http.StatusNotFound, api.GetCapabilityResponse{
				Error: "Capability not found: " + capabilityID,
			})
			return
		}
		s.writeJSON(w, http.StatusOK, api.GetCapabilityResponse{Capability: capability})
		return
	}

	// Load using discovered path
	capability, err := s.loadCapabilityFromPath(tracked.Path, capPath)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetCapabilityResponse{
			Error: "Capability not found: " + capabilityID,
		})
		return
	}

	s.writeJSON(w, http.StatusOK, api.GetCapabilityResponse{Capability: capability})
}

// loadCapability loads a capability stack by ID
func (s *Server) loadCapability(projectPath, capabilityID string) (api.CapabilityStack, error) {
	capabilityDir := filepath.Join(projectPath, "capability")

	// Try different file patterns
	patterns := []string{
		filepath.Join(capabilityDir, capabilityID+".json"),
		filepath.Join(capabilityDir, capabilityID+".capability.json"),
		filepath.Join(capabilityDir, "domains", capabilityID+".json"),
		filepath.Join(capabilityDir, "domains", capabilityID+".capability.json"),
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
		return api.CapabilityStack{}, err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return api.CapabilityStack{}, err
	}

	return s.parseCapability(raw), nil
}

// loadCapabilityFromPath loads a capability stack from a relative path
func (s *Server) loadCapabilityFromPath(projectPath, relPath string) (api.CapabilityStack, error) {
	fullPath := filepath.Join(projectPath, relPath)

	data, err := os.ReadFile(fullPath) //nolint:gosec // G703: Path from tracked project config
	if err != nil {
		return api.CapabilityStack{}, err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return api.CapabilityStack{}, err
	}

	return s.parseCapability(raw), nil
}

// parseCapability converts raw JSON to CapabilityStack struct
func (s *Server) parseCapability(raw map[string]any) api.CapabilityStack {
	capability := api.CapabilityStack{}

	if meta, ok := raw["metadata"].(map[string]any); ok {
		capability.Metadata = meta
	}

	if layers, ok := raw["layers"].([]any); ok {
		capability.Layers = toMapSlice(layers)
	}

	if categories, ok := raw["categories"].([]any); ok {
		capability.Categories = toMapSlice(categories)
	}

	if caps, ok := raw["capabilities"].([]any); ok {
		capability.Capabilities = toMapSlice(caps)
	}

	if prism, ok := raw["prismIntegration"].(map[string]any); ok {
		capability.PrismIntegration = prism
	}

	return capability
}
