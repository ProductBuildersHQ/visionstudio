package assignment

import (
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

var baseTime = time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)

func TestClaimNewAssignment(t *testing.T) {
	a, err := Claim("RMI-A-001", "session-1", 2*time.Hour, baseTime, nil)
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != StatusActive {
		t.Fatalf("expected active, got %s", a.Status)
	}
	if a.Worker != "session-1" {
		t.Fatalf("expected session-1, got %s", a.Worker)
	}
	expectedExpiry := baseTime.Add(2 * time.Hour)
	if !a.LeaseExpiresAt.Equal(expectedExpiry) {
		t.Fatalf("expected expiry %v, got %v", expectedExpiry, a.LeaseExpiresAt)
	}
}

func TestClaimConflict(t *testing.T) {
	existing := &store.Assignment{
		RMIID:          "RMI-A-001",
		Worker:         "session-1",
		Status:         StatusActive,
		LeaseExpiresAt: baseTime.Add(2 * time.Hour),
	}
	_, err := Claim("RMI-A-001", "session-2", 2*time.Hour, baseTime, existing)
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestClaimAfterLeaseExpiry(t *testing.T) {
	expired := &store.Assignment{
		RMIID:          "RMI-A-001",
		Worker:         "session-1",
		Status:         StatusActive,
		LeaseExpiresAt: baseTime.Add(-1 * time.Hour),
	}
	a, err := Claim("RMI-A-001", "session-2", 2*time.Hour, baseTime, expired)
	if err != nil {
		t.Fatal(err)
	}
	if a.Worker != "session-2" {
		t.Fatalf("expected session-2, got %s", a.Worker)
	}
}

func TestClaimDefaultLease(t *testing.T) {
	a, err := Claim("RMI-A-001", "session-1", 0, baseTime, nil)
	if err != nil {
		t.Fatal(err)
	}
	expected := baseTime.Add(DefaultLease)
	if !a.LeaseExpiresAt.Equal(expected) {
		t.Fatalf("expected default lease expiry %v, got %v", expected, a.LeaseExpiresAt)
	}
}

func TestRenew(t *testing.T) {
	a := &store.Assignment{Status: StatusActive, LeaseExpiresAt: baseTime}
	if err := Renew(a, 3*time.Hour, baseTime); err != nil {
		t.Fatal(err)
	}
	expected := baseTime.Add(3 * time.Hour)
	if !a.LeaseExpiresAt.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, a.LeaseExpiresAt)
	}
}

func TestRenewNonActive(t *testing.T) {
	a := &store.Assignment{Status: StatusReleased}
	if err := Renew(a, time.Hour, baseTime); err == nil {
		t.Fatal("expected error renewing non-active assignment")
	}
}

func TestExpireStale(t *testing.T) {
	a := &store.Assignment{
		Status:         StatusActive,
		LeaseExpiresAt: baseTime.Add(-1 * time.Minute),
	}
	if !ExpireStale(a, baseTime) {
		t.Fatal("expected assignment to be expired")
	}
	if a.Status != StatusExpired {
		t.Fatalf("expected expired, got %s", a.Status)
	}
}

func TestExpireStaleNotExpired(t *testing.T) {
	a := &store.Assignment{
		Status:         StatusActive,
		LeaseExpiresAt: baseTime.Add(1 * time.Hour),
	}
	if ExpireStale(a, baseTime) {
		t.Fatal("expected assignment not to be expired")
	}
}

func TestTrailerLine(t *testing.T) {
	got := TrailerLine("RMI-PRISMCONTROL-005")
	want := "Refs: RMI-PRISMCONTROL-005"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
