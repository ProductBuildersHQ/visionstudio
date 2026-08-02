package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/report"
	"github.com/ProductBuildersHQ/visionstudio/pkg/tokens"
)

func reportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate reports",
	}
	cmd.AddCommand(reportInitiativeCmd())
	cmd.AddCommand(reportTokensCmd())
	return cmd
}

func reportInitiativeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initiative <id>",
		Short: "Generate a full initiative report",
		Long: `Compute an end-to-end initiative report: duration, phases, RMI progress,
commit distribution by type and repo, releases, and unattributed residual.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			r, err := report.Generate(cmd.Context(), svc.Store, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(r)
			case "markdown":
				fmt.Print(r.Markdown())
				return nil
			default:
				return fmt.Errorf("unknown format: %s (use json or markdown)", format)
			}
		},
	}
	cmd.Flags().String("format", "json", "Output format: json or markdown")
	return cmd
}

func reportTokensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "Generate token attribution report",
		Long: `Generate a token attribution report mapping spend to initiatives and RMIs.

Modes:
  --initiative <id>   Report for a specific initiative
  --quarter 2026-Q3   Report for a calendar quarter
  --since/--until     Report for a custom date range

Examples:
  vistudio report tokens --initiative INIT-PRISM-001
  vistudio report tokens --quarter 2026-Q3
  vistudio report tokens --since 2026-07-01 --until 2026-07-31`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			// Create token source
			dataDir, _ := cmd.Flags().GetString("data-dir")
			source, err := tokens.NewJSONLSource(dataDir)
			if err != nil {
				return fmt.Errorf("create token source: %w", err)
			}

			var r *report.TokenReport

			initiativeID, _ := cmd.Flags().GetString("initiative")
			quarter, _ := cmd.Flags().GetString("quarter")
			since, _ := cmd.Flags().GetString("since")
			until, _ := cmd.Flags().GetString("until")

			switch {
			case initiativeID != "":
				// Initiative mode
				r, err = report.GenerateInitiativeTokenReport(
					cmd.Context(), svc.Store, source, initiativeID)
				if err != nil {
					return err
				}

			case quarter != "":
				// Quarter mode
				period, err := report.ParseQuarter(quarter)
				if err != nil {
					return err
				}
				r, err = report.GeneratePeriodTokenReport(
					cmd.Context(), svc.Store, source, period)
				if err != nil {
					return err
				}

			case since != "" || until != "":
				// Date range mode
				if since == "" || until == "" {
					return fmt.Errorf("both --since and --until are required for date range mode")
				}
				start, err := time.Parse("2006-01-02", since)
				if err != nil {
					return fmt.Errorf("invalid --since date: %w", err)
				}
				end, err := time.Parse("2006-01-02", until)
				if err != nil {
					return fmt.Errorf("invalid --until date: %w", err)
				}
				// End of day
				end = end.Add(24*time.Hour - time.Nanosecond)

				period := report.TokenPeriod{
					Start: start,
					End:   end,
					Label: fmt.Sprintf("%s to %s", since, until),
				}
				r, err = report.GeneratePeriodTokenReport(
					cmd.Context(), svc.Store, source, period)
				if err != nil {
					return err
				}

			default:
				return fmt.Errorf("specify --initiative, --quarter, or --since/--until")
			}

			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(r)
			case "markdown":
				fmt.Print(r.MarkdownTokenReport())
				return nil
			default:
				return fmt.Errorf("unknown format: %s (use json or markdown)", format)
			}
		},
	}

	cmd.Flags().String("initiative", "", "Initiative ID for initiative-scoped report")
	cmd.Flags().String("quarter", "", "Quarter (e.g., 2026-Q3) for period report")
	cmd.Flags().String("since", "", "Start date (YYYY-MM-DD) for custom period")
	cmd.Flags().String("until", "", "End date (YYYY-MM-DD) for custom period")
	cmd.Flags().String("format", "json", "Output format: json or markdown")
	cmd.Flags().String("data-dir", "", "omnidevx data directory (default: ~/.plexusone/omnidevx/data)")

	return cmd
}
