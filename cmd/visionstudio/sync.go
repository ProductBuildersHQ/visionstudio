package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
	"github.com/ProductBuildersHQ/visionstudio/pkg/cloudsync"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store/doltstore"
)

func syncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Push the local Dolt database to a VisionStudio Cloud tenant",
		Long: `Sync is currently WHOLE-DATABASE push and is dogfood-only: the local
database holds every registered repository in one set of tables, so a
full push carries fields the sync contract excludes (e.g. local_path)
unfiltered. Safe for our own tenants; a synced-entity projection is
required before this is offered to external tenants.

Requires server mode (sync uses Dolt's SQL-wire push, which needs a
running dolt sql-server — see the Dolt operational lesson in the
VisionStudio Cloud TRD). Configure the tenant remote first with
'cloud remote set <tenant-slug> <dolt-remote-url>'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tenant, _ := cmd.Flags().GetString("tenant")
			branch, _ := cmd.Flags().GetString("branch")
			if tenant == "" {
				return fmt.Errorf("--tenant is required")
			}

			if getDataDir(cmd) != "" {
				return fmt.Errorf("sync requires server mode (dolt sql-server) — the embedded engine does not support the concurrent-session workload sync assumes")
			}

			cfg, err := cliconfig.Load()
			if err != nil {
				return err
			}
			remoteURL, ok := cfg.TenantRemote(tenant)
			if !ok {
				return fmt.Errorf("no remote configured for tenant %q — run 'cloud remote set %s <dolt-remote-url>' first", tenant, tenant)
			}

			ds, err := doltstore.New(getDSN(cmd))
			if err != nil {
				return fmt.Errorf("connect: %w", err)
			}
			defer func() { _ = ds.Close() }()

			res, err := cloudsync.Push(cmd.Context(), ds.DB(), branch, tenant, remoteURL)
			if err != nil {
				return err
			}
			cmd.Printf("⚠ whole-database push (dogfood-only, see 'sync --help')\n")
			cmd.Printf("Pushed to tenant %s (%s): %s\n", res.TenantSlug, res.RemoteURL, res.Message)
			return nil
		},
	}
	cmd.Flags().String("tenant", "", "Tenant slug to sync to (required)")
	cmd.Flags().String("branch", "", "Branch to push (default: active branch)")
	return cmd
}
