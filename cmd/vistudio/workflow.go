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
		Short:   "Manage spec workflows",
	}
	cmd.AddCommand(workflowListCmd(), workflowGetCmd(), workflowSeedCmd())
	return cmd
}

func workflowListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all spec workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			workflows, err := svc.Store.ListSpecWorkflows(cmd.Context())
			if err != nil {
				return err
			}

			if len(workflows) == 0 {
				cmd.Println("No workflows found. Run 'vistudio workflow seed' to create built-in workflows.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tNAME\tREQUIRED\tOPTIONAL\tTYPES")
			for _, wf := range workflows {
				required := strings.Join(wf.SpecsRequired, ", ")
				optional := strings.Join(wf.SpecsOptional, ", ")
				if optional == "" {
					optional = "-"
				}
				types := strings.Join(wf.InitTypes, ", ")
				if types == "" {
					types = "(opt-in)"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", wf.ID, wf.Name, required, optional, types)
			}
			return w.Flush()
		},
	}
}

func workflowGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <workflow-id>",
		Short: "Show workflow details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			wf, err := svc.Store.GetSpecWorkflow(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			cmd.Printf("ID:          %s\n", wf.ID)
			cmd.Printf("Name:        %s\n", wf.Name)
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
			return nil
		},
	}
}

func workflowSeedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Create built-in spec workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			created, err := specworkflow.SeedBuiltIn(cmd.Context(), svc.Store)
			if err != nil {
				return err
			}

			if created == 0 {
				cmd.Println("All built-in workflows already exist.")
			} else {
				cmd.Printf("Created %d workflow(s).\n", created)
			}
			return nil
		},
	}
}
