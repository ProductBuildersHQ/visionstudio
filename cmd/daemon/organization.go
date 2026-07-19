package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ProductBuildersHQ/visionstudio/pkg/api"
	"github.com/ProductBuildersHQ/visionstudio/pkg/config"
)

// handleGetOrganization returns the current organization configuration
func (s *Server) handleGetOrganization(w http.ResponseWriter, r *http.Request) {
	org, err := config.LoadOrganization()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.GetOrganizationResponse{
			Error: err.Error(),
		})
		return
	}

	if org == nil {
		s.writeJSON(w, http.StatusOK, api.GetOrganizationResponse{
			Organization: nil,
		})
		return
	}

	s.writeJSON(w, http.StatusOK, api.GetOrganizationResponse{
		Organization: &api.Organization{
			ID:              org.ID,
			Name:            org.Name,
			Description:     org.Description,
			V2MOMPath:       org.V2MOMPath,
			FiscalYearStart: org.FiscalYearStart,
			CreatedAt:       org.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       org.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// handleCreateOrganization creates a new organization
func (s *Server) handleCreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req api.CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.CreateOrganizationResponse{
			Error: "Invalid request body",
		})
		return
	}

	if req.Name == "" {
		s.writeJSON(w, http.StatusBadRequest, api.CreateOrganizationResponse{
			Error: "Name is required",
		})
		return
	}

	org, err := config.CreateOrganization(req.Name)
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.CreateOrganizationResponse{
			Error: err.Error(),
		})
		return
	}

	// Set optional fields
	if req.Description != "" {
		org.Description = req.Description
	}
	if req.FiscalYearStart != "" {
		org.FiscalYearStart = req.FiscalYearStart
	}

	if err := config.SaveOrganization(org); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.CreateOrganizationResponse{
			Error: err.Error(),
		})
		return
	}

	s.logger.Info("Created organization", "name", org.Name, "id", org.ID)

	s.writeJSON(w, http.StatusCreated, api.CreateOrganizationResponse{
		Organization: &api.Organization{
			ID:              org.ID,
			Name:            org.Name,
			Description:     org.Description,
			V2MOMPath:       org.V2MOMPath,
			FiscalYearStart: org.FiscalYearStart,
			CreatedAt:       org.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       org.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// handleUpdateOrganization updates the organization configuration
func (s *Server) handleUpdateOrganization(w http.ResponseWriter, r *http.Request) {
	var req api.UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.UpdateOrganizationResponse{
			Error: "Invalid request body",
		})
		return
	}

	org, err := config.LoadOrganization()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.UpdateOrganizationResponse{
			Error: err.Error(),
		})
		return
	}

	if org == nil {
		s.writeJSON(w, http.StatusNotFound, api.UpdateOrganizationResponse{
			Error: "No organization exists. Create one first.",
		})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		org.Name = req.Name
	}
	if req.Description != "" {
		org.Description = req.Description
	}
	if req.V2MOMPath != "" {
		org.V2MOMPath = req.V2MOMPath
	}
	if req.FiscalYearStart != "" {
		org.FiscalYearStart = req.FiscalYearStart
	}

	if err := config.SaveOrganization(org); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.UpdateOrganizationResponse{
			Error: err.Error(),
		})
		return
	}

	s.logger.Info("Updated organization", "name", org.Name)

	s.writeJSON(w, http.StatusOK, api.UpdateOrganizationResponse{
		Organization: &api.Organization{
			ID:              org.ID,
			Name:            org.Name,
			Description:     org.Description,
			V2MOMPath:       org.V2MOMPath,
			FiscalYearStart: org.FiscalYearStart,
			CreatedAt:       org.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       org.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// handleDeleteOrganization removes the organization configuration
func (s *Server) handleDeleteOrganization(w http.ResponseWriter, r *http.Request) {
	if err := config.DeleteOrganization(); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	s.logger.Info("Deleted organization")
	s.writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleListOrganizationV2MOMs lists all organization-level V2MOMs
func (s *Server) handleListOrganizationV2MOMs(w http.ResponseWriter, r *http.Request) {
	v2momPath, err := config.GetOrganizationV2MOMPath()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.ListOrganizationV2MOMsResponse{
			Error: err.Error(),
		})
		return
	}

	v2moms, err := s.loadOrganizationV2MOMs(v2momPath)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.ListOrganizationV2MOMsResponse{
			Error: err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, api.ListOrganizationV2MOMsResponse{
		V2MOMs: v2moms,
	})
}

// handleGetOrganizationV2MOM gets a specific organization V2MOM by ID
func (s *Server) handleGetOrganizationV2MOM(w http.ResponseWriter, r *http.Request) {
	v2momID := chi.URLParam(r, "v2momId")

	v2momPath, err := config.GetOrganizationV2MOMPath()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.GetOrganizationV2MOMResponse{
			Error: err.Error(),
		})
		return
	}

	v2moms, err := s.loadOrganizationV2MOMs(v2momPath)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.GetOrganizationV2MOMResponse{
			Error: err.Error(),
		})
		return
	}

	for _, v := range v2moms {
		if v.ID == v2momID {
			s.writeJSON(w, http.StatusOK, api.GetOrganizationV2MOMResponse{
				V2MOM: v,
			})
			return
		}
	}

	s.writeJSON(w, http.StatusNotFound, api.GetOrganizationV2MOMResponse{
		Error: "V2MOM not found: " + v2momID,
	})
}

