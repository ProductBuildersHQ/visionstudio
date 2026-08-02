package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
	"github.com/spf13/cobra"
)

func parseSpecs(flags []string) map[string]string {
	if len(flags) == 0 {
		return nil
	}
	specs := make(map[string]string, len(flags))
	for _, f := range flags {
		k, v, ok := strings.Cut(f, "=")
		if ok {
			specs[k] = v
		}
	}
	return specs
}

func initiativeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "initiative",
		Aliases: []string{"init"},
		Short:   "Manage initiatives",
	}
	cmd.AddCommand(
		initiativeCreateCmd(),
		initiativeListCmd(),
		initiativeGetCmd(),
		initiativeUpdateCmd(),
		initiativeTransitionCmd(),
		initiativeDepCmd(),
	)
	return cmd
}

func initiativeCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new initiative",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			id, _ := cmd.Flags().GetString("id")
			org, _ := cmd.Flags().GetString("org")
			title, _ := cmd.Flags().GetString("title")
			desc, _ := cmd.Flags().GetString("description")
			priority, _ := cmd.Flags().GetString("priority")
			initType, _ := cmd.Flags().GetString("type")
			homeRepo, _ := cmd.Flags().GetString("home-repo")
			workspace, _ := cmd.Flags().GetString("workspace")
			program, _ := cmd.Flags().GetString("program")
			specFlags, _ := cmd.Flags().GetStringSlice("spec")

			if id == "" || title == "" {
				return fmt.Errorf("--id and --title are required")
			}
			if org == "" {
				org = "default"
			}

			specs := parseSpecs(specFlags)

			init, err := svc.CreateInitiative(cmd.Context(), id, org, title, desc, priority, initType)
			if err != nil {
				return err
			}

			if homeRepo != "" || workspace != "" || program != "" || len(specs) > 0 {
				init.HomeRepo = homeRepo
				init.Workspace = workspace
				init.ProgramID = program
				init.Specs = specs
				if err := svc.UpdateInitiative(cmd.Context(), init); err != nil {
					return err
				}
			}

			cmd.Printf("Created initiative: %s (%s)\n", init.ID, init.Status)
			return nil
		},
	}
	cmd.Flags().String("id", "", "Initiative ID (e.g. INIT-MYPROJECT-001) (required)")
	cmd.Flags().String("org", "", "Organization (default: 'default')")
	cmd.Flags().String("title", "", "Initiative title (required)")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("priority", "", "Priority (high, medium, low)")
	cmd.Flags().String("type", "", "Initiative type (feature, maintenance, migration, compliance, refactor); default: feature")
	cmd.Flags().String("home-repo", "", "Home repository ID (where specs live)")
	cmd.Flags().String("workspace", "", "Workspace identifier (e.g. tmux session name)")
	cmd.Flags().String("program", "", "Program ID (e.g. PROG-DELIVERY)")
	cmd.Flags().StringSlice("spec", nil, "Spec reference as key=path (repeatable, e.g. --spec prd=docs/specs/PRD.md)")
	return cmd
}

func initiativeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all initiatives",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			inits, err := svc.ListInitiatives(cmd.Context())
			if err != nil {
				return err
			}

			if len(inits) == 0 {
				cmd.Println("No initiatives found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tPROGRAM\tWORKSPACE")
			for _, i := range inits {
				ws := i.Workspace
				if ws == "" {
					ws = "-"
				}
				prog := i.ProgramID
				if prog == "" {
					prog = "-"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", i.ID, i.Title, i.Status, prog, ws)
			}
			return w.Flush()
		},
	}
}

func initiativeGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <initiative-id>",
		Short: "Show initiative details with phase status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			detail, err := svc.GetInitiativeDetail(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			init := detail.Initiative
			cmd.Printf("Initiative: %s\n", init.ID)
			cmd.Printf("Title:      %s\n", init.Title)
			cmd.Printf("Status:     %s\n", init.Status)
			if init.Description != "" {
				cmd.Printf("Desc:       %s\n", init.Description)
			}
			if init.Priority != "" {
				cmd.Printf("Priority:   %s\n", init.Priority)
			}
			if init.HomeRepo != "" {
				cmd.Printf("Home repo:  %s\n", init.HomeRepo)
			}
			if init.Workspace != "" {
				cmd.Printf("Workspace:  %s\n", init.Workspace)
			}
			if init.ProgramID != "" {
				cmd.Printf("Program:    %s\n", init.ProgramID)
			}
			if len(init.Specs) > 0 {
				cmd.Println("Specs:")
				for k, v := range init.Specs {
					cmd.Printf("  %s: %s\n", k, v)
				}
			}
			cmd.Printf("Created:    %s\n", init.CreatedAt.Format("2006-01-02 15:04"))
			if init.PlannedAt != nil {
				cmd.Printf("Planned:    %s\n", init.PlannedAt.Format("2006-01-02 15:04"))
			}
			if init.ExecutingAt != nil {
				cmd.Printf("Executing:  %s\n", init.ExecutingAt.Format("2006-01-02 15:04"))
			}
			if init.DeliveryCompleteAt != nil {
				cmd.Printf("Delivered:  %s\n", init.DeliveryCompleteAt.Format("2006-01-02 15:04"))
			}
			if init.ReleasedAt != nil {
				cmd.Printf("Released:   %s\n", init.ReleasedAt.Format("2006-01-02 15:04"))
			}
			if init.ClosedAt != nil {
				cmd.Printf("Closed:     %s\n", init.ClosedAt.Format("2006-01-02 15:04"))
			}

			if len(detail.Phases) > 0 {
				var totalRMIs, completedRMIs, requiredTotal, requiredCompleted int
				for _, p := range detail.Phases {
					for _, r := range p.RMIs {
						totalRMIs++
						if r.Status == "completed" {
							completedRMIs++
						}
						if r.Required {
							requiredTotal++
							if r.Status == "completed" {
								requiredCompleted++
							}
						}
					}
				}
				cmd.Printf("\nProgress:   %d/%d RMIs completed", completedRMIs, totalRMIs)
				if requiredTotal != totalRMIs {
					cmd.Printf(" (%d/%d required)", requiredCompleted, requiredTotal)
				}
				cmd.Println()

				cmd.Println("\nPhases:")
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				_, _ = fmt.Fprintln(w, "  #\tTITLE\tTHEME\tSTATUS\tRMIs")
				for _, p := range detail.Phases {
					theme := p.Phase.Theme
					if theme == "" {
						theme = "-"
					}
					_, _ = fmt.Fprintf(w, "  %d\t%s\t%s\t%s\t%d\n",
						p.Phase.SequenceNumber, p.Phase.Title, theme, p.Status, len(p.RMIs))
				}
				if err := w.Flush(); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func initiativeUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <initiative-id>",
		Short: "Update initiative fields (workspace, home-repo, description, priority)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			init, err := svc.Store.GetInitiative(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("workspace") {
				init.Workspace, _ = cmd.Flags().GetString("workspace")
			}
			if cmd.Flags().Changed("home-repo") {
				init.HomeRepo, _ = cmd.Flags().GetString("home-repo")
			}
			if cmd.Flags().Changed("description") {
				init.Description, _ = cmd.Flags().GetString("description")
			}
			if cmd.Flags().Changed("priority") {
				init.Priority, _ = cmd.Flags().GetString("priority")
			}
			if cmd.Flags().Changed("program") {
				init.ProgramID, _ = cmd.Flags().GetString("program")
			}

			if err := svc.UpdateInitiative(cmd.Context(), init); err != nil {
				return err
			}

			cmd.Printf("Updated %s\n", init.ID)
			if init.Workspace != "" {
				cmd.Printf("Workspace: %s\n", init.Workspace)
			}
			return nil
		},
	}
	cmd.Flags().String("workspace", "", "Workspace identifier (e.g. tmux session name)")
	cmd.Flags().String("home-repo", "", "Home repository ID")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("priority", "", "Priority (high, medium, low)")
	cmd.Flags().String("program", "", "Program ID (e.g. PROG-DELIVERY)")
	return cmd
}

func initiativeTransitionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "transition <initiative-id> <status>",
		Short: "Transition initiative to a new lifecycle status",
		Long: `Valid transitions:
  proposed -> planned -> executing -> delivery_complete -> releasing -> released -> closed
  Any active status -> cancelled`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			init, err := svc.TransitionInitiative(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			cmd.Printf("Transitioned %s to %s\n", init.ID, init.Status)
			return nil
		},
	}
}

func initiativeDepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dep",
		Aliases: []string{"dependency"},
		Short:   "Manage initiative dependencies",
	}
	cmd.AddCommand(initiativeDepAddCmd(), initiativeDepListCmd())
	return cmd
}

func initiativeDepAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a dependency between two initiatives",
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
			if rel == "" {
				rel = "requires"
			}

			dep := &store.InitiativeDependency{
				SourceInitiativeID: source,
				TargetInitiativeID: target,
				Relationship:       rel,
			}
			if err := svc.Store.CreateInitiativeDependency(cmd.Context(), dep); err != nil {
				return err
			}
			cmd.Printf("Added dependency: %s --%s--> %s\n", source, rel, target)
			return nil
		},
	}
	cmd.Flags().String("source", "", "Source initiative ID (required)")
	cmd.Flags().String("target", "", "Target initiative ID (required)")
	cmd.Flags().String("relationship", "requires", "Relationship type (requires, relates)")
	return cmd
}

func initiativeDepListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [initiative-id]",
		Short: "List initiative dependencies (all or for a specific initiative)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			var deps []*store.InitiativeDependency
			if len(args) == 1 {
				deps, err = svc.Store.ListInitiativeDependencies(cmd.Context(), args[0])
			} else {
				deps, err = svc.Store.ListAllInitiativeDependencies(cmd.Context())
			}
			if err != nil {
				return err
			}

			if len(deps) == 0 {
				cmd.Println("No initiative dependencies found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "SOURCE\tRELATIONSHIP\tTARGET")
			for _, d := range deps {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", d.SourceInitiativeID, d.Relationship, d.TargetInitiativeID)
			}
			return w.Flush()
		},
	}
}
