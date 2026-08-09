// Package main is the entry point for the visionstudio CLI — the VisionStudio
// control plane, successor to prismctl with the full command surface.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	config "github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
)

// defaultPort is the single default port for the local Dolt SQL server. The
// client DSN and the `db serve`/`start`/`restart` commands all fall back to it,
// so a freshly-started server and the CLI agree without any configuration. A
// saved config.json DSN takes precedence — see defaultServerPort and getDSN.
// Keep defaultDSN's port in sync with this value.
const defaultPort = 13306
const defaultDSN = "root:@tcp(127.0.0.1:13306)/visionstudio"
const defaultDataDir = "~/.productbuildershq/visionstudio"

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "visionstudio",
		Short: "VisionStudio — Product Delivery Control Plane",
		Long:  "Coordinate cross-repository initiatives, roadmap items, assignments, and delivery evidence.",
		// main() prints the returned error once. Silence cobra's own error and
		// usage dumps so runtime failures (e.g. the DB being down) show a clean,
		// actionable message instead of a duplicated error plus a usage tree.
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().String("dsn", "", "MySQL-compatible DSN for Dolt server mode (default: $VISIONSTUDIO_DSN or "+defaultDSN+")")
	cmd.PersistentFlags().String("data-dir", "", "Data directory for embedded Dolt (default: $VISIONSTUDIO_DATA or "+defaultDataDir+")")

	cmd.AddCommand(
		versionCmd(),
		configCmd(),
		dbCmd(),
		registryCmd(),
		programCmd(),
		initiativeCmd(),
		phaseCmd(),
		rmiCmd(),
		workCmd(),
		contextCmd(),
		exportCmd(),
		ingestCmd(),
		reportCmd(),
		validateCmd(),
		mcpCmd(),
		dashboardCmd(),
		roadmapCmd(),
		releaseCmd(),
		workflowCmd(),
		specCmd(),
		maturityCmd(),
	)

	return cmd
}

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
	if cfg, err := config.Load(); err == nil && cfg.DSN != "" {
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

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("visionstudio v0.1.0-dev")
		},
	}
}
