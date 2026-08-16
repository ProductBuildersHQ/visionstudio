package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/release"
	"github.com/ProductBuildersHQ/visionstudio/pkg/releasegate"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func releaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Release planning and management",
	}
	cmd.AddCommand(releasePlanCmd(), releaseRecordCmd(), releaseListCmd(), releaseShowCmd(), releaseAttachCmd(), releaseDetachCmd(), releaseDeleteCmd(), releaseUnshippedCmd(), releaseBackfillMatchCmd())
	return cmd
}

func releaseBackfillMatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backfill-match",
		Short: "AI-assisted historical release-to-initiative matching (review only, never auto-attaches)",
		Long: `Most repo history predates RMI adoption — releases here have no
trailer-derived initiative match, and that is expected, not a bug.
backfill-match surfaces evidence (release notes, tag, date) and candidate
initiatives for a human — or the reviewing agent session — to judge.
Any conclusion is Analyst inference, never fact: confirm a match with
'release attach', never write one automatically.`,
	}
	cmd.AddCommand(backfillMatchListCmd(), backfillMatchShowCmd())
	return cmd
}

func backfillMatchListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List releases with no initiative match (oldest first)",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoFilter, _ := cmd.Flags().GetString("repo")

			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			rels, err := svc.UnmatchedReleases(cmd.Context(), repoFilter)
			if err != nil {
				return err
			}
			if len(rels) == 0 {
				cmd.Println("No unmatched releases.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "RELEASE\tRELEASED\tHAS BODY")
			for _, r := range rels {
				hasBody := "no"
				if r.Body != "" {
					hasBody = "yes"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", r.ID, r.ReleasedAt.Format("2006-01-02"), hasBody)
			}
			return w.Flush()
		},
	}
	cmd.Flags().String("repo", "", "Filter by repository ID")
	return cmd
}

func backfillMatchShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <release-id>",
		Short: "Show a release's evidence and candidate initiatives for review",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			bc, err := svc.GetBackfillCandidates(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			cmd.Printf("=== Release evidence ===\n")
			cmd.Printf("ID:       %s\n", bc.Release.ID)
			cmd.Printf("Tag:      %s\n", bc.Release.Tag)
			cmd.Printf("Released: %s\n", bc.Release.ReleasedAt.Format("2006-01-02"))
			if bc.Release.URL != "" {
				cmd.Printf("URL:      %s\n", bc.Release.URL)
			}
			if bc.Release.Body != "" {
				cmd.Printf("Body:\n%s\n", bc.Release.Body)
			} else {
				cmd.Printf("Body:     (none captured — try 'ingest github-releases' first)\n")
			}

			cmd.Printf("\n=== Candidate initiatives (homed in or referencing this repository) ===\n")
			for _, c := range bc.Candidates {
				cmd.Printf("\n%s — %s [%s]\n", c.Initiative.ID, c.Initiative.Title, c.Initiative.Status)
				if c.Initiative.Description != "" {
					cmd.Printf("  %s\n", c.Initiative.Description)
				}
				if len(c.RMITitles) > 0 {
					cmd.Printf("  RMIs in this repo: %s\n", strings.Join(c.RMITitles, "; "))
				}
			}

			cmd.Printf("\n=== Review rules (CLAUDE.md \"AI-assisted historical backfill matching\") ===\n")
			cmd.Printf("- Any match here is Analyst inference, never Observed — regardless of confidence.\n")
			cmd.Printf("- Cite the specific evidence (a body/notes snippet) that supports a match; date\n")
			cmd.Printf("  proximity alone is not evidence.\n")
			cmd.Printf("- Nothing is written by this command. To confirm a match:\n")
			cmd.Printf("    visionstudio release attach %s --initiative <INIT-ID> [--rmi <RMI-ID>]\n", bc.Release.ID)
			return nil
		},
	}
}

func releaseUnshippedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unshipped",
		Short: "Initiatives claiming delivery with no release attached (stalest first)",
		Long: `The forcing-function queue: initiatives at delivery_complete or
releasing with zero associated releases. Working this queue to zero —
attach releases and transition to released, or consciously park — is
what activates the acceptance-mark quality signal.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			queue, err := svc.UnshippedQueue(cmd.Context(), time.Now())
			if err != nil {
				return err
			}
			if len(queue) == 0 {
				cmd.Println("Unshipped queue is empty — every delivered initiative has a release attached.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "INITIATIVE\tSTATUS\tDAYS STALE\tTITLE")
			for _, e := range queue {
				stale := "?"
				if e.StaleSince != nil {
					stale = fmt.Sprintf("%d", e.DaysStale)
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Initiative.ID, e.Initiative.Status, stale, e.Initiative.Title)
			}
			return w.Flush()
		},
	}
}

func releaseDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <release-id>",
		Short: "Delete a release record (associations are unlinked, nothing else is touched)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			if err := svc.Store.DeleteRelease(cmd.Context(), args[0]); err != nil {
				return err
			}
			cmd.Printf("Deleted: %s\n", args[0])
			return nil
		},
	}
}

func releaseRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Record a release (repo + tag) — the natural moment is right after updating CHANGELOG.json and tagging",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoID, _ := cmd.Flags().GetString("repo")
			tag, _ := cmd.Flags().GetString("tag")
			dateStr, _ := cmd.Flags().GetString("date")
			url, _ := cmd.Flags().GetString("url")
			notesRef, _ := cmd.Flags().GetString("notes-ref")
			inits, _ := cmd.Flags().GetStringSlice("initiative")
			rmis, _ := cmd.Flags().GetStringSlice("rmi")

			if repoID == "" || tag == "" {
				return fmt.Errorf("--repo and --tag are required")
			}
			var releasedAt time.Time
			if dateStr != "" {
				var err error
				releasedAt, err = time.Parse("2006-01-02", dateStr)
				if err != nil {
					return fmt.Errorf("parse --date (YYYY-MM-DD): %w", err)
				}
			}

			scaffold, _ := cmd.Flags().GetBool("scaffold")
			strict, _ := cmd.Flags().GetBool("strict")

			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			// Changelog release gate: check the convention, offer to
			// manufacture it — never presume the habit.
			if repo, repoErr := svc.GetRepository(cmd.Context(), repoID); repoErr == nil && repo.LocalPath != "" {
				gate := releasegate.Check(repo.LocalPath, tag)
				switch {
				case gate.VersionPresent:
					if notesRef == "" {
						notesRef = "CHANGELOG.json#" + tag
					}
				case scaffold:
					res, err := releasegate.Scaffold(cmd.Context(), repo.LocalPath, tag)
					if err != nil {
						return fmt.Errorf("scaffold changelog: %w", err)
					}
					cmd.Printf("Scaffolded %s: %d entries (%s) since %s — review and edit before publishing\n",
						res.ChangelogPath, res.Entries, strings.Join(res.Categories, ", "), res.SinceRef)
					if notesRef == "" {
						notesRef = "CHANGELOG.json#" + tag
					}
				case strict:
					return fmt.Errorf("changelog gate: %s has no entry for %s — re-run with --scaffold to generate one from conventional commits, or add it via schangelog", gate.ChangelogPath, tag)
				default:
					cmd.Printf("⚠ changelog gate: no CHANGELOG.json entry for %s — recording anyway; use --scaffold to generate one, or --strict to enforce\n", tag)
				}
			}

			rel, err := svc.RecordRelease(cmd.Context(), repoID, tag, releasedAt, url, notesRef, inits, rmis)
			if err != nil {
				return err
			}
			cmd.Printf("Recorded: %s (released %s)\n", rel.ID, rel.ReleasedAt.Format("2006-01-02"))
			if len(rel.InitiativeIDs) > 0 {
				cmd.Printf("  Initiatives: %s\n", strings.Join(rel.InitiativeIDs, ", "))
			}
			if len(rel.RMIIDs) > 0 {
				cmd.Printf("  RMIs: %s\n", strings.Join(rel.RMIIDs, ", "))
			}
			return nil
		},
	}
	cmd.Flags().String("repo", "", "Repository ID (required)")
	cmd.Flags().String("tag", "", "Release tag, e.g. v0.3.0 (required)")
	cmd.Flags().String("date", "", "Release date YYYY-MM-DD (default: now)")
	cmd.Flags().String("url", "", "GitHub release URL")
	cmd.Flags().String("notes-ref", "", "Changelog entry reference")
	cmd.Flags().StringSlice("initiative", nil, "Associated initiative ID (repeatable)")
	cmd.Flags().StringSlice("rmi", nil, "Associated RMI ID (repeatable)")
	cmd.Flags().Bool("scaffold", false, "Generate the CHANGELOG.json entry from conventional commits when missing")
	cmd.Flags().Bool("strict", false, "Refuse to record when CHANGELOG.json lacks an entry for the tag")
	return cmd
}

func releaseListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoID, _ := cmd.Flags().GetString("repo")
			initID, _ := cmd.Flags().GetString("initiative")

			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			rels, err := svc.ListReleases(cmd.Context(), repoID, initID)
			if err != nil {
				return err
			}
			if len(rels) == 0 {
				cmd.Println("No releases recorded.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tRELEASED\tINITIATIVES\tRMIS")
			for _, r := range rels {
				inits := "-"
				if len(r.InitiativeIDs) > 0 {
					inits = strings.Join(r.InitiativeIDs, ",")
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", r.ID, r.ReleasedAt.Format("2006-01-02"), inits, len(r.RMIIDs))
			}
			return w.Flush()
		},
	}
	cmd.Flags().String("repo", "", "Filter by repository ID")
	cmd.Flags().String("initiative", "", "Filter by initiative ID")
	return cmd
}

func releaseShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <release-id>",
		Short: "Show release details with associations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			rel, err := svc.GetRelease(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			cmd.Printf("Release:     %s\n", rel.ID)
			cmd.Printf("Repository:  %s\n", rel.RepositoryID)
			cmd.Printf("Tag:         %s\n", rel.Tag)
			cmd.Printf("Released:    %s\n", rel.ReleasedAt.Format("2006-01-02"))
			if rel.URL != "" {
				cmd.Printf("URL:         %s\n", rel.URL)
			}
			if rel.NotesRef != "" {
				cmd.Printf("Notes:       %s\n", rel.NotesRef)
			}
			if rel.Body != "" {
				cmd.Printf("Body:        %s\n", firstLine(rel.Body))
			}
			if len(rel.InitiativeIDs) > 0 {
				cmd.Printf("Initiatives: %s\n", strings.Join(rel.InitiativeIDs, ", "))
			}
			if len(rel.RMIIDs) > 0 {
				cmd.Printf("RMIs:        %s\n", strings.Join(rel.RMIIDs, ", "))
			}
			return nil
		},
	}
}

// firstLine returns the first non-empty line of s, truncated for
// one-line CLI display — full body text is available via the API/DB for
// RMI-VISIONSTUDIO-315's backfill-match review, not meant to dump here.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(line) > 100 {
			return line[:100] + "…"
		}
		return line
	}
	return ""
}

func releaseAttachCmd() *cobra.Command {
	return releaseAssociationCmd("attach", "Attached",
		"Attach initiative or RMI associations to a release",
		func(svc *service.Service, ctx context.Context, id string, inits, rmis []string) (*store.Release, error) {
			return svc.AttachRelease(ctx, id, inits, rmis)
		})
}

func releaseDetachCmd() *cobra.Command {
	return releaseAssociationCmd("detach", "Detached",
		"Detach initiative or RMI associations from a release",
		func(svc *service.Service, ctx context.Context, id string, inits, rmis []string) (*store.Release, error) {
			return svc.DetachRelease(ctx, id, inits, rmis)
		})
}

func releaseAssociationCmd(verb, done, short string, apply func(*service.Service, context.Context, string, []string, []string) (*store.Release, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   verb + " <release-id>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inits, _ := cmd.Flags().GetStringSlice("initiative")
			rmis, _ := cmd.Flags().GetStringSlice("rmi")
			if len(inits) == 0 && len(rmis) == 0 {
				return fmt.Errorf("provide --initiative and/or --rmi")
			}

			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			rel, err := apply(svc, cmd.Context(), args[0], inits, rmis)
			if err != nil {
				return err
			}
			cmd.Printf("%s. %s now has %d initiatives, %d RMIs\n", done, rel.ID, len(rel.InitiativeIDs), len(rel.RMIIDs))
			return nil
		},
	}
	cmd.Flags().StringSlice("initiative", nil, "Initiative ID (repeatable)")
	cmd.Flags().StringSlice("rmi", nil, "RMI ID (repeatable)")
	return cmd
}

func releasePlanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "plan <initiative-id>",
		Short: "Show dependency-ordered release plan for an initiative",
		Long: `Compute a topological release plan from repository dependencies.

