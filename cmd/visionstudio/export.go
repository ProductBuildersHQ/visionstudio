package main

import (
	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/export"
)

func exportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export JSONL snapshots of all tables",
		Long: `Write one JSONL file per table (initiatives, phases, roadmap_items,
rmi_dependencies, assignments, evidence, repositories, repository_dependencies)
plus a manifest.json with record counts and timestamp into the target directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cleanup, err := connectService(cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			dir, _ := cmd.Flags().GetString("dir")

			result, err := export.Run(cmd.Context(), svc.Store, dir)
			if err != nil {
				return err
			}

			cmd.Printf("Exported to %s:\n", result.Dir)
			for table, count := range result.Manifest.Counts {
				cmd.Printf("  %-25s %d\n", table, count)
			}
			cmd.Printf("  %-25s %s\n", "exported_at", result.Manifest.ExportedAt.Format("2006-01-02 15:04:05"))
			return nil
		},
	}
	cmd.Flags().String("dir", "exports", "Output directory for JSONL files")
	return cmd
}
