// Package initiative implements initiative lifecycle management and
// derived phase status computation.
package initiative

import (
	"fmt"
	"slices"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// Lifecycle statuses in order.
const (
	StatusProposed         = "proposed"
	StatusPlanned          = "planned"
	StatusExecuting        = "executing"
	StatusDeliveryComplete = "delivery_complete"
	StatusReleasing        = "releasing"
	StatusReleased         = "released"
	StatusClosed           = "closed"
	StatusCancelled        = "cancelled"
)

// Initiative types. Type determines the default spec workflow.
const (
	TypeFeature     = "feature"
	TypeMaintenance = "maintenance"
	TypeMigration   = "migration"
	TypeCompliance  = "compliance"
	TypeRefactor    = "refactor"
)

// Types lists all valid initiative types.
var Types = []string{TypeFeature, TypeMaintenance, TypeMigration, TypeCompliance, TypeRefactor}

// ValidType reports whether t is a recognized initiative type.
// The empty string is valid and means the default (feature).
func ValidType(t string) bool {
	return t == "" || slices.Contains(Types, t)
}

var forwardTransitions = map[string][]string{
	StatusProposed:         {StatusPlanned, StatusCancelled},
	StatusPlanned:          {StatusExecuting, StatusCancelled},
	StatusExecuting:        {StatusDeliveryComplete, StatusCancelled},
	StatusDeliveryComplete: {StatusReleasing, StatusClosed, StatusCancelled},
	StatusReleasing:        {StatusReleased, StatusCancelled},
	StatusReleased:         {StatusClosed},
	StatusClosed:           {},
	StatusCancelled:        {},
}

// ValidTransition reports whether transitioning from one status to another is allowed.
func ValidTransition(from, to string) bool {
	targets, ok := forwardTransitions[from]
	if !ok {
		return false
	}
	return slices.Contains(targets, to)
}

// Transition updates the initiative status and stamps the appropriate lifecycle
// timestamp. It returns an error if the transition is invalid.
func Transition(init *store.Initiative, to string, now time.Time) error {
	if !ValidTransition(init.Status, to) {
		return fmt.Errorf("invalid transition from %q to %q", init.Status, to)
	}
	init.Status = to
	init.UpdatedAt = now
	switch to {
	case StatusPlanned:
		init.PlannedAt = &now
	case StatusExecuting:
		init.ExecutingAt = &now
	case StatusDeliveryComplete:
		init.DeliveryCompleteAt = &now
	case StatusReleased:
		init.ReleasedAt = &now
	case StatusClosed:
		init.ClosedAt = &now
	}
	return nil
}

// Phase status constants (derived, never stored).
const (
	PhaseCompleted  = "completed"
	PhaseInProgress = "in_progress"
	PhaseBlocked    = "blocked"
	PhasePlanned    = "planned"
	PhasePartial    = "partial"
)

// DerivePhaseStatus computes phase status from its member RMIs.
func DerivePhaseStatus(rmis []*store.RoadmapItem) string {
	if len(rmis) == 0 {
		return PhasePlanned
	}

	var requiredTotal, requiredCompleted, anyActive, anyBlocked int
	for _, r := range rmis {
		if r.Required {
			requiredTotal++
			if r.Status == "completed" {
				requiredCompleted++
			}
		}
		switch r.Status {
		case "in_progress", "ready":
			anyActive++
		case "blocked":
			anyBlocked++
		}
	}

	if anyBlocked > 0 {
		return PhaseBlocked
	}
	if requiredTotal > 0 && requiredCompleted == requiredTotal {
		allComplete := true
		for _, r := range rmis {
			if r.Status != "completed" && r.Status != "cancelled" {
				allComplete = false
				break
			}
		}
		if allComplete {
			return PhaseCompleted
		}
		return PhasePartial
	}
	if anyActive > 0 || requiredCompleted > 0 {
		return PhaseInProgress
	}
	return PhasePlanned
}
