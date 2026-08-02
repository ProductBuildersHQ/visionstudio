package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func registryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Manage the repository catalog",
	}
	cmd.AddCommand(registryAddCmd(), registryListCmd())
	return cmd
}

func registryAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Register a repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			org, _ := cmd.Flags().GetString("org")
			name, _ := cmd.Flags().GetString("name")
			branch, _ := cmd.Flags().GetString("branch")
			localPath, _ := cmd.Flags().GetString("path")
			domain, _ := cmd.Flags().GetString("domain")

			if org == "" || name == "" {
				return fmt.Errorf("--org and --name are required")
			}

			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			repo, err := svc.RegisterRepository(cmd.Context(), org, name, branch, localPath, domain)
			if err != nil {
				return err
			}
			cmd.Printf("Registered: %s\n", repo.ID)
			return nil
		},
	}
	cmd.Flags().String("org", "", "GitHub organization (required)")
	cmd.Flags().String("name", "", "Repository name (required)")
	cmd.Flags().String("branch", "main", "Default branch")
	cmd.Flags().String("path", "", "Local filesystem path")
	cmd.Flags().String("domain", "", "Domain classification")
	return cmd
}

func registryListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			org, _ := cmd.Flags().GetString("org")

			var repos []*store.Repository
			if org != "" {
				repos, err = svc.ListRepositoriesByOrg(cmd.Context(), org)
			} else {
				repos, err = svc.ListRepositories(cmd.Context())
			}
			if err != nil {
				return err
			}

			if len(repos) == 0 {
				cmd.Println("No repositories registered.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tORG\tNAME\tSTATUS\tPATH")
			for _, r := range repos {
				path := r.LocalPath
				if path == "" {
					path = "-"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.ID, r.Organization, r.RepositoryName, r.Status, path)
			}
			return w.Flush()
		},
	}
	cmd.Flags().String("org", "", "Filter by organization")
	return cmd
}
