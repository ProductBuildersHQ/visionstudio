package main

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store/doltstore"
	"github.com/ProductBuildersHQ/visionstudio/pkg/tokens"
)

//go:embed views.sql
var viewsSQL string

func addDoltDBCommands(cmd *cobra.Command) {
	cmd.AddCommand(dbInitCmd(), dbCreateViewsCmd(), dbIngestTokensCmd(), dbCommitCmd())
}

func dbInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "Bootstrap a Dolt database for PRISM Control",
		Long: `Initialize a Dolt database and run schema migration.

In embedded mode (--data-dir or VISIONSTUDIO_DATA set), creates the database
directory and runs migration automatically — no server needed.

In server mode, requires a running Dolt SQL server (see 'visionstudio db serve')
and the --migrate flag.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir := getDataDir(cmd)

			if dataDir != "" {
				return dbInitEmbedded(cmd, dataDir)
			}

			dir := defaultDataDir
			if len(args) > 0 {
				dir = args[0]
			}
			return dbInitServer(cmd, dir)
		},
	}
	cmd.Flags().Bool("migrate", false, "Also run schema migration (server mode only)")
	return cmd
}


func dbInitServer(cmd *cobra.Command, dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	cmd.Printf("Initializing Dolt database at %s...\n", absDir)
	if err := service.DBInit(absDir); err != nil {
		return err
	}
	cmd.Println("Dolt database initialized.")

	migrateFlag, _ := cmd.Flags().GetBool("migrate")
	if !migrateFlag {
		cmd.Println("Run 'visionstudio db serve' then 'visionstudio db init --migrate' to create tables.")
		return nil
	}

	dsn := getDSN(cmd)
	cmd.Printf("Running schema migration (DSN: %s)...\n", dsn)
	ds, err := doltstore.New(dsn)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer func() { printCloseWarning(ds.Close()) }()

	if err := ds.Migrate(cmd.Context()); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	cmd.Println("Schema migration complete.")
	return nil
}

func dbCreateViewsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-views",
		Short: "Create read-only SQL views for VisionStudio and other consumers",
		Long: `Execute the bundled SQL view definitions against the database.

Views created:
  v_initiative_summary  — initiative rollup (RMI counts, phase count, repos)
  v_phase_progress      — per-phase progress with derived status
  v_rmi_detail          — flat RMI detail with denormalized references
  v_active_assignments  — currently active work assignments
  v_initiative_tokens   — token spend per initiative (requires devx.token_events)
  v_rmi_tokens          — token spend per RMI (requires devx.token_events)
  v_unattributed_tokens — unattributed token events (requires devx.token_events)

These views use only base tables (no Dolt system tables) and are safe
for read-only consumers such as VisionStudio.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir := getDataDir(cmd)

			var ds *doltstore.DoltStore
			var err error
			if dataDir != "" {
				ds, err = connectEmbedded(dataDir)
			} else {
				ds, err = doltstore.New(getDSN(cmd))
			}
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}
			defer func() { printCloseWarning(ds.Close()) }()

			stmts := splitSQLStatements(viewsSQL)
			for _, stmt := range stmts {
				if err := ds.ExecSQL(cmd.Context(), stmt); err != nil {
					return fmt.Errorf("execute view statement: %w", err)
				}
			}

			cmd.Println("Created views: v_initiative_summary, v_phase_progress, v_rmi_detail, v_active_assignments, v_initiative_tokens, v_rmi_tokens, v_unattributed_tokens")
			return nil
		},
	}
}

func dbIngestTokensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingest-tokens",
		Short: "Ingest token events from omnidevx JSONL files into devx.token_events",
		Long: `Read AI token usage events from the local omnidevx JSONL store and
write them to the devx.token_events table in Dolt.

This enables SQL-based access to token data for VisionStudio and other
consumers that cannot read the local JSONL files directly.

The ingest is idempotent — duplicate event IDs are skipped via INSERT IGNORE.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir := getDataDir(cmd)
			omnidevxDir, _ := cmd.Flags().GetString("omnidevx-dir")
			since, _ := cmd.Flags().GetString("since")
			until, _ := cmd.Flags().GetString("until")

			// Parse time range
			var period tokens.Period
			if since != "" {
				t, err := parseDate(since)
				if err != nil {
					return fmt.Errorf("parse --since: %w", err)
				}
				period.Start = t
			} else {
				// Default to 30 days ago
				period.Start = time.Now().AddDate(0, 0, -30)
			}
			if until != "" {
				t, err := parseDate(until)
				if err != nil {
					return fmt.Errorf("parse --until: %w", err)
				}
				period.End = t
			} else {
				period.End = time.Now()
			}

			// Connect to Dolt
			var ds *doltstore.DoltStore
			var err error
			if dataDir != "" {
				ds, err = connectEmbedded(dataDir)
			} else {
				ds, err = doltstore.New(getDSN(cmd))
			}
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}
			defer func() { printCloseWarning(ds.Close()) }()

			// Create JSONL source
			source, err := tokens.NewJSONLSource(omnidevxDir)
			if err != nil {
				return fmt.Errorf("create JSONL source: %w", err)
			}

			cmd.Printf("Ingesting token events from %s to %s...\n",
				period.Start.Format("2006-01-02"),
				period.End.Format("2006-01-02"))

			// Run ingest
			n, err := tokens.Ingest(cmd.Context(), ds.DB(), source, tokens.Query{Period: period})
			if err != nil {
				return fmt.Errorf("ingest: %w", err)
			}

			cmd.Printf("Ingested %d token events into devx.token_events\n", n)
			return nil
		},
	}
	cmd.Flags().String("omnidevx-dir", "", "omnidevx data directory (default: ~/.plexusone/omnidevx/data)")
	cmd.Flags().String("since", "", "Start date (YYYY-MM-DD, default: 30 days ago)")
	cmd.Flags().String("until", "", "End date (YYYY-MM-DD, default: today)")
	return cmd
}

func parseDate(s string) (time.Time, error) {
	// Try full datetime first
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// Try date only
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date format: %s (expected YYYY-MM-DD)", s)
}

func dbCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit [message]",
		Short: "Commit any uncommitted changes in the Dolt working set",
		Long: `Stage and commit all uncommitted changes in the Dolt database.

By default, only commits if there are changes. Use --force to create
an empty commit even when there are no changes.

This is useful when changes have been made without auto-commit enabled,
or after bulk operations that don't commit individually.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir := getDataDir(cmd)
			force, _ := cmd.Flags().GetBool("force")

			var ds *doltstore.DoltStore
			var err error
			if dataDir != "" {
				ds, err = connectEmbedded(dataDir)
			} else {
				ds, err = doltstore.New(getDSN(cmd))
			}
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}
			defer func() { printCloseWarning(ds.Close()) }()

			message := "visionstudio: manual commit"
			if len(args) > 0 {
				message = args[0]
			}

			if force {
				if err := ds.Commit(cmd.Context(), message); err != nil {
					return fmt.Errorf("commit: %w", err)
				}
				cmd.Println("Committed (forced).")
				return nil
			}

			committed, err := ds.CommitIfDirty(cmd.Context(), message)
			if err != nil {
				return fmt.Errorf("commit: %w", err)
			}
			if committed {
				cmd.Println("Committed uncommitted changes.")
			} else {
				cmd.Println("Nothing to commit (working set is clean).")
			}
			return nil
		},
	}
	cmd.Flags().Bool("force", false, "Create commit even if there are no changes")
	return cmd
}
