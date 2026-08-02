// Package main is the entry point for the visionstudio CLI — DB-backed
// orchestration of programs, initiatives, RMIs, and work assignments.
// It succeeds the prismctl DB commands, which now point here.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
		Use:   "visionstudio",
		Short: "VisionStudio — Product Delivery Control Plane",
		Long:  "Coordinate cross-repository initiatives, roadmap items, assignments, and delivery evidence.",
	}

	cmd.PersistentFlags().String("dsn", "", "MySQL-compatible DSN for Dolt server mode (default: $VISIONSTUDIO_DSN or "+defaultDSN+")")
	cmd.PersistentFlags().String("data-dir", "", "Data directory for embedded Dolt (default: $VISIONSTUDIO_DATA or "+defaultDataDir+")")

	cmd.AddCommand(
		versionCmd(),
		dbCmd(),
		registryCmd(),
		programCmd(),
		initiativeCmd(),
		phaseCmd(),
		rmiCmd(),
		workCmd(),
	)
	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("visionstudio v0.1.0-dev")
		},
	}
}
