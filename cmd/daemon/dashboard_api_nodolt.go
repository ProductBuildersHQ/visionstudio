//go:build !dolt

package main

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// registerDashboardAPI is a stub when built without the dolt tag.
// The endpoints exist but report that DB support is unavailable so the
// web app shows a clear error instead of a 404.
func registerDashboardAPI(r chi.Router, _ *slog.Logger) {
	unavailable := func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "dashboard API requires a daemon built with -tags dolt", http.StatusNotImplemented)
	}
	r.Get("/execution", unavailable)
	r.Get("/spend", unavailable)
	r.Get("/maturity", unavailable)
	r.Get("/specs", unavailable)
}
