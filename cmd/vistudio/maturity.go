package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/maturity"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func maturityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "maturity",
		Short: "Manage capability maturity models and assessments",
	}
	cmd.AddCommand(
		maturityModelCmd(),
		maturityAssessCmd(),
		maturityReportCmd(),
	)
	return cmd
}

// ---------------------------------------------------------------------------
// maturity model subcommands
// ---------------------------------------------------------------------------

func maturityModelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Manage capability models",
	}
	cmd.AddCommand(
		maturityModelListCmd(),
		maturityModelGetCmd(),
		maturityModelSeedCmd(),
	)
	return cmd
}

func maturityModelListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all capability models",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			models, err := svc.Store.ListCapabilityModels(cmd.Context())
			if err != nil {
				return err
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tNAME\tDIMENSIONS\tMAX LEVEL")
			for _, m := range models {
				fmt.Fprintf(tw, "%s\t%s\t%d\t%d\n", m.ID, m.Name, len(m.Dimensions), m.MaxLevel)
			}
			return tw.Flush()
		},
	}
}

func maturityModelGetCmd() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "get <model-id>",
		Short: "Show capability model details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			model, err := svc.Store.GetCapabilityModel(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outputJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(model)
			}

			fmt.Printf("ID:          %s\n", model.ID)
			fmt.Printf("Name:        %s\n", model.Name)
			fmt.Printf("Description: %s\n", model.Description)
			fmt.Printf("Max Level:   %d\n", model.MaxLevel)
			fmt.Printf("Dimensions:  %d\n\n", len(model.Dimensions))

			for _, dim := range model.Dimensions {
				sources := ""
				if len(dim.Sources) > 0 {
					sources = " (" + strings.Join(dim.Sources, ", ") + ")"
				}
				fmt.Printf("  %s - %s%s\n", dim.Key, dim.Name, sources)
				if dim.Description != "" {
					fmt.Printf("    %s\n", dim.Description)
				}
				for _, lvl := range dim.Levels {
					fmt.Printf("    L%d: %s - %s\n", lvl.Level, lvl.Name, lvl.Description)
				}
				fmt.Println()
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func maturityModelSeedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Seed built-in capability models",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			created, err := maturity.SeedBuiltIn(cmd.Context(), svc.Store)
			if err != nil {
				return err
			}
			fmt.Printf("Seeded %d capability model(s)\n", created)
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// maturity assess subcommands
// ---------------------------------------------------------------------------

func maturityAssessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assess",
		Short: "Manage maturity assessments",
	}
	cmd.AddCommand(
		maturityAssessShowCmd(),
		maturityAssessRecordCmd(),
		maturityAssessListCmd(),
	)
	return cmd
}

func maturityAssessShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <initiative-id> <model-id>",
		Short: "Show capability model dimensions for agent-driven assessment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			ctx := cmd.Context()
			initID := args[0]
			modelID := args[1]

			init, err := svc.Store.GetInitiative(ctx, initID)
			if err != nil {
				return fmt.Errorf("initiative %s: %w", initID, err)
			}

			model, err := svc.Store.GetCapabilityModel(ctx, modelID)
			if err != nil {
				return fmt.Errorf("capability model %s: %w", modelID, err)
			}

			fmt.Println("=== MATURITY ASSESSMENT ===")
			fmt.Println()
			fmt.Printf("Initiative:  %s\n", init.ID)
			fmt.Printf("Title:       %s\n", init.Title)
			fmt.Printf("Model:       %s (%s)\n", model.Name, model.ID)
			fmt.Printf("Max Level:   %d\n", model.MaxLevel)
			fmt.Println()
			fmt.Println("DIMENSIONS TO ASSESS:")
			fmt.Println()

			for i, dim := range model.Dimensions {
				sources := ""
				if len(dim.Sources) > 0 {
					sources = " [" + strings.Join(dim.Sources, ", ") + "]"
				}
				fmt.Printf("%d. %s (%s)%s\n", i+1, dim.Name, dim.Key, sources)
				if dim.Description != "" {
					fmt.Printf("   %s\n", dim.Description)
				}
				fmt.Println("   Levels:")
				for _, lvl := range dim.Levels {
					fmt.Printf("   %d - %s: %s\n", lvl.Level, lvl.Name, lvl.Description)
				}
				fmt.Println()
			}

			fmt.Println("---")
			fmt.Println("After assessment, use 'vistudio maturity assess record' to persist results.")
			return nil
		},
	}
}

