package service

import (
	"context"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// WorkReadyFilters controls which RMIs are considered by WorkReady.
type WorkReadyFilters struct {
	InitiativeID string
	RepoID       string
}

// WorkReady returns RMIs that are ready to be worked on:
// status="ready", all "requires" dependencies completed, and no active assignment.
func (s *Service) WorkReady(ctx context.Context, filters WorkReadyFilters) ([]*store.RoadmapItem, error) {
	candidates, err := s.Store.ListRMIsByStatus(ctx, RMIStatusReady)
	if err != nil {
		return nil, err
	}

	if filters.InitiativeID != "" {
		var filtered []*store.RoadmapItem
		for _, r := range candidates {
			if r.InitiativeID == filters.InitiativeID {
				filtered = append(filtered, r)
			}
		}
		candidates = filtered
	}
	if filters.RepoID != "" {
		var filtered []*store.RoadmapItem
		for _, r := range candidates {
			if r.RepositoryID == filters.RepoID {
				filtered = append(filtered, r)
			}
		}
		candidates = filtered
	}

	var ready []*store.RoadmapItem
	for _, rmi := range candidates {
		blocked, err := s.isBlockedByDeps(ctx, rmi.ID)
		if err != nil {
			return nil, err
		}
		if blocked {
			continue
		}

		claimed, err := s.isClaimed(ctx, rmi.ID)
		if err != nil {
			return nil, err
		}
		if claimed {
			continue
		}

		ready = append(ready, rmi)
	}
	return ready, nil
}

func (s *Service) isBlockedByDeps(ctx context.Context, rmiID string) (bool, error) {
	deps, err := s.Store.ListDependencies(ctx, rmiID)
	if err != nil {
		return false, err
	}
	for _, d := range deps {
		if d.Relationship != "requires" {
			continue
		}
		if d.SourceRMIID != rmiID {
			continue
		}
		target, err := s.Store.GetRMI(ctx, d.TargetRMIID)
		if err != nil {
			return true, nil
		}
		if target.Status != RMIStatusCompleted {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) isClaimed(ctx context.Context, rmiID string) (bool, error) {
	a, err := s.Store.GetActiveAssignment(ctx, rmiID)
	if err != nil {
		return false, err
	}
	return a != nil, nil
}
