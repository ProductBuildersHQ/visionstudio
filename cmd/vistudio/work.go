package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func defaultWorker() string {
	return os.Getenv("CLAUDE_CODE_SESSION_ID")
}

func workCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "work",
		Short: "Find and manage work assignments",
	}
	cmd.AddCommand(
		workReadyCmd(),
		workClaimCmd(),
		workClaimPhaseCmd(),
		workRenewCmd(),
		workReleaseCmd(),
		workCompleteCmd(),
		workCompletePhaseCmd(),
		workUpdateCmd(),
		workStatusCmd(),
	)
	return cmd
}

func workReadyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ready",
		Short: "List RMIs that are ready, unblocked, and unclaimed",
		Long: `Show roadmap items that can be worked on right now:
  - Status is "ready"
  - All "requires" dependencies are completed
  - No active assignment (unclaimed)

Use --initiative or --repo to narrow results.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			initiative, _ := cmd.Flags().GetString("initiative")
			repo, _ := cmd.Flags().GetString("repo")

			filters := service.WorkReadyFilters{
				InitiativeID: initiative,
				RepoID:       repo,
			}

			ready, err := svc.WorkReady(cmd.Context(), filters)
			if err != nil {
				return err
			}

			if len(ready) == 0 {
				cmd.Println("No ready work items found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tTITLE\tTYPE\tPRIORITY\tREPO\tINITIATIVE")
			for _, r := range ready {
				priority := r.Priority
				if priority == "" {
					priority = "-"
				}
				repoShort := r.RepositoryID
				if idx := strings.LastIndex(repoShort, "/"); idx >= 0 {
					repoShort = repoShort[idx+1:]
				}
				initID := r.InitiativeID
				if initID == "" {
					initID = "-"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					r.ID, r.Title, r.ItemType, priority, repoShort, initID)
			}
			return w.Flush()
		},
	}
	cmd.Flags().String("initiative", "", "Filter by initiative ID")
	cmd.Flags().String("repo", "", "Filter by repository ID")
	return cmd
}

func workClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim <rmi-id>",
		Short: "Claim an RMI for work — prints the git trailer",
		Long: `Claim a roadmap item, creating a lease-based assignment.
The RMI transitions to in_progress and the git trailer is printed for use in commits.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			worker, _ := cmd.Flags().GetString("worker")
			workspace, _ := cmd.Flags().GetString("workspace")
			leaseHours, _ := cmd.Flags().GetInt("lease-hours")

			if worker == "" {
				worker = defaultWorker()
			}
			if worker == "" {
				return fmt.Errorf("--worker is required (or set CLAUDE_CODE_SESSION_ID)")
			}

			lease := time.Duration(leaseHours) * time.Hour

			result, err := svc.ClaimRMI(cmd.Context(), args[0], worker, workspace, lease)
			if err != nil {
				return err
			}

			cmd.Printf("Claimed %s (assignment: %s)\n", args[0], result.Assignment.ID)
			cmd.Printf("Worker:  %s\n", result.Assignment.Worker)
			cmd.Printf("Lease:   %s\n", result.Assignment.LeaseExpiresAt.Format("2006-01-02 15:04"))
			cmd.Printf("\nGit trailer for commits:\n  %s\n", result.TrailerLine)
			return nil
		},
	}
	cmd.Flags().String("worker", "", "Worker/session ID (auto-detected from CLAUDE_CODE_SESSION_ID)")
	cmd.Flags().String("workspace", "", "Workspace path or identifier")
	cmd.Flags().Int("lease-hours", 4, "Lease duration in hours")
	return cmd
}

func workRenewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "renew <rmi-or-assignment-id>",
		Short: "Extend the lease on an active assignment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			leaseHours, _ := cmd.Flags().GetInt("lease-hours")
			lease := time.Duration(leaseHours) * time.Hour

			a, err := svc.RenewLeaseByRef(cmd.Context(), args[0], lease)
			if err != nil {
				return err
			}
			cmd.Printf("Renewed %s — new expiry: %s\n", a.ID, a.LeaseExpiresAt.Format("2006-01-02 15:04"))
			return nil
		},
	}
	cmd.Flags().Int("lease-hours", 4, "Lease extension in hours")
	return cmd
}

func workReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release <rmi-or-assignment-id>",
		Short: "Release a work claim (RMI returns to ready)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			handoffJSON, _ := cmd.Flags().GetString("handoff")
			var handoff *store.Handoff
			if handoffJSON != "" {
				handoff = &store.Handoff{}
				if err := json.Unmarshal([]byte(handoffJSON), handoff); err != nil {
					return fmt.Errorf("parse handoff JSON: %w", err)
				}
			}

			a, err := svc.ReleaseWorkByRef(cmd.Context(), args[0], handoff)
			if err != nil {
				return err
			}
			cmd.Printf("Released %s — RMI %s is now ready\n", a.ID, a.RMIID)
			return nil
		},
	}
	cmd.Flags().String("handoff", "", `Handoff state as JSON (e.g. '{"completed":["step1"],"remaining":["step2"],"next_action":"continue"}')`)
	return cmd
}

func workCompleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete <rmi-or-assignment-id> [more-ids...]",
		Short: "Mark work as completed (accepts RMI IDs or assignment IDs)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			handoffJSON, _ := cmd.Flags().GetString("handoff")
			var handoff *store.Handoff
			if handoffJSON != "" {
				handoff = &store.Handoff{}
				if err := json.Unmarshal([]byte(handoffJSON), handoff); err != nil {
					return fmt.Errorf("parse handoff JSON: %w", err)
				}
			}

			transition, _ := cmd.Flags().GetBool("transition")

			initiativesSeen := map[string]bool{}
			for _, ref := range args {
				a, err := svc.CompleteWorkByRef(cmd.Context(), ref, handoff, transition)
				if err != nil {
					return fmt.Errorf("%s: %w", ref, err)
				}
				cmd.Printf("Completed %s — RMI %s is done\n", a.ID, a.RMIID)
				if transition {
					cmd.Printf("Transitioned %s → completed\n", a.RMIID)
					rmi, err := svc.GetRMI(cmd.Context(), a.RMIID)
					if err == nil && rmi.InitiativeID != "" {
						initiativesSeen[rmi.InitiativeID] = true
					}
				}
			}

			if transition {
				for initID := range initiativesSeen {
					if svc.CheckInitiativeAllComplete(cmd.Context(), initID) {
						cmd.Printf("\n✓ All required RMIs in %s are now completed.\n", initID)
						cmd.Printf("  Transition to delivery_complete with:\n")
						cmd.Printf("    visionstudio initiative transition %s delivery_complete\n", initID)
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().String("handoff", "", "Final handoff state as JSON")
	cmd.Flags().Bool("transition", false, "Also transition the RMI to completed")
	return cmd
}

func workUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <rmi-or-assignment-id>",
		Short: "Update handoff state on an active assignment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			handoffJSON, _ := cmd.Flags().GetString("handoff")
			if handoffJSON == "" {
				return fmt.Errorf("--handoff is required")
			}

			handoff := &store.Handoff{}
			if err := json.Unmarshal([]byte(handoffJSON), handoff); err != nil {
				return fmt.Errorf("parse handoff JSON: %w", err)
			}

			a, err := svc.UpdateHandoffByRef(cmd.Context(), args[0], handoff)
			if err != nil {
				return err
			}
			cmd.Printf("Updated handoff on %s\n", a.ID)
			return nil
		},
	}
	cmd.Flags().String("handoff", "", `Handoff state as JSON (required)`)
	return cmd
}

func workClaimPhaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-phase <phase-id>",
		Short: "Claim all ready, unblocked RMIs in a phase",
		Long: `Claim every RMI in the given phase that is ready, unblocked, and unclaimed.
Phase ID format: INITIATIVE-ID/phase-N (e.g. INIT-PRISMCONTROL-001/phase-5).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			worker, _ := cmd.Flags().GetString("worker")
			workspace, _ := cmd.Flags().GetString("workspace")
			leaseHours, _ := cmd.Flags().GetInt("lease-hours")

			if worker == "" {
				worker = defaultWorker()
			}
			if worker == "" {
				return fmt.Errorf("--worker is required (or set CLAUDE_CODE_SESSION_ID)")
			}

			readyFirst, _ := cmd.Flags().GetBool("ready")
			if readyFirst {
				updated, _, err := svc.UpdatePhaseStatus(cmd.Context(), args[0], "", "ready")
				if err != nil {
					return fmt.Errorf("auto-transition to ready: %w", err)
				}
				if len(updated) > 0 {
					cmd.Printf("Transitioned %d RMIs to ready:\n", len(updated))
					for _, id := range updated {
						cmd.Printf("  %s\n", id)
					}
				}
			}

			lease := time.Duration(leaseHours) * time.Hour
			result, err := svc.ClaimPhase(cmd.Context(), args[0], worker, workspace, lease)
			if err != nil {
				return err
			}

			if len(result.Claimed) == 0 {
				cmd.Println("No claimable RMIs in phase.")
				if len(result.Blocked) > 0 {
					cmd.Printf("  Blocked by dependencies: %s\n", strings.Join(result.Blocked, ", "))
				}
				if len(result.AlreadyOwned) > 0 {
					cmd.Printf("  Already claimed: %s\n", strings.Join(result.AlreadyOwned, ", "))
				}
				return nil
			}

			cmd.Printf("Claimed %d RMIs:\n", len(result.Claimed))
			for _, r := range result.Claimed {
				cmd.Printf("  %s (assignment: %s)\n", r.Assignment.RMIID, r.Assignment.ID)
			}
			if len(result.Blocked) > 0 {
				cmd.Printf("Blocked by dependencies (%d): %s\n", len(result.Blocked), strings.Join(result.Blocked, ", "))
			}
			if len(result.AlreadyOwned) > 0 {
				cmd.Printf("Already claimed (%d): %s\n", len(result.AlreadyOwned), strings.Join(result.AlreadyOwned, ", "))
			}
			cmd.Printf("\nGit trailer for commits:\n")
			for _, r := range result.Claimed {
				cmd.Printf("  %s\n", r.TrailerLine)
			}
			return nil
		},
	}
	cmd.Flags().String("worker", "", "Worker/session ID (auto-detected from CLAUDE_CODE_SESSION_ID)")
	cmd.Flags().String("workspace", "", "Workspace path or identifier")
	cmd.Flags().Int("lease-hours", 4, "Lease duration in hours")
	cmd.Flags().Bool("ready", false, "Auto-transition proposed/planned RMIs to ready before claiming")
	return cmd
}

func workCompletePhaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete-phase <phase-id>",
		Short: "Complete all in-progress RMIs in a phase",
		Long: `Complete every in-progress RMI with an active assignment in the given phase.
Phase ID format: INITIATIVE-ID/phase-N (e.g. INIT-PRISMCONTROL-001/phase-5).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			handoffJSON, _ := cmd.Flags().GetString("handoff")
			var handoff *store.Handoff
			if handoffJSON != "" {
				handoff = &store.Handoff{}
				if err := json.Unmarshal([]byte(handoffJSON), handoff); err != nil {
					return fmt.Errorf("parse handoff JSON: %w", err)
				}
			}

			transition, _ := cmd.Flags().GetBool("transition")
			result, err := svc.CompletePhase(cmd.Context(), args[0], handoff, transition)
			if err != nil {
				return err
			}

			if len(result.Completed) == 0 {
				cmd.Println("No in-progress RMIs with active assignments in phase.")
				if len(result.NoAssignment) > 0 {
					cmd.Printf("  In-progress but no active assignment: %s\n", strings.Join(result.NoAssignment, ", "))
					cmd.Println("  These RMIs may have expired leases — reclaim with: visionstudio work claim <RMI-ID> --worker <session-id>")
				}
				return nil
			}

			cmd.Printf("Completed %d RMIs:\n", len(result.Completed))
			for _, a := range result.Completed {
				cmd.Printf("  %s — %s\n", a.RMIID, a.ID)
				if transition {
					cmd.Printf("  Transitioned %s → completed\n", a.RMIID)
				}
			}
			if len(result.NoAssignment) > 0 {
				cmd.Printf("Skipped (no active assignment): %s\n", strings.Join(result.NoAssignment, ", "))
			}
			if result.InitiativeAllComplete {
				cmd.Printf("\n✓ All required RMIs in %s are now completed.\n", result.InitiativeID)
				cmd.Printf("  Transition to delivery_complete with:\n")
				cmd.Printf("    visionstudio initiative transition %s delivery_complete\n", result.InitiativeID)
			}
			return nil
		},
	}
	cmd.Flags().String("handoff", "", "Final handoff state as JSON")
	cmd.Flags().Bool("transition", false, "Also transition each RMI to completed")
	return cmd
}

func workStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "List all active assignments",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			active, err := svc.ListActiveAssignments(cmd.Context())
			if err != nil {
				return err
			}

			if len(active) == 0 {
				cmd.Println("No active assignments.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ASSIGNMENT\tRMI\tWORKER\tEXPIRES\tWORKSPACE")
			for _, a := range active {
				ws := a.Workspace
				if ws == "" {
					ws = "-"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					a.ID, a.RMIID, a.Worker, a.LeaseExpiresAt.Format("2006-01-02 15:04"), ws)
			}
			return w.Flush()
		},
	}
}