// handleGetOrganizationCascade returns the full organization → projects cascade
func (s *Server) handleGetOrganizationCascade(w http.ResponseWriter, r *http.Request) {
	// Load organization
	org, err := config.LoadOrganization()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.GetOrganizationCascadeResponse{
			Error: err.Error(),
		})
		return
	}

	cascade := api.OrganizationCascade{
		ProjectV2MOMs:   make(map[string][]api.V2MOMSummary),
		AlignmentScores: make(map[string]float64),
	}

	// Add organization if it exists
	if org != nil {
		cascade.Organization = &api.Organization{
			ID:              org.ID,
			Name:            org.Name,
			Description:     org.Description,
			V2MOMPath:       org.V2MOMPath,
			FiscalYearStart: org.FiscalYearStart,
			CreatedAt:       org.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       org.UpdatedAt.Format(time.RFC3339),
		}

		// Load org V2MOMs
		v2momPath, err := config.GetOrganizationV2MOMPath()
		if err == nil {
			v2moms, err := s.loadOrganizationV2MOMs(v2momPath)
			if err == nil {
				cascade.OrgV2MOMs = v2moms
			}
		}
	}

	// Load project V2MOMs
	projects, err := config.LoadProjects()
	if err == nil {
		for _, p := range projects.Projects {
			summaries := s.loadProjectV2MOMSummaries(p.Path)
			if len(summaries) > 0 {
				cascade.ProjectV2MOMs[p.Name] = summaries
			}
		}
	}

	// TODO: Load alignments from project configs

	s.writeJSON(w, http.StatusOK, api.GetOrganizationCascadeResponse{
		Cascade: cascade,
	})
}

