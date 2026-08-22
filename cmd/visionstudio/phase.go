package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/initiative"
)

func phaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "phase",
		Short: "Manage phases within an initiative",
	}
	cmd.AddCommand(phaseAddCmd(), phaseListCmd(), phaseRemoveCmd())
	return cmd
}

func phaseRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <phase-id>",
		Short: "Remove an empty phase (fails if it still has member RMIs)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			if err := svc.RemovePhase(cmd.Context(), args[0]); err != nil {
				return err
			}
			cmd.Printf("Removed phase %s\n", args[0])
			return nil
		},
	}
}

func phaseAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a phase to an initiative",
		Long: `Add a phase — a themed grouping of ~5 RMIs within an initiative.

Phase IDs follow '<INIT-ID>/phase-N'. Phase status is always derived from its
member RMIs (completed and cancelled both count as resolved), never set
directly. 'roadmap import' creates phases from '## Phase N — Title' headings
automatically; use this command when adding phases individually.`,
		Example: `  visionstudio phase add --id INIT-MYPROJECT-001/phase-2 \
    --initiative INIT-MYPROJECT-001 --sequence 2 \
    --title "Persistence layer" --theme "omniroadmap integration"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			id, _ := cmd.Flags().GetString("id")
			initID, _ := cmd.Flags().GetString("initiative")
			seq, _ := cmd.Flags().GetInt("sequence")
			title, _ := cmd.Flags().GetString("title")
			theme, _ := cmd.Flags().GetString("theme")

			if id == "" || initID == "" || title == "" {
				return fmt.Errorf("--id, --initiative, and --title are required")
			}

			p, err := svc.CreatePhase(cmd.Context(), id, initID, seq, title, theme)
			if err != nil {
				return err
			}
			cmd.Printf("Created phase: %s (seq %d)\n", p.ID, p.SequenceNumber)
			return nil
		},
	}
	cmd.Flags().String("id", "", "Phase ID (e.g. INIT-FOO-001/phase-1) (required)")
	cmd.Flags().String("initiative", "", "Parent initiative ID (required)")
	cmd.Flags().Int("sequence", 1, "Sequence number")
	cmd.Flags().String("title", "", "Phase title (required)")
	cmd.Flags().String("theme", "", "Phase theme")
	return cmd
}

func phaseListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List phases for an initiative with derived status",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			initID, _ := cmd.Flags().GetString("initiative")
			if initID == "" {
				return fmt.Errorf("--initiative is required")
			}

			detail, err := svc.GetInitiativeDetail(cmd.Context(), initID)
			if err != nil {
				return err
			}

			if len(detail.Phases) == 0 {
				cmd.Println("No phases found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "#\tID\tTITLE\tTHEME\tSTATUS\tRMIs")
			for _, p := range detail.Phases {
				theme := p.Phase.Theme
				if theme == "" {
					theme = "-"
				}
				completed := 0
				for _, r := range p.RMIs {
					if r.Status == "completed" {
						completed++
					}
				}
				status := p.Status
				rmiSummary := fmt.Sprintf("%d/%d", completed, len(p.RMIs))
				if len(p.RMIs) == 0 {
					rmiSummary = "-"
					status = initiative.PhasePlanned
				}
				_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
					p.Phase.SequenceNumber, p.Phase.ID, p.Phase.Title, theme, status, rmiSummary)
			}
			return w.Flush()
		},
	}
	cmd.Flags().String("initiative", "", "Initiative ID (required)")
	return cmd
}
