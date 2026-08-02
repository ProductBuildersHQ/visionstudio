package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func programCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "program",
		Aliases: []string{"prog"},
		Short:   "Manage programs",
	}
	cmd.AddCommand(
		programCreateCmd(),
		programListCmd(),
		programGetCmd(),
		programUpdateCmd(),
		programHideCmd(),
		programShowCmd(),
	)
	return cmd
}

// setProgramHidden loads a program, sets its hidden flag, and persists it.
func setProgramHidden(cmd *cobra.Command, id string, hidden bool) error {
	svc, cleanup, err := connectService(cmd)
	if err != nil {
		return err
	}
	defer cleanup()

	prog, err := svc.Store.GetProgram(cmd.Context(), id)
	if err != nil {
		return err
	}
	if prog.Hidden == hidden {
		state := "shown"
		if hidden {
			state = "hidden"
		}
		cmd.Printf("Program %s is already %s\n", prog.ID, state)
		return nil
	}
	prog.Hidden = hidden
	prog.UpdatedAt = time.Now()
	if err := svc.UpdateProgram(cmd.Context(), prog); err != nil {
		return err
	}
	verb := "shown on"
	if hidden {
		verb = "hidden from"
	}
	cmd.Printf("Program %s %s the dashboard homepage\n", prog.ID, verb)
	return nil
}

func programHideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hide <program-id>",
		Short: "Hide a program from the dashboard homepage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setProgramHidden(cmd, args[0], true)
		},
	}
}

func programShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <program-id>",
		Short: "Show a previously hidden program on the dashboard homepage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setProgramHidden(cmd, args[0], false)
		},
	}
}

func programCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new program",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")
			org, _ := cmd.Flags().GetString("org")
			desc, _ := cmd.Flags().GetString("description")

			if id == "" || name == "" {
				return fmt.Errorf("--id and --name are required")
			}
			if org == "" {
				org = "default"
			}

			prog, err := svc.CreateProgram(cmd.Context(), id, name, org, desc)
			if err != nil {
				return err
			}
			cmd.Printf("Created program: %s (%s)\n", prog.ID, prog.Name)
			return nil
		},
	}
	cmd.Flags().String("id", "", "Program ID (e.g. PROG-DELIVERY) (required)")
	cmd.Flags().String("name", "", "Program display name (required)")
	cmd.Flags().String("org", "", "Organization (default: 'default')")
	cmd.Flags().String("description", "", "Description")
	return cmd
}

func programListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all programs",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			progs, err := svc.ListPrograms(cmd.Context())
			if err != nil {
				return err
			}
			if len(progs) == 0 {
				cmd.Println("No programs found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tNAME\tORG\tHIDDEN\tDESCRIPTION")
			for _, p := range progs {
				desc := p.Description
				if desc == "" {
					desc = "-"
				}
				hidden := "no"
				if p.Hidden {
					hidden = "yes"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", p.ID, p.Name, p.Organization, hidden, desc)
			}
			return w.Flush()
		},
	}
}

func programGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <program-id>",
		Short: "Show program details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			prog, err := svc.GetProgram(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			cmd.Printf("Program:      %s\n", prog.ID)
			cmd.Printf("Name:         %s\n", prog.Name)
			cmd.Printf("Organization: %s\n", prog.Organization)
			if prog.Description != "" {
				cmd.Printf("Description:  %s\n", prog.Description)
			}
			cmd.Printf("Hidden:       %v\n", prog.Hidden)
			cmd.Printf("Created:      %s\n", prog.CreatedAt.Format("2006-01-02 15:04"))
			return nil
		},
	}
}

func programUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <program-id>",
		Short: "Update program fields",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			prog, err := svc.Store.GetProgram(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("name") {
				prog.Name, _ = cmd.Flags().GetString("name")
			}
			if cmd.Flags().Changed("description") {
				prog.Description, _ = cmd.Flags().GetString("description")
			}
			if err := svc.UpdateProgram(cmd.Context(), prog); err != nil {
				return err
			}
			cmd.Printf("Updated %s\n", prog.ID)
			return nil
		},
	}
	cmd.Flags().String("name", "", "Program display name")
	cmd.Flags().String("description", "", "Description")
	return cmd
}
