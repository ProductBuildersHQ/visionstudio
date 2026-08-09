package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/release"
)

func releaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Release planning and management",
	}
	cmd.AddCommand(releasePlanCmd())
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
