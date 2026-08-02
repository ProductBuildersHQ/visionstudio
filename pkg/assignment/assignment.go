// Package assignment implements lease-based work claims for agent sessions.
package assignment

import (
	"fmt"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

const (
	StatusActive    = "active"
	StatusReleased  = "released"
	StatusExpired   = "expired"
	StatusCompleted = "completed"
)

// DefaultLease is the default lease duration when none is specified.
const DefaultLease = 4 * time.Hour

// Claim creates an assignment for the given RMI and worker. It returns
// an error if the RMI already has an active, non-expired assignment.
func Claim(rmiID, worker string, lease time.Duration, now time.Time, existing *store.Assignment) (*store.Assignment, error) {
	if existing != nil && existing.Status == StatusActive && existing.LeaseExpiresAt.After(now) {
		return nil, fmt.Errorf("RMI %s already claimed by %s (expires %s)", rmiID, existing.Worker, existing.LeaseExpiresAt.Format(time.RFC3339))
	}
	if lease <= 0 {
		lease = DefaultLease
	}
	return &store.Assignment{
		RMIID:          rmiID,
		Worker:         worker,
		Status:         StatusActive,
		LeaseExpiresAt: now.Add(lease),
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// Renew extends an active assignment's lease. Returns an error if the
// assignment is not active.
func Renew(a *store.Assignment, lease time.Duration, now time.Time) error {
	if a.Status != StatusActive {
		return fmt.Errorf("cannot renew assignment in status %q", a.Status)
	}
	if lease <= 0 {
		lease = DefaultLease
	}
	a.LeaseExpiresAt = now.Add(lease)
	a.UpdatedAt = now
	return nil
}

// Release marks an active assignment as released.
func Release(a *store.Assignment, now time.Time) error {
	if a.Status != StatusActive {
		return fmt.Errorf("cannot release assignment in status %q", a.Status)
	}
	a.Status = StatusReleased
	a.UpdatedAt = now
	return nil
}

// Complete marks an active assignment as completed.
func Complete(a *store.Assignment, now time.Time) error {
	if a.Status != StatusActive {
		return fmt.Errorf("cannot complete assignment in status %q", a.Status)
	}
	a.Status = StatusCompleted
	a.CompletedAt = &now
	a.UpdatedAt = now
	return nil
}

// ExpireStale transitions any active assignment whose lease has passed to expired.
func ExpireStale(a *store.Assignment, now time.Time) bool {
	if a.Status == StatusActive && !a.LeaseExpiresAt.After(now) {
		a.Status = StatusExpired
		a.UpdatedAt = now
		return true
	}
	return false
}

// TrailerLine returns the git trailer string for a claimed RMI.
func TrailerLine(rmiID string) string {
	return fmt.Sprintf("Refs: %s", rmiID)
}
