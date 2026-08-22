package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/plexusone/structured-evaluation/rubric"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/ProductBuildersHQ/visionstudio/pkg/specworkflow"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func specCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "Manage initiative specification documents",
	}
	cmd.AddCommand(specInitCmd(), specValidateCmd(), specJudgeCmd(), specSyncCmd())
	return cmd
}

func specInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <initiative-id>",
		Short: "Scaffold spec files for an initiative's workflow",
		Long: `Scaffold the spec documents an initiative's workflow requires.

Files are created in the initiative's HOME REPO working tree at
  <home-repo local path>/docs/specs/initiatives/<INIT-ID>/
using the registry's local path, so the initiative must have --home-repo set
('initiative update <id> --home-repo <repo-id>') and that repo must be
registered with a local path.

Which files depends on the initiative's workflow ('workflow list' shows all):
pbhq-lite (default for type 'feature') requires PRD.md, TRD.md, PLAN.md,
ROADMAP.md; quick-fix (default for maintenance/refactor/migration) requires
only ROADMAP.md. Existing files are never overwritten. The scaffolded paths
are also recorded on the initiative record so other tools can find them.

After filling in ROADMAP.md, run 'roadmap import' to create its phases and RMIs.`,
		Example: `  visionstudio spec init INIT-MYPROJECT-001
  visionstudio spec init INIT-MYPROJECT-001 --with-optional`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			withOptional, _ := cmd.Flags().GetBool("with-optional")

			init, err := svc.Store.GetInitiative(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if init.HomeRepo == "" {
				return fmt.Errorf("initiative %s has no home-repo set; run 'visionstudio initiative update %s --home-repo <repo-id>' first", init.ID, init.ID)
			}
			repo, err := svc.Store.GetRepository(cmd.Context(), init.HomeRepo)
			if err != nil {
				return fmt.Errorf("get home repo: %w", err)
			}
			if repo.LocalPath == "" {
				return fmt.Errorf("repository %s has no local path set", repo.ID)
			}

			wf, err := specworkflow.Resolve(specworkflow.DefaultLoader(), init)
			if err != nil {
				return err
			}

			specDir := filepath.Join(repo.LocalPath, specworkflow.SpecDir(init.ID))
			if err := os.MkdirAll(specDir, 0o755); err != nil {
				return fmt.Errorf("create spec dir: %w", err)
			}

			specs := append([]string{}, wf.SpecsRequired...)
			if withOptional {
				specs = append(specs, wf.SpecsOptional...)
			}

			var created, skipped []string
			for _, name := range specs {
				path := filepath.Join(specDir, name)
				if _, err := os.Stat(path); err == nil {
					skipped = append(skipped, name)
					continue
				}
				if err := os.WriteFile(path, []byte(scaffoldTemplate(name, init)), 0o600); err != nil {
					return fmt.Errorf("write %s: %w", name, err)
				}
				created = append(created, name)
			}

			// RMI-112: standardize the canonical spec path onto the initiative
			// record so contextbuild and other consumers resolve these files
			// without needing --spec flags set by hand.
			if init.Specs == nil {
				init.Specs = make(map[string]string)
			}
			var specsChanged bool
			for _, name := range specs {
				key := strings.ToLower(strings.TrimSuffix(name, filepath.Ext(name)))
				relPath := filepath.Join(specworkflow.SpecDir(init.ID), name)
				if init.Specs[key] != relPath {
					init.Specs[key] = relPath
					specsChanged = true
				}
			}
			if specsChanged {
				if err := svc.UpdateInitiative(cmd.Context(), init); err != nil {
					return fmt.Errorf("update initiative specs: %w", err)
				}
			}

			cmd.Printf("Workflow: %s (%s)\n", wf.ID, wf.Name)
			cmd.Printf("Spec dir: %s\n", specDir)
			if len(created) > 0 {
				cmd.Printf("Created:  %s\n", strings.Join(created, ", "))
			}
			if len(skipped) > 0 {
				cmd.Printf("Skipped (exists): %s\n", strings.Join(skipped, ", "))
			}
			return nil
		},
	}
	cmd.Flags().Bool("with-optional", false, "Also scaffold optional specs")
	return cmd
}

