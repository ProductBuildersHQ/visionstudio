package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/ingest"
)

func ingestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Scan external sources for delivery evidence",
	}
	cmd.AddCommand(ingestGitCmd(), ingestChangelogCmd(), ingestIRCmd(), ingestReleasesCmd(), ingestGitHubReleasesCmd())
	return cmd
}

func ingestGitHubReleasesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "github-releases [repo-id]",
		Short: "Fetch GitHub Releases via gh and reconcile against the store",
		Long: `Closes the gap between local git tags and real GitHub Release history:
tag-based ingest can miss releases that exist on GitHub but aren't
reachable as local tags. Gap-filled releases are created with no
initiative/RMI associations (no local commit range to walk for
trailers). Releases already known get URL/body backfilled only where
currently empty — released_at, tag, and every association are never
touched. Requires the gh CLI, authenticated.

Use --all for every repository with a local path.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")

			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			var results []*ingest.GitHubReleaseIngestResult
			if all {
				results, err = ingest.GitHubReleasesAll(cmd.Context(), svc, ingest.GHReleasesLookup)
				if err != nil {
					return err
				}
			} else {
				if len(args) == 0 {
					return fmt.Errorf("specify a repository ID or use --all")
				}
				res, err := ingest.GitHubReleases(cmd.Context(), svc, args[0], ingest.GHReleasesLookup)
				if err != nil {
					return err
				}
				results = []*ingest.GitHubReleaseIngestResult{res}
			}

			var totFetched, totGapFilled, totEnriched, totDrafts, errCount int
			for _, r := range results {
				if r.Err != nil {
					cmd.Printf("! %s: %v\n", r.RepoID, r.Err)
					errCount++
					continue
				}
				if r.Fetched == 0 {
					continue
				}
				cmd.Printf("%s: %d fetched, %d gap-filled, %d enriched, %d unchanged, %d drafts skipped\n",
					r.RepoID, r.Fetched, r.GapFilled, r.Enriched, r.Unchanged, r.DraftsSkipped)
				totFetched += r.Fetched
				totGapFilled += r.GapFilled
				totEnriched += r.Enriched
				totDrafts += r.DraftsSkipped
			}
			cmd.Printf("\nTotal: %d fetched, %d gap-filled, %d enriched, %d drafts skipped, %d repos with errors\n",
				totFetched, totGapFilled, totEnriched, totDrafts, errCount)
			return nil
		},
	}
	cmd.Flags().Bool("all", false, "Run for all repositories with a local path")
	return cmd
}

func ingestReleasesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "releases [repo-id]",
		Short: "Ingest per-repo releases from tags + CHANGELOG.json with trailer-chain association",
		Long: `Walk each repository's semver-like tags (oldest first). CHANGELOG.json is
the primary metadata source (date, notes ref); bare git tags are the
fallback. Commits in each tag range are scanned for Refs: trailers —
RMI IDs resolve to initiatives, auto-associating every release with the
work it shipped. No time-proximity guessing: untrailered commits
contribute nothing, and the coverage report says so.

Use --all for the historical backfill across every repository with a
local path.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			allTags, _ := cmd.Flags().GetBool("all-tags")
			workers, _ := cmd.Flags().GetInt("workers")

			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			var results []*ingest.ReleaseIngestResult
			if all {
				results, err = ingest.ReleasesAll(cmd.Context(), svc, allTags, workers)
				if err != nil {
					return err
				}
			} else {
				if len(args) == 0 {
					return fmt.Errorf("specify a repository ID or use --all")
				}
				res, err := ingest.Releases(cmd.Context(), svc, args[0], allTags)
				if err != nil {
					return err
				}
				results = []*ingest.ReleaseIngestResult{res}
			}

			for _, r := range results {
				if r.Err != nil {
					cmd.Printf("! %s: %v\n", r.RepoID, r.Err)
					continue
				}
				if r.ReleasesUpserted == 0 && r.CommitsWalked == 0 {
					continue
				}
				cmd.Printf("%s: %d releases (%d w/changelog), coverage %d/%d commits trailered (%.0f%%), links: %d initiatives, %d RMIs\n",
					r.RepoID, r.ReleasesUpserted, r.WithChangelog,
					r.CommitsWithTrailers, r.CommitsWalked, r.Coverage()*100,
					r.InitiativeLinks, r.RMILinks)
			}
			sum := ingest.SummarizeReleaseResults(results)
			cmd.Printf("\nTotal: %d releases; %d initiative links, %d RMI links; trailer coverage %d/%d commits (%.0f%%); unresolved RMI refs: %d; repos with errors: %d\n",
				sum.Releases, sum.InitiativeLinks, sum.RMILinks,
				sum.CommitsWithTrailers, sum.CommitsWalked, sum.Coverage()*100,
				sum.UnresolvedRMIRefs, sum.ReposWithErrors)
			return nil
		},
	}
	cmd.Flags().Bool("all", false, "Backfill all repositories with a local path")
	cmd.Flags().Bool("all-tags", false, "Include non-semver-like tags")
	cmd.Flags().Int("workers", runtime.NumCPU(), "Parallel workers with --all")
	return cmd
}

func ingestGitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git <repo-id>",
		Short: "Scan git commits for Refs: trailers and create evidence rows",
		Long: `Walk commits since the repository's high-water mark, parse Refs: trailers
and conventional commit type/scope, create delivery_evidence rows, and
advance the high-water mark.

Use --all to ingest all registered repositories with a local_path.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")

			if all {
				return ingestAllRepos(cmd)
			}

			if len(args) == 0 {
				return fmt.Errorf("specify a repository ID or use --all")
			}

			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			result, err := ingest.Git(cmd.Context(), svc, args[0])
			if err != nil {
				return err
			}
			printIngestResult(cmd, result)
			return nil
		},
	}
	cmd.Flags().Bool("all", false, "Ingest all repositories with a local_path")
	cmd.Flags().Int("workers", 0, "Parallel workers for --all (0 = GOMAXPROCS)")
	return cmd
}

func ingestAllRepos(cmd *cobra.Command) error {
	svc, cleanup, err := connectService(cmd)
	if err != nil {
		return err
	}
	defer cleanup()

	workers, _ := cmd.Flags().GetInt("workers")
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}

	cmd.Printf("Ingesting all repos (%d workers)...\n", workers)
	results, err := ingest.GitAll(cmd.Context(), svc, workers)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		cmd.Println("No repositories with local_path found.")
		return nil
	}

	var errCount int
	for _, r := range results {
		if r.Err != nil {
			cmd.PrintErrf("  %s: error: %v\n", r.RepoID, r.Err)
			errCount++
			continue
		}
		printIngestResult(cmd, r)
		cmd.Println()
	}
	if errCount > 0 {
		cmd.Printf("%d of %d repos had errors.\n", errCount, len(results))
	}
	return nil
}

func ingestChangelogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changelog <repo-id>",
		Short: "Ingest CHANGELOG.json entries with RMI references",
		Long: `Read a repository's CHANGELOG.json (structured-changelog format) and
create delivery_evidence rows for entries that have an rmi_ref field.

Optionally specify --file to ingest from a specific path.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			file, _ := cmd.Flags().GetString("file")

			var result *ingest.ChangelogResult
			if file != "" {
				result, err = ingest.ChangelogFromFile(cmd.Context(), svc, args[0], file)
			} else {
				result, err = ingest.Changelog(cmd.Context(), svc, args[0])
			}
			if err != nil {
				return err
			}

			cmd.Printf("Repository:  %s\n", result.RepoID)
			cmd.Printf("Releases:    %d read\n", result.ReleasesRead)
			cmd.Printf("Entries:     %d read\n", result.EntriesRead)
			cmd.Printf("Evidence:    %d added\n", result.EvidenceAdded)
			return nil
		},
	}
	cmd.Flags().String("file", "", "Path to CHANGELOG.json (default: <repo-local-path>/CHANGELOG.json)")
	return cmd
}

func printIngestResult(cmd *cobra.Command, r *ingest.GitResult) {
	cmd.Printf("Repository:   %s\n", r.RepoID)
	cmd.Printf("Commits:      %d walked\n", r.CommitsWalked)
	cmd.Printf("Evidence:     %d added\n", r.EvidenceAdded)
	cmd.Printf("Unattributed: %d\n", r.Unattributed)
	if r.NewHighWater != "" {
		cmd.Printf("High-water:   %s\n", r.NewHighWater)
	}
}

func ingestIRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ir <repo-id> [file]",
		Short: "Ingest *.ir.json files into Dolt",
		Long: `Import IR (intermediate representation) JSON files into the database.

If a file path is provided, ingests that single file.
Otherwise, scans the repository for all *.ir.json files.

IR files contain multi-domain snapshots including:
- DevX metrics (developer experience period reports)
- PRISM roadmaps and goals
- PRISM maturity documents
- Execution data (initiatives, phases, RMIs)`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			repoID := args[0]

			if len(args) == 2 {
				result, err := ingest.IR(cmd.Context(), svc, repoID, args[1])
				if err != nil {
					return err
				}
				printIRResult(cmd, result)
				return nil
			}

			results, err := ingest.IRFromRepo(cmd.Context(), svc, repoID)
			if err != nil {
				return err
			}

			if len(results) == 0 {
				cmd.Println("No *.ir.json files found.")
				return nil
			}

			var errCount int
			for _, r := range results {
				if r.Err != nil {
					cmd.PrintErrf("  %s: error: %v\n", r.FilePath, r.Err)
					errCount++
					continue
				}
				printIRResult(cmd, r)
				cmd.Println()
			}
			if errCount > 0 {
				cmd.Printf("%d of %d files had errors.\n", errCount, len(results))
			}
			return nil
		},
	}
	return cmd
}

func printIRResult(cmd *cobra.Command, r *ingest.IRResult) {
	cmd.Printf("Repository:      %s\n", r.RepoID)
	cmd.Printf("File:            %s\n", r.FilePath)
	cmd.Printf("DevX Reports:    %d\n", r.DevXReports)
	cmd.Printf("PRISM Roadmaps:  %d\n", r.PRISMRoadmaps)
	cmd.Printf("PRISM Goals:     %d\n", r.PRISMGoals)
	cmd.Printf("PRISM Documents: %d\n", r.PRISMDocuments)
}
