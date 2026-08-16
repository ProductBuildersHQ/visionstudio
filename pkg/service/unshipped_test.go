package service

import (
	"context"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestUnshippedQueue(t *testing.T) {
	s := store.NewMemStore()
	svc := New(s)
	ctx := context.Background()
	now := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
	old := now.AddDate(0, 0, -30)
	recent := now.AddDate(0, 0, -2)

	if err := s.CreateRepository(ctx, &store.Repository{ID: "github.com/x/r", Organization: "x", RepositoryName: "r", Status: "active"}); err != nil {
		t.Fatal(err)
	}
	inits := []*store.Initiative{
		{ID: "INIT-U-001", Title: "old unshipped", Status: "delivery_complete", Organization: "default", DeliveryCompleteAt: &old},
		{ID: "INIT-U-002", Title: "recent unshipped", Status: "delivery_complete", Organization: "default", DeliveryCompleteAt: &recent},
		{ID: "INIT-U-003", Title: "shipped", Status: "delivery_complete", Organization: "default", DeliveryCompleteAt: &old},
		{ID: "INIT-U-004", Title: "executing", Status: "executing", Organization: "default"},
		{ID: "INIT-U-005", Title: "no timestamp", Status: "releasing", Organization: "default"},
	}
	for _, in := range inits {
		if err := s.CreateInitiative(ctx, in); err != nil {
			t.Fatal(err)
		}
	}
	// INIT-U-003 has a release attached — off the queue.
	if _, err := svc.RecordRelease(ctx, "github.com/x/r", "v1.0.0", old, "", "", []string{"INIT-U-003"}, nil); err != nil {
		t.Fatal(err)
	}

	queue, err := svc.UnshippedQueue(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(queue) != 3 {
		t.Fatalf("queue = %d, want 3 (001, 002, 005)", len(queue))
	}
	// Unknown staleness first, then stalest.
	if queue[0].Initiative.ID != "INIT-U-005" {
		t.Fatalf("first = %s, want INIT-U-005 (unknown staleness)", queue[0].Initiative.ID)
	}
	if queue[1].Initiative.ID != "INIT-U-001" || queue[1].DaysStale != 30 {
		t.Fatalf("second = %s (%d days), want INIT-U-001 at 30", queue[1].Initiative.ID, queue[1].DaysStale)
	}
	if queue[2].Initiative.ID != "INIT-U-002" {
		t.Fatalf("third = %s, want INIT-U-002", queue[2].Initiative.ID)
	}
}
