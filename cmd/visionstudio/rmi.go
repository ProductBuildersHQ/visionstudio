package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	rmidomain "github.com/ProductBuildersHQ/visionstudio/pkg/rmi"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func rmiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rmi",
		Short: "Manage roadmap items (RMIs)",
	}
	cmd.AddCommand(
		rmiCreateCmd(),
		rmiGetCmd(),
		rmiListCmd(),
		rmiUpdateCmd(),
		rmiUpdatePhaseCmd(),
		rmiMoveCmd(),
		rmiDepCmd(),
		rmiBulkUpdateCmd(),
	)
	return cmd
}

// matchingRMIsForRepo returns the RMIs currently on repoID, optionally
// narrowed to a single initiative.
func matchingRMIsForRepo(ctx context.Context, svc *service.Service, repoID, initiativeFilter string) ([]*store.RoadmapItem, error) {
	candidates, err := svc.ListRMIsByRepo(ctx, repoID)
	if err != nil {
		return nil, err
	}
	if initiativeFilter == "" {
		return candidates, nil
	}
	var filtered []*store.RoadmapItem
	for _, r := range candidates {
		if r.InitiativeID == initiativeFilter {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

// reassignRMIRepo repoints every given RMI to toID and persists it.
func reassignRMIRepo(ctx context.Context, svc *service.Service, candidates []*store.RoadmapItem, toID string) error {
	for _, r := range candidates {
		r.RepositoryID = toID
		if err := svc.UpdateRMI(ctx, r); err != nil {
			return fmt.Errorf("update %s: %w", r.ID, err)
		}
	}
	return nil
}

func rmiBulkUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk-update",
		Short: "Reassign the repository field across matching RMIs",
		Long: `Repoints every RMI currently on --repo to --set-repo in one call.
Use --initiative to narrow the scope to a single initiative, and
--dry-run to preview which RMIs would change without persisting.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			fromRepo, _ := cmd.Flags().GetString("repo")
			toRepo, _ := cmd.Flags().GetString("set-repo")
			initiative, _ := cmd.Flags().GetString("initiative")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			if fromRepo == "" || toRepo == "" {
				return fmt.Errorf("--repo and --set-repo are required")
			}

			fromID, err := resolveRepoID(cmd.Context(), svc, fromRepo)
			if err != nil {
				return fmt.Errorf("--repo: %w", err)
			}
			toID, err := resolveRepoID(cmd.Context(), svc, toRepo)
			if err != nil {
				return fmt.Errorf("--set-repo: %w", err)
			}
			if fromID == toID {
				return fmt.Errorf("--repo and --set-repo resolve to the same repository (%s)", fromID)
			}

			candidates, err := matchingRMIsForRepo(cmd.Context(), svc, fromID, initiative)
			if err != nil {
				return err
			}
			if len(candidates) == 0 {
				cmd.Printf("No matching RMIs on %s\n", fromID)
				return nil
			}

			verb := "Reassigning"
			if dryRun {
				verb = "Would reassign"
			}
			cmd.Printf("%s %d RMI(s) from %s to %s:\n", verb, len(candidates), fromID, toID)
			for _, r := range candidates {
				cmd.Printf("  %s (%s)\n", r.ID, r.Title)
			}
			if dryRun {
				cmd.Println("Dry run — no changes made. Re-run without --dry-run to apply.")
				return nil
			}

			if err := reassignRMIRepo(cmd.Context(), svc, candidates, toID); err != nil {
				return err
			}
			cmd.Printf("Reassigned %d RMI(s)\n", len(candidates))
			return nil
		},
	}
	cmd.Flags().String("repo", "", "Repository ID to reassign from (required)")
	cmd.Flags().String("set-repo", "", "Repository ID to reassign to (required)")
	cmd.Flags().String("initiative", "", "Only reassign RMIs in this initiative")
	cmd.Flags().Bool("dry-run", false, "Preview without persisting")
	return cmd
}

func rmiMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <rmi-id>",
		Short: "Move an RMI to another phase (and its initiative)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			phaseID, err := cmd.Flags().GetString("phase")
			if err != nil {
				return err
			}
			seq, err := cmd.Flags().GetInt("seq")
			if err != nil {
				return err
			}
			rmi, err := svc.MoveRMI(cmd.Context(), args[0], phaseID, seq)
			if err != nil {
				return err
			}
			cmd.Printf("Moved %s to %s (initiative %s)\n", rmi.ID, rmi.PhaseID, rmi.InitiativeID)
			return nil
		},
	}
	cmd.Flags().String("phase", "", "Target phase ID (INITIATIVE-ID/phase-N)")
	cmd.Flags().Int("seq", 0, "Sequence number within the target phase (optional)")
	if err := cmd.MarkFlagRequired("phase"); err != nil {
		panic(err)
	}
	return cmd
}

func rmiCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new roadmap item",
		Long: `Create a roadmap item (RMI) — the unit of trackable work, tied to one repository.

ID convention: RMI-<REPOSLUG>-NNN (regex ^RMI-[A-Z0-9]+-\d{3}$), where REPOSLUG is
the uppercased repository name with separators removed (repo 'prism-roadmap' →
RMI-PRISMROADMAP-001). Numbers are per-repo: check 'rmi list --repo <repo-id>'
for the next free one. Commits implementing an RMI carry the git trailer
'Refs: <RMI-ID>' ('work claim' prints it).

The --repo repository must already be registered ('registry list' / 'registry add').
--initiative and --phase attach the RMI to its parents; an initiative's RMIs may
span multiple repositories, each RMI naming its own --repo.

--origin records how the scope was identified (for spec-completeness telemetry):
  spec                in the initiative's original PRD/ROADMAP (default)
  implementation      discovered while implementing another RMI
  acceptance_testing  a human found the gap using the shipped result
  discussion          proposed directly by a human in conversation`,
		Example: `  visionstudio rmi create \
    --id RMI-MYREPO-007 \
    --repo github.com/myorg/myrepo \
    --initiative INIT-MYPROJECT-001 --phase INIT-MYPROJECT-001/phase-2 \
    --title "Add evidence entity" --type capability --priority high \
    --acceptance "schema generated; tests pass" --origin implementation`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			id, _ := cmd.Flags().GetString("id")
			repo, _ := cmd.Flags().GetString("repo")
			initiative, _ := cmd.Flags().GetString("initiative")
			phase, _ := cmd.Flags().GetString("phase")
			title, _ := cmd.Flags().GetString("title")
			desc, _ := cmd.Flags().GetString("description")
			itemType, _ := cmd.Flags().GetString("type")
			priority, _ := cmd.Flags().GetString("priority")
			required, _ := cmd.Flags().GetBool("required")
			seq, _ := cmd.Flags().GetInt("sequence")
			acceptanceRaw, _ := cmd.Flags().GetString("acceptance")
			origin, _ := cmd.Flags().GetString("origin")

			if id == "" || repo == "" || title == "" || itemType == "" {
				return fmt.Errorf("--id, --repo, --title, and --type are required")
			}
			if !rmidomain.ValidOrigin(origin) {
				return fmt.Errorf("invalid --origin %q (want one of: %s)", origin, strings.Join(rmidomain.Origins, ", "))
			}

			repo, err = resolveRepoID(cmd.Context(), svc, repo)
			if err != nil {
				return err
			}

			var acceptance []string
			if acceptanceRaw != "" {
				acceptance = strings.Split(acceptanceRaw, ";")
				for i := range acceptance {
					acceptance[i] = strings.TrimSpace(acceptance[i])
				}
			}

			rmi, err := svc.CreateRMI(cmd.Context(), id, repo, initiative, phase, title, desc, itemType, priority, required, seq, acceptance)
			if err != nil {
				return err
			}

			// Handle context_spec and origin, if provided -- both are set
			// post-create via UpdateRMI rather than added to CreateRMI's
			// already-long positional signature.
			contextSpecJSON, _ := cmd.Flags().GetString("context-spec")
			needsUpdate := false
			if contextSpecJSON != "" {
				var spec store.ContextSpec
				if err := json.Unmarshal([]byte(contextSpecJSON), &spec); err != nil {
					return fmt.Errorf("invalid --context-spec JSON: %w", err)
				}
				rmi.ContextSpec = &spec
				needsUpdate = true
			}
			if origin != "" {
				rmi.Origin = origin
				needsUpdate = true
			}
			if needsUpdate {
				if err := svc.UpdateRMI(cmd.Context(), rmi); err != nil {
					return err
				}
			}

			cmd.Printf("Created RMI: %s (%s)\n", rmi.ID, rmi.Status)
			return nil
		},
	}
	cmd.Flags().String("id", "", "RMI ID (e.g. RMI-MYREPO-001) (required)")
	cmd.Flags().String("repo", "", "Repository ID (e.g. github.com/org/repo) (required)")
	cmd.Flags().String("initiative", "", "Parent initiative ID")
	cmd.Flags().String("phase", "", "Parent phase ID")
	cmd.Flags().String("title", "", "RMI title (required)")
	cmd.Flags().String("description", "", "Detailed description")
	cmd.Flags().String("type", "capability", "Item type (capability, fix, chore, spike) (required)")
	cmd.Flags().String("priority", "", "Priority (high, medium, low)")
	cmd.Flags().Bool("required", true, "Whether this RMI is required for phase completion")
	cmd.Flags().Int("sequence", 0, "Sequence number within phase")
	cmd.Flags().String("acceptance", "", "Acceptance criteria (semicolon-separated)")
	cmd.Flags().String("context-spec", "", "Context spec JSON: {\"extra_repos\":[...],\"include_specs\":[...],\"exclude_specs\":[...]}")
	cmd.Flags().String("origin", "", "How this RMI's scope was identified: spec (default), implementation, acceptance_testing, discussion")
	return cmd
}

func rmiGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <rmi-id>",
		Short: "Show RMI details with dependencies",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			detail, err := svc.GetRMIDetail(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(detail)
			case "text":
				// falls through to the text rendering below
			default:
				return fmt.Errorf("unknown format: %s (use text or json)", format)
			}

			rmi := detail.RMI
			cmd.Printf("RMI:         %s\n", rmi.ID)
			cmd.Printf("Title:       %s\n", rmi.Title)
			cmd.Printf("Status:      %s\n", rmi.Status)
			cmd.Printf("Type:        %s\n", rmi.ItemType)
			cmd.Printf("Repository:  %s\n", rmi.RepositoryID)
			if rmi.InitiativeID != "" {
				cmd.Printf("Initiative:  %s\n", rmi.InitiativeID)
			}
			if rmi.PhaseID != "" {
				cmd.Printf("Phase:       %s\n", rmi.PhaseID)
			}
			if rmi.Description != "" {
				cmd.Printf("Description: %s\n", rmi.Description)
			}
			if rmi.Priority != "" {
				cmd.Printf("Priority:    %s\n", rmi.Priority)
			}
			cmd.Printf("Required:    %v\n", rmi.Required)
			if rmi.Origin != "" && rmi.Origin != rmidomain.OriginSpec {
				cmd.Printf("Origin:      %s\n", rmi.Origin)
			}
			if rmi.SequenceNumber != 0 {
				cmd.Printf("Sequence:    %d\n", rmi.SequenceNumber)
			}
			cmd.Printf("Created:     %s\n", rmi.CreatedAt.Format("2006-01-02 15:04"))
			if rmi.CompletedAt != nil {
				cmd.Printf("Completed:   %s\n", rmi.CompletedAt.Format("2006-01-02 15:04"))
			}

			if len(rmi.AcceptanceCriteria) > 0 {
				cmd.Println("\nAcceptance Criteria:")
				for i, ac := range rmi.AcceptanceCriteria {
					cmd.Printf("  %d. %s\n", i+1, ac)
				}
			}

			if rmi.ContextSpec != nil {
				cmd.Println("\nContext Spec:")
				if len(rmi.ContextSpec.ExtraRepos) > 0 {
					cmd.Printf("  Extra Repos:   %s\n", strings.Join(rmi.ContextSpec.ExtraRepos, ", "))
				}
				if len(rmi.ContextSpec.IncludeSpecs) > 0 {
					cmd.Printf("  Include Specs: %s\n", strings.Join(rmi.ContextSpec.IncludeSpecs, ", "))
				}
				if len(rmi.ContextSpec.ExcludeSpecs) > 0 {
					cmd.Printf("  Exclude Specs: %s\n", strings.Join(rmi.ContextSpec.ExcludeSpecs, ", "))
				}
			}

			if len(detail.Dependencies) > 0 {
				cmd.Println("\nDependencies:")
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				_, _ = fmt.Fprintln(w, "  DIRECTION\tRMI\tRELATIONSHIP")
				for _, d := range detail.Dependencies {
					if d.SourceRMIID == rmi.ID {
						_, _ = fmt.Fprintf(w, "  depends on\t%s\t%s\n", d.TargetRMIID, d.Relationship)
					} else {
						_, _ = fmt.Fprintf(w, "  depended by\t%s\t%s\n", d.SourceRMIID, d.Relationship)
					}
				}
				if err := w.Flush(); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().String("format", "text", "Output format: text or json")
	return cmd
}

func rmiListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List roadmap items",
		Long: `List roadmap items (RMIs).

Scope the listing with --initiative or --repo. Use --grep <term> to search
RMI IDs and titles (case-insensitive substring) — useful for finding which
RMI describes a given change or feature. --grep works on its own, searching
across every initiative and repository, or combines with --initiative/--repo
to search within that scope. --origin narrows by how the scope was identified.

Examples:
  visionstudio rmi list --initiative INIT-MYPROJECT-001
  visionstudio rmi list --repo github.com/myorg/myrepo --origin acceptance_testing
  visionstudio rmi list --grep "grokifyql"          # which RMI covers this?
  visionstudio rmi list --grep question --repo github.com/myorg/myrepo`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			initiative, _ := cmd.Flags().GetString("initiative")
			repo, _ := cmd.Flags().GetString("repo")
			originFilter, _ := cmd.Flags().GetString("origin")
			grep, _ := cmd.Flags().GetString("grep")

			// --grep enables cross-initiative discovery, so it relaxes the
			// usual requirement that a scope filter be given.
			if initiative == "" && repo == "" && grep == "" {
				return fmt.Errorf("either --initiative, --repo, or --grep is required")
			}
			if !rmidomain.ValidOrigin(originFilter) {
				return fmt.Errorf("invalid --origin %q (want one of: %s)", originFilter, strings.Join(rmidomain.Origins, ", "))
			}

			if repo != "" {
				repo, err = resolveRepoID(cmd.Context(), svc, repo)
				if err != nil {
					return err
				}
			}

			var rmis []*store.RoadmapItem
			switch {
			case initiative != "":
				rmis, err = svc.ListRMIs(cmd.Context(), initiative)
			case repo != "":
				rmis, err = svc.ListRMIsByRepo(cmd.Context(), repo)
			default:
				// --grep only: search across every initiative and repo.
				rmis, err = svc.ListAllRMIs(cmd.Context())
			}
			if err != nil {
				return err
			}

			if grep != "" {
				needle := strings.ToLower(grep)
				matched := make([]*store.RoadmapItem, 0, len(rmis))
				for _, r := range rmis {
					if strings.Contains(strings.ToLower(r.ID), needle) ||
						strings.Contains(strings.ToLower(r.Title), needle) {
						matched = append(matched, r)
					}
				}
				rmis = matched
			}

			if originFilter != "" {
				filtered := make([]*store.RoadmapItem, 0, len(rmis))
				for _, r := range rmis {
					o := r.Origin
					if o == "" {
						o = rmidomain.OriginSpec
					}
					if o == originFilter {
						filtered = append(filtered, r)
					}
				}
				rmis = filtered
			}

			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(rmis)
			case "text":
				// falls through to the tabwriter rendering below
			default:
				return fmt.Errorf("unknown format: %s (use text or json)", format)
			}

			if len(rmis) == 0 {
				cmd.Println("No RMIs found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tTYPE\tREQUIRED\tORIGIN\tREPO")
			for _, r := range rmis {
				req := "yes"
				if !r.Required {
					req = "no"
				}
				origin := r.Origin
				if origin == "" {
					origin = rmidomain.OriginSpec
				}
				repoShort := r.RepositoryID
				if idx := strings.LastIndex(repoShort, "/"); idx >= 0 {
					repoShort = repoShort[idx+1:]
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					r.ID, r.Title, r.Status, r.ItemType, req, origin, repoShort)
			}
			return w.Flush()
		},
	}
	cmd.Flags().String("initiative", "", "Filter by initiative ID")
	cmd.Flags().String("repo", "", "Filter by repository ID")
	cmd.Flags().String("origin", "", "Filter by origin: spec, implementation, acceptance_testing, discussion")
	cmd.Flags().String("grep", "", "Search RMI IDs and titles (case-insensitive substring); works across all RMIs when no scope is given")
	cmd.Flags().String("format", "text", "Output format: text or json")
	return cmd
}

func rmiUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <rmi-id>",
		Short: "Update an RMI's status or fields",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			status, _ := cmd.Flags().GetString("status")
			title, _ := cmd.Flags().GetString("title")
			description, _ := cmd.Flags().GetString("description")
			priority, _ := cmd.Flags().GetString("priority")
			repo, _ := cmd.Flags().GetString("repo")
			contextSpecJSON, _ := cmd.Flags().GetString("context-spec")
			origin, _ := cmd.Flags().GetString("origin")

			if status == "" && title == "" && repo == "" && !cmd.Flags().Changed("description") && !cmd.Flags().Changed("priority") && !cmd.Flags().Changed("required") && contextSpecJSON == "" && !cmd.Flags().Changed("origin") {
				return fmt.Errorf("at least one of --status, --title, --repo, --description, --priority, --required, --context-spec, or --origin is required")
			}
			if !rmidomain.ValidOrigin(origin) {
				return fmt.Errorf("invalid --origin %q (want one of: %s)", origin, strings.Join(rmidomain.Origins, ", "))
			}

			if repo != "" {
				repo, err = resolveRepoID(cmd.Context(), svc, repo)
				if err != nil {
					return err
				}
			}

			if status != "" {
				rmi, err := svc.UpdateRMIStatus(cmd.Context(), args[0], status)
				if err != nil {
					return err
				}
				cmd.Printf("Updated %s status to %s\n", rmi.ID, rmi.Status)
			}

			if title != "" || repo != "" || cmd.Flags().Changed("description") || cmd.Flags().Changed("priority") || cmd.Flags().Changed("required") || contextSpecJSON != "" || cmd.Flags().Changed("origin") {
				rmi, err := svc.GetRMI(cmd.Context(), args[0])
				if err != nil {
					return err
				}
				if title != "" {
					rmi.Title = title
				}
				if repo != "" {
					rmi.RepositoryID = repo
				}
				if cmd.Flags().Changed("description") {
					rmi.Description = description
				}
				if cmd.Flags().Changed("priority") {
					rmi.Priority = priority
				}
				if cmd.Flags().Changed("required") {
					required, err := cmd.Flags().GetBool("required")
					if err != nil {
						return err
					}
					rmi.Required = required
				}
				if contextSpecJSON != "" {
					if contextSpecJSON == "null" || contextSpecJSON == "{}" {
						rmi.ContextSpec = nil
					} else {
						var spec store.ContextSpec
						if err := json.Unmarshal([]byte(contextSpecJSON), &spec); err != nil {
							return fmt.Errorf("invalid --context-spec JSON: %w", err)
						}
						rmi.ContextSpec = &spec
					}
				}
				if cmd.Flags().Changed("origin") {
					rmi.Origin = origin
				}
				if err := svc.UpdateRMI(cmd.Context(), rmi); err != nil {
					return err
				}
				var fields []string
				if title != "" {
					fields = append(fields, "title")
				}
				if repo != "" {
					fields = append(fields, "repo")
				}
				if cmd.Flags().Changed("description") {
					fields = append(fields, "description")
				}
				if cmd.Flags().Changed("priority") {
					fields = append(fields, "priority")
				}
				if cmd.Flags().Changed("required") {
					fields = append(fields, "required")
				}
				if contextSpecJSON != "" {
					fields = append(fields, "context_spec")
				}
				if cmd.Flags().Changed("origin") {
					fields = append(fields, "origin")
				}
				cmd.Printf("Updated %s fields: %s\n", args[0], strings.Join(fields, ", "))
			}
			return nil
		},
	}
	cmd.Flags().String("status", "", "New status (proposed, planned, ready, in_progress, completed, blocked, cancelled)")
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().String("repo", "", "New repository ID (e.g. github.com/org/repo)")
	cmd.Flags().String("description", "", "New description (use empty string to clear)")
	cmd.Flags().String("priority", "", "New priority (use empty string to clear)")
	cmd.Flags().Bool("required", false, "Whether the RMI is required for phase completion")
	cmd.Flags().String("context-spec", "", "Context spec JSON (use \"null\" or \"{}\" to clear)")
	cmd.Flags().String("origin", "", "How this RMI's scope was identified: spec, implementation, acceptance_testing, discussion")
	return cmd
}

func rmiUpdatePhaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-phase <phase-id>",
		Short: "Bulk-update status of all RMIs in a phase",
		Long: `Transition all RMIs in a phase to the target status.
Use --from to only transition RMIs currently in a specific status.

Phase ID format: INITIATIVE-ID/phase-N (e.g. INIT-PRISMCONTROL-001/phase-5).

Examples:
  visionstudio rmi update-phase INIT-X-001/phase-3 --status ready
  visionstudio rmi update-phase INIT-X-001/phase-3 --status ready --from proposed`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			status, _ := cmd.Flags().GetString("status")
			from, _ := cmd.Flags().GetString("from")
			if status == "" {
				return fmt.Errorf("--status is required")
			}

			updated, skipped, err := svc.UpdatePhaseStatus(cmd.Context(), args[0], from, status)
			if err != nil {
				return err
			}

			if len(updated) == 0 {
				cmd.Println("No RMIs were updated.")
			} else {
				cmd.Printf("Updated %d RMIs to %s:\n", len(updated), status)
				for _, id := range updated {
					cmd.Printf("  %s\n", id)
				}
			}
			if len(skipped) > 0 {
				cmd.Printf("Skipped %d (already %s or filtered by --from): %s\n",
					len(skipped), status, strings.Join(skipped, ", "))
			}
			return nil
		},
	}
	cmd.Flags().String("status", "", "Target status (required)")
	cmd.Flags().String("from", "", "Only transition RMIs currently in this status (optional)")
	return cmd
}

func rmiDepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dep",
		Short: "Manage RMI dependencies",
	}
	cmd.AddCommand(rmiDepAddCmd(), rmiDepListCmd())
	return cmd
}

func rmiDepAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a dependency between two RMIs",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			source, _ := cmd.Flags().GetString("source")
			target, _ := cmd.Flags().GetString("target")
			rel, _ := cmd.Flags().GetString("relationship")

			if source == "" || target == "" {
				return fmt.Errorf("--source and --target are required")
			}

			if err := svc.CreateDependency(cmd.Context(), source, target, rel); err != nil {
				return err
			}
			cmd.Printf("Added dependency: %s -> %s (%s)\n", source, target, rel)
			return nil
		},
	}
	cmd.Flags().String("source", "", "Source RMI ID (the one that depends) (required)")
	cmd.Flags().String("target", "", "Target RMI ID (the one depended upon) (required)")
	cmd.Flags().String("relationship", "requires", "Relationship type (requires, relates)")
	return cmd
}

func rmiDepListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <rmi-id>",
		Short: "List dependencies for an RMI",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			deps, err := svc.ListDependencies(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if len(deps) == 0 {
				cmd.Println("No dependencies found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "SOURCE\tTARGET\tRELATIONSHIP")
			for _, d := range deps {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", d.SourceRMIID, d.TargetRMIID, d.Relationship)
			}
			return w.Flush()
		},
	}
}
