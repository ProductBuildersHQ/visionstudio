package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/roadmap"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func roadmapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roadmap",
		Short: "Sync ROADMAP.md files with the database",
	}
	cmd.AddCommand(
		roadmapDiffCmd(),
		roadmapGenerateCmd(),
		roadmapImportCmd(),
	)
	return cmd
}

func roadmapDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff <roadmap.md>",
		Short: "Compare a ROADMAP.md file against the database",
		Long: `Parse a ROADMAP.md file and show differences with the database.
Reports status mismatches, title differences, and items missing from either side.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			f, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}
			defer func() {
				if err := f.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "warning: close file: %v\n", err)
				}
			}()

			parsed, err := roadmap.Parse(f)
			if err != nil {
				return fmt.Errorf("parse roadmap: %w", err)
			}

			if parsed.InitiativeID == "" {
				return fmt.Errorf("no **Initiative:** found in %s", args[0])
			}

			cmd.Printf("Initiative: %s\n", parsed.InitiativeID)
			cmd.Printf("Phases:     %d\n", len(parsed.Phases))
			totalItems := 0
			for _, p := range parsed.Phases {
				totalItems += len(p.Items)
			}
			cmd.Printf("RMIs:       %d\n\n", totalItems)

			dbInput, err := loadDiffInput(cmd, svc, parsed.InitiativeID)
			if err != nil {
				return err
			}

			diffs := roadmap.Diff(parsed, dbInput)

			if len(diffs) == 0 {
				cmd.Println("No differences found — file and database are in sync.")
				return nil
			}

			cmd.Printf("Found %d difference(s):\n\n", len(diffs))
			for _, d := range diffs {
				icon := diffIcon(d.Kind)
				cmd.Printf("  %s %s\n", icon, d.Message)
			}
			return nil
		},
	}
}

func roadmapGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate <initiative-id>",
		Short: "Generate a ROADMAP.md from database state",
		Long: `Generate a complete ROADMAP.md file from the current database state.
Writes to stdout by default; use --output to write to a file.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			output, _ := cmd.Flags().GetString("output")
			input, err := loadGenerateInput(cmd, svc, args[0])
			if err != nil {
				return err
			}

			var w *os.File
			if output != "" {
				w, err = os.Create(output)
				if err != nil {
					return fmt.Errorf("create output: %w", err)
				}
				defer func() {
					if err := w.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "warning: close output: %v\n", err)
					}
				}()
			} else {
				w = os.Stdout
			}

			if err := roadmap.Generate(w, input); err != nil {
				return fmt.Errorf("generate: %w", err)
			}

			if output != "" {
				cmd.Printf("Generated %s\n", output)
			}
			return nil
		},
	}
	cmd.Flags().StringP("output", "o", "", "Output file path (default: stdout)")
	return cmd
}

func loadGenerateInput(cmd *cobra.Command, svc *service.Service, initiativeID string) (*roadmap.GenerateInput, error) {
	init, err := svc.GetInitiative(cmd.Context(), initiativeID)
	if err != nil {
		return nil, fmt.Errorf("get initiative: %w", err)
	}

	phases, err := svc.ListPhases(cmd.Context(), initiativeID)
	if err != nil {
		return nil, fmt.Errorf("list phases: %w", err)
	}
	sort.Slice(phases, func(i, j int) bool {
		return phases[i].SequenceNumber < phases[j].SequenceNumber
	})

	allRMIs, err := svc.ListRMIs(cmd.Context(), initiativeID)
	if err != nil {
		return nil, fmt.Errorf("list RMIs: %w", err)
	}
	rmisByPhase := map[string][]*store.RoadmapItem{}
	for _, r := range allRMIs {
		rmisByPhase[r.PhaseID] = append(rmisByPhase[r.PhaseID], r)
	}

	var genPhases []roadmap.GeneratePhase
	for _, p := range phases {
		genPhases = append(genPhases, roadmap.GeneratePhase{
			Phase: p,
			RMIs:  rmisByPhase[p.ID],
		})
	}

	allDeps, err := svc.Store.ListAllDependencies(cmd.Context())
	if err != nil {
		return nil, fmt.Errorf("list dependencies: %w", err)
	}

	return &roadmap.GenerateInput{
		Initiative: init,
		Phases:     genPhases,
		Deps:       allDeps,
	}, nil
}

func loadDiffInput(cmd *cobra.Command, svc *service.Service, initiativeID string) (*roadmap.DiffInput, error) {
	input, err := loadGenerateInput(cmd, svc, initiativeID)
	if err != nil {
		return nil, err
	}
	return &roadmap.DiffInput{Phases: input.Phases}, nil
}

func roadmapImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <roadmap.md>",
		Short: "Import a ROADMAP.md into the database (create or update)",
		Long: `Parse a ROADMAP.md file and create or update phases, RMIs, and dependencies.

Existing entities are updated (title, phase, sequence); new ones are created.
The initiative must already exist in the database ('initiative create' first).

The parser recognizes exactly this structure (tables do NOT parse):

  # <Title>
  **Initiative:** ` + "`INIT-MYPROJECT-001`" + `
  **Repository:** ` + "`github.com/myorg/myrepo`" + `

  ## Phase 1 — <Phase Title>
  **Theme:** <optional theme>

  - [ ] ` + "`RMI-MYREPO-001`" + ` First item title
    - Depends on: ` + "`RMI-MYREPO-000`" + `
  - [x] ` + "`RMI-MYREPO-002`" + ` Completed item title

RMI IDs must be backticked; '[x]' marks completed. 'roadmap generate' emits this
exact format from database state, so generate → edit → import round-trips.

Use --dry-run to preview changes without writing.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			f, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}
			defer func() {
				if err := f.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "warning: close file: %v\n", err)
				}
			}()

			parsed, err := roadmap.Parse(f)
			if err != nil {
				return fmt.Errorf("parse roadmap: %w", err)
			}

			if parsed.InitiativeID == "" {
				return fmt.Errorf("no **Initiative:** found in %s", args[0])
			}

			if _, err := svc.GetInitiative(cmd.Context(), parsed.InitiativeID); err != nil {
				return fmt.Errorf("initiative %s not found in DB (create it first): %w", parsed.InitiativeID, err)
			}

			repo, _ := cmd.Flags().GetString("repo")
			itemType, _ := cmd.Flags().GetString("type")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			imp := &roadmap.Importer{
				Store: svc.Store,
				CreateRMI: func(ctx context.Context, id, repoID, initiativeID, phaseID, title, desc, iType, priority string, required bool, seq int, acceptance []string) error {
					_, err := svc.CreateRMI(ctx, id, repoID, initiativeID, phaseID, title, desc, iType, priority, required, seq, acceptance)
					return err
				},
				CreatePhase: func(ctx context.Context, id, initiativeID string, seq int, title, theme string) error {
					_, err := svc.CreatePhase(ctx, id, initiativeID, seq, title, theme)
					return err
				},
				CreateDep: svc.CreateDependency,
				UpdateRMI: func(ctx context.Context, rmi *store.RoadmapItem) error {
					return svc.UpdateRMI(ctx, rmi)
				},
			}

			if dryRun {
				cmd.Println("DRY RUN — no changes will be written")
			}

			actions, err := imp.Import(cmd.Context(), parsed, repo, itemType, dryRun)
			if err != nil {
				return fmt.Errorf("import: %w", err)
			}

			counts := map[string]int{}
			for _, a := range actions {
				key := a.Entity + ":" + a.Action
				counts[key]++
			}

			created := counts["phase:created"] + counts["rmi:created"]
			updated := counts["phase:updated"] + counts["rmi:updated"]

			cmd.Printf("\nImport summary for %s:\n", parsed.InitiativeID)
			cmd.Printf("  Phases:       %d created, %d updated, %d unchanged\n",
				counts["phase:created"], counts["phase:updated"], counts["phase:unchanged"])
			cmd.Printf("  RMIs:         %d created, %d updated, %d unchanged\n",
				counts["rmi:created"], counts["rmi:updated"], counts["rmi:unchanged"])
			cmd.Printf("  Dependencies: %d\n", counts["dependency:created"])

			if created > 0 || updated > 0 {
				cmd.Println("\nDetails:")
				for _, a := range actions {
					if a.Action == "unchanged" {
						continue
					}
					if a.Entity == "dependency" {
						continue
					}
					detail := ""
					if a.Detail != "" {
						detail = " (" + a.Detail + ")"
					}
					cmd.Printf("  [%s] %s %s%s\n", a.Action, a.Entity, a.ID, detail)
				}
			}
			return nil
		},
	}
	cmd.Flags().String("repo", "", "Default repository ID for new RMIs (overrides file's **Repository:**)")
	cmd.Flags().String("type", "capability", "Default item type for new RMIs")
	cmd.Flags().Bool("dry-run", false, "Preview changes without writing to the database")
	return cmd
}

func diffIcon(kind roadmap.DiffKind) string {
	switch kind {
	case roadmap.DiffStatusMismatch:
		return "[STATUS]"
	case roadmap.DiffTitleMismatch:
		return "[TITLE]"
	case roadmap.DiffMissingInDB:
		return "[+FILE]"
	case roadmap.DiffMissingInFile:
		return "[+DB]"
	case roadmap.DiffPhaseMissing:
		return "[PHASE]"
	default:
		return "[?]"
	}
}
