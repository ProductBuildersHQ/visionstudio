//go:build dolt

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store/doltstore"
)

func getDataDir(cmd *cobra.Command) string {
	dir, _ := cmd.Flags().GetString("data-dir")
	if dir != "" {
		return expandHome(dir)
	}
	if env := os.Getenv("VISIONSTUDIO_DATA"); env != "" {
		return expandHome(env)
	}
	if dsn, _ := cmd.Flags().GetString("dsn"); dsn != "" {
		return ""
	}
	if os.Getenv("VISIONSTUDIO_DSN") != "" {
		return ""
	}
	return expandHome(defaultDataDir)
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func getDSN(cmd *cobra.Command) string {
	dsn, _ := cmd.Flags().GetString("dsn")
	if dsn != "" {
		return dsn
	}
	if env := os.Getenv("VISIONSTUDIO_DSN"); env != "" {
		return env
	}
	return defaultDSN
}

func connectService(cmd *cobra.Command) (*service.Service, func(), error) {
	dataDir := getDataDir(cmd)
	if dataDir != "" {
		ds, err := doltstore.NewEmbedded(dataDir)
		if err != nil {
			return nil, nil, fmt.Errorf("connect (embedded): %w", err)
		}
		return service.New(ds), func() { printCloseWarning(ds.Close()) }, nil
	}

	ds, err := doltstore.New(getDSN(cmd))
	if err != nil {
		return nil, nil, fmt.Errorf("connect (server): %w", err)
	}
	return service.New(ds), func() { printCloseWarning(ds.Close()) }, nil
}

func printCloseWarning(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: close database: %v\n", err)
	}
}

func dbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management",
	}
	cmd.AddCommand(dbInitCmd())
	return cmd
}

func dbInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Bootstrap the VisionStudio Dolt database",
		Long: `Initialize the embedded Dolt database and run schema migration.

Uses --data-dir / $VISIONSTUDIO_DATA (default ` + defaultDataDir + `).
With --dsn / $VISIONSTUDIO_DSN set, migrates against the Dolt SQL server instead.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir := getDataDir(cmd)

			var ds *doltstore.DoltStore
			var err error
			if dataDir != "" {
				cmd.Printf("Initializing embedded Dolt database at %s...\n", dataDir)
				ds, err = doltstore.NewEmbedded(dataDir)
			} else {
				dsn := getDSN(cmd)
				cmd.Printf("Connecting to Dolt server (DSN: %s)...\n", dsn)
				ds, err = doltstore.New(dsn)
			}
			if err != nil {
				return fmt.Errorf("connect: %w", err)
			}
			defer func() { printCloseWarning(ds.Close()) }()

			cmd.Println("Running schema migration...")
			if err := ds.Migrate(cmd.Context()); err != nil {
				return fmt.Errorf("migrate: %w", err)
			}
			cmd.Println("Database initialized and migrated.")
			return nil
		},
	}
}
