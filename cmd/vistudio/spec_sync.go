package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/specworkflow"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func specSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Reconcile spec files on disk with Initiative.Specs in the database",
		Long: `Scan docs/specs/initiatives/{INIT-ID}/ in each registered repository,
backfill the Initiative.Specs map for any specs found on disk but missing
from the database, and report legacy-location spec files that should be migrated.

This keeps the specs panel honest when agents write specs by hand without
using 'vistudio spec init'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			verbose, _ := cmd.Flags().GetBool("verbose")

			ctx := cmd.Context()

			// Get all repositories with local paths
			repos, err := svc.Store.ListRepositories(ctx)
			if err != nil {
				return fmt.Errorf("list repositories: %w", err)
			}

			// Get all initiatives
			initiatives, err := svc.Store.ListInitiatives(ctx)
			if err != nil {
				return fmt.Errorf("list initiatives: %w", err)
			}

			initByID := make(map[string]*store.Initiative)
			for _, init := range initiatives {
				initByID[init.ID] = init
			}

			var (
				specsFound   int
				specsAdded   int
				legacySpecs  []string
				missingRepos []string
			)

			for _, repo := range repos {
				if repo.LocalPath == "" {
					if verbose {
						cmd.Printf("Skipping %s: no local path\n", repo.ID)
					}
					continue
				}

				if _, err := os.Stat(repo.LocalPath); err != nil {
					missingRepos = append(missingRepos, repo.ID)
					continue
				}

				// Scan docs/specs/initiatives/
				specsBase := filepath.Join(repo.LocalPath, "docs", "specs", "initiatives")
				if _, err := os.Stat(specsBase); err != nil {
					if verbose {
						cmd.Printf("No specs directory in %s\n", repo.ID)
					}
					continue
				}

				entries, err := os.ReadDir(specsBase)
				if err != nil {
					return fmt.Errorf("read %s: %w", specsBase, err)
				}

				for _, entry := range entries {
					if !entry.IsDir() {
						continue
					}

					initID := entry.Name()
					init, ok := initByID[initID]
					if !ok {
						if verbose {
							cmd.Printf("Directory %s/%s does not match any initiative\n", specsBase, initID)
						}
						continue
					}

					// Scan spec files in this directory
					specDir := filepath.Join(specsBase, initID)
					specFiles, err := os.ReadDir(specDir)
					if err != nil {
						return fmt.Errorf("read %s: %w", specDir, err)
					}

					if init.Specs == nil {
						init.Specs = make(map[string]string)
					}

					specsChanged := false
					for _, sf := range specFiles {
						if sf.IsDir() || !strings.HasSuffix(sf.Name(), ".md") {
							continue
						}

						specsFound++
						specName := sf.Name()
						key := strings.ToLower(strings.TrimSuffix(specName, ".md"))
						relPath := filepath.Join(specworkflow.SpecDir(initID), specName)

						if existing, ok := init.Specs[key]; ok {
							if existing != relPath && verbose {
								cmd.Printf("  %s: %s already mapped to %s (disk has %s)\n",
									initID, key, existing, relPath)
							}
							continue
						}

						// New spec found on disk
						specsAdded++
						init.Specs[key] = relPath
						specsChanged = true

						if verbose {
							cmd.Printf("  %s: adding %s -> %s\n", initID, key, relPath)
						}
					}

					if specsChanged && !dryRun {
						if err := svc.UpdateInitiative(ctx, init); err != nil {
							return fmt.Errorf("update initiative %s: %w", initID, err)
						}
					}
				}

				// Check for legacy spec locations (docs/specs/*.md at repo root)
				legacyDir := filepath.Join(repo.LocalPath, "docs", "specs")
				legacyFiles, err := os.ReadDir(legacyDir)
				if err == nil {
					for _, lf := range legacyFiles {
						if lf.IsDir() && lf.Name() != "initiatives" {
							continue
						}
						if !lf.IsDir() && strings.HasSuffix(lf.Name(), ".md") {
							// Check if it looks like an initiative spec
							name := strings.ToLower(lf.Name())
							if name == "prd.md" || name == "trd.md" || name == "plan.md" ||
								name == "roadmap.md" || name == "prfaq.md" {
								legacySpecs = append(legacySpecs,
									fmt.Sprintf("%s: %s", repo.ID, filepath.Join("docs/specs", lf.Name())))
							}
						}
					}
				}
			}

			// Summary
			cmd.Printf("\nSpec sync summary:\n")
			cmd.Printf("  Specs found on disk: %d\n", specsFound)
			if dryRun {
				cmd.Printf("  Specs to add (dry-run): %d\n", specsAdded)
			} else {
				cmd.Printf("  Specs added to DB: %d\n", specsAdded)
			}

			if len(missingRepos) > 0 {
				cmd.Printf("\nRepositories with missing local paths:\n")
				for _, r := range missingRepos {
					cmd.Printf("  - %s\n", r)
				}
			}

			if len(legacySpecs) > 0 {
				cmd.Printf("\nLegacy spec locations (consider migrating to docs/specs/initiatives/{INIT-ID}/):\n")
				for _, ls := range legacySpecs {
					cmd.Printf("  - %s\n", ls)
				}
			}

			return nil
		},
	}
	cmd.Flags().Bool("dry-run", false, "Show what would be synced without making changes")
	cmd.Flags().BoolP("verbose", "v", false, "Show detailed progress")
	return cmd
}
