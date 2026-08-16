package validate

import (
	"context"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func validRMI(id, repoID, status string) *store.RoadmapItem {
	return &store.RoadmapItem{
		ID: id, RepositoryID: repoID, InitiativeID: "INIT-TEST-001",
		PhaseID: "INIT-TEST-001/phase-1", Title: id, ItemType: "capability",
		Status: status, Required: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}

func TestCleanStore(t *testing.T) {
	s := store.NewMemStore()
	ctx := context.Background()
	now := time.Now()

	if err := s.CreateInitiative(ctx, &store.Initiative{
		ID: "INIT-TEST-001", Organization: "test", Title: "Clean",
		Status: "executing", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePhase(ctx, &store.Phase{
		ID: "INIT-TEST-001/phase-1", InitiativeID: "INIT-TEST-001",
		SequenceNumber: 1, Title: "P1",
	}); err != nil {
		t.Fatal(err)
	}
	rmi := validRMI("RMI-TEST-001", "github.com/test/repo", "ready")
	if err := s.CreateRMI(ctx, rmi); err != nil {
		t.Fatal(err)
	}

	r, err := Run(ctx, s, now)
	if err != nil {
		t.Fatal(err)
	}
	if !r.OK() {
		t.Fatalf("expected clean store, got %d findings: %v", len(r.Findings), r.Findings)
	}
}

func TestIDFormatViolation(t *testing.T) {
	s := store.NewMemStore()
	ctx := context.Background()
	now := time.Now()

	if err := s.CreateInitiative(ctx, &store.Initiative{
		ID: "INIT-TEST-001", Organization: "test", Title: "Test",
		Status: "executing", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePhase(ctx, &store.Phase{
		ID: "INIT-TEST-001/phase-1", InitiativeID: "INIT-TEST-001",
		SequenceNumber: 1, Title: "P1",
	}); err != nil {
		t.Fatal(err)
	}

	bad := validRMI("BAD-ID", "repo", "ready")
	if err := s.CreateRMI(ctx, bad); err != nil {
		t.Fatal(err)
	}

	r, err := Run(ctx, s, now)
	if err != nil {
		t.Fatal(err)
	}
	if r.OK() {
		t.Fatal("expected id_format error")
	}
	found := false
	for _, f := range r.Findings {
		if f.Check == "id_format" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing id_format finding")
	}
}

func TestDanglingDependency(t *testing.T) {
	s := store.NewMemStore()
	ctx := context.Background()
	now := time.Now()

	if err := s.CreateInitiative(ctx, &store.Initiative{
		ID: "INIT-TEST-001", Organization: "test", Title: "Test",
		Status: "executing", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePhase(ctx, &store.Phase{
		ID: "INIT-TEST-001/phase-1", InitiativeID: "INIT-TEST-001",
		SequenceNumber: 1, Title: "P1",
	}); err != nil {
		t.Fatal(err)
	}
	rmi := validRMI("RMI-TEST-001", "repo", "ready")
	if err := s.CreateRMI(ctx, rmi); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateDependency(ctx, &store.RMIDependency{
		SourceRMIID: "RMI-TEST-001", TargetRMIID: "RMI-GHOST-999", Relationship: "requires",
	}); err != nil {
		t.Fatal(err)
	}

	r, err := Run(ctx, s, now)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range r.Findings {
		if f.Check == "dangling_dep" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected dangling_dep finding")
	}
}

func TestDependencyCycle(t *testing.T) {
	s := store.NewMemStore()
	ctx := context.Background()
	now := time.Now()

	if err := s.CreateInitiative(ctx, &store.Initiative{
		ID: "INIT-TEST-001", Organization: "test", Title: "Test",
		Status: "executing", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePhase(ctx, &store.Phase{
		ID: "INIT-TEST-001/phase-1", InitiativeID: "INIT-TEST-001",
		SequenceNumber: 1, Title: "P1",
	}); err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"RMI-TEST-001", "RMI-TEST-002", "RMI-TEST-003"} {
		rmi := validRMI(id, "repo", "ready")
		if err := s.CreateRMI(ctx, rmi); err != nil {
			t.Fatal(err)
		}
	}

	// A → B → C → A
	for _, dep := range []store.RMIDependency{
		{SourceRMIID: "RMI-TEST-001", TargetRMIID: "RMI-TEST-002", Relationship: "requires"},
		{SourceRMIID: "RMI-TEST-002", TargetRMIID: "RMI-TEST-003", Relationship: "requires"},
		{SourceRMIID: "RMI-TEST-003", TargetRMIID: "RMI-TEST-001", Relationship: "requires"},
	} {
		if err := s.CreateDependency(ctx, &dep); err != nil {
			t.Fatal(err)
		}
	}

	r, err := Run(ctx, s, now)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range r.Findings {
		if f.Check == "dependency_cycle" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected dependency_cycle finding")
	}
}

func TestExpiredLease(t *testing.T) {
	s := store.NewMemStore()
	ctx := context.Background()
	now := time.Now()

	if err := s.CreateInitiative(ctx, &store.Initiative{
		ID: "INIT-TEST-001", Organization: "test", Title: "Test",
		Status: "executing", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePhase(ctx, &store.Phase{
		ID: "INIT-TEST-001/phase-1", InitiativeID: "INIT-TEST-001",
		SequenceNumber: 1, Title: "P1",
	}); err != nil {
		t.Fatal(err)
	}
	rmi := validRMI("RMI-TEST-001", "repo", "in_progress")
	if err := s.CreateRMI(ctx, rmi); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateAssignment(ctx, &store.Assignment{
		ID: "assign-1", RMIID: "RMI-TEST-001", Worker: "session-1",
		Status: "active", LeaseExpiresAt: now.Add(-1 * time.Hour),
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	r, err := Run(ctx, s, now)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range r.Findings {
		if f.Check == "expired_lease" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected expired_lease finding")
	}
}

func TestStatusCoherence(t *testing.T) {
	s := store.NewMemStore()
	ctx := context.Background()
	now := time.Now()
	dc := now.Add(-1 * time.Hour)

	if err := s.CreateInitiative(ctx, &store.Initiative{
		ID: "INIT-TEST-001", Organization: "test", Title: "Test",
		Status: "delivery_complete", DeliveryCompleteAt: &dc,
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePhase(ctx, &store.Phase{
		ID: "INIT-TEST-001/phase-1", InitiativeID: "INIT-TEST-001",
		SequenceNumber: 1, Title: "P1",
	}); err != nil {
		t.Fatal(err)
	}

	// Required RMI still in_progress under a delivery_complete initiative.
	rmi := validRMI("RMI-TEST-001", "repo", "in_progress")
	if err := s.CreateRMI(ctx, rmi); err != nil {
		t.Fatal(err)
	}

	r, err := Run(ctx, s, now)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range r.Findings {
		if f.Check == "status_coherence" && f.Level == "error" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected status_coherence error for open required RMI under delivery_complete initiative")
	}
}

func TestCompletedRMINoEvidence(t *testing.T) {
	s := store.NewMemStore()
	ctx := context.Background()
	now := time.Now()
	completed := now.Add(-1 * time.Hour)

	if err := s.CreateInitiative(ctx, &store.Initiative{
		ID: "INIT-TEST-001", Organization: "test", Title: "Test",
		Status: "executing", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePhase(ctx, &store.Phase{
		ID: "INIT-TEST-001/phase-1", InitiativeID: "INIT-TEST-001",
		SequenceNumber: 1, Title: "P1",
	}); err != nil {
		t.Fatal(err)
	}

	rmi := validRMI("RMI-TEST-001", "repo", "completed")
	rmi.CompletedAt = &completed
	if err := s.CreateRMI(ctx, rmi); err != nil {
		t.Fatal(err)
	}

	r, err := Run(ctx, s, now)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range r.Findings {
		if f.Check == "status_coherence" && f.Level == "warning" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected status_coherence warning for completed RMI with no evidence")
	}
}

func TestEvidenceRefsMissingRMI(t *testing.T) {
	s := store.NewMemStore()
	ctx := context.Background()
	now := time.Now()

	if err := s.CreateInitiative(ctx, &store.Initiative{
		ID: "INIT-TEST-001", Organization: "test", Title: "Test",
		Status: "executing", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.CreateEvidence(ctx, &store.DeliveryEvidence{
		ID: "ev-1", RMIID: "RMI-GHOST-999", EvidenceType: "commit",
		Reference: "abc123", CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	r, err := Run(ctx, s, now)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range r.Findings {
		if f.Check == "evidence_ref" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected evidence_ref finding")
	}
}

func TestContextSpecInvalidRepo(t *testing.T) {
	s := store.NewMemStore()
	ctx := context.Background()
	now := time.Now()

	if err := s.CreateRepository(ctx, &store.Repository{
		ID: "github.com/test/repo", Organization: "test",
		RepositoryName: "repo", DefaultBranch: "main", Status: "active",
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.CreateInitiative(ctx, &store.Initiative{
		ID: "INIT-TEST-001", Organization: "test", Title: "Test",
		Status: "executing", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePhase(ctx, &store.Phase{
		ID: "INIT-TEST-001/phase-1", InitiativeID: "INIT-TEST-001",
		SequenceNumber: 1, Title: "P1",
	}); err != nil {
		t.Fatal(err)
	}

	rmi := validRMI("RMI-TEST-001", "github.com/test/repo", "ready")
	rmi.ContextSpec = &store.ContextSpec{
		ExtraRepos: []string{"github.com/nonexistent/repo"},
	}
	if err := s.CreateRMI(ctx, rmi); err != nil {
		t.Fatal(err)
	}

	r, err := Run(ctx, s, now)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range r.Findings {
		if f.Check == "context_spec" && f.Level == "error" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected context_spec error for non-existent repo in extra_repos")
	}
}

func TestUnshippedInitiativeWarning(t *testing.T) {
	s := store.NewMemStore()
	ctx := context.Background()
	now := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
	old := now.AddDate(0, 0, -30)
	recent := now.AddDate(0, 0, -2)

	seed := []*store.Initiative{
		{ID: "INIT-UNS-001", Organization: "t", Title: "stale", Status: "delivery_complete", DeliveryCompleteAt: &old, CreatedAt: now, UpdatedAt: now},
		{ID: "INIT-UNS-002", Organization: "t", Title: "fresh", Status: "delivery_complete", DeliveryCompleteAt: &recent, CreatedAt: now, UpdatedAt: now},
		{ID: "INIT-UNS-003", Organization: "t", Title: "shipped", Status: "delivery_complete", DeliveryCompleteAt: &old, CreatedAt: now, UpdatedAt: now},
		{ID: "INIT-UNS-004", Organization: "t", Title: "no ts", Status: "releasing", CreatedAt: now, UpdatedAt: now},
	}
	for _, in := range seed {
		if err := s.CreateInitiative(ctx, in); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.CreateRepository(ctx, &store.Repository{ID: "github.com/t/r", Organization: "t", RepositoryName: "r", Status: "active"}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRelease(ctx, &store.Release{
		ID: "github.com/t/r@v1.0.0", RepositoryID: "github.com/t/r", Tag: "v1.0.0",
		ReleasedAt: old, InitiativeIDs: []string{"INIT-UNS-003"}, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	r, err := Run(ctx, s, now)
	if err != nil {
		t.Fatal(err)
	}
	var unshipped []Finding
	for _, f := range r.Findings {
		if f.Check == "unshipped" {
			unshipped = append(unshipped, f)
		}
	}
	// stale (30d > threshold) warns; no-timestamp warns; fresh (2d) and
	// shipped do not.
	if len(unshipped) != 2 {
		t.Fatalf("unshipped findings = %d (%v), want 2", len(unshipped), unshipped)
	}
	for _, f := range unshipped {
		if f.Level != "warning" {
			t.Fatalf("level = %s, want warning (never error — soft gate)", f.Level)
		}
	}
}
