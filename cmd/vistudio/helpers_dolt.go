//go:build dolt

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	config "github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
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
