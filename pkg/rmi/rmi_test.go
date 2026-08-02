package rmi

import (
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestValidID(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"RMI-PRISMCONTROL-001", true},
		{"RMI-OMNIDEVXCORE-003", true},
		{"RMI-VIDEOASCODE-019", true},
		{"rmi-lower-001", false},
		{"RMI-MISSING", false},
		{"RMI--001", false},
		{"INIT-FOO-001", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := ValidID(tt.id); got != tt.want {
				t.Fatalf("ValidID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestDependencyGraphIsBlocked(t *testing.T) {
	deps := []*store.RMIDependency{
		{SourceRMIID: "RMI-A-002", TargetRMIID: "RMI-A-001", Relationship: "requires"},
	}
	g := NewDependencyGraph(deps)

	// blocked: dependency not completed
	statuses := map[string]string{
		"RMI-A-001": StatusPlanned,
		"RMI-A-002": StatusPlanned,
	}
	if !g.IsBlocked("RMI-A-002", statuses) {
		t.Fatal("expected RMI-A-002 to be blocked")
	}

	// unblocked: dependency completed
	statuses["RMI-A-001"] = StatusCompleted
	if g.IsBlocked("RMI-A-002", statuses) {
		t.Fatal("expected RMI-A-002 to be unblocked")
	}
}

func TestDetectCycles(t *testing.T) {
	t.Run("no cycle", func(t *testing.T) {
		deps := []*store.RMIDependency{
			{SourceRMIID: "RMI-A-002", TargetRMIID: "RMI-A-001", Relationship: "requires"},
			{SourceRMIID: "RMI-A-003", TargetRMIID: "RMI-A-002", Relationship: "requires"},
		}
		g := NewDependencyGraph(deps)
		if cycle := g.DetectCycles(); cycle != nil {
			t.Fatalf("expected no cycle, got %v", cycle)
		}
	})

	t.Run("has cycle", func(t *testing.T) {
		deps := []*store.RMIDependency{
			{SourceRMIID: "RMI-A-001", TargetRMIID: "RMI-A-002", Relationship: "requires"},
			{SourceRMIID: "RMI-A-002", TargetRMIID: "RMI-A-001", Relationship: "requires"},
		}
		g := NewDependencyGraph(deps)
		if cycle := g.DetectCycles(); cycle == nil {
			t.Fatal("expected cycle")
		}
	})
}

func TestReadyWork(t *testing.T) {
	rmis := []*store.RoadmapItem{
		{ID: "RMI-A-001", Status: StatusPlanned, Required: true},
		{ID: "RMI-A-002", Status: StatusPlanned, Required: true},
		{ID: "RMI-A-003", Status: StatusCompleted, Required: true},
		{ID: "RMI-A-004", Status: StatusPlanned, Required: true},
	}
	deps := []*store.RMIDependency{
		{SourceRMIID: "RMI-A-002", TargetRMIID: "RMI-A-001", Relationship: "requires"},
	}
	activeAssignments := map[string]bool{
		"RMI-A-004": true,
	}

	ready := ReadyWork(rmis, deps, activeAssignments)

	// RMI-A-001: planned, no deps, not assigned -> ready
	// RMI-A-002: planned, dep on 001 (not completed) -> blocked
	// RMI-A-003: completed -> excluded
	// RMI-A-004: planned, no deps, but assigned -> excluded
	if len(ready) != 1 || ready[0] != "RMI-A-001" {
		t.Fatalf("expected [RMI-A-001], got %v", ready)
	}
}
