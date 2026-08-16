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
		// forward skipping not allowed
		{StatusProposed, StatusExecuting, false},
		{StatusExecuting, StatusReleased, false},
		// backwards reopens to any earlier pipeline status
		{StatusExecuting, StatusPlanned, true},
		{StatusDeliveryComplete, StatusExecuting, true},
		{StatusDeliveryComplete, StatusProposed, true},
		{StatusReleasing, StatusExecuting, true},
		{StatusReleased, StatusExecuting, true},
		{StatusClosed, StatusProposed, true},
		{StatusClosed, StatusReleased, true},
		// backwards never fabricates later stages
		{StatusReleasing, StatusReleased, true}, // forward, for contrast
		{StatusDeliveryComplete, StatusReleased, false},
		// cancellation from any active state
		{StatusProposed, StatusCancelled, true},
		{StatusPlanned, StatusCancelled, true},
		{StatusExecuting, StatusCancelled, true},
		{StatusDeliveryComplete, StatusCancelled, true},
		{StatusReleasing, StatusCancelled, true},
		// released cannot be cancelled, only closed or reopened
		{StatusReleased, StatusCancelled, false},
		// cancelled reopens to any pre-release status, never to released/closed
		{StatusCancelled, StatusProposed, true},
		{StatusCancelled, StatusExecuting, true},
		{StatusCancelled, StatusReleasing, true},
		{StatusCancelled, StatusReleased, false},
		{StatusCancelled, StatusClosed, false},
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
		Status: StatusExecuting,
	}
	err := Transition(init, StatusReleased, time.Now())
	if err == nil {
		t.Fatal("expected error for forward-skipping transition")
	}
}

func TestTransitionBackwardsClearsLaterStamps(t *testing.T) {
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	init := &store.Initiative{
		ID:     "INIT-TEST-001",
		Status: StatusProposed,
	}

	// Walk forward to delivery_complete.
	for _, s := range []string{StatusPlanned, StatusExecuting, StatusDeliveryComplete} {
		if err := Transition(init, s, now); err != nil {
			t.Fatal(err)
		}
	}
	if init.DeliveryCompleteAt == nil {
		t.Fatal("expected DeliveryCompleteAt to be stamped")
	}

	// Scope grew: reopen to executing. The delivery-complete stamp must clear;
	// the earlier stamps must survive.
	later := now.Add(time.Hour)
	if err := Transition(init, StatusExecuting, later); err != nil {
		t.Fatal(err)
	}
	if init.Status != StatusExecuting {
		t.Fatalf("status = %q, want %q", init.Status, StatusExecuting)
	}
	if init.DeliveryCompleteAt != nil {
		t.Fatal("expected DeliveryCompleteAt cleared after reopening to executing")
	}
	if init.PlannedAt == nil {
		t.Fatal("expected PlannedAt to survive a reopen to executing")
	}
	if init.ExecutingAt == nil || !init.ExecutingAt.Equal(later) {
		t.Fatal("expected ExecutingAt re-stamped at the reopen time")
	}

	// Completing again re-stamps.
	if err := Transition(init, StatusDeliveryComplete, later); err != nil {
		t.Fatal(err)
	}
	if init.DeliveryCompleteAt == nil {
		t.Fatal("expected DeliveryCompleteAt re-stamped")
	}
}

func TestTransitionCancelPreservesStamps(t *testing.T) {
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	init := &store.Initiative{
		ID:     "INIT-TEST-001",
		Status: StatusProposed,
	}
	for _, s := range []string{StatusPlanned, StatusExecuting, StatusCancelled} {
		if err := Transition(init, s, now); err != nil {
			t.Fatal(err)
		}
	}
	if init.ExecutingAt == nil || init.PlannedAt == nil {
		t.Fatal("expected cancellation to preserve lifecycle stamps")
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