func maturityAssessRecordCmd() *cobra.Command {
	var (
		scoresJSON   string
		overallScore float64
		summary      string
		assessedBy   string
		modelUsed    string
	)
	cmd := &cobra.Command{
		Use:   "record <initiative-id> <model-id>",
		Short: "Record a maturity assessment result",
		Long: `Record a maturity assessment result for an initiative.

Scores must be provided as JSON mapping dimension keys to scores:
  --scores '{"customer-obsession": {"level": 3, "rationale": "..."}, ...}'

Example:
  vistudio maturity assess record INIT-FOO-001 big-tech-essentials \
    --scores '{"customer-obsession": {"level": 3, "rationale": "Regular user research"}}' \
    --overall 3.2 \
    --summary "Good customer focus, needs work on API design" \
    --assessed-by "Claude Code" \
    --model "claude-opus-4.5"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			ctx := cmd.Context()
			initID := args[0]
			modelID := args[1]

			if _, err := svc.Store.GetInitiative(ctx, initID); err != nil {
				return fmt.Errorf("initiative %s: %w", initID, err)
			}
			if _, err := svc.Store.GetCapabilityModel(ctx, modelID); err != nil {
				return fmt.Errorf("capability model %s: %w", modelID, err)
			}

			var scores map[string]store.DimensionScore
			if err := json.Unmarshal([]byte(scoresJSON), &scores); err != nil {
				return fmt.Errorf("parse scores JSON: %w", err)
			}

			assessment := &store.MaturityAssessment{
				ID:           fmt.Sprintf("MA-%s-%d", initID, time.Now().Unix()),
				ModelID:      modelID,
				InitiativeID: initID,
				Scores:       scores,
				Summary:      summary,
				AssessedBy:   assessedBy,
				Model:        modelUsed,
				AssessedAt:   time.Now(),
			}
			if overallScore > 0 {
				assessment.OverallScore = &overallScore
			}

			if err := svc.Store.CreateMaturityAssessment(ctx, assessment); err != nil {
				return fmt.Errorf("create assessment: %w", err)
			}

			fmt.Printf("Recorded maturity assessment %s\n", assessment.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&scoresJSON, "scores", "{}", "JSON object mapping dimension keys to scores")
	cmd.Flags().Float64Var(&overallScore, "overall", 0, "Overall maturity score")
	cmd.Flags().StringVar(&summary, "summary", "", "Summary of the assessment")
	cmd.Flags().StringVar(&assessedBy, "assessed-by", "", "Who performed the assessment")
	cmd.Flags().StringVar(&modelUsed, "model", "", "LLM model used for assessment")
	return cmd
}

func maturityAssessListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <initiative-id>",
		Short: "List maturity assessments for an initiative",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			assessments, err := svc.Store.ListMaturityAssessments(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tMODEL\tOVERALL\tASSESSED BY\tDATE")
			for _, a := range assessments {
				overall := "-"
				if a.OverallScore != nil {
					overall = fmt.Sprintf("%.1f", *a.OverallScore)
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					a.ID, a.ModelID, overall, a.AssessedBy,
					a.AssessedAt.Format("2006-01-02"))
			}
			return tw.Flush()
		},
	}
}

// ---------------------------------------------------------------------------
// maturity report
// ---------------------------------------------------------------------------

func maturityReportCmd() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "report <initiative-id>",
		Short: "Generate a maturity report for an initiative",
		Long: `Generate a maturity report showing the latest assessment for an initiative.

The report shows dimension-by-dimension scores and the overall maturity level.
Use --json for machine-readable output suitable for visualization.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			ctx := cmd.Context()
			initID := args[0]

			init, err := svc.Store.GetInitiative(ctx, initID)
			if err != nil {
				return fmt.Errorf("initiative %s: %w", initID, err)
			}

			assessments, err := svc.Store.ListMaturityAssessments(ctx, initID)
			if err != nil {
				return err
			}
			if len(assessments) == 0 {
				return fmt.Errorf("no maturity assessments found for %s", initID)
			}

			var latest *store.MaturityAssessment
			for _, a := range assessments {
				if latest == nil || a.AssessedAt.After(latest.AssessedAt) {
					latest = a
				}
			}

			model, err := svc.Store.GetCapabilityModel(ctx, latest.ModelID)
			if err != nil {
				return fmt.Errorf("capability model %s: %w", latest.ModelID, err)
			}

			if outputJSON {
				report := map[string]any{
					"initiative_id": init.ID,
					"initiative":    init.Title,
					"model_id":      model.ID,
					"model_name":    model.Name,
					"assessed_at":   latest.AssessedAt,
					"assessed_by":   latest.AssessedBy,
					"overall_score": latest.OverallScore,
					"summary":       latest.Summary,
					"dimensions":    []map[string]any{},
				}
				dims := []map[string]any{}
				for _, dim := range model.Dimensions {
					score := latest.Scores[dim.Key]
					dims = append(dims, map[string]any{
						"key":       dim.Key,
						"name":      dim.Name,
						"level":     score.Level,
						"max_level": model.MaxLevel,
						"rationale": score.Rationale,
						"evidence":  score.Evidence,
					})
				}
				report["dimensions"] = dims
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			}

			fmt.Println("=== MATURITY REPORT ===")
			fmt.Println()
			fmt.Printf("Initiative:  %s (%s)\n", init.Title, init.ID)
			fmt.Printf("Model:       %s\n", model.Name)
			fmt.Printf("Assessed:    %s by %s\n", latest.AssessedAt.Format("2006-01-02 15:04"), latest.AssessedBy)
			if latest.OverallScore != nil {
				fmt.Printf("Overall:     %.1f / %d\n", *latest.OverallScore, model.MaxLevel)
			}
			fmt.Println()

			if latest.Summary != "" {
				fmt.Println("Summary:")
				fmt.Printf("  %s\n\n", latest.Summary)
			}

			fmt.Println("Dimensions:")
			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "  KEY\tNAME\tLEVEL\tRATIONALE")
			for _, dim := range model.Dimensions {
				score := latest.Scores[dim.Key]
				rationale := score.Rationale
				if len(rationale) > 40 {
					rationale = rationale[:37] + "..."
				}
				fmt.Fprintf(tw, "  %s\t%s\t%d/%d\t%s\n",
					dim.Key, dim.Name, score.Level, model.MaxLevel, rationale)
			}
			tw.Flush()

			fmt.Println()
			fmt.Println("Use --json for detailed output suitable for radar chart visualization.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}
