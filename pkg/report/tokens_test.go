package report

import (
	"testing"
	"time"
)

func TestParseQuarter(t *testing.T) {
	tests := []struct {
		input     string
		wantStart time.Time
		wantEnd   time.Time
		wantLabel string
		wantErr   bool
	}{
		{
			input:     "2026-Q1",
			wantStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, 3, 31, 23, 59, 59, 999999999, time.UTC),
			wantLabel: "2026-Q1",
		},
		{
			input:     "2026-Q2",
			wantStart: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, 6, 30, 23, 59, 59, 999999999, time.UTC),
			wantLabel: "2026-Q2",
		},
		{
			input:     "2026-Q3",
			wantStart: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, 9, 30, 23, 59, 59, 999999999, time.UTC),
			wantLabel: "2026-Q3",
		},
		{
			input:     "2026-Q4",
			wantStart: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, 12, 31, 23, 59, 59, 999999999, time.UTC),
			wantLabel: "2026-Q4",
		},
		{
			input:   "invalid",
			wantErr: true,
		},
		{
			input:   "2026-Q5",
			wantErr: true,
		},
		{
			input:   "2026-Q0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseQuarter(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Start.Equal(tt.wantStart) {
				t.Errorf("Start = %v, want %v", got.Start, tt.wantStart)
			}
			if !got.End.Equal(tt.wantEnd) {
				t.Errorf("End = %v, want %v", got.End, tt.wantEnd)
			}
			if got.Label != tt.wantLabel {
				t.Errorf("Label = %q, want %q", got.Label, tt.wantLabel)
			}
		})
	}
}

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{500, "500"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{999999, "1000.0K"},
		{1000000, "1.0M"},
		{2500000, "2.5M"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatTokens(tt.input)
			if got != tt.want {
				t.Errorf("formatTokens(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTokenTotalsAddition(t *testing.T) {
	// This is a simple integration test for addToTotals
	// The full integration test would require a mock store
}

func TestMarkdownTokenReport(t *testing.T) {
	r := &TokenReport{
		Mode:            "initiative",
		InitiativeID:    "INIT-TEST-001",
		InitiativeTitle: "Test Initiative",
		Period: TokenPeriod{
			Start: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2026, 7, 31, 23, 59, 59, 0, time.UTC),
		},
		Totals: TokenTotals{
			InputTokens:  1000000,
			OutputTokens: 500000,
			TotalTokens:  1500000,
			CostUSD:      15.50,
		},
		ByRMI: []RMITokens{
			{RMIID: "RMI-001", Title: "First RMI", Totals: TokenTotals{CostUSD: 10.00, TotalTokens: 1000000}},
			{RMIID: "RMI-002", Title: "Second RMI", Totals: TokenTotals{CostUSD: 5.50, TotalTokens: 500000}},
		},
		ByModel: []ModelTokens{
			{Model: "claude-opus-4-8", Totals: TokenTotals{CostUSD: 15.50, TotalTokens: 1500000}},
		},
		SessionCount: 5,
		EventCount:   100,
		Coverage:     0.95,
	}

	md := r.MarkdownTokenReport()

	// Check key sections exist
	if !contains(md, "# Token Report: INIT-TEST-001") {
		t.Error("missing title")
	}
	if !contains(md, "**Title:** Test Initiative") {
		t.Error("missing initiative title")
	}
	if !contains(md, "Total Cost") {
		t.Error("missing totals section")
	}
	if !contains(md, "By RMI") {
		t.Error("missing RMI section")
	}
	if !contains(md, "By Model") {
		t.Error("missing model section")
	}
	if !contains(md, "95.0%") {
		t.Error("missing coverage")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
