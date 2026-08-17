package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/initiative"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// CreateInitiative creates a new initiative in "proposed" status.
// initType defaults to "feature" if empty.
// workflowID is optional; if provided, the workflow is selected for the initiative.
func (s *Service) CreateInitiative(ctx context.Context, id, org, title, description, priority, initType, workflowID string) (*store.Initiative, error) {
	if !initiative.ValidType(initType) {
		return nil, fmt.Errorf("invalid initiative type %q", initType)
	}
	if initType == "" {
		initType = initiative.TypeFeature
	}
	now := time.Now()
	init := &store.Initiative{
		ID:           id,
		Organization: org,
		Title:        title,
		Description:  description,
		Status:       initiative.StatusProposed,
		InitType:     initType,
		Priority:     priority,
		WorkflowID:   workflowID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.Store.CreateInitiative(ctx, init); err != nil {
		return nil, err
	}

	// If workflowID provided, also select it for the initiative
	if workflowID != "" {
		if err := s.Store.SelectWorkflowForInitiative(ctx, id, workflowID); err != nil {
			// Log but don't fail - initiative was created
			_ = err
		}
	}

	return init, nil
}

// GetInitiative returns an initiative by ID.
func (s *Service) GetInitiative(ctx context.Context, id string) (*store.Initiative, error) {
	return s.Store.GetInitiative(ctx, id)
}

// UpdateInitiative persists changes to an initiative.
func (s *Service) UpdateInitiative(ctx context.Context, init *store.Initiative) error {
	return s.Store.UpdateInitiative(ctx, init)
}

// ListInitiatives returns all initiatives.
func (s *Service) ListInitiatives(ctx context.Context) ([]*store.Initiative, error) {
	return s.Store.ListInitiatives(ctx)
}

// TransitionInitiative changes an initiative's lifecycle status,
// validating the transition and stamping the appropriate timestamp.
func (s *Service) TransitionInitiative(ctx context.Context, id, toStatus string) (*store.Initiative, error) {
	init, err := s.Store.GetInitiative(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if err := initiative.Transition(init, toStatus, now); err != nil {
		return nil, err
	}

	if err := s.Store.UpdateInitiative(ctx, init); err != nil {
		return nil, fmt.Errorf("save transition: %w", err)
	}
	return init, nil
}

// InitiativeDetail is an initiative with its phases and derived status.
type InitiativeDetail struct {
	Initiative *store.Initiative
	Phases     []PhaseDetail
	// Releases lists every repo@tag release this initiative is attached
	// to -- an initiative's RMIs can span multiple repos, so this can
	// have entries from more than one repository.
	Releases []*store.Release
}

// PhaseDetail is a phase with its derived status from member RMIs.
type PhaseDetail struct {
	Phase  *store.Phase
	Status string
	RMIs   []*store.RoadmapItem
}

// GetInitiativeDetail returns an initiative with phases and derived status.
func (s *Service) GetInitiativeDetail(ctx context.Context, id string) (*InitiativeDetail, error) {
	init, err := s.Store.GetInitiative(ctx, id)
	if err != nil {
		return nil, err
	}

	phases, err := s.Store.ListPhases(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list phases: %w", err)
	}

	allRMIs, err := s.Store.ListRMIs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list RMIs: %w", err)
	}

	rmisByPhase := make(map[string][]*store.RoadmapItem)
	for _, r := range allRMIs {
		rmisByPhase[r.PhaseID] = append(rmisByPhase[r.PhaseID], r)
	}

	releases, err := s.ListReleases(ctx, "", id)
	if err != nil {
		return nil, fmt.Errorf("list releases: %w", err)
	}

	detail := &InitiativeDetail{Initiative: init, Releases: releases}
	for _, p := range phases {
		phaseRMIs := rmisByPhase[p.ID]
		detail.Phases = append(detail.Phases, PhaseDetail{
			Phase:  p,
			Status: initiative.DerivePhaseStatus(phaseRMIs),
			RMIs:   phaseRMIs,
		})
	}
	return detail, nil
}

// CreatePhase adds a phase to an initiative.
func (s *Service) CreatePhase(ctx context.Context, id, initiativeID string, seq int, title, theme string) (*store.Phase, error) {
	if _, err := s.Store.GetInitiative(ctx, initiativeID); err != nil {
		return nil, fmt.Errorf("initiative %s: %w", initiativeID, err)
	}

	p := &store.Phase{
		ID:             id,
		InitiativeID:   initiativeID,
		SequenceNumber: seq,
		Title:          title,
		Theme:          theme,
	}
	if err := s.Store.CreatePhase(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ListPhases returns phases for an initiative.
func (s *Service) ListPhases(ctx context.Context, initiativeID string) ([]*store.Phase, error) {
	return s.Store.ListPhases(ctx, initiativeID)
}

// RemovePhase deletes a phase that has no member RMIs. Phases with members
// must have their RMIs moved (or the phase kept) — deletion never cascades.
func (s *Service) RemovePhase(ctx context.Context, phaseID string) error {
	parts := strings.SplitN(phaseID, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid phase ID %q (expected INITIATIVE-ID/phase-N)", phaseID)
	}
	rmis, err := s.Store.ListRMIs(ctx, parts[0])
	if err != nil {
		return fmt.Errorf("list RMIs: %w", err)
	}
	for _, r := range rmis {
		if r.PhaseID == phaseID {
			return fmt.Errorf("phase %s still has member RMIs (e.g. %s); move them first", phaseID, r.ID)
		}
	}
	return s.Store.DeletePhase(ctx, phaseID)
}
