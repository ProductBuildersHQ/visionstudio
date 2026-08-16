package service

import (
	"context"
	"sort"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// UnshippedEntry is an initiative that claims delivery but has no release
// attached — the forcing-function queue for the roadmap board.
type UnshippedEntry struct {
	Initiative *store.Initiative
	// StaleSince is when the initiative reached delivery_complete (nil if
	// the timestamp was never recorded).
	StaleSince *time.Time
	DaysStale  int
}

// UnshippedQueue returns initiatives at delivery_complete or releasing
// with zero associated releases, stalest first. Working the queue to zero
// is what activates the acceptance-mark quality signal.
func (s *Service) UnshippedQueue(ctx context.Context, now time.Time) ([]*UnshippedEntry, error) {
	inits, err := s.Store.ListInitiatives(ctx)
	if err != nil {
		return nil, err
	}
	var queue []*UnshippedEntry
	for _, in := range inits {
		if in.Status != "delivery_complete" && in.Status != "releasing" {
			continue
		}
		rels, err := s.Store.ListReleasesByInitiative(ctx, in.ID)
		if err != nil {
			return nil, err
		}
		if len(rels) > 0 {
			continue
		}
		entry := &UnshippedEntry{Initiative: in, StaleSince: in.DeliveryCompleteAt}
		if entry.StaleSince != nil {
			entry.DaysStale = int(now.Sub(*entry.StaleSince).Hours() / 24)
		}
		queue = append(queue, entry)
	}
	sort.Slice(queue, func(i, j int) bool {
		a, b := queue[i].StaleSince, queue[j].StaleSince
		switch {
		case a == nil && b == nil:
			return queue[i].Initiative.ID < queue[j].Initiative.ID
		case a == nil:
			return true // unknown staleness sorts first: oldest unknowns are the riskiest
		case b == nil:
			return false
		case !a.Equal(*b):
			return a.Before(*b)
		}
		return queue[i].Initiative.ID < queue[j].Initiative.ID
	})
	return queue, nil
}