Repos are grouped into stages: stage 0 has no in-initiative dependencies
(release first), stage 1 depends only on stage 0, and so on.
Repos in the same stage can be released in parallel.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			initID := args[0]
			rmis, err := svc.ListRMIs(cmd.Context(), initID)
			if err != nil {
				return fmt.Errorf("list RMIs: %w", err)
			}

			repoDeps, err := svc.Store.ListAllRepoDependencies(cmd.Context())
			if err != nil {
				return fmt.Errorf("list repo deps: %w", err)
			}

			rs, err := release.Plan(initID, rmis, repoDeps)
			if err != nil {
				return fmt.Errorf("plan: %w", err)
			}

			stages := rs.Stages()
			if len(stages) == 0 {
				cmd.Println("No completed RMIs — nothing to release.")
				return nil
			}

			cmd.Printf("Release plan for %s (%d repos, %d stages):\n\n",
				initID, len(rs.Components), len(stages))

			for _, stage := range stages {
				cmd.Printf("Stage %d", stage.Number)
				if stage.Number == 0 {
					cmd.Printf(" (no in-initiative deps — release first)")
				}
				cmd.Println(":")

				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				for _, c := range stage.Repos {
					repo := c.RepositoryID
					if idx := strings.LastIndex(repo, "/"); idx >= 0 {
						repo = repo[idx+1:]
					}
					_, _ = fmt.Fprintf(w, "  %s\t%d RMIs\t%s\n",
						repo, len(c.RMIs), strings.Join(c.RMIs, ", "))
				}
				if err := w.Flush(); err != nil {
					return err
				}
				cmd.Println()
			}
			return nil
		},
	}
}
