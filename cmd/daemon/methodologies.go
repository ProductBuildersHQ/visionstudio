package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ProductBuildersHQ/visionspec/pkg/types"
	"github.com/ProductBuildersHQ/visionstudio/pkg/api"
	"github.com/ProductBuildersHQ/visionstudio/pkg/config"
)

// Note: This file implements methodology management endpoints:
// GET  /api/methodologies/requirements     - List requirements methodologies (profiles)
// GET  /api/methodologies/implementation   - List implementation methodologies
// GET  /api/projects/{project}/methodology - Get project's methodology config
// PUT  /api/projects/{project}/methodology - Update methodology selection

// handleListRequirementsMethodologies returns available requirements methodologies (profiles)
func (s *Server) handleListRequirementsMethodologies(w http.ResponseWriter, r *http.Request) {
	// Requirements methodologies are the same as profiles
	profiles := availableProfiles()

	// Add type field to indicate these are requirements methodologies
	for i := range profiles {
		profiles[i].Type = "requirements"
	}

	s.writeJSON(w, http.StatusOK, api.ListProfilesResponse{Profiles: profiles})
}

// handleListImplementationMethodologies returns available implementation methodologies
func (s *Server) handleListImplementationMethodologies(w http.ResponseWriter, r *http.Request) {
	methodologies := types.AvailableImplementationMethodologies()

	// Convert to API types
	apiMethodologies := make([]api.ImplementationMethodologySummary, len(methodologies))
	for i, m := range methodologies {
		apiMethodologies[i] = api.ImplementationMethodologySummary{
			ID:          string(m.ID),
			Name:        m.Name,
			Description: m.Description,
		}
	}

	s.writeJSON(w, http.StatusOK, api.ListImplementationMethodologiesResponse{
		Methodologies: apiMethodologies,
	})
}

// handleGetProjectMethodology returns a project's methodology configuration
func (s *Server) handleGetProjectMethodology(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	// Get project from config
	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetProjectMethodologyResponse{
			Error: err.Error(),
		})
		return
	}

	// Determine requirements methodology (use profile if not explicitly set)
	reqMethodology := tracked.RequirementsMethodology
	if reqMethodology == "" {
		reqMethodology = tracked.Profile
	}

	// Determine implementation methodology
	implMethodology := tracked.ImplementationMethodology
	if implMethodology == "" {
		implMethodology = "none"
	}

	s.writeJSON(w, http.StatusOK, api.GetProjectMethodologyResponse{
		Config: api.ProjectMethodologyConfig{
			RequirementsMethodology:   reqMethodology,
			ImplementationMethodology: implMethodology,
		},
	})
}

// handleUpdateProjectMethodology updates a project's methodology configuration
func (s *Server) handleUpdateProjectMethodology(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	var req api.UpdateProjectMethodologyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.UpdateProjectMethodologyResponse{
			Error: "Invalid request body",
		})
		return
	}

	// Get project from config
	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.UpdateProjectMethodologyResponse{
			Error: err.Error(),
		})
		return
	}

	// Validate implementation methodology if provided
	if req.ImplementationMethodology != "" {
		implMethod := types.ImplementationMethodology(req.ImplementationMethodology)
		if !types.IsValidImplementationMethodology(implMethod) {
			s.writeJSON(w, http.StatusBadRequest, api.UpdateProjectMethodologyResponse{
				Error: "Invalid implementation methodology: " + req.ImplementationMethodology,
			})
			return
		}
	}

	// Update the configuration
	if req.RequirementsMethodology != "" {
		tracked.RequirementsMethodology = req.RequirementsMethodology
		// Also update profile for backwards compatibility
		tracked.Profile = req.RequirementsMethodology
	}

	if req.ImplementationMethodology != "" {
		tracked.ImplementationMethodology = req.ImplementationMethodology
	}

	// Save to config
	if err := config.UpdateProject(tracked); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.UpdateProjectMethodologyResponse{
			Error: "Failed to update project: " + err.Error(),
		})
		return
	}

	// Return updated configuration
	reqMethodology := tracked.RequirementsMethodology
	if reqMethodology == "" {
		reqMethodology = tracked.Profile
	}

	implMethodology := tracked.ImplementationMethodology
	if implMethodology == "" {
		implMethodology = "none"
	}

	s.logger.Info("Updated project methodology",
		"project", projectName,
		"requirementsMethodology", reqMethodology,
		"implementationMethodology", implMethodology,
	)

	s.writeJSON(w, http.StatusOK, api.UpdateProjectMethodologyResponse{
		Config: api.ProjectMethodologyConfig{
			RequirementsMethodology:   reqMethodology,
			ImplementationMethodology: implMethodology,
		},
	})
}
