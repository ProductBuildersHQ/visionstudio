package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
)

// cloudCmd is the local client's connection to VisionStudio Cloud
// (INIT-VISIONSTUDIO-002). Per the open-core split, the cloud product
// itself lives in the private visionstudio-cloud repo; this surface is
// only the local record of tenant assignments and remotes, plus sync.
func cloudCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Connect the local app to VisionStudio Cloud (tenant assignment, sync)",
	}
	cmd.AddCommand(cloudTenantCmd(), cloudRemoteCmd())
	return cmd
}

func cloudTenantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Manage which registered repositories sync to which cloud tenant",
	}

	set := &cobra.Command{
		Use:   "set <repo-id> <tenant-slug>",
		Short: "Assign a repository's synced data to a tenant",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cliconfig.Load()
			if err != nil {
				return err
			}
			cfg.AssignTenant(args[0], args[1])
			if err := cfg.Save(); err != nil {
				return err
			}
			cmd.Printf("Assigned %s -> tenant %s\n", args[0], args[1])
			return nil
		},
	}

	unset := &cobra.Command{
		Use:   "unset <repo-id>",
		Short: "Remove a repository's tenant assignment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cliconfig.Load()
			if err != nil {
				return err
			}
			if !cfg.UnassignTenant(args[0]) {
				cmd.Printf("%s had no tenant assignment.\n", args[0])
				return nil
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			cmd.Printf("Unassigned %s\n", args[0])
			return nil
		},
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List repository-to-tenant assignments",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cliconfig.Load()
			if err != nil {
				return err
			}
			if len(cfg.Cloud.TenantAssignments) == 0 {
				cmd.Println("No tenant assignments. Use 'cloud tenant set <repo-id> <tenant-slug>'.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "REPO\tTENANT")
			for repo, slug := range cfg.Cloud.TenantAssignments {
				_, _ = fmt.Fprintf(w, "%s\t%s\n", repo, slug)
			}
			return w.Flush()
		},
	}

	cmd.AddCommand(set, unset, list)
	return cmd
}

func cloudRemoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remote",
		Short: "Manage tenant Dolt remote URLs",
	}

	set := &cobra.Command{
		Use:   "set <tenant-slug> <dolt-remote-url>",
		Short: "Configure the Dolt remote URL a tenant syncs to",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cliconfig.Load()
			if err != nil {
				return err
			}
			cfg.SetTenantRemote(args[0], args[1])
			if err := cfg.Save(); err != nil {
				return err
			}
			cmd.Printf("Tenant %s remote set to %s\n", args[0], args[1])
			return nil
		},
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List configured tenant remotes",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cliconfig.Load()
			if err != nil {
				return err
			}
			if len(cfg.Cloud.TenantRemotes) == 0 {
				cmd.Println("No tenant remotes configured. Use 'cloud remote set <tenant-slug> <url>'.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "TENANT\tREMOTE URL")
			for slug, url := range cfg.Cloud.TenantRemotes {
				_, _ = fmt.Fprintf(w, "%s\t%s\n", slug, url)
			}
			return w.Flush()
		},
	}

	cmd.AddCommand(set, list)
	return cmd
}
