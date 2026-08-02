package contextbuild

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestBuildPhaseHandoff_Basic(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupHandoffTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	proj, err := builder.BuildPhaseHandoff(ctx, "INIT-TEST-001/phase-1")
	if err != nil {
		t.Fatalf("build handoff: %v", err)
	}

	if proj.Version != SchemaVersion {
		t.Errorf("version = %q, want %q", proj.Version, SchemaVersion)
	}

	if proj.PhaseID != "INIT-TEST-001/phase-1" {
		t.Errorf("phase_id = %q, want %q", proj.PhaseID, "INIT-TEST-001/phase-1")
	}

	if proj.PhaseTitle != "Phase 1" {
		t.Errorf("phase_title = %q, want %q", proj.PhaseTitle, "Phase 1")
	}

	if proj.InitiativeID != "INIT-TEST-001" {
		t.Errorf("initiative_id = %q, want %q", proj.InitiativeID, "INIT-TEST-001")
	}
}

func TestBuildPhaseHandoff_Summary(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupHandoffTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	proj, err := builder.BuildPhaseHandoff(ctx, "INIT-TEST-001/phase-1")
	if err != nil {
		t.Fatalf("build handoff: %v", err)
	}

	if proj.Summary.TotalRMIs != 3 {
		t.Errorf("total_rmis = %d, want 3", proj.Summary.TotalRMIs)
	}

	if proj.Summary.CompletedRMIs != 1 {
		t.Errorf("completed_rmis = %d, want 1", proj.Summary.CompletedRMIs)
	}

	if proj.Summary.RequiredRMIs != 2 {
		t.Errorf("required_rmis = %d, want 2", proj.Summary.RequiredRMIs)
	}

	if proj.Summary.InProgress != 1 {
		t.Errorf("in_progress = %d, want 1", proj.Summary.InProgress)
	}
}

func TestBuildPhaseHandoff_RMIOrdering(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupHandoffTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	proj, err := builder.BuildPhaseHandoff(ctx, "INIT-TEST-001/phase-1")
	if err != nil {
		t.Fatalf("build handoff: %v", err)
	}

	if len(proj.RMIHandoffs) != 3 {
		t.Fatalf("rmi_handoffs count = %d, want 3", len(proj.RMIHandoffs))
	}

	for i := 1; i < len(proj.RMIHandoffs); i++ {
		if proj.RMIHandoffs[i].Sequence < proj.RMIHandoffs[i-1].Sequence {
			t.Errorf("RMI handoffs not ordered by sequence: %d before %d",
				proj.RMIHandoffs[i-1].Sequence, proj.RMIHandoffs[i].Sequence)
		}
	}
}

func TestBuildPhaseHandoff_AggregatedDecisions(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupHandoffTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	proj, err := builder.BuildPhaseHandoff(ctx, "INIT-TEST-001/phase-1")
	if err != nil {
		t.Fatalf("build handoff: %v", err)
	}

	if len(proj.AggregatedDecisions) != 2 {
		t.Errorf("aggregated_decisions count = %d, want 2", len(proj.AggregatedDecisions))
	}
}

func TestBuildPhaseHandoff_Evidence(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupHandoffTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	proj, err := builder.BuildPhaseHandoff(ctx, "INIT-TEST-001/phase-1")
	if err != nil {
		t.Fatalf("build handoff: %v", err)
	}

	if len(proj.Evidence) != 2 {
		t.Errorf("evidence count = %d, want 2", len(proj.Evidence))
	}

	for _, rmi := range proj.RMIHandoffs {
		if rmi.RMIID == "RMI-TESTREPO-001" && rmi.EvidenceCount != 2 {
			t.Errorf("RMI-TESTREPO-001 evidence_count = %d, want 2", rmi.EvidenceCount)
		}
	}
}

func TestBuildPhaseHandoff_ByteIdentical(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupHandoffTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	proj1, err := builder.BuildPhaseHandoff(ctx, "INIT-TEST-001/phase-1")
	if err != nil {
		t.Fatalf("first build: %v", err)
	}

	proj2, err := builder.BuildPhaseHandoff(ctx, "INIT-TEST-001/phase-1")
	if err != nil {
		t.Fatalf("second build: %v", err)
	}

	proj1.BuildTimestamp = time.Time{}
	proj2.BuildTimestamp = time.Time{}

	json1, err := json.MarshalIndent(proj1, "", "  ")
	if err != nil {
		t.Fatalf("marshal proj1: %v", err)
	}

	json2, err := json.MarshalIndent(proj2, "", "  ")
	if err != nil {
		t.Fatalf("marshal proj2: %v", err)
	}

	if string(json1) != string(json2) {
		t.Errorf("outputs not byte-identical")
	}
}

