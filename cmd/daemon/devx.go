package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
)

// devxDashboardPath is where VisionStudio looks for the OmniDevX dashboard
// export. It shares the `~/.plexusone/omnidevx/` home used by the
// omnidevx-core local event store; the file itself is produced out of band
// by `devfolio devx dashboard -o <path>`.
//
// VisionStudio deliberately never reads the OmniDevX event store or
// canonical report types directly — only this already-generated, already
// disclosure-scoped dashboard-IR file. That boundary is the point: the
// producer (devfolio) decides what's safe to show; VisionStudio is a
// read-only file consumer, never a live query path into someone else's
// local data store.
func devxDashboardPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".plexusone", "omnidevx", "dashboard.json"), nil
}

// handleGetDevXDashboard serves the OmniDevX dashboard-IR file as-is. The
// file is dashforge dashboardir.Dashboard JSON; VisionStudio does not parse
// or interpret its structure server-side, only passes it through for the
// frontend to render.
func (s *Server) handleGetDevXDashboard(w http.ResponseWriter, r *http.Request) {
	path, err := devxDashboardPath()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "resolving dashboard path: " + err.Error(),
		})
		return
	}

	data, err := os.ReadFile(path) //nolint:gosec // G304: fixed, non-user-controlled path
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "no dashboard found at " + path + " — generate one with `devfolio devx dashboard -o " + path + "`",
			})
			return
		}
		s.logger.Error("Failed to read DevX dashboard", "path", path, "error", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "reading dashboard file: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		s.logger.Error("Failed to write DevX dashboard response", "error", err)
	}
}

// devxReportsDir returns the base directory for period report files.
func devxReportsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".plexusone", "omnidevx", "reports"), nil
}

// DevXPeriodEntry describes one available period report.
type DevXPeriodEntry struct {
	Type  string `json:"type"`  // "weekly", "monthly", "quarterly"
	Label string `json:"label"` // e.g., "2026-W30", "2026-07", "2026-Q3"
	Path  string `json:"path"`  // relative path from reports dir
}

// validPeriodType checks if a period type is valid.
func validPeriodType(t string) bool {
	return t == "weekly" || t == "monthly" || t == "quarterly"
}

// validPeriodLabel checks if a label is safe (no path traversal).
var validLabelPattern = regexp.MustCompile(`^[0-9]{4}(-W[0-9]{2}|-[0-9]{2}|-Q[1-4])$`)

func validPeriodLabel(label string) bool {
	return validLabelPattern.MatchString(label)
}

// handleListDevXPeriods lists available period reports.
// GET /api/devx/periods
func (s *Server) handleListDevXPeriods(w http.ResponseWriter, r *http.Request) {
	reportsDir, err := devxReportsDir()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "resolving reports directory: " + err.Error(),
		})
		return
	}

	var entries []DevXPeriodEntry

	periodTypes := []string{"weekly", "monthly", "quarterly"}
	for _, pt := range periodTypes {
		dir := filepath.Join(reportsDir, pt)
		files, err := os.ReadDir(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			s.logger.Warn("Failed to read period directory", "dir", dir, "error", err)
			continue
		}

		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
				continue
			}
			label := strings.TrimSuffix(f.Name(), ".json")
			entries = append(entries, DevXPeriodEntry{
				Type:  pt,
				Label: label,
				Path:  filepath.Join(pt, f.Name()),
			})
		}
	}

	// Sort by label descending (most recent first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Label > entries[j].Label
	})

	s.writeJSON(w, http.StatusOK, entries)
}

// handleGetDevXPeriodDashboard serves a specific period report.
// GET /api/devx/reports/{periodType}/{label}
func (s *Server) handleGetDevXPeriodDashboard(w http.ResponseWriter, r *http.Request) {
	periodType := chi.URLParam(r, "periodType")
	label := chi.URLParam(r, "label")

	// Validate to prevent path traversal
	if !validPeriodType(periodType) {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid period type: must be weekly, monthly, or quarterly",
		})
		return
	}
	if !validPeriodLabel(label) {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid label format",
		})
		return
	}

	reportsDir, err := devxReportsDir()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "resolving reports directory: " + err.Error(),
		})
		return
	}

	path := filepath.Join(reportsDir, periodType, label+".json")

	data, err := os.ReadFile(path) //nolint:gosec // G304: path validated above
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "period report not found: " + periodType + "/" + label,
			})
			return
		}
		s.logger.Error("Failed to read period report", "path", path, "error", err)
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "reading report file: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		s.logger.Error("Failed to write period report response", "error", err)
	}
}
