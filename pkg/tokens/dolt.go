package tokens

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// DoltSource reads token events from the devx.token_events table on a Dolt server.
// This enables SQL-based access for VisionStudio and other consumers that
// cannot read the local JSONL files directly.
type DoltSource struct {
	db *sql.DB
}

// NewDoltSource creates a DoltSource from an existing database connection.
// The connection should be to a Dolt server with the devx database accessible.
func NewDoltSource(db *sql.DB) *DoltSource {
	return &DoltSource{db: db}
}

// Read implements Source by querying devx.token_events.
func (s *DoltSource) Read(ctx context.Context, q Query) (*ReadResult, error) {
	result := &ReadResult{
		Events: []Event{},
	}

	query, args := s.buildQuery(q)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query token_events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var evt Event
		var ts time.Time
		if err := rows.Scan(
			&evt.ID,
			&ts,
			&evt.SessionID,
			&evt.Workspace,
			&evt.Model,
			&evt.InputTokens,
			&evt.OutputTokens,
			&evt.CacheReadTokens,
			&evt.CacheCreationTokens,
		); err != nil {
			result.Diagnostics = append(result.Diagnostics,
				fmt.Sprintf("scan row: %v", err))
			continue
		}
		evt.Timestamp = ts
		result.Events = append(result.Events, evt)
	}

	if err := rows.Err(); err != nil {
		result.Diagnostics = append(result.Diagnostics,
			fmt.Sprintf("iterate rows: %v", err))
	}

	return result, nil
}

// buildQuery constructs a SELECT query with optional filters.
func (s *DoltSource) buildQuery(q Query) (string, []any) {
	var conditions []string
	var args []any

	// Period filter
	if !q.Period.Start.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, q.Period.Start)
	}
	if !q.Period.End.IsZero() {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, q.Period.End)
	}

	// Session ID filter
	if len(q.SessionIDs) > 0 {
		placeholders := make([]string, len(q.SessionIDs))
		for i, id := range q.SessionIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		conditions = append(conditions,
			fmt.Sprintf("session_id IN (%s)", strings.Join(placeholders, ", ")))
	}

	// Workspace filter
	if len(q.Workspaces) > 0 {
		placeholders := make([]string, len(q.Workspaces))
		for i, ws := range q.Workspaces {
			placeholders[i] = "?"
			args = append(args, ws)
		}
		conditions = append(conditions,
			fmt.Sprintf("workspace IN (%s)", strings.Join(placeholders, ", ")))
	}

	query := `SELECT
		event_id, timestamp, session_id, workspace, model,
		input_tokens, output_tokens, cache_read_tokens, cache_creation_tokens
	FROM devx.token_events`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args
}

// DevxSchema contains the DDL for the devx database and token_events table.
const DevxSchema = `
-- Create devx database if not exists
CREATE DATABASE IF NOT EXISTS devx;

-- Token events table for AI token usage tracking
-- Stores events from omnidevx JSONL files for SQL-based access
CREATE TABLE IF NOT EXISTS devx.token_events (
    event_id              VARCHAR(128) PRIMARY KEY,
    timestamp             DATETIME(6) NOT NULL,
    session_id            VARCHAR(128) NOT NULL,
    workspace             VARCHAR(512) NOT NULL,
    model                 VARCHAR(128) NOT NULL,
    input_tokens          BIGINT NOT NULL DEFAULT 0,
    output_tokens         BIGINT NOT NULL DEFAULT 0,
    cache_read_tokens     BIGINT NOT NULL DEFAULT 0,
    cache_creation_tokens BIGINT NOT NULL DEFAULT 0,

    INDEX idx_timestamp (timestamp),
    INDEX idx_session_id (session_id),
    INDEX idx_workspace (workspace)
);
`

// Ingest reads events from a Source and writes them to devx.token_events.
// It uses INSERT IGNORE for idempotency - duplicate event IDs are skipped.
func Ingest(ctx context.Context, db *sql.DB, source Source, q Query) (int, error) {
	// Ensure schema exists
	stmts := splitStatements(DevxSchema)
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return 0, fmt.Errorf("create schema: %w", err)
		}
	}

	// Read events from source
	result, err := source.Read(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("read events: %w", err)
	}

	if len(result.Events) == 0 {
		return 0, nil
	}

	// Batch insert with INSERT IGNORE for idempotency
	const batchSize = 1000
	inserted := 0

	for i := 0; i < len(result.Events); i += batchSize {
		end := i + batchSize
		if end > len(result.Events) {
			end = len(result.Events)
		}
		batch := result.Events[i:end]

		n, err := insertBatch(ctx, db, batch)
		if err != nil {
			return inserted, fmt.Errorf("insert batch: %w", err)
		}
		inserted += n
	}

	return inserted, nil
}

func insertBatch(ctx context.Context, db *sql.DB, events []Event) (int, error) {
	if len(events) == 0 {
		return 0, nil
	}

	// Build multi-value INSERT IGNORE
	var values []string
	var args []any

	for _, e := range events {
		values = append(values, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args,
			e.ID,
			e.Timestamp,
			e.SessionID,
			e.Workspace,
			e.Model,
			e.InputTokens,
			e.OutputTokens,
			e.CacheReadTokens,
			e.CacheCreationTokens,
		)
	}

	//nolint:gosec // G202: values are placeholders, not user input; args holds actual values
	query := `INSERT IGNORE INTO devx.token_events
		(event_id, timestamp, session_id, workspace, model,
		 input_tokens, output_tokens, cache_read_tokens, cache_creation_tokens)
		VALUES ` + strings.Join(values, ", ")

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	n, _ := result.RowsAffected()
	return int(n), nil
}

func splitStatements(sql string) []string {
	var stmts []string
	var current strings.Builder

	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comment-only lines
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		// Check if line ends with semicolon (statement boundary)
		if strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(current.String())
			stmt = strings.TrimSuffix(stmt, ";")
			stmt = strings.TrimSpace(stmt)
			if stmt != "" {
				stmts = append(stmts, stmt)
			}
			current.Reset()
		}
	}

	// Handle any remaining content without trailing semicolon
	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		stmts = append(stmts, remaining)
	}

	return stmts
}
