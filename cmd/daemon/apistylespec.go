package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/go-chi/chi/v5"
)

// ---------------------------------------------------------------------------
// Request / Response types
// ---------------------------------------------------------------------------

type apiStyleLintRequest struct {
	Spec    string `json:"spec"`
	Profile string `json:"profile"`
}

type apiStyleViolation struct {
	RuleID     string `json:"ruleId"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Path       string `json:"path"`
	Line       int    `json:"line,omitempty"`
	Column     int    `json:"column,omitempty"`
	EndLine    int    `json:"endLine,omitempty"`
	EndColumn  int    `json:"endColumn,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Category   string `json:"category,omitempty"`
	RuleTitle  string `json:"ruleTitle,omitempty"`
}

type apiStyleViolationSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Infos    int `json:"infos"`
	Hints    int `json:"hints"`
	Total    int `json:"total"`
}

type apiStyleLintResponse struct {
	Status     string                   `json:"status"`
	Violations []apiStyleViolation      `json:"violations"`
	Summary    apiStyleViolationSummary `json:"summary"`
	Profile    string                   `json:"profile"`
}

type apiStyleProfileInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type apiStyleFixRequest struct {
	Spec    string `json:"spec"`
	Profile string `json:"profile"`
}

type apiStyleFixSuggestion struct {
	RuleID         string  `json:"ruleId"`
	Path           string  `json:"path"`
	CurrentValue   string  `json:"currentValue,omitempty"`
	SuggestedValue string  `json:"suggestedValue"`
	Diff           string  `json:"diff,omitempty"`
	Confidence     float64 `json:"confidence"`
	Reasoning      string  `json:"reasoning,omitempty"`
	Breaking       bool    `json:"breaking,omitempty"`
	BreakingReason string  `json:"breakingReason,omitempty"`
}

