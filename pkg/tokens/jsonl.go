package tokens

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// JSONLSource reads token events from the omnidevx JSONL store layout:
//
//	<dir>/events/YYYY/MM/DD/<product>.jsonl
//
// It reads ai.message.completed events and extracts token usage data.
type JSONLSource struct {
	// Dir is the omnidevx data directory. Defaults to ~/.plexusone/omnidevx/data.
	Dir string
}

// NewJSONLSource creates a JSONLSource with the given directory. If dir is
// empty, it defaults to ~/.plexusone/omnidevx/data.
func NewJSONLSource(dir string) (*JSONLSource, error) {
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("tokens: resolve home directory: %w", err)
		}
		dir = filepath.Join(home, ".plexusone", "omnidevx", "data")
	}
	return &JSONLSource{Dir: dir}, nil
}

// Read implements Source by scanning JSONL files within the query period.
func (s *JSONLSource) Read(ctx context.Context, q Query) (*ReadResult, error) {
	eventsDir := filepath.Join(s.Dir, "events")

	result := &ReadResult{
		Events: []Event{},
	}

	// Build the list of day directories to scan based on the query period.
	days := daysInPeriod(q.Period)
	for _, day := range days {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		dayDir := filepath.Join(eventsDir, day.Format("2006/01/02"))
		if _, err := os.Stat(dayDir); os.IsNotExist(err) {
			continue
		}

		// Scan all JSONL files in the day directory.
		files, err := filepath.Glob(filepath.Join(dayDir, "*.jsonl"))
		if err != nil {
			result.Diagnostics = append(result.Diagnostics,
				fmt.Sprintf("glob %s: %v", dayDir, err))
			continue
		}

		for _, file := range files {
			if err := ctx.Err(); err != nil {
				return result, err
			}

			events, diags := s.readFile(file, q)
			result.Events = append(result.Events, events...)
			result.Diagnostics = append(result.Diagnostics, diags...)
		}
	}

	return result, nil
}

// readFile parses one JSONL file and returns matching events.
func (s *JSONLSource) readFile(path string, q Query) ([]Event, []string) {
	f, err := os.Open(path)
	if err != nil {
		return nil, []string{fmt.Sprintf("open %s: %v", path, err)}
	}
	defer f.Close()

	var events []Event
	var diags []string

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // Support long lines

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		event, ok, err := s.parseLine(line, q)
		if err != nil {
			diags = append(diags, fmt.Sprintf("%s:%d: %v", path, lineNo, err))
			continue
		}
		if ok {
			events = append(events, event)
		}
	}

	if err := scanner.Err(); err != nil {
		diags = append(diags, fmt.Sprintf("scan %s: %v", path, err))
	}

	return events, diags
}

// omnidevxEvent is the subset of an omnidevx event we need to parse.
type omnidevxEvent struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Timestamp  time.Time      `json:"timestamp"`
	Context    eventContext   `json:"context"`
	Attributes map[string]any `json:"attributes"`
}

type eventContext struct {
	SessionID string `json:"sessionId"`
	Workspace string `json:"workspace"`
}

// parseLine attempts to parse a JSONL line as a token event. Returns
// (event, true, nil) if the line is a matching ai.message.completed event,
// (Event{}, false, nil) if the line is valid but not a match, or
// (Event{}, false, err) if the line is malformed.
func (s *JSONLSource) parseLine(line []byte, q Query) (Event, bool, error) {
	var raw omnidevxEvent
	if err := json.Unmarshal(line, &raw); err != nil {
		return Event{}, false, fmt.Errorf("parse JSON: %w", err)
	}

	// Only process ai.message.completed events (these have token counts).
	if raw.Type != "ai.message.completed" {
		return Event{}, false, nil
	}

	// Check period filter.
	if !q.Period.Contains(raw.Timestamp) {
		return Event{}, false, nil
	}

	// Check session filter.
	if len(q.SessionIDs) > 0 && !contains(q.SessionIDs, raw.Context.SessionID) {
		return Event{}, false, nil
	}

	// Check workspace filter.
	if len(q.Workspaces) > 0 && !contains(q.Workspaces, raw.Context.Workspace) {
		return Event{}, false, nil
	}

	event := Event{
		ID:                  raw.ID,
		Timestamp:           raw.Timestamp,
		SessionID:           raw.Context.SessionID,
		Workspace:           raw.Context.Workspace,
		Model:               getString(raw.Attributes, "model"),
		InputTokens:         getInt64(raw.Attributes, "input_tokens"),
		OutputTokens:        getInt64(raw.Attributes, "output_tokens"),
		CacheReadTokens:     getInt64(raw.Attributes, "cache_read_tokens"),
		CacheCreationTokens: getInt64(raw.Attributes, "cache_creation_tokens"),
	}

	return event, true, nil
}

// daysInPeriod returns all calendar days (UTC) within the period.
func daysInPeriod(p Period) []time.Time {
	var days []time.Time
	start := time.Date(p.Start.Year(), p.Start.Month(), p.Start.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(p.End.Year(), p.End.Month(), p.End.Day(), 0, 0, 0, 0, time.UTC)

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}
	return days
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt64(m map[string]any, key string) int64 {
	switch v := m[key].(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	default:
		return 0
	}
}
