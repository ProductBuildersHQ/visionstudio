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
	cmd.AddCommand(ingestGitCmd(), ingestChangelogCmd())
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
