package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/reposcan"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store/doltstore"
)

func registryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Manage the repository catalog",
	}
	cmd.AddCommand(registryAddCmd(), registryListCmd(), registryScanCmd(), registryDepsCmd(), registryUnpushedCmd(), registryOrgCmd(), registryPersonCmd(), registryVisibilityCmd())
	return cmd
}

func registryVisibilityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "visibility",
		Short: "Manage GitHub-derived repository visibility",
	}

	refresh := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh visibility (public|private|unknown) from GitHub via gh",
		Long: `Queries GitHub (gh CLI) for each registered repository and records
its visibility. Lookup failures leave the stored value untouched;
"unknown" is never treated as public by any export path.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repoID, _ := cmd.Flags().GetString("repo")

			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			res, err := svc.RefreshVisibility(cmd.Context(), service.GHVisibilityLookup, repoID)
			if err != nil {
				return err
			}
			cmd.Printf("Updated: %d, unchanged: %d, errors: %d\n", res.Updated, res.Unchanged, len(res.Errors))
			for _, e := range res.Errors {
				cmd.Printf("  ! %s\n", e)
			}
			return nil
		},
	}
	refresh.Flags().String("repo", "", "Refresh a single repository ID (default: all)")

	cmd.AddCommand(refresh)
	return cmd
}

func registryOrgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Manage first-class organizations",
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List organizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			orgs, err := svc.ListOrganizations(cmd.Context())
			if err != nil {
				return err
			}
			if len(orgs) == 0 {
				cmd.Println("No organizations. Run 'registry org backfill' to create them from registered repositories.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tLOGIN\tKIND\tWEBSITE")
			for _, o := range orgs {
				website := o.Website
				if website == "" {
					website = "-"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", o.ID, o.Login, o.Kind, website)
			}
			return w.Flush()
		},
	}

	backfill := &cobra.Command{
		Use:   "backfill",
		Short: "Create organization rows from registered repositories and link them",
		Long: `Creates an Organization row for every distinct organization string on
registered repositories and links each repository to its organization.
Idempotent. Use --user to mark logins that are GitHub user accounts
(e.g. --user grokify) rather than organizations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			users, _ := cmd.Flags().GetStringSlice("user")
			userLogins := make(map[string]bool, len(users))
			for _, u := range users {
				userLogins[u] = true
			}

			res, err := svc.BackfillOrganizations(cmd.Context(), userLogins)
			if err != nil {
				return err
			}
			cmd.Printf("Organizations created: %d\n", len(res.OrgsCreated))
			for _, id := range res.OrgsCreated {
				cmd.Printf("  %s\n", id)
			}
			cmd.Printf("Repositories linked: %d (skipped/already linked: %d)\n", res.ReposLinked, res.ReposSkipped)
			return nil
		},
	}
	backfill.Flags().StringSlice("user", nil, "Login that is a GitHub user account, not an organization (repeatable)")

	cmd.AddCommand(list, backfill)
	return cmd
}

func registryPersonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "person",
		Short: "Manage identities",
	}

	add := &cobra.Command{
		Use:   "add",
		Short: "Register or update an identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			githubLogin, _ := cmd.Flags().GetString("github-login")
			id, _ := cmd.Flags().GetString("id")
			displayName, _ := cmd.Flags().GetString("display-name")
			emails, _ := cmd.Flags().GetStringSlice("email")
			orgs, _ := cmd.Flags().GetStringSlice("org")

			if githubLogin == "" {
				return fmt.Errorf("--github-login is required")
			}

			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			p, err := svc.RegisterPerson(cmd.Context(), id, githubLogin, displayName, emails, orgs)
			if err != nil {
				return err
			}
			cmd.Printf("Registered: %s (orgs: %s)\n", p.ID, strings.Join(p.OrgIDs, ", "))
			return nil
		},
	}
	add.Flags().String("id", "", "Person ID (default person:<github-login>)")
	add.Flags().String("github-login", "", "GitHub login (required)")
	add.Flags().String("display-name", "", "Display name")
	add.Flags().StringSlice("email", nil, "Commit-author email identity (repeatable)")
	add.Flags().StringSlice("org", nil, "Affiliated organization login (repeatable)")

	list := &cobra.Command{
		Use:   "list",
		Short: "List identities",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			people, err := svc.ListPeople(cmd.Context())
			if err != nil {
				return err
			}
			if len(people) == 0 {
				cmd.Println("No identities registered.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tGITHUB\tNAME\tORGS")
			for _, p := range people {
				name := p.DisplayName
				if name == "" {
					name = "-"
				}
				orgs := "-"
				if len(p.OrgIDs) > 0 {
					orgs = strings.Join(p.OrgIDs, ", ")
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID, p.GitHubLogin, name, orgs)
			}
			return w.Flush()
		},
	}

	cmd.AddCommand(add, list)
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
			_, _ = fmt.Fprintln(w, "ID\tORG\tNAME\tSTATUS\tVIS\tPATH")
			for _, r := range repos {
				path := r.LocalPath
				if path == "" {
					path = "-"
				}
				vis := r.Visibility
				if vis == "" {
					vis = "unknown"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", r.ID, r.Organization, r.RepositoryName, r.Status, vis, path)
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

func connectService(cmd *cobra.Command) (*service.Service, func(), error) {
	dataDir := getDataDir(cmd)
	if dataDir != "" {
		ds, err := connectEmbedded(dataDir)
		if err != nil {
			return nil, nil, fmt.Errorf("connect (embedded): %w", err)
		}
		svc := service.New(ds)
		return svc, func() {
			if err := ds.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: close database: %v\n", err)
			}
		}, nil
	}

	dsn := getDSN(cmd)
	ds, err := doltstore.New(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("connect (server): %w", err)
	}
	// sql.Open is lazy, so New succeeds even when Dolt is down. Ping now to
	// surface an actionable "the server is not running" message instead of a
	// raw driver error deep inside the first query.
	baseCtx := cmd.Context()
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	pingCtx, cancel := context.WithTimeout(baseCtx, 5*time.Second)
	defer cancel()
	if err := ds.Ping(pingCtx); err != nil {
		if closeErr := ds.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close database: %v\n", closeErr)
		}
		return nil, nil, diagnoseDBError(dsn, err)
	}
	svc := service.New(ds)
	return svc, func() {
		if err := ds.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: close database: %v\n", err)
		}
	}, nil
}
