package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
	cmd.AddCommand(registryAddCmd(), registryListCmd(), registryUpdateCmd(), registryArchiveCmd(), registryRemoveCmd(), registryScanCmd(), registryDepsCmd(), registryUnpushedCmd(), registryOrgCmd(), registryPersonCmd(), registryVisibilityCmd(), registryFocusCmd())
	return cmd
}

func registryRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <repo-id>",
		Short: "Hard-delete a repository record",
		Long: `Permanently deletes a repository's registry row. For true mistakes
only (e.g. a typo'd 'registry add') — for a merge or rename where the
record should be kept for history, use 'registry archive
--superseded-by' instead.

Always refuses while RMIs, releases, or spec documents still reference
the repository (their foreign keys require it to exist) — reassign
them first with 'rmi bulk-update', or archive instead. Otherwise
requires --force to confirm.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			id, err := resolveRepoID(cmd.Context(), svc, args[0])
			if err != nil {
				return err
			}
			if _, err := svc.GetRepository(cmd.Context(), id); err != nil {
				return err
			}

			var blockers []string
			if rmis, err := svc.ListRMIsByRepo(cmd.Context(), id); err != nil {
				return fmt.Errorf("check referencing RMIs: %w", err)
			} else if len(rmis) > 0 {
				blockers = append(blockers, fmt.Sprintf("%d RMI(s)", len(rmis)))
			}
			if rels, err := svc.Store.ListReleasesByRepo(cmd.Context(), id); err != nil {
				return fmt.Errorf("check referencing releases: %w", err)
			} else if len(rels) > 0 {
				blockers = append(blockers, fmt.Sprintf("%d release(s)", len(rels)))
			}
			if docs, err := svc.ListSpecDocumentsByRepo(cmd.Context(), id); err != nil {
				return fmt.Errorf("check referencing spec documents: %w", err)
			} else if len(docs) > 0 {
				blockers = append(blockers, fmt.Sprintf("%d spec document(s)", len(docs)))
			}
			if len(blockers) > 0 {
				return fmt.Errorf("refusing to remove %s: still referenced by %s (reassign with 'rmi bulk-update' or use 'registry archive' instead)",
					id, strings.Join(blockers, ", "))
			}

			force, _ := cmd.Flags().GetBool("force")
			if !force {
				return fmt.Errorf("refusing to remove %s without --force", id)
			}

			if err := svc.Store.DeleteRepository(cmd.Context(), id); err != nil {
				return err
			}
			cmd.Printf("Removed %s\n", id)
			return nil
		},
	}
	cmd.Flags().Bool("force", false, "Confirm the hard delete")
	return cmd
}

func registryArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <repo-id>",
		Short: "Archive a repository (status change, record preserved)",
		Long: `Marks a repository archived without deleting its record or the
RMIs/releases/spec documents that reference it. Preferred over
'registry remove' for a merge or rename — pair with --superseded-by to
point at the repository that replaced it.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			id, err := resolveRepoID(cmd.Context(), svc, args[0])
			if err != nil {
				return err
			}
			repo, err := svc.GetRepository(cmd.Context(), id)
			if err != nil {
				return err
			}

			supersededBy, _ := cmd.Flags().GetString("superseded-by")
			if supersededBy != "" {
				supersededBy, err = resolveRepoID(cmd.Context(), svc, supersededBy)
				if err != nil {
					return fmt.Errorf("--superseded-by: %w", err)
				}
				if supersededBy == repo.ID {
					return fmt.Errorf("--superseded-by cannot be the repository being archived")
				}
			}

			if repo.Status == "archived" && repo.SupersededBy == supersededBy {
				cmd.Printf("Repository %s is already archived\n", repo.ID)
				return nil
			}

			repo.Status = "archived"
			repo.SupersededBy = supersededBy
			if err := svc.Store.UpdateRepository(cmd.Context(), repo); err != nil {
				return err
			}

			cmd.Printf("Archived %s\n", repo.ID)
			if supersededBy != "" {
				cmd.Printf("Superseded by: %s\n", supersededBy)
			}
			if reason, _ := cmd.Flags().GetString("reason"); reason != "" {
				cmd.Printf("Reason: %s\n", reason)
			}
			return nil
		},
	}
	cmd.Flags().String("reason", "", "Why this repository is being archived (echoed to output)")
	cmd.Flags().String("superseded-by", "", "Repository ID that replaced this one")
	return cmd
}

func registryUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <repo-id>",
		Short: "Repoint or edit a registered repository",
		Long: `Edits an existing repository's metadata (path, org, branch, name).

The repository's ID stays "github.com/<org>/<name>" as originally
registered — --org/--name correct metadata on this same record, they do
not rename the ID. For an actual repository merge/rename, use
'registry archive --superseded-by <new-id>' followed by 'registry add'
for the new ID (see 'registry archive --help').`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			id, err := resolveRepoID(cmd.Context(), svc, args[0])
			if err != nil {
				return err
			}
			repo, err := svc.GetRepository(cmd.Context(), id)
			if err != nil {
				return err
			}

			var changed []string
			if cmd.Flags().Changed("path") {
				v, _ := cmd.Flags().GetString("path")
				if repo.LocalPath != v {
					repo.LocalPath = v
					changed = append(changed, "path")
				}
			}
			if cmd.Flags().Changed("org") {
				v, _ := cmd.Flags().GetString("org")
				if repo.Organization != v {
					repo.Organization = v
					changed = append(changed, "org")
				}
			}
			if cmd.Flags().Changed("branch") {
				v, _ := cmd.Flags().GetString("branch")
				if repo.DefaultBranch != v {
					repo.DefaultBranch = v
					changed = append(changed, "branch")
				}
			}
			if cmd.Flags().Changed("name") {
				v, _ := cmd.Flags().GetString("name")
				if repo.RepositoryName != v {
					repo.RepositoryName = v
					changed = append(changed, "name")
				}
			}

			if len(changed) == 0 {
				cmd.Printf("Repository %s is already up to date\n", repo.ID)
				return nil
			}

			if err := svc.Store.UpdateRepository(cmd.Context(), repo); err != nil {
				return err
			}
			cmd.Printf("Updated %s: %s\n", repo.ID, strings.Join(changed, ", "))
			if slices.Contains(changed, "org") || slices.Contains(changed, "name") {
				cmd.Printf("Note: the repository ID is still %s (org/name metadata only; see --help for renames)\n", repo.ID)
			}
			return nil
		},
	}
	cmd.Flags().String("path", "", "New local filesystem path")
	cmd.Flags().String("org", "", "New GitHub organization (metadata only, does not change the ID)")
	cmd.Flags().String("branch", "", "New default branch")
	cmd.Flags().String("name", "", "New repository name (metadata only, does not change the ID)")
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

	stats := &cobra.Command{
		Use:   "stats",
		Short: "Per-organization rollup: repos, visibility, initiatives, open RMIs",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			rollup, err := svc.OrgRollup(cmd.Context())
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ORG\tKIND\tREPOS\tPUB\tPRIV\tUNK\tINITS\tACTIVE\tOPEN-RMIS\tPEOPLE")
			for _, st := range rollup {
				people := "-"
				if len(st.PeopleLogins) > 0 {
					people = strings.Join(st.PeopleLogins, ",")
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%s\n",
					st.Org.Login, st.Org.Kind, st.Repos, st.Public, st.Private, st.Unknown,
					st.Initiatives, st.ActiveInits, st.OpenRMIs, people)
			}
			return w.Flush()
		},
	}

	cmd.AddCommand(list, backfill, stats)
	return cmd
}

func registryFocusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "focus",
		Short: "Private repositories and their active initiatives",
		Long: `The confirmed-private focus list: repositories with visibility=private
(unknown is excluded — this view never guesses) and the active
initiatives homed in them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			entries, err := svc.FocusList(cmd.Context())
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				cmd.Println("No confirmed-private repositories. Run 'registry visibility refresh' first.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "REPO\tACTIVE INITIATIVES")
			for _, e := range entries {
				inits := "-"
				if len(e.Initiatives) > 0 {
					var titles []string
					for _, in := range e.Initiatives {
						titles = append(titles, fmt.Sprintf("%s (%s)", in.ID, in.Status))
					}
					inits = strings.Join(titles, "; ")
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\n", e.Repo.ID, inits)
			}
			return w.Flush()
		},
	}
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

	repolist := &cobra.Command{
		Use:   "repos <person-id-or-login>",
		Short: "Repositories across all of a person's organizations (practitioner lens)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			p, repos, err := svc.PersonRepositories(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			cmd.Printf("%s — %d repositories across %d organizations\n", p.ID, len(repos), len(p.OrgIDs))
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "REPO\tVIS\tSTATUS")
			for _, r := range repos {
				vis := r.Visibility
				if vis == "" {
					vis = "unknown"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", r.ID, vis, r.Status)
			}
			return w.Flush()
		},
	}

	cmd.AddCommand(add, list, repolist)
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

			warnAddConflicts(cmd, svc, localPath)

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

// resolveRepoID resolves a possibly-short repository reference — a bare
// name ("visionstudio"), "org/name", or the full "github.com/org/name"
// ID — to a registered repository's full ID. Repository IDs are always
// "github.com/org/name" (see Service.RegisterRepository), so a bare short
// name silently matching nothing used to return an empty result set with
// no indication why; this gives an exact match or a did-you-mean error
// instead.
func resolveRepoID(ctx context.Context, svc *service.Service, ref string) (string, error) {
	if ref == "" {
		return "", nil
	}
	if strings.HasPrefix(ref, "github.com/") {
		if _, err := svc.GetRepository(ctx, ref); err != nil {
			return "", fmt.Errorf("repository %q not found in registry", ref)
		}
		return ref, nil
	}

	repos, err := svc.ListRepositories(ctx)
	if err != nil {
		return "", err
	}

	var exact []*store.Repository
	for _, r := range repos {
		shortID := r.Organization + "/" + r.RepositoryName
		if r.RepositoryName == ref || shortID == ref {
			exact = append(exact, r)
		}
	}
	switch len(exact) {
	case 1:
		return exact[0].ID, nil
	case 0:
		var suggestions []string
		for _, r := range repos {
			if strings.Contains(r.RepositoryName, ref) || strings.Contains(ref, r.RepositoryName) {
				suggestions = append(suggestions, r.ID)
			}
		}
		if len(suggestions) > 0 {
			return "", fmt.Errorf("no repository matches %q — did you mean: %s?", ref, strings.Join(suggestions, ", "))
		}
		return "", fmt.Errorf("no repository matches %q (register it with 'registry add', or use org/name or the full github.com/org/name ID)", ref)
	default:
		var ids []string
		for _, r := range exact {
			ids = append(ids, r.ID)
		}
		return "", fmt.Errorf("%q is ambiguous — matches: %s (use the full ID)", ref, strings.Join(ids, ", "))
	}
}

// warnAddConflicts prints non-blocking warnings for 'registry add' when
// --path looks wrong (missing, not a directory, not a git working tree) or
// collides with an already-registered repository (same local path, or the
// same git remote origin URL). Registration proceeds regardless — these are
// warnings, not validation errors.
func warnAddConflicts(cmd *cobra.Command, svc *service.Service, localPath string) {
	if localPath == "" {
		return
	}

	info, statErr := os.Stat(localPath)
	switch {
	case statErr != nil:
		fmt.Fprintf(os.Stderr, "warning: --path %s does not exist\n", localPath)
	case !info.IsDir():
		fmt.Fprintf(os.Stderr, "warning: --path %s is not a directory\n", localPath)
	default:
		if _, err := os.Stat(filepath.Join(localPath, ".git")); err != nil {
			fmt.Fprintf(os.Stderr, "warning: --path %s is not a git working tree\n", localPath)
		}
	}

	repos, err := svc.ListRepositories(cmd.Context())
	if err != nil {
		return
	}
	newRemote := gitRemoteURL(localPath)
	for _, r := range repos {
		if r.LocalPath != "" && r.LocalPath == localPath {
			fmt.Fprintf(os.Stderr, "warning: %s is already registered at this path\n", r.ID)
		}
		if newRemote != "" && r.LocalPath != "" && r.LocalPath != localPath {
			if existingRemote := gitRemoteURL(r.LocalPath); existingRemote != "" && existingRemote == newRemote {
				fmt.Fprintf(os.Stderr, "warning: %s already has this git remote (%s)\n", r.ID, newRemote)
			}
		}
	}
}

// gitRemoteURL returns the origin remote URL for the git working tree at
// path, or "" if path isn't a git working tree, has no origin, or the
// lookup otherwise fails. Best-effort — never treated as fatal.
func gitRemoteURL(path string) string {
	// #nosec G204 -- path is a caller-supplied local directory for a local
	// dev CLI, not untrusted network input (same posture as godolt's exec.go).
	out, err := exec.Command("git", "-C", path, "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
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
