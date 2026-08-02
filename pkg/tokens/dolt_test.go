package tokens

import (
	"testing"
	"time"
)

func TestBuildQuery(t *testing.T) {
	source := &DoltSource{}

	tests := []struct {
		name     string
		query    Query
		wantSQL  string
		wantArgs int
	}{
		{
			name:  "no filters",
			query: Query{},
			wantSQL: `SELECT
		event_id, timestamp, session_id, workspace, model,
		input_tokens, output_tokens, cache_read_tokens, cache_creation_tokens
	FROM devx.token_events`,
			wantArgs: 0,
		},
		{
			name: "period only",
			query: Query{
				Period: Period{
					Start: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2026, 7, 31, 23, 59, 59, 0, time.UTC),
				},
			},
			wantArgs: 2,
		},
		{
			name: "session filter",
			query: Query{
				SessionIDs: []string{"session-1", "session-2"},
			},
			wantArgs: 2,
		},
		{
			name: "workspace filter",
			query: Query{
				Workspaces: []string{"/home/user/project"},
			},
			wantArgs: 1,
		},
		{
			name: "all filters",
			query: Query{
				Period: Period{
					Start: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2026, 7, 31, 23, 59, 59, 0, time.UTC),
				},
				SessionIDs: []string{"session-1"},
				Workspaces: []string{"/home/user/project"},
			},
			wantArgs: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := source.buildQuery(tt.query)
			if sql == "" {
				t.Error("empty SQL returned")
			}
			if len(args) != tt.wantArgs {
				t.Errorf("got %d args, want %d", len(args), tt.wantArgs)
			}
		})
	}
}

func TestSplitStatements(t *testing.T) {
	sql := `
-- Comment
CREATE DATABASE IF NOT EXISTS devx;

-- Another comment
CREATE TABLE IF NOT EXISTS devx.token_events (
    event_id VARCHAR(128) PRIMARY KEY
);
`
	stmts := splitStatements(sql)
	if len(stmts) != 2 {
		t.Errorf("got %d statements, want 2", len(stmts))
	}
}

func TestDevxSchemaNotEmpty(t *testing.T) {
	if DevxSchema == "" {
		t.Error("DevxSchema is empty")
	}

	stmts := splitStatements(DevxSchema)
	if len(stmts) < 2 {
		t.Errorf("expected at least 2 statements (CREATE DATABASE, CREATE TABLE), got %d", len(stmts))
	}
}
