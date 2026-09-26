// Package initiative implements initiative lifecycle management and
// derived phase status computation.
package initiative

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
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

// idParsePattern captures the slug and sequence number of an initiative ID for
// allocation. It accepts three-or-more digits so it keeps working past 999.
var idParsePattern = regexp.MustCompile(`^INIT-([A-Z0-9]+)-(\d{3,})$`)

// ParseID splits an initiative ID into its slug and sequence number. ok is
// false if id is not a well-formed initiative ID.
func ParseID(id string) (slug string, num int, ok bool) {
	m := idParsePattern.FindStringSubmatch(id)
	if m == nil {
		return "", 0, false
	}
	n, err := strconv.Atoi(m[2])
	if err != nil {
		return "", 0, false
	}
	return m[1], n, true
}

// FormatID builds an initiative ID from a slug and sequence number,
// zero-padding the number to three digits.
func FormatID(slug string, num int) string {
	return fmt.Sprintf("INIT-%s-%03d", slug, num)
}

// NormalizeSlug uppercases s and strips every non-alphanumeric character,
// matching the slug form used in initiative and RMI IDs. If s looks like a
// repository ID (contains '/'), only its final path segment is used.
func NormalizeSlug(s string) string {
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[i+1:]
	}
	var b strings.Builder
	for _, r := range strings.ToUpper(s) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
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

// pipelineRank orders the pipeline statuses for backwards-transition checks.
// Cancelled sits outside the pipeline and has no rank.
var pipelineRank = map[string]int{
	StatusProposed:         0,
	StatusPlanned:          1,
	StatusExecuting:        2,
	StatusDeliveryComplete: 3,
	StatusReleasing:        4,
	StatusReleased:         5,
	StatusClosed:           6,
}

// ValidTransition reports whether transitioning from one status to another is
// allowed. Forward transitions follow the pipeline one step at a time (plus
// cancellation from any active status). Backwards transitions reopen the
// initiative to any earlier pipeline status — an initiative's scope can grow
// after delivery or release (new phases land in its roadmap), and the status
// should be able to follow it back. A cancelled initiative reopens the same
// way, to any pre-release status; it cannot jump to released or closed, which
// would fabricate lifecycle history that never happened.
func ValidTransition(from, to string) bool {
	if targets, ok := forwardTransitions[from]; ok && slices.Contains(targets, to) {
		return true
	}
	toRank, ok := pipelineRank[to]
	if !ok || to == StatusClosed {
		// Closed is reachable only through the forward map.
		return false
	}
	if from == StatusCancelled {
		return toRank < pipelineRank[StatusReleased]
	}
	fromRank, ok := pipelineRank[from]
	if !ok {
		return false
	}
	return toRank < fromRank
}

// Transition updates the initiative status and stamps the appropriate lifecycle
// timestamp. It returns an error if the transition is invalid. A backwards
// transition clears the stamps of every stage later than the new status — an
// initiative reopened to executing is no longer delivery-complete, and its
// timestamps must not claim otherwise; re-entering a stage later re-stamps it.
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
	if rank, ok := pipelineRank[to]; ok {
		clearStampsAfter(init, rank)
	}
	return nil
}

// clearStampsAfter clears lifecycle timestamps for pipeline stages ranked
// strictly later than rank. Cancellation (no rank) preserves all stamps.
func clearStampsAfter(init *store.Initiative, rank int) {
	if rank < pipelineRank[StatusClosed] {
		init.ClosedAt = nil
	}
	if rank < pipelineRank[StatusReleased] {
		init.ReleasedAt = nil
	}
	if rank < pipelineRank[StatusDeliveryComplete] {
		init.DeliveryCompleteAt = nil
	}
	if rank < pipelineRank[StatusExecuting] {
		init.ExecutingAt = nil
	}
	if rank < pipelineRank[StatusPlanned] {
		init.PlannedAt = nil
	}
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

	var requiredTotal, requiredResolved, requiredCompleted, anyActive, anyBlocked int
	for _, r := range rmis {
		if r.Required {
			requiredTotal++
			switch r.Status {
			case "completed":
				requiredCompleted++
				requiredResolved++
			case "cancelled":
				requiredResolved++
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
	// A phase is resolved once every required RMI is either completed or
	// cancelled -- cancelled work leaves nothing pending, the same as done.
	if requiredTotal > 0 && requiredResolved == requiredTotal {
		allResolved := true
		for _, r := range rmis {
			if r.Status != "completed" && r.Status != "cancelled" {
				allResolved = false
				break
			}
		}
		if allResolved {
			return PhaseCompleted
		}
		return PhasePartial
	}
	if anyActive > 0 || requiredCompleted > 0 {
		return PhaseInProgress
	}
	return PhasePlanned
}