type apiStyleFixResponse struct {
	Suggestions  []apiStyleFixSuggestion `json:"suggestions"`
	FixedCount   int                     `json:"fixedCount"`
	UnfixedCount int                     `json:"unfixedCount"`
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// apiStyleBinary returns the path to the api-style CLI binary, or an error
// if it cannot be found.
func apiStyleBinary() (string, error) {
	return exec.LookPath("api-style")
}

// runAPIStyle executes the api-style CLI with the given arguments, writing
// specContent to a temp file. It returns the combined stdout+stderr and any
// exec error. The caller does NOT need to clean up the temp file.
func (s *Server) runAPIStyle(ctx context.Context, specContent string, args ...string) ([]byte, error) {
	bin, err := apiStyleBinary()
	if err != nil {
		return nil, err
	}

	tmp, err := os.CreateTemp("", "apistyle-*.yaml")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(specContent); err != nil {
		tmp.Close()
		return nil, err
	}
	tmp.Close()

	// Append the temp file path to args.
	args = append(args, tmp.Name())

	cmd := exec.CommandContext(ctx, bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// api-style lint exits 1 when violations are found — that's not a
	// fatal error for us, so we only propagate exec failures.
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, err
		}
	}

	// Prefer stdout; fall back to stderr when stdout is empty.
	if stdout.Len() > 0 {
		return stdout.Bytes(), nil
	}
	return stderr.Bytes(), nil
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// handleAPIStyleLint lints an OpenAPI spec against a style profile.
//
//	POST /api/extensions/api-style-spec/lint
func (s *Server) handleAPIStyleLint(w http.ResponseWriter, r *http.Request) {
	var req apiStyleLintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body: " + err.Error()})
		return
	}
	if req.Spec == "" {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "spec is required"})
		return
	}

	prof := req.Profile
	if prof == "" {
		prof = "default"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// --config /dev/null bypasses any local include patterns that would
	// reject the temp file name.
	output, err := s.runAPIStyle(ctx, req.Spec, "lint", "--format", "json", "--config", "/dev/null", "--profile", prof)
	if err != nil {
		if _, lookErr := apiStyleBinary(); lookErr != nil {
			s.writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "api-style CLI not found; install with: go install github.com/plexusone/api-style-spec/cmd/api-style@latest"})
			return
		}
		s.logger.Error("api-style lint failed", "error", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lint failed: " + err.Error()})
		return
	}

	// Parse the JSON output from api-style.
	var raw struct {
		Status     string                   `json:"status"`
		Summary    apiStyleViolationSummary `json:"summary"`
		Violations []apiStyleViolation      `json:"violations"`
	}
	if err := json.Unmarshal(output, &raw); err != nil {
		s.logger.Error("failed to parse api-style output", "error", err, "output", string(output))
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to parse lint output"})
		return
	}

	resp := apiStyleLintResponse{
		Status:     raw.Status,
		Violations: raw.Violations,
		Summary:    raw.Summary,
		Profile:    prof,
	}
	if resp.Violations == nil {
		resp.Violations = []apiStyleViolation{}
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// handleAPIStyleListProfiles returns the list of built-in style profiles.
//
//	GET /api/extensions/api-style-spec/profiles
func (s *Server) handleAPIStyleListProfiles(w http.ResponseWriter, _ *http.Request) {
	profiles := []apiStyleProfileInfo{
		{Name: "default", Description: "Balanced API style rules for general-purpose REST APIs"},
		{Name: "azure", Description: "Microsoft Azure REST API guidelines"},
		{Name: "comprehensive", Description: "Exhaustive rule set covering all categories"},
		{Name: "google", Description: "Google Cloud API design guidelines"},
		{Name: "microsoft-graph", Description: "Microsoft Graph API guidelines"},
		{Name: "microsoft-rest", Description: "Microsoft REST API guidelines"},
		{Name: "minimal", Description: "Minimal rule set for quick checks"},
		{Name: "zalando", Description: "Zalando RESTful API guidelines"},
	}
	s.writeJSON(w, http.StatusOK, profiles)
}

// handleAPIStyleGetProfile returns details for a specific profile.
//
//	GET /api/extensions/api-style-spec/profiles/{name}
func (s *Server) handleAPIStyleGetProfile(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	knownProfiles := map[string]apiStyleProfileInfo{
		"default":         {Name: "default", Description: "Balanced API style rules for general-purpose REST APIs"},
		"azure":           {Name: "azure", Description: "Microsoft Azure REST API guidelines"},
		"comprehensive":   {Name: "comprehensive", Description: "Exhaustive rule set covering all categories"},
		"google":          {Name: "google", Description: "Google Cloud API design guidelines"},
		"microsoft-graph": {Name: "microsoft-graph", Description: "Microsoft Graph API guidelines"},
		"microsoft-rest":  {Name: "microsoft-rest", Description: "Microsoft REST API guidelines"},
		"minimal":         {Name: "minimal", Description: "Minimal rule set for quick checks"},
		"zalando":         {Name: "zalando", Description: "Zalando RESTful API guidelines"},
	}

	info, ok := knownProfiles[name]
	if !ok {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": "profile not found: " + name})
		return
	}

	s.writeJSON(w, http.StatusOK, info)
}

// handleAPIStyleSuggestFixes generates fix suggestions for an OpenAPI spec.
//
//	POST /api/extensions/api-style-spec/suggest-fixes
func (s *Server) handleAPIStyleSuggestFixes(w http.ResponseWriter, r *http.Request) {
	var req apiStyleFixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body: " + err.Error()})
		return
	}
	if req.Spec == "" {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "spec is required"})
		return
	}

	prof := req.Profile
	if prof == "" {
		prof = "default"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	output, err := s.runAPIStyle(ctx, req.Spec, "lint", "--format", "json", "--config", "/dev/null", "--suggest-fixes", "--profile", prof)
	if err != nil {
		if _, lookErr := apiStyleBinary(); lookErr != nil {
			s.writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "api-style CLI not found"})
			return
		}
		s.logger.Error("api-style suggest-fixes failed", "error", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "suggest-fixes failed: " + err.Error()})
		return
	}

	// The --suggest-fixes flag adds a "fixSuggestions" field to the lint output.
	var raw struct {
		FixSuggestions []apiStyleFixSuggestion `json:"fixSuggestions"`
		FixReport      *struct {
			Suggestions  []apiStyleFixSuggestion `json:"suggestions"`
			FixedCount   int                     `json:"fixedCount"`
			UnfixedCount int                     `json:"unfixedCount"`
		} `json:"fixReport"`
	}
	if err := json.Unmarshal(output, &raw); err != nil {
		s.logger.Error("failed to parse suggest-fixes output", "error", err, "output", string(output))
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to parse fix output"})
		return
	}

	resp := apiStyleFixResponse{}
	if raw.FixReport != nil {
		resp.Suggestions = raw.FixReport.Suggestions
		resp.FixedCount = raw.FixReport.FixedCount
		resp.UnfixedCount = raw.FixReport.UnfixedCount
	} else if raw.FixSuggestions != nil {
		resp.Suggestions = raw.FixSuggestions
		resp.FixedCount = len(raw.FixSuggestions)
	}
	if resp.Suggestions == nil {
		resp.Suggestions = []apiStyleFixSuggestion{}
	}

	s.writeJSON(w, http.StatusOK, resp)
}