func scaffoldTemplate(specFile string, init *store.Initiative) string {
	title := strings.TrimSuffix(specFile, ".md")
	return fmt.Sprintf("# %s — %s\n\n<!-- %s for %s. Fill in before judging. -->\n", title, init.Title, title, init.ID)
}

func specValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <initiative-id>",
		Short: "Check that required specs exist for an initiative's workflow",
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
			if init.HomeRepo == "" {
				return fmt.Errorf("initiative %s has no home-repo set", init.ID)
			}
			repo, err := svc.Store.GetRepository(cmd.Context(), init.HomeRepo)
			if err != nil {
				return fmt.Errorf("get home repo: %w", err)
			}

			wf, err := specworkflow.Resolve(specworkflow.DefaultLoader(), init)
			if err != nil {
				return err
			}

			specDir := filepath.Join(repo.LocalPath, specworkflow.SpecDir(init.ID))

			var missing []string
			for _, name := range wf.SpecsRequired {
				if _, err := os.Stat(filepath.Join(specDir, name)); err != nil {
					missing = append(missing, name)
				}
			}

			cmd.Printf("Workflow: %s (%s)\n", wf.ID, wf.Name)
			cmd.Printf("Spec dir: %s\n", specDir)
			if len(missing) == 0 {
				cmd.Println("All required specs present.")
				return nil
			}
			cmd.Printf("Missing required specs: %s\n", strings.Join(missing, ", "))
			return fmt.Errorf("%d required spec(s) missing", len(missing))
		},
	}
}

func specJudgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "judge",
		Short: "Evaluate spec quality via interactive agent judging",
		Long: `Judging is performed by an interactive agent session (e.g. Claude Code),
not a remote API call. 'spec judge show' prints a spec's content and rubric
for the agent to read; 'spec judge record' persists the agent's verdict.`,
	}
	cmd.AddCommand(specJudgeShowCmd(), specJudgeRecordCmd(), specJudgeListCmd())
	return cmd
}

func specJudgeShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <initiative-id> <spec-file>",
		Short: "Print a spec's content and rubric for agent evaluation",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			initID, specFile := args[0], args[1]
			init, err := svc.Store.GetInitiative(cmd.Context(), initID)
			if err != nil {
				return err
			}
			if init.HomeRepo == "" {
				return fmt.Errorf("initiative %s has no home-repo set", init.ID)
			}
			repo, err := svc.Store.GetRepository(cmd.Context(), init.HomeRepo)
			if err != nil {
				return fmt.Errorf("get home repo: %w", err)
			}

			loader := specworkflow.DefaultLoader()
			wf, err := specworkflow.Resolve(loader, init)
			if err != nil {
				return err
			}

			specPath := filepath.Join(repo.LocalPath, specworkflow.SpecDir(init.ID), specFile)
			content, err := os.ReadFile(specPath)
			if err != nil {
				return fmt.Errorf("read spec: %w", err)
			}

			// Resolve the rubric live from the visionspec catalog (the single
			// source of truth), deriving the spec type from the filename
			// (PRD.md -> prd). The judge_rubrics DB table is never populated
			// from the catalog, so we resolve through the workflow loader here.
			specType := strings.ToLower(strings.TrimSuffix(specFile, filepath.Ext(specFile)))
			rs, rubricErr := loader.GetRubric(wf.ID, specType)

			cmd.Printf("=== Spec: %s (initiative %s, workflow %s) ===\n\n", specFile, init.ID, wf.ID)
			if rubricErr != nil || rs == nil {
				cmd.Printf("--- No rubric defined for spec type %q in workflow %s; use general judgment ---\n\n", specType, wf.ID)
			} else {
				rubricYAML, err := yaml.Marshal(rs)
				if err != nil {
					return fmt.Errorf("render rubric %s: %w", rs.ID, err)
				}
				cmd.Printf("--- Rubric (%s) ---\n%s\n", rs.ID, string(rubricYAML))
			}
			cmd.Printf("--- Content (%s) ---\n%s\n", specPath, string(content))
			cmd.Printf("\n--- To record your evaluation ---\n")
			cmd.Printf("visionstudio spec judge record %s %s --score <0-10> --rationale \"<why>\" --model <your-model-id>\n", initID, specFile)
			return nil
		},
	}
}

func specJudgeRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record <initiative-id> <spec-file>",
		Short: "Record a judge evaluation result",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			initID, specFile := args[0], args[1]

			if !cmd.Flags().Changed("score") {
				return fmt.Errorf("--score is required")
			}
			score, _ := cmd.Flags().GetInt("score")
			rationale, _ := cmd.Flags().GetString("rationale")
			model, _ := cmd.Flags().GetString("model")
			if rationale == "" {
				return fmt.Errorf("--rationale is required")
			}

			init, err := svc.Store.GetInitiative(cmd.Context(), initID)
			if err != nil {
				return err
			}
			loader := specworkflow.DefaultLoader()
			wf, err := specworkflow.Resolve(loader, init)
			if err != nil {
				return err
			}

			// Resolve the rubric ID from the catalog (single source of truth)
			// so the recorded result references the same layered rubric that
			// 'spec judge show' displays. Spec type derives from the filename
			// (PRD.md -> prd).
			specType := strings.ToLower(strings.TrimSuffix(specFile, filepath.Ext(specFile)))
			var rubricID string
			if rs, rerr := loader.GetRubric(wf.ID, specType); rerr == nil && rs != nil {
				rubricID = rs.ID
			}

			now := time.Now()

			// Build structured-evaluation report
			intScore := rubric.IntegerScore(score)
			if intScore < 1 {
				intScore = 1
			}
			if intScore > 5 {
				intScore = 5
			}

			report := &rubric.Rubric{
				Metadata: rubric.ReportMetadata{
					Document:    filepath.Join(specworkflow.SpecDir(initID), specFile),
					GeneratedAt: now,
				},
				ReviewType: specType,
				RubricID:   rubricID,
				IntScore:   intScore,
				Pass:       intScore >= 4,
				Summary:    rationale,
			}
			if model != "" {
				report.Judge = &rubric.JudgeMetadata{Model: model}
			}

			result := &store.JudgeResult{
				ID:           fmt.Sprintf("%s-%s-%d", initID, specType, now.Unix()),
				InitiativeID: initID,
				SpecPath:     filepath.Join(specworkflow.SpecDir(initID), specFile),
				SpecType:     specType,
				RubricID:     rubricID,
				EvaluatedAt:  now,
				Report:       report,
			}
			if err := svc.Store.CreateJudgeResult(cmd.Context(), result); err != nil {
				return err
			}
			cmd.Printf("Recorded judge result %s (score: %d/5)\n", result.ID, result.Score())
			return nil
		},
	}
	cmd.Flags().Int("score", 0, "Score 1-5 (required)")
	cmd.Flags().String("rationale", "", "Rationale for the score (required)")
	cmd.Flags().String("model", "", "Model/agent identifier that produced this judgment")
	return cmd
}

func specJudgeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <initiative-id>",
		Short: "List judge results for an initiative",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			results, err := svc.Store.ListJudgeResults(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				cmd.Println("No judge results found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tSPEC\tSCORE\tMODEL\tEVALUATED")
			for _, r := range results {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n", r.ID, r.SpecPath, r.Score(), r.Model(), r.EvaluatedAt.Format("2006-01-02 15:04"))
			}
			return w.Flush()
		},
	}
}
