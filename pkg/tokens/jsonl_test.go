package tokens

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJSONLSource_Read(t *testing.T) {
	// Create a temp directory with the omnidevx store layout.
	tmpDir := t.TempDir()
	eventsDir := filepath.Join(tmpDir, "events", "2026", "07", "15")
	if err := os.MkdirAll(eventsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write fixture JSONL with ai.message.completed events.
	fixture := `{"id":"evt-1","type":"ai.message.completed","timestamp":"2026-07-15T10:00:00Z","context":{"sessionId":"session-abc","workspace":"/home/user/project"},"attributes":{"model":"claude-opus-4-8","input_tokens":100,"output_tokens":500,"cache_read_tokens":1000,"cache_creation_tokens":50}}
{"id":"evt-2","type":"ai.message.completed","timestamp":"2026-07-15T11:00:00Z","context":{"sessionId":"session-abc","workspace":"/home/user/project"},"attributes":{"model":"claude-opus-4-8","input_tokens":200,"output_tokens":600,"cache_read_tokens":2000,"cache_creation_tokens":100}}
{"id":"evt-3","type":"ai.session.started","timestamp":"2026-07-15T09:00:00Z","context":{"sessionId":"session-abc"}}
{"id":"evt-4","type":"ai.message.completed","timestamp":"2026-07-15T12:00:00Z","context":{"sessionId":"session-xyz","workspace":"/home/user/other"},"attributes":{"model":"claude-sonnet-5","input_tokens":50,"output_tokens":300,"cache_read_tokens":500,"cache_creation_tokens":25}}
`
	if err := os.WriteFile(filepath.Join(eventsDir, "claude-code.jsonl"), []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}

	source, err := NewJSONLSource(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	t.Run("reads all events in period", func(t *testing.T) {
		result, err := source.Read(ctx, Query{
			Period: Period{
				Start: time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2026, 7, 15, 23, 59, 59, 0, time.UTC),
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Should have 3 ai.message.completed events (evt-3 is ai.session.started).
		if got := len(result.Events); got != 3 {
			t.Errorf("got %d events, want 3", got)
		}

		// Verify first event fields.
		evt := findEvent(result.Events, "evt-1")
		if evt == nil {
			t.Fatal("evt-1 not found")
		}
		if evt.SessionID != "session-abc" {
			t.Errorf("SessionID = %q, want %q", evt.SessionID, "session-abc")
		}
		if evt.Workspace != "/home/user/project" {
			t.Errorf("Workspace = %q, want %q", evt.Workspace, "/home/user/project")
		}
		if evt.Model != "claude-opus-4-8" {
			t.Errorf("Model = %q, want %q", evt.Model, "claude-opus-4-8")
		}
		if evt.InputTokens != 100 {
			t.Errorf("InputTokens = %d, want %d", evt.InputTokens, 100)
		}
		if evt.OutputTokens != 500 {
			t.Errorf("OutputTokens = %d, want %d", evt.OutputTokens, 500)
		}
		if evt.CacheReadTokens != 1000 {
			t.Errorf("CacheReadTokens = %d, want %d", evt.CacheReadTokens, 1000)
		}
		if evt.CacheCreationTokens != 50 {
			t.Errorf("CacheCreationTokens = %d, want %d", evt.CacheCreationTokens, 50)
		}
		if evt.TotalTokens() != 1650 {
			t.Errorf("TotalTokens() = %d, want %d", evt.TotalTokens(), 1650)
		}
	})

	t.Run("filters by session ID", func(t *testing.T) {
		result, err := source.Read(ctx, Query{
			Period: Period{
				Start: time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2026, 7, 15, 23, 59, 59, 0, time.UTC),
			},
			SessionIDs: []string{"session-abc"},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Should have 2 events from session-abc.
		if got := len(result.Events); got != 2 {
			t.Errorf("got %d events, want 2", got)
		}
	})

	t.Run("filters by workspace", func(t *testing.T) {
		result, err := source.Read(ctx, Query{
			Period: Period{
				Start: time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2026, 7, 15, 23, 59, 59, 0, time.UTC),
			},
			Workspaces: []string{"/home/user/other"},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Should have 1 event from /home/user/other.
		if got := len(result.Events); got != 1 {
			t.Errorf("got %d events, want 1", got)
		}
		if result.Events[0].ID != "evt-4" {
			t.Errorf("got event %q, want %q", result.Events[0].ID, "evt-4")
		}
	})

	t.Run("filters by period", func(t *testing.T) {
		result, err := source.Read(ctx, Query{
			Period: Period{
				Start: time.Date(2026, 7, 15, 10, 30, 0, 0, time.UTC),
				End:   time.Date(2026, 7, 15, 11, 30, 0, 0, time.UTC),
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Should have 1 event (evt-2 at 11:00).
		if got := len(result.Events); got != 1 {
			t.Errorf("got %d events, want 1", got)
		}
	})

	t.Run("empty result for no matching period", func(t *testing.T) {
		result, err := source.Read(ctx, Query{
			Period: Period{
				Start: time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2026, 7, 16, 23, 59, 59, 0, time.UTC),
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		if got := len(result.Events); got != 0 {
			t.Errorf("got %d events, want 0", got)
		}
	})
}

func TestJSONLSource_ReadMultipleDays(t *testing.T) {
	tmpDir := t.TempDir()

	// Create events across multiple days.
	days := []struct {
		date    string
		eventID string
	}{
		{"2026/07/14", "evt-day14"},
		{"2026/07/15", "evt-day15"},
		{"2026/07/16", "evt-day16"},
	}

	for _, d := range days {
		eventsDir := filepath.Join(tmpDir, "events", d.date)
		if err := os.MkdirAll(eventsDir, 0o755); err != nil {
			t.Fatal(err)
		}

		fixture := `{"id":"` + d.eventID + `","type":"ai.message.completed","timestamp":"` + d.date[:4] + `-` + d.date[5:7] + `-` + d.date[8:10] + `T10:00:00Z","context":{"sessionId":"session-1","workspace":"/project"},"attributes":{"model":"claude-opus-4-8","input_tokens":100,"output_tokens":200}}`
		if err := os.WriteFile(filepath.Join(eventsDir, "claude-code.jsonl"), []byte(fixture), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	source, err := NewJSONLSource(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	result, err := source.Read(context.Background(), Query{
		Period: Period{
			Start: time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2026, 7, 16, 23, 59, 59, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if got := len(result.Events); got != 3 {
		t.Errorf("got %d events, want 3", got)
	}
}

func TestJSONLSource_MalformedLine(t *testing.T) {
	tmpDir := t.TempDir()
	eventsDir := filepath.Join(tmpDir, "events", "2026", "07", "15")
	if err := os.MkdirAll(eventsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Include a malformed line.
	fixture := `{"id":"evt-1","type":"ai.message.completed","timestamp":"2026-07-15T10:00:00Z","context":{"sessionId":"s1","workspace":"/p"},"attributes":{"model":"opus","input_tokens":100}}
not valid json
{"id":"evt-2","type":"ai.message.completed","timestamp":"2026-07-15T11:00:00Z","context":{"sessionId":"s1","workspace":"/p"},"attributes":{"model":"opus","input_tokens":200}}
`
	if err := os.WriteFile(filepath.Join(eventsDir, "claude-code.jsonl"), []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}

	source, err := NewJSONLSource(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	result, err := source.Read(context.Background(), Query{
		Period: Period{
			Start: time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2026, 7, 15, 23, 59, 59, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Should still get 2 valid events.
	if got := len(result.Events); got != 2 {
		t.Errorf("got %d events, want 2", got)
	}

	// Should have 1 diagnostic for the malformed line.
	if got := len(result.Diagnostics); got != 1 {
		t.Errorf("got %d diagnostics, want 1", got)
	}
}

func findEvent(events []Event, id string) *Event {
	for i := range events {
		if events[i].ID == id {
			return &events[i]
		}
	}
	return nil
}
