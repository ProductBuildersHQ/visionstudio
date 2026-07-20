package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
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
