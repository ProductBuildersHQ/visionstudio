package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/contextbuild"
)

func contextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Assemble deterministic context packages",
		Long: `Build context packages for agent sessions from the PRISM execution graph.
Context packages contain program/initiative/phase/RMI data, spec references,
and assignment state, ordered stable→volatile for cache efficiency.`,
	}
	cmd.AddCommand(contextBuildCmd())
	return cmd
}

func contextBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build <rmi-id|phase-id>",
		Short: "Build a context package for an RMI or phase",
		Long: `Assemble a deterministic context package for a phase or RMI.

For a phase (e.g., INIT-FOO-001/phase-1):
  - Program and initiative context (stable)
  - Phase objectives and member RMIs (phase_stable)
  - Prerequisite phase handoffs (phase_stable)
  - Spec file references with git revisions (rmi_stable)
  - Derived repository set

For an RMI (e.g., RMI-REPO-001):
  - All of the above, plus:
  - Current RMI details with acceptance criteria (rmi_stable)
  - Active assignment state and evidence (volatile)

Output is byte-identical across runs at the same Dolt/git revisions.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			targetID := args[0]
			format, _ := cmd.Flags().GetString("format")
			outFile, _ := cmd.Flags().GetString("out")

			doltCommit := "unknown"

			builder := contextbuild.NewBuilder(svc.Store, doltCommit)

			var output []byte

			if isPhaseID(targetID) {
				pkg, err := builder.BuildForPhase(cmd.Context(), targetID)
				if err != nil {
					return fmt.Errorf("build phase context: %w", err)
				}
				output, err = renderContextPackage(pkg, format)
				if err != nil {
					return err
				}
			} else {
				pkg, err := builder.BuildForRMI(cmd.Context(), targetID)
				if err != nil {
					return fmt.Errorf("build RMI context: %w", err)
				}
				output, err = renderContextPackage(pkg, format)
				if err != nil {
					return err
				}
			}

			if outFile != "" {
				if err := os.WriteFile(outFile, output, 0o600); err != nil {
					return fmt.Errorf("write output: %w", err)
				}
				cmd.Printf("Wrote context package to %s\n", outFile)
				return nil
			}

			fmt.Print(string(output))
			return nil
		},
	}
	cmd.Flags().StringP("format", "f", "json", "Output format: json or markdown")
	cmd.Flags().StringP("out", "o", "", "Write output to file instead of stdout")
	return cmd
}

func isPhaseID(id string) bool {
	return strings.Contains(id, "/phase-")
}

func renderContextPackage(pkg *contextbuild.ContextPackage, format string) ([]byte, error) {
	switch format {
	case "json":
		return pkg.RenderJSON()
	case "markdown":
		return []byte(pkg.RenderMarkdown()), nil
	default:
		return nil, fmt.Errorf("unknown format: %s (use json or markdown)", format)
	}
}
