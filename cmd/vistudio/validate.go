package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/validate"
)

func validateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Run consistency checks on the store",
		Long: `Check for: trailer↔RMI consistency, dangling/cyclic dependencies,
expired leases, status coherence, duplicate IDs, and ID-format violations.

Exit code 1 if any errors are found; warnings alone exit 0.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			format, _ := cmd.Flags().GetString("format")

			result, err := validate.Run(cmd.Context(), svc.Store, time.Now())
			if err != nil {
				return err
			}

			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			case "text":
				if len(result.Findings) == 0 {
					fmt.Println("No issues found.")
					return nil
				}
				for _, f := range result.Findings {
					fmt.Printf("[%s] %s: %s\n", f.Level, f.Check, f.Message)
				}
			default:
				return fmt.Errorf("unknown format: %s (use json or text)", format)
			}

			if !result.OK() {
				os.Exit(1)
			}
			return nil
		},
	}
	cmd.Flags().String("format", "text", "Output format: text or json")
	return cmd
}
