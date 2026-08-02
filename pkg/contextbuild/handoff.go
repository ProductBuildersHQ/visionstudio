package contextbuild

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/initiative"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// PhaseHandoffProjection is the complete handoff artifact for a phase.
// It aggregates member-RMI assignment handoffs, evidence, and derived status.
type PhaseHandoffProjection struct {
	// Version is the schema version for forwards compatibility.
	Version string `json:"version"`

	// PhaseID is the phase this handoff is for.
	PhaseID string `json:"phase_id"`

	// PhaseTitle is the human-readable phase title.
	PhaseTitle string `json:"phase_title"`

	// Theme is the phase's grouping rationale.
	Theme string `json:"theme,omitempty"`

	// InitiativeID is the parent initiative.
	InitiativeID string `json:"initiative_id"`

	// DerivedStatus is the phase status computed from member RMIs.
	DerivedStatus string `json:"derived_status"`

	// BuildTimestamp is when this projection was assembled.
	BuildTimestamp time.Time `json:"build_timestamp"`

	// Provenance tracks the Dolt revision used.
	Provenance Provenance `json:"provenance"`

	// Summary contains aggregated counts and completion stats.
	Summary HandoffSummary `json:"summary"`

	// RMIHandoffs contains per-RMI handoff data, ordered by sequence.
	RMIHandoffs []RMIHandoffEntry `json:"rmi_handoffs"`

	// AggregatedDecisions collects all decisions across RMIs.
	AggregatedDecisions []string `json:"aggregated_decisions,omitempty"`

	// Evidence contains all delivery evidence for the phase.
	Evidence []EvidenceEntry `json:"evidence,omitempty"`
}

// HandoffSummary contains aggregated statistics for the phase.
type HandoffSummary struct {
	TotalRMIs     int `json:"total_rmis"`
	CompletedRMIs int `json:"completed_rmis"`
	RequiredRMIs  int `json:"required_rmis"`
	RequiredDone  int `json:"required_done"`
	InProgress    int `json:"in_progress"`
	Blocked       int `json:"blocked"`
}

