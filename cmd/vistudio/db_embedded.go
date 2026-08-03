//go:build dolt_embedded

package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store/doltstore"
)

// connectEmbedded returns a DoltStore using embedded Dolt, or nil if not available.
func connectEmbedded(dataDir string) (*doltstore.DoltStore, error) {
	return doltstore.NewEmbedded(dataDir)
}

// hasEmbeddedSupport returns true when built with dolt_embedded tag.
func hasEmbeddedSupport() bool {
	return true
}

func dbInitEmbedded(cmd *cobra.Command, dataDir string) error {
	absDir, err := filepath.Abs(dataDir)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	cmd.Printf("Initializing embedded Dolt database at %s...\n", absDir)
	ds, err := doltstore.NewEmbedded(dataDir)
	if err != nil {
		return fmt.Errorf("init embedded: %w", err)
	}
	defer func() { printCloseWarning(ds.Close()) }()

	cmd.Println("Running schema migration...")
	if err := ds.Migrate(cmd.Context()); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	cmd.Println("Embedded Dolt database initialized and migrated.")
	return nil
}
