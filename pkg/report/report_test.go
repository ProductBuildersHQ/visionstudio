package report

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func seedStore(t *testing.T, s store.Store) {
	t.Helper()
	ctx := context.Background()
	now := time.Now()
	executing := now.Add(-10 * 24 * time.Hour)

	if err := s.CreateInitiative(ctx, &store.Initiative{
		ID: "INIT-TEST-001", Organization: "test", Title: "Test Initiative",
		Status: "executing", CreatedAt: now, ExecutingAt: &executing, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePhase(ctx, &store.Phase{
		ID: "INIT-TEST-001/phase-1", InitiativeID: "INIT-TEST-001",
		SequenceNumber: 1, Title: "Foundation", Theme: "Prove the stack",
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePhase(ctx, &store.Phase{
		ID: "INIT-TEST-001/phase-2", InitiativeID: "INIT-TEST-001",
		SequenceNumber: 2, Title: "Coordination", Theme: "Build coordination",
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRepository(ctx, &store.Repository{
		ID: "github.com/test/repo", Organization: "test",
		RepositoryName: "repo", DefaultBranch: "main", Status: "active",
	}); err != nil {
		t.Fatal(err)
	}

	completed := now.Add(-5 * 24 * time.Hour)
	rmis := []*store.RoadmapItem{
		{ID: "RMI-TEST-001", RepositoryID: "github.com/test/repo", InitiativeID: "INIT-TEST-001", PhaseID: "INIT-TEST-001/phase-1", Title: "Done item", ItemType: "capability", Status: "completed", Required: true, CompletedAt: &completed, CreatedAt: now, UpdatedAt: now},
		{ID: "RMI-TEST-002", RepositoryID: "github.com/test/repo", InitiativeID: "INIT-TEST-001", PhaseID: "INIT-TEST-001/phase-1", Title: "Also done", ItemType: "task", Status: "completed", Required: true, CompletedAt: &completed, CreatedAt: now, UpdatedAt: now},
		{ID: "RMI-TEST-003", RepositoryID: "github.com/test/repo", InitiativeID: "INIT-TEST-001", PhaseID: "INIT-TEST-001/phase-2", Title: "In progress", ItemType: "capability", Status: "in_progress", Required: true, CreatedAt: now, UpdatedAt: now},
		{ID: "RMI-TEST-004", RepositoryID: "github.com/test/repo", InitiativeID: "INIT-TEST-001", PhaseID: "INIT-TEST-001/phase-2", Title: "Ready item", ItemType: "task", Status: "ready", Required: false, CreatedAt: now, UpdatedAt: now},
	}
	for _, rmi := range rmis {
		if err := s.CreateRMI(ctx, rmi); err != nil {
			t.Fatal(err)
		}
	}

	evidence := []*store.DeliveryEvidence{
		{ID: "ev-1", RMIID: "RMI-TEST-001", EvidenceType: "commit", Reference: "aaa", CommitType: "feat", CreatedAt: now},
		{ID: "ev-2", RMIID: "RMI-TEST-001", EvidenceType: "commit", Reference: "bbb", CommitType: "test", CreatedAt: now},
		{ID: "ev-3", RMIID: "RMI-TEST-002", EvidenceType: "commit", Reference: "ccc", CommitType: "feat", CreatedAt: now},
		{ID: "ev-4", RMIID: "RMI-TEST-001", EvidenceType: "release", Reference: "v1.0.0", CreatedAt: now},
	}
	for _, ev := range evidence {
		if err := s.CreateEvidence(ctx, ev); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGenerate(t *testing.T) {
	s := store.NewMemStore()
	seedStore(t, s)

	r, err := Generate(context.Background(), s, "INIT-TEST-001")
	if err != nil {
		t.Fatal(err)
	}

	if r.InitiativeID != "INIT-TEST-001" {
		t.Fatalf("initiative_id: got %s", r.InitiativeID)
	}
	if r.Status != "executing" {
		t.Fatalf("status: got %s", r.Status)
	}
	if r.Duration.DaysExecuting < 10 {
		t.Fatalf("days_executing: got %d, want >= 10", r.Duration.DaysExecuting)
	}

	if len(r.Phases) != 2 {
		t.Fatalf("phases: got %d, want 2", len(r.Phases))
	}
	if r.Phases[0].Status != "completed" {
		t.Fatalf("phase-1 status: got %s, want completed", r.Phases[0].Status)
	}
	if r.Phases[0].RMIsComplete != 2 {
		t.Fatalf("phase-1 completed: got %d, want 2", r.Phases[0].RMIsComplete)
	}

	if r.RMIs.Total != 4 {
		t.Fatalf("total RMIs: got %d, want 4", r.RMIs.Total)
	}
	if r.RMIs.Completed != 2 {
		t.Fatalf("completed RMIs: got %d, want 2", r.RMIs.Completed)
	}
	if r.RMIs.RequiredCompleted != 2 {
		t.Fatalf("required completed: got %d, want 2", r.RMIs.RequiredCompleted)
	}
	if r.RMIs.InProgress != 1 {
		t.Fatalf("in_progress: got %d, want 1", r.RMIs.InProgress)
	}
	if r.RMIs.Ready != 1 {
		t.Fatalf("ready: got %d, want 1", r.RMIs.Ready)
	}

	if r.Repos.Count != 1 {
		t.Fatalf("repos: got %d, want 1", r.Repos.Count)
	}

	if r.Commits.Total != 3 {
		t.Fatalf("commits: got %d, want 3", r.Commits.Total)
	}
	if r.Commits.ByType["feat"] != 2 {
		t.Fatalf("feat commits: got %d, want 2", r.Commits.ByType["feat"])
	}

	if len(r.Releases) != 1 {
		t.Fatalf("releases: got %d, want 1", len(r.Releases))
	}
	if r.Releases[0].Version != "v1.0.0" {
		t.Fatalf("release version: got %s", r.Releases[0].Version)
	}
}

func TestMarkdown(t *testing.T) {
	s := store.NewMemStore()
	seedStore(t, s)

	r, err := Generate(context.Background(), s, "INIT-TEST-001")
	if err != nil {
		t.Fatal(err)
	}

	md := r.Markdown()
	if !strings.Contains(md, "# Initiative Report: INIT-TEST-001") {
		t.Fatal("missing header")
	}
	if !strings.Contains(md, "Foundation") {
		t.Fatal("missing phase title")
	}
	if !strings.Contains(md, "v1.0.0") {
		t.Fatal("missing release")
	}
}