func TestPhaseHandoffProjection_RenderMarkdown(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupHandoffTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	proj, err := builder.BuildPhaseHandoff(ctx, "INIT-TEST-001/phase-1")
	if err != nil {
		t.Fatalf("build handoff: %v", err)
	}

	md := proj.RenderMarkdown()

	if !strings.Contains(md, "# Phase Handoff: Phase 1") {
		t.Error("markdown missing title")
	}

	if !strings.Contains(md, "## Summary") {
		t.Error("markdown missing summary section")
	}

	if !strings.Contains(md, "## RMI Status") {
		t.Error("markdown missing RMI status section")
	}

	if !strings.Contains(md, "## Decisions") {
		t.Error("markdown missing decisions section")
	}
}

func TestPhaseHandoffProjection_RenderJSON(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupHandoffTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	proj, err := builder.BuildPhaseHandoff(ctx, "INIT-TEST-001/phase-1")
	if err != nil {
		t.Fatalf("build handoff: %v", err)
	}

	jsonBytes, err := proj.RenderJSON()
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}

	var parsed PhaseHandoffProjection
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}

	if parsed.PhaseID != proj.PhaseID {
		t.Errorf("parsed phase_id = %q, want %q", parsed.PhaseID, proj.PhaseID)
	}
}

func setupHandoffTestData(t *testing.T, ctx context.Context, ms *store.MemStore) {
	t.Helper()

	if err := ms.CreateRepository(ctx, &store.Repository{
		ID:             "github.com/test/repo",
		Organization:   "test",
		RepositoryName: "repo",
		DefaultBranch:  "main",
		LocalPath:      "/tmp/test/repo",
		Status:         "active",
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateInitiative(ctx, &store.Initiative{
		ID:          "INIT-TEST-001",
		Title:       "Test Initiative",
		Description: "A test initiative",
		Status:      "executing",
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreatePhase(ctx, &store.Phase{
		ID:             "INIT-TEST-001/phase-1",
		InitiativeID:   "INIT-TEST-001",
		SequenceNumber: 1,
		Title:          "Phase 1",
		Theme:          "Foundation",
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateRMI(ctx, &store.RoadmapItem{
		ID:             "RMI-TESTREPO-001",
		RepositoryID:   "github.com/test/repo",
		InitiativeID:   "INIT-TEST-001",
		PhaseID:        "INIT-TEST-001/phase-1",
		Title:          "First RMI",
		ItemType:       "capability",
		Status:         "completed",
		Required:       true,
		SequenceNumber: 1,
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateRMI(ctx, &store.RoadmapItem{
		ID:             "RMI-TESTREPO-002",
		RepositoryID:   "github.com/test/repo",
		InitiativeID:   "INIT-TEST-001",
		PhaseID:        "INIT-TEST-001/phase-1",
		Title:          "Second RMI",
		ItemType:       "capability",
		Status:         "in_progress",
		Required:       true,
		SequenceNumber: 2,
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateRMI(ctx, &store.RoadmapItem{
		ID:             "RMI-TESTREPO-003",
		RepositoryID:   "github.com/test/repo",
		InitiativeID:   "INIT-TEST-001",
		PhaseID:        "INIT-TEST-001/phase-1",
		Title:          "Third RMI (optional)",
		ItemType:       "capability",
		Status:         "proposed",
		Required:       false,
		SequenceNumber: 3,
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateAssignment(ctx, &store.Assignment{
		ID:             "assign-001",
		RMIID:          "RMI-TESTREPO-001",
		Worker:         "session-abc",
		Status:         "completed",
		LeaseExpiresAt: time.Now().Add(time.Hour),
		CreatedAt:      time.Now().Add(-time.Hour),
		Handoff: &store.Handoff{
			Completed:  []string{"Implemented core logic", "Added tests"},
			Remaining:  []string{},
			Decisions:  []string{"Used interface pattern for extensibility", "Chose JSON over YAML for config"},
			NextAction: "",
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateAssignment(ctx, &store.Assignment{
		ID:             "assign-002",
		RMIID:          "RMI-TESTREPO-002",
		Worker:         "session-def",
		Status:         "active",
		LeaseExpiresAt: time.Now().Add(time.Hour),
		CreatedAt:      time.Now(),
		Handoff:        nil,
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateEvidence(ctx, &store.DeliveryEvidence{
		ID:           "ev-001",
		RMIID:        "RMI-TESTREPO-001",
		EvidenceType: "commit",
		Reference:    "abc123",
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateEvidence(ctx, &store.DeliveryEvidence{
		ID:           "ev-002",
		RMIID:        "RMI-TESTREPO-001",
		EvidenceType: "pr",
		Reference:    "https://github.com/test/repo/pull/1",
	}); err != nil {
		t.Fatal(err)
	}
}
