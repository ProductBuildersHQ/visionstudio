package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// RMI status constants.
const (
	RMIStatusProposed   = "proposed"
	RMIStatusPlanned    = "planned"
	RMIStatusReady      = "ready"
	RMIStatusInProgress = "in_progress"
	RMIStatusCompleted  = "completed"
	RMIStatusBlocked    = "blocked"
	RMIStatusCancelled  = "cancelled"
)

var validRMIStatuses = map[string]bool{
	RMIStatusProposed:   true,
	RMIStatusPlanned:    true,
	RMIStatusReady:      true,
	RMIStatusInProgress: true,
	RMIStatusCompleted:  true,
	RMIStatusBlocked:    true,
	RMIStatusCancelled:  true,
}

// CreateRMI creates a new roadmap item in "proposed" status.
func (s *Service) CreateRMI(ctx context.Context, id, repoID, initiativeID, phaseID, title, description, itemType, priority string, required bool, seq int, acceptance []string) (*store.RoadmapItem, error) {
	if id == "" || repoID == "" || title == "" || itemType == "" {
		return nil, fmt.Errorf("id, repo, title, and type are required")
	}

	now := time.Now()
	rmi := &store.RoadmapItem{
		ID:                 id,
		RepositoryID:       repoID,
		InitiativeID:       initiativeID,
		PhaseID:            phaseID,
		Title:              title,
		Description:        description,
		ItemType:           itemType,
		Status:             RMIStatusProposed,
		Priority:           priority,
		Required:           required,
		SequenceNumber:     seq,
		AcceptanceCriteria: acceptance,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := s.Store.CreateRMI(ctx, rmi); err != nil {
		return nil, err
	}
	return rmi, nil
}

// GetRMI returns a roadmap item by ID.
func (s *Service) GetRMI(ctx context.Context, id string) (*store.RoadmapItem, error) {
	return s.Store.GetRMI(ctx, id)
}

// ListRMIs returns all RMIs for an initiative.
func (s *Service) ListRMIs(ctx context.Context, initiativeID string) ([]*store.RoadmapItem, error) {
	return s.Store.ListRMIs(ctx, initiativeID)
}

// ListRMIsByRepo returns all RMIs for a repository.
func (s *Service) ListRMIsByRepo(ctx context.Context, repoID string) ([]*store.RoadmapItem, error) {
	return s.Store.ListRMIsByRepo(ctx, repoID)
}

// ListAllRMIs returns every RMI across all initiatives and repositories. Used
// for cross-initiative discovery (e.g. keyword search to map a change to the
// RMI that describes it).
func (s *Service) ListAllRMIs(ctx context.Context) ([]*store.RoadmapItem, error) {
	return s.Store.ListAllRMIs(ctx)
}

// UpdateRMIStatus changes an RMI's status, stamping CompletedAt when appropriate.
func (s *Service) UpdateRMIStatus(ctx context.Context, id, status string) (*store.RoadmapItem, error) {
	if !validRMIStatuses[status] {
		return nil, fmt.Errorf("invalid RMI status: %q", status)
	}

	rmi, err := s.Store.GetRMI(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	rmi.Status = status
	rmi.UpdatedAt = now
	if status == RMIStatusCompleted {
		rmi.CompletedAt = &now
	}
	if err := s.Store.UpdateRMI(ctx, rmi); err != nil {
		return nil, fmt.Errorf("update RMI status: %w", err)
	}
	return rmi, nil
}

// UpdateRMI updates mutable fields of an existing RMI.
func (s *Service) UpdateRMI(ctx context.Context, rmi *store.RoadmapItem) error {
	rmi.UpdatedAt = time.Now()
	return s.Store.UpdateRMI(ctx, rmi)
}

// MoveRMI re-parents an RMI to another phase (and its initiative). The target
// phase ID must be of the form INITIATIVE-ID/phase-N; the initiative and phase
// must both exist. A seq > 0 also updates the RMI's sequence number.
func (s *Service) MoveRMI(ctx context.Context, id, phaseID string, seq int) (*store.RoadmapItem, error) {
	parts := strings.SplitN(phaseID, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid phase ID %q (expected INITIATIVE-ID/phase-N)", phaseID)
	}
	initiativeID := parts[0]

	if _, err := s.Store.GetInitiative(ctx, initiativeID); err != nil {
		return nil, fmt.Errorf("target initiative %s: %w", initiativeID, err)
	}
	phases, err := s.Store.ListPhases(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("list phases for %s: %w", initiativeID, err)
	}
	found := false
	for _, p := range phases {
		if p.ID == phaseID {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("phase %s not found in initiative %s", phaseID, initiativeID)
	}

	rmi, err := s.Store.GetRMI(ctx, id)
	if err != nil {
		return nil, err
	}
	rmi.InitiativeID = initiativeID
	rmi.PhaseID = phaseID
	if seq > 0 {
		rmi.SequenceNumber = seq
	}
	rmi.UpdatedAt = time.Now()
	if err := s.Store.UpdateRMI(ctx, rmi); err != nil {
		return nil, fmt.Errorf("move RMI %s: %w", id, err)
	}
	return rmi, nil
}

// CreateDependency adds a dependency edge between two RMIs.
func (s *Service) CreateDependency(ctx context.Context, sourceID, targetID, relationship string) error {
	if relationship == "" {
		relationship = "requires"
	}
	if relationship != "requires" && relationship != "relates" {
		return fmt.Errorf("invalid relationship: %q (must be requires or relates)", relationship)
	}
	dep := &store.RMIDependency{
		SourceRMIID:  sourceID,
		TargetRMIID:  targetID,
		Relationship: relationship,
	}
	if err := s.Store.CreateDependency(ctx, dep); err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") {
			return nil
		}
		return err
	}
	return nil
}

// UpdatePhaseStatus transitions all RMIs in a phase from one status to another.
// Returns the count of updated RMIs and the IDs that were skipped (already in target status).
func (s *Service) UpdatePhaseStatus(ctx context.Context, phaseID, fromStatus, toStatus string) (updated []string, skipped []string, err error) {
	if !validRMIStatuses[toStatus] {
		return nil, nil, fmt.Errorf("invalid target status: %q", toStatus)
	}

	parts := strings.SplitN(phaseID, "/", 2)
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("invalid phase ID %q (expected INITIATIVE-ID/phase-N)", phaseID)
	}

	allRMIs, err := s.Store.ListRMIs(ctx, parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("list RMIs: %w", err)
	}

	for _, r := range allRMIs {
		if r.PhaseID != phaseID {
			continue
		}
		if r.Status == toStatus {
			skipped = append(skipped, r.ID)
			continue
		}
		if fromStatus != "" && r.Status != fromStatus {
			skipped = append(skipped, r.ID)
			continue
		}
		if _, err := s.UpdateRMIStatus(ctx, r.ID, toStatus); err != nil {
			return updated, skipped, fmt.Errorf("update %s: %w", r.ID, err)
		}
		updated = append(updated, r.ID)
	}
	return updated, skipped, nil
}

// ListDependencies returns all dependency edges for an RMI.
func (s *Service) ListDependencies(ctx context.Context, rmiID string) ([]*store.RMIDependency, error) {
	return s.Store.ListDependencies(ctx, rmiID)
}

// RMIDetail is an RMI with its dependency edges.
type RMIDetail struct {
	RMI          *store.RoadmapItem
	Dependencies []*store.RMIDependency
}

// GetRMIDetail returns an RMI with its dependency edges.
func (s *Service) GetRMIDetail(ctx context.Context, id string) (*RMIDetail, error) {
	rmi, err := s.Store.GetRMI(ctx, id)
	if err != nil {
		return nil, err
	}
	deps, err := s.Store.ListDependencies(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list dependencies: %w", err)
	}
	return &RMIDetail{RMI: rmi, Dependencies: deps}, nil
}
