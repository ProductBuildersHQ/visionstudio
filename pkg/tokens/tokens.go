// Package tokens provides interfaces and implementations for reading token
// spend data from various sources. It supports the Token Attribution
// Reporting feature (TRD §16) by abstracting over the omnidevx JSONL store
// and future SQL-based sources.
package tokens

import (
	"context"
	"time"
)

// Event represents a single token usage observation. It contains the subset
// of omnidevx event fields needed for attribution and reporting.
type Event struct {
	// ID is the unique event identifier (deterministic, dedup-safe).
	ID string

	// Timestamp is when the token usage occurred.
	Timestamp time.Time

	// SessionID is the Claude Code session UUID (used for assignment attribution).
	SessionID string

	// Workspace is the working directory path (used for repository attribution).
	Workspace string

	// Model is the model name (e.g., "claude-opus-4-8").
	Model string

	// Token counts by category.
	InputTokens         int64
	OutputTokens        int64
	CacheReadTokens     int64
	CacheCreationTokens int64
}

// TotalTokens returns the sum of all token categories.
func (e *Event) TotalTokens() int64 {
	return e.InputTokens + e.OutputTokens + e.CacheReadTokens + e.CacheCreationTokens
}

// Period defines a time range for querying events.
type Period struct {
	Start time.Time
	End   time.Time
}

// Contains returns true if t is within the period [Start, End].
func (p Period) Contains(t time.Time) bool {
	return !t.Before(p.Start) && !t.After(p.End)
}

// Query specifies filters for reading token events.
type Query struct {
	// Period limits events to those within the time range.
	Period Period

	// SessionIDs, if non-empty, limits events to those sessions.
	SessionIDs []string

	// Workspaces, if non-empty, limits events to those workspaces.
	Workspaces []string
}

// ReadResult contains the events and metadata from a read operation.
type ReadResult struct {
	Events []Event

	// Diagnostics reports any issues encountered during reading (e.g.,
	// unparseable lines). Diagnostics are informational; the Events slice
	// contains all successfully parsed events.
	Diagnostics []string
}

// Source reads token events from a backing store. Implementations include
// the JSONL reader (omnidevx store) and future SQL readers (devx database).
type Source interface {
	// Read returns token events matching the query. The returned events are
	// not guaranteed to be in any particular order.
	Read(ctx context.Context, q Query) (*ReadResult, error)
}
