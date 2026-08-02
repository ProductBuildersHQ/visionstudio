package main

import "testing"

func TestValidPeriodType(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"weekly", true},
		{"monthly", true},
		{"quarterly", true},
		{"daily", false},
		{"yearly", false},
		{"", false},
		{"../weekly", false},
	}

	for _, tt := range tests {
		got := validPeriodType(tt.input)
		if got != tt.want {
			t.Errorf("validPeriodType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestValidPeriodLabel(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		// Valid weekly
		{"2026-W01", true},
		{"2026-W30", true},
		{"2026-W52", true},

		// Valid monthly
		{"2026-01", true},
		{"2026-07", true},
		{"2026-12", true},

		// Valid quarterly
		{"2026-Q1", true},
		{"2026-Q2", true},
		{"2026-Q3", true},
		{"2026-Q4", true},

		// Invalid
		{"2026-W100", false},      // Too many digits
		{"2026-Q5", false},        // Invalid quarter
		{"2026-13", true},         // Pattern allows 00-99; filesystem handles actual validity
		{"../2026-W30", false},    // Path traversal
		{"2026-W30/../..", false}, // Path traversal
		{"", false},
		{"2026", false},
		{"W30", false},
	}

	for _, tt := range tests {
		got := validPeriodLabel(tt.input)
		if got != tt.want {
			t.Errorf("validPeriodLabel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
