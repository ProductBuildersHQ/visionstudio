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
		Long: `Coordinate cross-repository initiatives, roadmap items, assignments, and delivery evidence.

Entity model (parent → child):

  Program     PROG-<SLUG>              optional grouping of initiatives
  Initiative  INIT-<SLUG>-NNN          a body of work; has a home repo (where specs live),
                                       a spec workflow, and a lifecycle:
                                       proposed → planned → executing → delivery_complete
                                       → releasing → released → closed
  Phase       <INIT-ID>/phase-N        themed grouping of ~5 RMIs within an initiative
  RMI         RMI-<REPOSLUG>-NNN       roadmap item: the unit of work, tied to ONE repository;
                                       an initiative's RMIs may span multiple repositories

Conventions:
  - Repositories must be registered ('registry add') before RMIs reference them.
    Repository IDs are 'github.com/<org>/<repo>'.
  - Git commits reference work via the trailer 'Refs: RMI-<REPOSLUG>-NNN'
    (RMI only, never the INIT ID; 'work claim' prints the exact line).
  - Initiative spec documents live in the home repo at docs/specs/initiatives/<INIT-ID>/
    ('spec init' scaffolds them; which files depends on the initiative's workflow).

Typical sequence to log a new body of work:

  visionstudio registry add --org myorg --name myrepo --path ~/src/myrepo   # once per repo
  visionstudio initiative create --id INIT-MYPROJECT-001 --title "..." \
      --home-repo github.com/myorg/myrepo --type feature --workflow pbhq-lite
  visionstudio spec init INIT-MYPROJECT-001        # scaffold PRD/TRD/PLAN/ROADMAP.md
  # ...fill in the specs, then either:
  visionstudio roadmap import <path>/ROADMAP.md    # bulk-create phases + RMIs from the doc
  # or create them individually:
  visionstudio phase add --id INIT-MYPROJECT-001/phase-1 --initiative INIT-MYPROJECT-001 --title "..."
  visionstudio rmi create --id RMI-MYREPO-001 --repo github.com/myorg/myrepo \
      --initiative INIT-MYPROJECT-001 --phase INIT-MYPROJECT-001/phase-1 --title "..." --type capability

Executing work: 'work ready' lists unblocked RMIs, 'work claim' leases one and prints the
git trailer, 'work complete' marks it done. 'initiative sweep' finds initiatives whose RMIs
are all resolved but whose status hasn't been advanced.

Discovery: 'initiative list', 'rmi list', 'registry list', 'workflow list' show current state;
most list commands accept '--format json'. Each subcommand's --help documents its contract.`,
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
		appCmd(),
		uiCmd(),
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
		cloudCmd(),
		syncCmd(),
		pullCmd(),
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
