// Package main is the entry point for the vistudio CLI — the VisionStudio
// control plane, successor to prismctl with the full command surface.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	config "github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
)

const defaultDSN = "root:@tcp(127.0.0.1:3306)/visionstudio"
const defaultDataDir = "~/.productbuildershq/visionstudio"

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vistudio",
		Short: "VisionStudio — Product Delivery Control Plane",
		Long:  "Coordinate cross-repository initiatives, roadmap items, assignments, and delivery evidence.",
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
			fmt.Println("vistudio v0.1.0-dev")
		},
	}
}
