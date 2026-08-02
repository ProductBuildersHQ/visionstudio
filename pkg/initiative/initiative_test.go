package initiative

import (
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestValidTransition(t *testing.T) {
	tests := []struct {
		from string
		to   string
		want bool
	}{
		{StatusProposed, StatusPlanned, true},
		{StatusPlanned, StatusExecuting, true},
		{StatusExecuting, StatusDeliveryComplete, true},
		{StatusDeliveryComplete, StatusReleasing, true},
		{StatusReleasing, StatusReleased, true},
		{StatusReleased, StatusClosed, true},
		// backwards not allowed
		{StatusExecuting, StatusPlanned, false},
		{StatusReleased, StatusExecuting, false},
		// cancellation from any active state
		{StatusProposed, StatusCancelled, true},
		{StatusPlanned, StatusCancelled, true},
		{StatusExecuting, StatusCancelled, true},
		{StatusDeliveryComplete, StatusCancelled, true},
		{StatusReleasing, StatusCancelled, true},
		// terminal states have no transitions
		{StatusClosed, StatusProposed, false},
		{StatusCancelled, StatusProposed, false},
	}
	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			got := ValidTransition(tt.from, tt.to)
			if got != tt.want {
				t.Fatalf("ValidTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestTransitionStampsTimestamp(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	init := &store.Initiative{
		ID:     "INIT-TEST-001",
		Status: StatusProposed,
	}

	if err := Transition(init, StatusPlanned, now); err != nil {
		t.Fatal(err)
	}
	if init.PlannedAt == nil || !init.PlannedAt.Equal(now) {
		t.Fatal("expected PlannedAt to be stamped")
	}

	if err := Transition(init, StatusExecuting, now); err != nil {
		t.Fatal(err)
	}
	if init.ExecutingAt == nil || !init.ExecutingAt.Equal(now) {
		t.Fatal("expected ExecutingAt to be stamped")
	}
}

func TestTransitionInvalidReturnsError(t *testing.T) {
	init := &store.Initiative{
		ID:     "INIT-TEST-001",
		Status: StatusClosed,
	}
	err := Transition(init, StatusExecuting, time.Now())
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
}

func TestDerivePhaseStatus(t *testing.T) {
	tests := []struct {
		name string
		rmis []*store.RoadmapItem
		want string
	}{
		{
			name: "empty",
			rmis: nil,
			want: PhasePlanned,
		},
		{
			name: "all required completed, no optional",
			rmis: []*store.RoadmapItem{
				{Status: "completed", Required: true},
				{Status: "completed", Required: true},
			},
			want: PhaseCompleted,
		},
		{
			name: "all required completed, optional open",
			rmis: []*store.RoadmapItem{
				{Status: "completed", Required: true},
				{Status: "in_progress", Required: false},
			},
			want: PhasePartial,
		},
		{
			name: "one blocked",
			rmis: []*store.RoadmapItem{
				{Status: "completed", Required: true},
				{Status: "blocked", Required: true},
			},
			want: PhaseBlocked,
		},
		{
			name: "some active",
			rmis: []*store.RoadmapItem{
				{Status: "completed", Required: true},
				{Status: "in_progress", Required: true},
			},
			want: PhaseInProgress,
		},
		{
			name: "all planned",
			rmis: []*store.RoadmapItem{
				{Status: "planned", Required: true},
				{Status: "planned", Required: true},
			},
			want: PhasePlanned,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DerivePhaseStatus(tt.rmis)
			if got != tt.want {
				t.Fatalf("DerivePhaseStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