// loadOrganizationV2MOMs loads all V2MOMs from the organization V2MOM directory
func (s *Server) loadOrganizationV2MOMs(v2momPath string) ([]api.OrganizationV2MOM, error) {
	v2moms := []api.OrganizationV2MOM{}

	// Check if directory exists
	if _, err := os.Stat(v2momPath); os.IsNotExist(err) {
		return v2moms, nil
	}

	// Walk directory for .json files
	err := filepath.Walk(v2momPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only process .json files
		if !strings.HasSuffix(info.Name(), ".json") && !strings.HasSuffix(info.Name(), ".v2mom.json") {
			return nil
		}

		// Load and parse the V2MOM
		data, err := os.ReadFile(path) //nolint:gosec // G122: Path from Walk in trusted project directory
		if err != nil {
			s.logger.Warn("Failed to read V2MOM file", "path", path, "error", err)
			return nil
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			s.logger.Warn("Failed to parse V2MOM file", "path", path, "error", err)
			return nil
		}

		v2mom := s.parseOrganizationV2MOM(raw, path, info.ModTime())
		v2moms = append(v2moms, v2mom)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return v2moms, nil
}

// parseOrganizationV2MOM parses a raw V2MOM map into an OrganizationV2MOM
func (s *Server) parseOrganizationV2MOM(raw map[string]any, path string, modTime time.Time) api.OrganizationV2MOM {
	v2mom := api.OrganizationV2MOM{
		Scope:       api.V2MOMScopeOrganization,
		Path:        path,
		LastUpdated: modTime.Format(time.RFC3339),
	}

	// Extract metadata
	if metadata, ok := raw["metadata"].(map[string]any); ok {
		if id, ok := metadata["id"].(string); ok {
			v2mom.ID = id
		}
		if name, ok := metadata["name"].(string); ok {
			v2mom.Name = name
		}
		if fy, ok := metadata["fiscalYear"].(string); ok {
			v2mom.FiscalYear = fy
		}
		if owner, ok := metadata["owner"].(string); ok {
			v2mom.Owner = owner
		}
		if parentID, ok := metadata["parentId"].(string); ok {
			v2mom.ParentID = parentID
		}
	}

	// Use filename as ID if not in metadata
	if v2mom.ID == "" {
		base := filepath.Base(path)
		v2mom.ID = strings.TrimSuffix(strings.TrimSuffix(base, ".json"), ".v2mom")
	}
	if v2mom.Name == "" {
		v2mom.Name = v2mom.ID
	}

	// Extract vision
	if vision, ok := raw["vision"].(string); ok {
		v2mom.Vision = vision
	}

	// Extract values
	if values, ok := raw["values"].([]any); ok {
		for i, v := range values {
			if val, ok := v.(map[string]any); ok {
				value := api.V2MOMValue{
					ID:       getString(val, "id"),
					Name:     getString(val, "name"),
					Priority: i + 1,
				}
				if desc, ok := val["description"].(string); ok {
					value.Description = desc
				}
				v2mom.Values = append(v2mom.Values, value)
			}
		}
	}

	// Extract methods
	if methods, ok := raw["methods"].([]any); ok {
		for i, m := range methods {
			if meth, ok := m.(map[string]any); ok {
				method := api.V2MOMMethod{
					ID:       getString(meth, "id"),
					Name:     getString(meth, "name"),
					Priority: i + 1,
				}
				if desc, ok := meth["description"].(string); ok {
					method.Description = desc
				}
				if status, ok := meth["status"].(string); ok {
					method.Status = status
				}
				if owner, ok := meth["owner"].(string); ok {
					method.Owner = owner
				}
				v2mom.Methods = append(v2mom.Methods, method)
			}
		}
	}

	// Extract obstacles
	if obstacles, ok := raw["obstacles"].([]any); ok {
		for _, o := range obstacles {
			if obs, ok := o.(map[string]any); ok {
				obstacle := api.V2MOMItem{
					ID:   getString(obs, "id"),
					Name: getString(obs, "name"),
				}
				if desc, ok := obs["description"].(string); ok {
					obstacle.Description = desc
				}
				if sev, ok := obs["severity"].(string); ok {
					obstacle.Severity = sev
				}
				v2mom.Obstacles = append(v2mom.Obstacles, obstacle)
			}
		}
	}

	// Extract measures
	if measures, ok := raw["measures"].([]any); ok {
		for _, m := range measures {
			if meas, ok := m.(map[string]any); ok {
				measure := api.V2MOMMeasure{
					ID:   getString(meas, "id"),
					Name: getString(meas, "name"),
				}
				if desc, ok := meas["description"].(string); ok {
					measure.Description = desc
				}
				if target, ok := meas["target"].(string); ok {
					measure.Target = target
				}
				if current, ok := meas["current"].(string); ok {
					measure.Current = current
				}
				if prog, ok := meas["progress"].(float64); ok {
					measure.Progress = prog
				}
				v2mom.Measures = append(v2mom.Measures, measure)
			}
		}
	}

	return v2mom
}

// loadProjectV2MOMSummaries loads V2MOM summaries for a project
func (s *Server) loadProjectV2MOMSummaries(projectPath string) []api.V2MOMSummary {
	summaries := []api.V2MOMSummary{}

	v2momDir := filepath.Join(projectPath, "v2mom")
	if _, err := os.Stat(v2momDir); os.IsNotExist(err) {
		return summaries
	}

	err := filepath.Walk(v2momDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		if !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		data, err := os.ReadFile(path) //nolint:gosec // G122: Path from Walk in trusted project directory
		if err != nil {
			return nil
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil
		}

		// Extract summary info
		summary := api.V2MOMSummary{
			Path: path,
		}

		if metadata, ok := raw["metadata"].(map[string]any); ok {
			if id, ok := metadata["id"].(string); ok {
				summary.ID = id
			}
			if name, ok := metadata["name"].(string); ok {
				summary.Name = name
			}
			if parentID, ok := metadata["parentId"].(string); ok {
				summary.ParentID = parentID
			}
		}

		if summary.ID == "" {
			base := filepath.Base(path)
			summary.ID = strings.TrimSuffix(strings.TrimSuffix(base, ".json"), ".v2mom")
		}
		if summary.Name == "" {
			summary.Name = summary.ID
		}

		summaries = append(summaries, summary)
		return nil
	})

	if err != nil {
		s.logger.Warn("Error walking V2MOM directory", "path", v2momDir, "error", err)
	}

	return summaries
}

// getString safely gets a string from a map
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
