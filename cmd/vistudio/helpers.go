package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	config "github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store/doltstore"
)

func getDSN(cmd *cobra.Command) string {
	dsn, _ := cmd.Flags().GetString("dsn")
	if dsn != "" {
		return dsn
	}
	if env := os.Getenv("VISIONSTUDIO_DSN"); env != "" {
		return env
	}
	if cfg, err := config.Load(); err == nil && cfg.DSN != "" {
		return cfg.DSN
	}
	return defaultDSN
}

func splitSQLStatements(sql string) []string {
	raw := strings.Split(sql, ";")
	var stmts []string
	for _, s := range raw {
		trimmed := strings.TrimSpace(s)
		if trimmed == "" {
			continue
		}
		lines := strings.Split(trimmed, "\n")
		hasSQL := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "--") {
				hasSQL = true
				break
			}
		}
		if hasSQL {
			stmts = append(stmts, trimmed)
		}
	}
	return stmts
}

func printCloseWarning(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: close database: %v\n", err)
	}
}

// PeriodicCommitter commits uncommitted Dolt changes at regular intervals.
// This prevents data loss when mutations don't auto-commit.
type PeriodicCommitter struct {
	dsn      string
	dataDir  string
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewPeriodicCommitter creates a committer that runs every interval.
// Provide either dsn (server mode) or dataDir (embedded mode), not both.
func NewPeriodicCommitter(dsn, dataDir string, interval time.Duration) *PeriodicCommitter {
	return &PeriodicCommitter{
		dsn:      dsn,
		dataDir:  dataDir,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the periodic commit loop in the background.
func (pc *PeriodicCommitter) Start(ctx context.Context) {
	pc.wg.Add(1)
	go func() {
		defer pc.wg.Done()
		ticker := time.NewTicker(pc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				pc.commitOnce(context.Background()) // Final commit on shutdown
				return
			case <-pc.stopCh:
				pc.commitOnce(context.Background()) // Final commit on stop
				return
			case <-ticker.C:
				pc.commitOnce(ctx)
			}
		}
	}()
}

// Stop signals the committer to stop and waits for it to finish.
func (pc *PeriodicCommitter) Stop() {
	close(pc.stopCh)
	pc.wg.Wait()
}

func (pc *PeriodicCommitter) commitOnce(ctx context.Context) {
	var ds *doltstore.DoltStore
	var err error

	if pc.dataDir != "" {
		ds, err = connectEmbedded(pc.dataDir)
	} else {
		ds, err = doltstore.New(pc.dsn)
	}
	if err != nil {
		log.Printf("periodic commit: connect: %v", err)
		return
	}
	defer ds.Close()

	committed, err := ds.CommitIfDirty(ctx, "visionstudio: periodic auto-commit")
	if err != nil {
		log.Printf("periodic commit: %v", err)
		return
	}
	if committed {
		log.Printf("periodic commit: committed uncommitted changes")
	}
}

// startPeriodicCommitter starts the periodic commit loop for the dashboard.
// Returns a cleanup function to call on shutdown.
func startPeriodicCommitter(ctx context.Context, cmd *cobra.Command) func() {
	dsn := getDSN(cmd)
	dataDir := getDataDir(cmd)

	interval := 5 * time.Minute // Commit every 5 minutes
	pc := NewPeriodicCommitter(dsn, dataDir, interval)
	pc.Start(ctx)

	return func() {
		pc.Stop()
	}
}
