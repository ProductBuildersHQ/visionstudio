package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/specworkflow"
)

func workflowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workflow",
		Aliases: []string{"wf"},
		Short:   "Manage spec workflows (definitions from visionspec)",
	}
	cmd.AddCommand(workflowListCmd(), workflowGetCmd(), workflowSyncCmd(), workflowSeedCmd())
	return cmd
}

func workflowListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all spec workflows from the catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			loader := specworkflow.DefaultLoader()
			infos, err := loader.List()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tREQUIRED\tOPTIONAL\tTYPES")
			for _, info := range infos {
				required := strings.Join(info.SpecsRequired, ", ")
				optional := strings.Join(info.SpecsOptional, ", ")
				if optional == "" {
					optional = "-"
				}
				types := strings.Join(specworkflow.DefaultInitTypes(info.ID), ", ")
				if types == "" {
					types = "(opt-in)"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", info.ID, required, optional, types)
			}
			return w.Flush()
		},
	}
}

func workflowGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <workflow-id>",
		Short: "Show workflow details from the catalog",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			loader := specworkflow.DefaultLoader()
			lw, err := loader.Load(args[0])
			if err != nil {
				return err
			}
			wf := specworkflow.StoreWorkflow(args[0], lw)

			cmd.Printf("ID:          %s\n", wf.ID)
			cmd.Printf("Name:        %s\n", wf.Name)
			if lw.Workflow.Extends != "" {
				cmd.Printf("Extends:     %s\n", lw.Workflow.Extends)
			}
			if wf.Description != "" {
				cmd.Printf("Description: %s\n", wf.Description)
			}
			cmd.Printf("Required:    %s\n", strings.Join(wf.SpecsRequired, ", "))
			if len(wf.SpecsOptional) > 0 {
				cmd.Printf("Optional:    %s\n", strings.Join(wf.SpecsOptional, ", "))
			}
			if len(wf.InitTypes) > 0 {
				cmd.Printf("Types:       %s\n", strings.Join(wf.InitTypes, ", "))
			} else {
				cmd.Printf("Types:       (opt-in only)\n")
			}
			if exec := lw.Workflow.Execution; exec != nil {
				if len(exec.Sequence) > 0 {
					cmd.Printf("Sequence:    %s\n", strings.Join(exec.Sequence, " → "))
				}
				for _, p := range exec.Phases {
					cmd.Printf("Phase:       %s (%s)\n", p.Name, strings.Join(p.Specs, ", "))
				}
			}
			return nil
		},
	}
}

func workflowSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync the database workflow index from the visionspec catalog",
		Long: `Upsert a database row for every workflow in the catalog, remap initiatives
referencing retired workflow IDs to their canonical replacement, and delete
retired rows that are no longer referenced. Idempotent.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			res, err := specworkflow.SyncFromCatalog(cmd.Context(), svc.Store, specworkflow.DefaultLoader())
			if err != nil {
				return err
			}

			cmd.Printf("Synced: %d created, %d updated\n", res.Created, res.Updated)
			for initID, wfID := range res.Remapped {
				cmd.Printf("Remapped: %s → %s\n", initID, wfID)
			}
			for _, id := range res.Deleted {
				cmd.Printf("Deleted retired workflow: %s\n", id)
			}
			for _, id := range res.Retained {
				cmd.Printf("Retained (not in catalog, still referenced): %s\n", id)
			}
			return nil
		},
	}
}

func workflowSeedCmd() *cobra.Command {
	cmd := workflowSyncCmd()
	cmd.Use = "seed"
	cmd.Short = "Deprecated: use 'workflow sync'"
	cmd.Deprecated = "use 'workflow sync' instead"
	return cmd
}
