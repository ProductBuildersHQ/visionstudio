package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/reposcan"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func registryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Manage the repository catalog",
	}
	cmd.AddCommand(registryAddCmd(), registryListCmd(), registryScanCmd(), registryDepsCmd(), registryUnpushedCmd())
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

func registryScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan <directory> [directory...]",
		Short: "Scan org directories and import repos into the registry",
		Long: `Scan one or more organization directories (e.g. ~/go/src/github.com/plexusone)
and auto-register all git repos found. Also populates dependency edges from go.mod.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			checkUnpushed, _ := cmd.Flags().GetBool("check-unpushed")

			for _, dir := range args {
				cmd.Printf("Scanning %s...\n", dir)
				opts := reposcan.ScanOptions{
					CheckUnpushed: checkUnpushed,
					Progress: func(current, total int, name string) {
						_, _ = fmt.Fprintf(os.Stderr, "\r  [%d/%d] %s", current, total, name)
					},
				}
				res, err := reposcan.ScanAndImport(cmd.Context(), svc, dir, opts)
				if err != nil {
					return err
				}
				cmd.Printf("\n  Scanned: %d, Git repos: %d, Imported: %d, Deps created: %d\n",
					res.TotalScanned, res.GitRepos, res.Imported, res.DepsCreated)
			}
			return nil
		},
	}
	cmd.Flags().Bool("check-unpushed", false, "Also check for unpushed commits (slower)")
	return cmd
}

func registryDepsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deps <directory>",
		Short: "Show repository dependency order (topological sort)",
		Long:  "Scan an org directory and display repos in dependency order (dependencies first).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ordered, err := reposcan.DependencyOrder(args[0])
			if err != nil {
				return err
			}

			if len(ordered) == 0 {
				cmd.Println("No repos with Go modules found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "#\tNAME\tDEPENDS ON")
			for _, r := range ordered {
				deps := "-"
				if len(r.Deps) > 0 {
					deps = strings.Join(r.Deps, ", ")
				}
				_, _ = fmt.Fprintf(w, "%d\t%s\t%s\n", r.Position, r.Name, deps)
			}
			return w.Flush()
		},
	}
}

func registryUnpushedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unpushed <directory> [directory...]",
		Short: "List repos with uncommitted or unpushed work",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var allUnpushed []reposcan.UnpushedRepo
			for _, dir := range args {
				unpushed, err := reposcan.FindUnpushed(dir)
				if err != nil {
					return fmt.Errorf("scan %s: %w", dir, err)
				}
				allUnpushed = append(allUnpushed, unpushed...)
			}

			if len(allUnpushed) == 0 {
				cmd.Println("All repos are clean.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "REPO\tUNCOMMITTED\tUNPUSHED")
			for _, r := range allUnpushed {
				uncommitted := "-"
				if r.HasUncommittedChanges {
					uncommitted = "yes"
				}
				unpushed := "-"
				if r.HasUnpushedCommits {
					unpushed = "yes"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", r.Name, uncommitted, unpushed)
			}
			return w.Flush()
		},
	}
}
