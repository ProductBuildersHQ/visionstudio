package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
	"github.com/ProductBuildersHQ/visionstudio/pkg/cloudsync"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store/doltstore"
)

func pullCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Fetch and merge a VisionStudio Cloud tenant's branch into the local database",
		Long: `Pull is the other half of the fast-forward-only sync policy: sync
push is rejected whenever the tenant remote has commits the local
branch doesn't have, so pull first, resolve any conflicts locally,
then push again. The cloud never merges divergent user work
server-side — merging always happens on your machine, where Dolt's
conflict tables are available if the merge needs help.

Requires server mode, same as sync. Configure the tenant remote first
with 'cloud remote set <tenant-slug> <dolt-remote-url>'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tenant, _ := cmd.Flags().GetString("tenant")
			branch, _ := cmd.Flags().GetString("branch")
			if tenant == "" {
				return fmt.Errorf("--tenant is required")
			}

			if getDataDir(cmd) != "" {
				return fmt.Errorf("pull requires server mode (dolt sql-server) — the embedded engine does not support the concurrent-session workload sync assumes")
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

			res, err := cloudsync.Pull(cmd.Context(), ds.DB(), branch, tenant, remoteURL)
			if err != nil {
				return err
			}
			cmd.Printf("Pulled from tenant %s (%s): %s\n", res.TenantSlug, res.RemoteURL, res.Message)
			if res.Conflicts > 0 {
				cmd.Printf("⚠ %d table(s) have merge conflicts — resolve dolt_conflicts_<table> rows before pushing again\n", res.Conflicts)
			}
			return nil
		},
	}
	cmd.Flags().String("tenant", "", "Tenant slug to pull from (required)")
	cmd.Flags().String("branch", "", "Branch to pull (default: active branch)")
	return cmd
}