// RMIHandoffEntry contains handoff data for a single RMI.
type RMIHandoffEntry struct {
	RMIID    string `json:"rmi_id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Required bool   `json:"required"`
	Sequence int    `json:"sequence"`
	RepoID   string `json:"repo_id"`

	// Handoff is the structured handoff from the most recent assignment.
	Handoff *HandoffData `json:"handoff,omitempty"`

	// EvidenceCount is the number of evidence items for this RMI.
	EvidenceCount int `json:"evidence_count"`
}

// EvidenceEntry is a delivery evidence item with RMI context.
type EvidenceEntry struct {
	RMIID      string     `json:"rmi_id"`
	RMITitle   string     `json:"rmi_title"`
	Type       string     `json:"type"`
	Reference  string     `json:"reference"`
	OccurredAt *time.Time `json:"occurred_at,omitempty"`
}

// BuildPhaseHandoff assembles a handoff projection for a phase.
func (b *Builder) BuildPhaseHandoff(ctx context.Context, phaseID string) (*PhaseHandoffProjection, error) {
	parts := strings.Split(phaseID, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid phase ID format: %s", phaseID)
	}
	initiativeID := parts[0]

	phases, err := b.store.ListPhases(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("list phases: %w", err)
	}

	var targetPhase *store.Phase
	for _, p := range phases {
		if p.ID == phaseID {
			targetPhase = p
			break
		}
	}
	if targetPhase == nil {
		return nil, fmt.Errorf("phase not found: %s", phaseID)
	}

	rmis, err := b.store.ListRMIs(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("list RMIs: %w", err)
	}

	phaseRMIs := filterRMIsByPhase(rmis, phaseID)

	allAssignments, err := b.store.ListAllAssignments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list assignments: %w", err)
	}

	allEvidence, err := b.store.ListEvidenceByInitiative(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("list evidence: %w", err)
	}

	proj := &PhaseHandoffProjection{
		Version:        SchemaVersion,
		PhaseID:        phaseID,
		PhaseTitle:     targetPhase.Title,
		Theme:          targetPhase.Theme,
		InitiativeID:   initiativeID,
		DerivedStatus:  initiative.DerivePhaseStatus(phaseRMIs),
		BuildTimestamp: time.Now().UTC().Truncate(time.Second),
		Provenance:     Provenance{Source: "dolt", Revision: b.doltCommit},
	}

	rmiHandoffs, decisions := buildRMIHandoffs(phaseRMIs, allAssignments, allEvidence)
	proj.RMIHandoffs = rmiHandoffs
	proj.AggregatedDecisions = decisions

	proj.Summary = computeSummary(phaseRMIs)

	proj.Evidence = buildEvidenceEntries(phaseRMIs, allEvidence)

	return proj, nil
}

func buildRMIHandoffs(rmis []*store.RoadmapItem, assignments []*store.Assignment, evidence []*store.DeliveryEvidence) ([]RMIHandoffEntry, []string) {
	assignmentByRMI := make(map[string]*store.Assignment)
	for _, a := range assignments {
		existing, ok := assignmentByRMI[a.RMIID]
		if !ok || a.CreatedAt.After(existing.CreatedAt) {
			assignmentByRMI[a.RMIID] = a
		}
	}

	evidenceCountByRMI := make(map[string]int)
	for _, ev := range evidence {
		evidenceCountByRMI[ev.RMIID]++
	}

	var entries []RMIHandoffEntry
	var allDecisions []string

	for _, rmi := range rmis {
		entry := RMIHandoffEntry{
			RMIID:         rmi.ID,
			Title:         rmi.Title,
			Status:        rmi.Status,
			Required:      rmi.Required,
			Sequence:      rmi.SequenceNumber,
			RepoID:        rmi.RepositoryID,
			EvidenceCount: evidenceCountByRMI[rmi.ID],
		}

		if a, ok := assignmentByRMI[rmi.ID]; ok && a.Handoff != nil {
			entry.Handoff = &HandoffData{
				Completed:  a.Handoff.Completed,
				Remaining:  a.Handoff.Remaining,
				Decisions:  a.Handoff.Decisions,
				NextAction: a.Handoff.NextAction,
			}
			allDecisions = append(allDecisions, a.Handoff.Decisions...)
		}

		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Sequence != entries[j].Sequence {
			return entries[i].Sequence < entries[j].Sequence
		}
		return entries[i].RMIID < entries[j].RMIID
	})

	return entries, allDecisions
}

func computeSummary(rmis []*store.RoadmapItem) HandoffSummary {
	var s HandoffSummary
	s.TotalRMIs = len(rmis)

	for _, rmi := range rmis {
		if rmi.Required {
			s.RequiredRMIs++
			if rmi.Status == "completed" {
				s.RequiredDone++
			}
		}
		switch rmi.Status {
		case "completed":
			s.CompletedRMIs++
		case "in_progress":
			s.InProgress++
		case "blocked":
			s.Blocked++
		}
	}

	return s
}

func buildEvidenceEntries(rmis []*store.RoadmapItem, evidence []*store.DeliveryEvidence) []EvidenceEntry {
	rmiTitles := make(map[string]string)
	rmiIDs := make(map[string]bool)
	for _, rmi := range rmis {
		rmiTitles[rmi.ID] = rmi.Title
		rmiIDs[rmi.ID] = true
	}

	var entries []EvidenceEntry
	for _, ev := range evidence {
		if !rmiIDs[ev.RMIID] {
			continue
		}
		entries = append(entries, EvidenceEntry{
			RMIID:      ev.RMIID,
			RMITitle:   rmiTitles[ev.RMIID],
			Type:       ev.EvidenceType,
			Reference:  ev.Reference,
			OccurredAt: ev.OccurredAt,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].RMIID != entries[j].RMIID {
			return entries[i].RMIID < entries[j].RMIID
		}
		if entries[i].Type != entries[j].Type {
			return entries[i].Type < entries[j].Type
		}
		return entries[i].Reference < entries[j].Reference
	})

	return entries
}

// RenderHandoffMarkdown renders the handoff projection as Markdown.
func (p *PhaseHandoffProjection) RenderMarkdown() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "# Phase Handoff: %s\n\n", p.PhaseTitle)
	if p.Theme != "" {
		fmt.Fprintf(&buf, "**Theme:** %s\n\n", p.Theme)
	}
	fmt.Fprintf(&buf, "**Phase ID:** `%s`\n", p.PhaseID)
	fmt.Fprintf(&buf, "**Initiative:** `%s`\n", p.InitiativeID)
	fmt.Fprintf(&buf, "**Status:** %s\n", p.DerivedStatus)
	fmt.Fprintf(&buf, "**Generated:** %s\n\n", p.BuildTimestamp.Format(time.RFC3339))

	fmt.Fprintf(&buf, "## Summary\n\n")
	fmt.Fprintf(&buf, "| Metric | Count |\n")
	fmt.Fprintf(&buf, "|--------|-------|\n")
	fmt.Fprintf(&buf, "| Total RMIs | %d |\n", p.Summary.TotalRMIs)
	fmt.Fprintf(&buf, "| Completed | %d |\n", p.Summary.CompletedRMIs)
	fmt.Fprintf(&buf, "| Required | %d |\n", p.Summary.RequiredRMIs)
	fmt.Fprintf(&buf, "| Required Done | %d |\n", p.Summary.RequiredDone)
	fmt.Fprintf(&buf, "| In Progress | %d |\n", p.Summary.InProgress)
	fmt.Fprintf(&buf, "| Blocked | %d |\n", p.Summary.Blocked)
	fmt.Fprintf(&buf, "\n")

	fmt.Fprintf(&buf, "## RMI Status\n\n")
	fmt.Fprintf(&buf, "| # | RMI | Status | Required | Evidence |\n")
	fmt.Fprintf(&buf, "|---|-----|--------|----------|----------|\n")
	for _, rmi := range p.RMIHandoffs {
		req := "no"
		if rmi.Required {
			req = "yes"
		}
		fmt.Fprintf(&buf, "| %d | %s | %s | %s | %d |\n",
			rmi.Sequence, rmi.Title, rmi.Status, req, rmi.EvidenceCount)
	}
	fmt.Fprintf(&buf, "\n")

	if len(p.AggregatedDecisions) > 0 {
		fmt.Fprintf(&buf, "## Decisions\n\n")
		for _, d := range p.AggregatedDecisions {
			fmt.Fprintf(&buf, "- %s\n", d)
		}
		fmt.Fprintf(&buf, "\n")
	}

	completedRMIs := make([]RMIHandoffEntry, 0)
	for _, rmi := range p.RMIHandoffs {
		if rmi.Handoff != nil && len(rmi.Handoff.Completed) > 0 {
			completedRMIs = append(completedRMIs, rmi)
		}
	}
	if len(completedRMIs) > 0 {
		fmt.Fprintf(&buf, "## Completed Work\n\n")
		for _, rmi := range completedRMIs {
			fmt.Fprintf(&buf, "### %s\n\n", rmi.Title)
			for _, c := range rmi.Handoff.Completed {
				fmt.Fprintf(&buf, "- %s\n", c)
			}
			fmt.Fprintf(&buf, "\n")
		}
	}

	if len(p.Evidence) > 0 {
		fmt.Fprintf(&buf, "## Delivery Evidence\n\n")
		fmt.Fprintf(&buf, "| RMI | Type | Reference |\n")
		fmt.Fprintf(&buf, "|-----|------|----------|\n")
		for _, ev := range p.Evidence {
			fmt.Fprintf(&buf, "| %s | %s | `%s` |\n", ev.RMITitle, ev.Type, ev.Reference)
		}
		fmt.Fprintf(&buf, "\n")
	}

	return buf.String()
}

// RenderHandoffJSON renders the handoff projection as indented JSON.
func (p *PhaseHandoffProjection) RenderJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}
