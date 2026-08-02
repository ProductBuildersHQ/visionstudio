//go:build dolt

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store/doltstore"
	"github.com/ProductBuildersHQ/visionstudio/pkg/webapi"
)

// registerDashboardAPI mounts the unified dashboard JSON API
// (/api/execution, /api/spend, /api/maturity, /api/specs) backed by Dolt.
//
// Connection resolution order:
//  1. VISIONSTUDIO_DSN — MySQL-compatible DSN for Dolt server mode
//  2. VISIONSTUDIO_DATA — directory for embedded Dolt (default if unset)
func registerDashboardAPI(r chi.Router, logger *slog.Logger) {
	closeStore := func(ds *doltstore.DoltStore) func() {
		return func() {
			// The HTTP response has already been written by the time the
			// cleanup runs, so the close error can only be logged.
			if cerr := ds.Close(); cerr != nil {
				logger.Error("close doltstore", "error", cerr)
			}
		}
	}

	// Run schema migration once on the first successful connection so a
	// fresh data directory serves empty responses instead of table errors.
	var migrateOnce sync.Once
	migrate := func(ds *doltstore.DoltStore) error {
		var err error
		migrateOnce.Do(func() {
			err = ds.Migrate(context.Background())
		})
		return err
	}

	connect := func() (*service.Service, func(), error) {
		var ds *doltstore.DoltStore
		var err error
		if dsn := os.Getenv("VISIONSTUDIO_DSN"); dsn != "" {
			ds, err = doltstore.New(dsn)
		} else {
			ds, err = doltstore.NewEmbedded(os.Getenv("VISIONSTUDIO_DATA"))
		}
		if err != nil {
			return nil, nil, err
		}
		cleanup := closeStore(ds)
		if err := migrate(ds); err != nil {
			cleanup()
			return nil, nil, err
		}
		return service.New(ds), cleanup, nil
	}

	mux := http.NewServeMux()
	webapi.RegisterRoutes(mux, connect, os.Getenv("OMNIDEVX_DATA"))

	r.Get("/execution", mux.ServeHTTP)
	r.Get("/spend", mux.ServeHTTP)
	r.Get("/maturity", mux.ServeHTTP)
	r.Get("/specs", mux.ServeHTTP)
}
