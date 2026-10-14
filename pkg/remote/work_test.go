package remote_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/remote"
	"github.com/ProductBuildersHQ/visionstudio/pkg/remote/remotetest"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// seedWork creates an initiative with one phase and two RMIs server-side.
func seedWork(t *testing.T, fake *remotetest.Server) {
	t.Helper()
	ctx := context.Background()
	if _, err := fake.Service.CreateInitiative(ctx, "INIT-ACME-001", "acme", "Launch", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := fake.Service.CreatePhase(ctx, "INIT-ACME-001/phase-1", "INIT-ACME-001", 1, "Foundation", ""); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"RMI-ACME-001", "RMI-ACME-002"} {
		if _, err := fake.Service.CreateRMI(ctx, id, "github.com/acme/app", "INIT-ACME-001", "INIT-ACME-001/phase-1", "t "+id, "", "task", "", true, 1, nil); err != nil {
			t.Fatal(err)
		}
	}
}

func lastRequest(t *testing.T, fake *remotetest.Server) remotetest.Request {
	t.Helper()
	reqs := fake.Requests()
	if len(reqs) == 0 {
		t.Fatal("no requests")
	}
	return reqs[len(reqs)-1]
}

func TestPrograms(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	ctx := context.Background()

	p := &store.Program{ID: "PROG-X", Name: "X", Organization: "acme", Hidden: true}
	if err := rs.CreateProgram(ctx, p); err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	if p.CreatedAt.IsZero() || !p.Hidden {
		t.Fatalf("created = %+v, want server timestamps and hidden", p)
	}
	if err := rs.CreateProgram(ctx, &store.Program{ID: "PROG-X", Name: "X"}); !errors.Is(err, remote.ErrConflict) {
		t.Fatalf("duplicate CreateProgram err = %v, want ErrConflict", err)
	}
	got, err := rs.GetProgram(ctx, "PROG-X")
	if err != nil || got.Name != "X" {
		t.Fatalf("GetProgram = %+v, %v", got, err)
	}
	if r := lastRequest(t, fake); r.Method != http.MethodGet || r.Path != "/o/acme/api/v1/programs/PROG-X" {
		t.Fatalf("GetProgram request = %s %s", r.Method, r.Path)
	}
	if _, err := rs.GetProgram(ctx, "PROG-NOPE"); !errors.Is(err, remote.ErrNotFound) {
		t.Fatalf("GetProgram missing err = %v, want ErrNotFound", err)
	}
	got.Description = "updated"
	if err := rs.UpdateProgram(ctx, got); err != nil {
		t.Fatalf("UpdateProgram: %v", err)
	}
	if sp, _ := fake.Service.GetProgram(ctx, "PROG-X"); sp.Description != "updated" {
		t.Fatalf("server-side program = %+v", sp)
	}
	progs, err := rs.ListPrograms(ctx)
	if err != nil || len(progs) != 1 {
		t.Fatalf("ListPrograms = %+v, %v", progs, err)
	}
}

func TestInitiativeUpdateAndTransition(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	svc := service.New(rs)
	ctx := context.Background()
	seedWork(t, fake)

	in, err := svc.TransitionInitiative(ctx, "INIT-ACME-001", "planned")
	if err != nil {
		t.Fatalf("TransitionInitiative: %v", err)
	}
	if in.PlannedAt == nil {
		t.Fatalf("planned_at not stamped: %+v", in)
	}
	r := lastRequest(t, fake)
	if r.Method != http.MethodPut || r.Path != "/o/acme/api/v1/initiatives/INIT-ACME-001" {
		t.Fatalf("UpdateInitiative request = %s %s", r.Method, r.Path)
	}
	srv, err := fake.Service.GetInitiative(ctx, "INIT-ACME-001")
	if err != nil || srv.Status != "planned" || srv.PlannedAt == nil {
		t.Fatalf("server-side initiative = %+v, %v", srv, err)
	}
	if err := rs.UpdateInitiative(ctx, &store.Initiative{ID: "INIT-NOPE-001", Title: "x", Status: "proposed"}); !errors.Is(err, remote.ErrNotFound) {
		t.Fatalf("UpdateInitiative missing err = %v, want ErrNotFound", err)
	}
}

func TestCreateWithExtraFieldsWritesFullEntity(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	ctx := context.Background()

	in := &store.Initiative{ID: "INIT-ACME-002", Organization: "acme", Title: "t", Status: "executing", HomeRepo: "github.com/acme/app", Specs: map[string]string{"prd": "PRD.md"}}
	if err := rs.CreateInitiative(ctx, in); err != nil {
		t.Fatalf("CreateInitiative with extras: %v", err)
	}
	srv, err := fake.Service.GetInitiative(ctx, "INIT-ACME-002")
	if err != nil || srv.Status != "executing" || srv.HomeRepo != "github.com/acme/app" || srv.Specs["prd"] != "PRD.md" || srv.CreatedAt.IsZero() {
		t.Fatalf("server-side initiative = %+v, %v", srv, err)
	}
	if in.Status != "executing" || in.CreatedAt.IsZero() {
		t.Fatalf("caller copy = %+v", in)
	}

	now := time.Now().UTC().Truncate(time.Second)
	rmi := &store.RoadmapItem{ID: "RMI-ACME-009", RepositoryID: "github.com/acme/app", Title: "t", ItemType: "task", Status: "completed", Origin: "discussion", CompletedAt: &now}
	if err := rs.CreateRMI(ctx, rmi); err != nil {
		t.Fatalf("CreateRMI with extras: %v", err)
	}
	got, err := fake.Service.GetRMI(ctx, "RMI-ACME-009")
	if err != nil || got.Status != "completed" || got.Origin != "discussion" || got.CompletedAt == nil {
		t.Fatalf("server-side RMI = %+v, %v", got, err)
	}

	// Plain creates stay a single request.
	before := len(fake.Requests())
	if err := rs.CreateInitiative(ctx, &store.Initiative{ID: "INIT-ACME-003", Organization: "acme", Title: "plain"}); err != nil {
		t.Fatal(err)
	}
	if n := len(fake.Requests()) - before; n != 1 {
		t.Fatalf("plain create sent %d requests, want 1", n)
	}
}

func TestRMIUpdateFiltersAndMove(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	svc := service.New(rs)
	ctx := context.Background()
	seedWork(t, fake)
	if _, err := svc.CreatePhase(ctx, "INIT-ACME-001/phase-2", "INIT-ACME-001", 2, "Next", ""); err != nil {
		t.Fatalf("CreatePhase: %v", err)
	}

	rmi, err := svc.UpdateRMIStatus(ctx, "RMI-ACME-001", "ready")
	if err != nil || rmi.Status != "ready" {
		t.Fatalf("UpdateRMIStatus = %+v, %v", rmi, err)
	}
	assertIDs(t, "ListRMIsByStatus", mustRMIs(t)(rs.ListRMIsByStatus(ctx, "ready")), "RMI-ACME-001")
	if q := lastRequest(t, fake).Query; q != "status=ready" {
		t.Fatalf("ListRMIsByStatus query = %q", q)
	}
	assertIDs(t, "ListRMIsByRepo", mustRMIs(t)(rs.ListRMIsByRepo(ctx, "github.com/acme/app")), "RMI-ACME-001", "RMI-ACME-002")
	if q := lastRequest(t, fake).Query; q != "repository=github.com%2Facme%2Fapp" {
		t.Fatalf("ListRMIsByRepo query = %q", q)
	}

	moved, err := svc.MoveRMI(ctx, "RMI-ACME-002", "INIT-ACME-001/phase-2", 3)
	if err != nil {
		t.Fatalf("MoveRMI: %v", err)
	}
	if moved.PhaseID != "INIT-ACME-001/phase-2" || moved.SequenceNumber != 3 {
		t.Fatalf("moved = %+v", moved)
	}
	if srv, _ := fake.Service.GetRMI(ctx, "RMI-ACME-002"); srv.PhaseID != "INIT-ACME-001/phase-2" {
		t.Fatalf("server-side RMI = %+v", srv)
	}
	if err := rs.UpdateRMI(ctx, &store.RoadmapItem{ID: "RMI-NOPE-001", Title: "x"}); !errors.Is(err, remote.ErrNotFound) {
		t.Fatalf("UpdateRMI missing err = %v", err)
	}
}

func TestPhaseCreateDelete(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	svc := service.New(rs)
	ctx := context.Background()
	seedWork(t, fake)

	if _, err := svc.CreatePhase(ctx, "INIT-ACME-001/phase-2", "INIT-ACME-001", 2, "Empty", "theme"); err != nil {
		t.Fatalf("CreatePhase: %v", err)
	}
	err := rs.CreatePhase(ctx, &store.Phase{ID: "INIT-NOPE-001/phase-1", InitiativeID: "INIT-NOPE-001"})
	var apiErr *remote.APIError
	if !errors.Is(err, remote.ErrInvalidReference) || !errors.Is(err, remote.ErrInvalid) || !errors.As(err, &apiErr) || apiErr.Code != remote.CodeInvalidReference {
		t.Fatalf("CreatePhase unknown initiative err = %v, want INVALID_REFERENCE", err)
	}
	if err := rs.CreatePhase(ctx, &store.Phase{ID: "INIT-ACME-001/phase-2", InitiativeID: "INIT-ACME-001"}); !errors.Is(err, remote.ErrConflict) {
		t.Fatalf("duplicate CreatePhase err = %v, want ErrConflict", err)
	}

	// A phase with member RMIs is refused (409); an empty one is deleted.
	if err := rs.DeletePhase(ctx, "INIT-ACME-001/phase-1"); !errors.Is(err, remote.ErrConflict) {
		t.Fatalf("DeletePhase with members err = %v, want ErrConflict", err)
	}
	if err := svc.RemovePhase(ctx, "INIT-ACME-001/phase-2"); err != nil {
		t.Fatalf("RemovePhase: %v", err)
	}
	r := lastRequest(t, fake)
	if r.Method != http.MethodDelete || r.RawPath != "/o/acme/api/v1/phases/INIT-ACME-001%2Fphase-2" || r.Path != "/o/acme/api/v1/phases/INIT-ACME-001/phase-2" {
		t.Fatalf("DeletePhase request = %s %s (raw %s)", r.Method, r.Path, r.RawPath)
	}
	phases, err := rs.ListPhases(ctx, "INIT-ACME-001")
	if err != nil || len(phases) != 1 {
		t.Fatalf("ListPhases = %+v, %v", phases, err)
	}
	if err := rs.DeletePhase(ctx, "INIT-ACME-001/phase-9"); !errors.Is(err, remote.ErrNotFound) {
		t.Fatalf("DeletePhase missing err = %v, want ErrNotFound", err)
	}
}

func TestDependencies(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	svc := service.New(rs)
	ctx := context.Background()
	seedWork(t, fake)

	if err := svc.CreateDependency(ctx, "RMI-ACME-002", "RMI-ACME-001", ""); err != nil {
		t.Fatalf("CreateDependency: %v", err)
	}
	dep := &store.RMIDependency{SourceRMIID: "RMI-ACME-002", TargetRMIID: "RMI-ACME-001"}
	if err := rs.CreateDependency(ctx, dep); err != nil || dep.Relationship != "requires" {
		t.Fatalf("idempotent CreateDependency = %+v, %v", dep, err)
	}
	if err := rs.CreateDependency(ctx, &store.RMIDependency{SourceRMIID: "RMI-ACME-002", TargetRMIID: "RMI-NOPE-001"}); !errors.Is(err, remote.ErrInvalidReference) {
		t.Fatalf("CreateDependency unknown target err = %v, want ErrInvalidReference", err)
	}
	deps, err := rs.ListDependencies(ctx, "RMI-ACME-002")
	if err != nil || len(deps) != 1 || deps[0].TargetRMIID != "RMI-ACME-001" {
		t.Fatalf("ListDependencies = %+v, %v", deps, err)
	}
	all, err := rs.ListAllDependencies(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAllDependencies = %+v, %v", all, err)
	}
	none, err := rs.ListDependencies(ctx, "RMI-NOPE-001")
	if err != nil || none == nil || len(none) != 0 {
		t.Fatalf("ListDependencies(none) = %#v, %v; want empty non-nil", none, err)
	}
}

func TestAssignmentLifecycle(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	svc := service.New(rs)
	ctx := context.Background()
	seedWork(t, fake)

	a, err := rs.GetActiveAssignment(ctx, "RMI-ACME-001")
	if err != nil || a != nil {
		t.Fatalf("GetActiveAssignment(none) = %+v, %v; want nil, nil", a, err)
	}

	claim, err := svc.ClaimRMI(ctx, "RMI-ACME-001", "session-1", "/ws", time.Hour)
	if err != nil {
		t.Fatalf("ClaimRMI: %v", err)
	}
	id := claim.Assignment.ID
	if claim.Assignment.CreatedAt.IsZero() || claim.TrailerLine == "" {
		t.Fatalf("claim = %+v", claim)
	}
	if srv, _ := fake.Service.GetRMI(ctx, "RMI-ACME-001"); srv.Status != "in_progress" {
		t.Fatalf("server-side RMI status = %q, want in_progress", srv.Status)
	}
	active, err := rs.GetActiveAssignment(ctx, "RMI-ACME-001")
	if err != nil || active == nil || active.ID != id {
		t.Fatalf("GetActiveAssignment = %+v, %v", active, err)
	}
	got, err := rs.GetAssignment(ctx, id)
	if err != nil || got.Worker != "session-1" || got.Workspace != "/ws" {
		t.Fatalf("GetAssignment = %+v, %v", got, err)
	}
	if _, err := rs.GetAssignment(ctx, "assign-nope"); !errors.Is(err, remote.ErrNotFound) {
		t.Fatalf("GetAssignment missing err = %v", err)
	}

	// Lease rule 1: a second active assignment while a live lease exists.
	now := time.Now()
	err = rs.CreateAssignment(ctx, &store.Assignment{ID: "assign-other", RMIID: "RMI-ACME-001", Worker: "session-2", Status: "active", LeaseExpiresAt: now.Add(time.Hour)})
	if !errors.Is(err, remote.ErrLeaseConflict) || !errors.Is(err, remote.ErrConflict) {
		t.Fatalf("competing CreateAssignment err = %v, want ErrLeaseConflict", err)
	}
	if err := rs.CreateAssignment(ctx, &store.Assignment{ID: "assign-x", RMIID: "RMI-NOPE-001", Status: "active"}); !errors.Is(err, remote.ErrInvalidReference) {
		t.Fatalf("CreateAssignment unknown RMI err = %v, want ErrInvalidReference", err)
	}
	if err := rs.CreateAssignment(ctx, &store.Assignment{ID: "assign-y", RMIID: "RMI-ACME-001", Status: "bogus"}); !errors.Is(err, remote.ErrInvalid) {
		t.Fatalf("CreateAssignment bad status err = %v, want ErrInvalid", err)
	}

	renewed, err := svc.RenewLease(ctx, id, 2*time.Hour)
	if err != nil || !renewed.LeaseExpiresAt.After(now.Add(90*time.Minute)) {
		t.Fatalf("RenewLease = %+v, %v", renewed, err)
	}
	if r := lastRequest(t, fake); r.Method != http.MethodPut || r.Path != "/o/acme/api/v1/assignments/"+id {
		t.Fatalf("RenewLease request = %s %s", r.Method, r.Path)
	}

	activeList, err := rs.ListActiveAssignments(ctx)
	if err != nil || len(activeList) != 1 {
		t.Fatalf("ListActiveAssignments = %+v, %v", activeList, err)
	}
	if q := lastRequest(t, fake).Query; q != "status=active" {
		t.Fatalf("ListActiveAssignments query = %q", q)
	}

	handoff := &store.Handoff{Completed: []string{"a"}, Remaining: []string{"b"}, NextAction: "c"}
	released, err := svc.ReleaseWork(ctx, id, handoff)
	if err != nil || released.Status != "released" {
		t.Fatalf("ReleaseWork = %+v, %v", released, err)
	}
	if srv, _ := fake.Service.GetRMI(ctx, "RMI-ACME-001"); srv.Status != "ready" {
		t.Fatalf("server-side RMI after release = %q, want ready", srv.Status)
	}
	if srv, _ := fake.Service.GetAssignment(ctx, id); srv.Handoff == nil || srv.Handoff.NextAction != "c" {
		t.Fatalf("server-side handoff = %+v", srv.Handoff)
	}

	// Lease rule 2: an assignment that is no longer active cannot change.
	released.Status = "active"
	if err := rs.UpdateAssignment(ctx, released); !errors.Is(err, remote.ErrLeaseConflict) {
		t.Fatalf("UpdateAssignment on released err = %v, want ErrLeaseConflict", err)
	}
	if _, err := svc.RenewLease(ctx, id, time.Hour); err == nil {
		t.Fatal("RenewLease on a released assignment succeeded")
	}

	// Reclaim, then complete.
	claim2, err := svc.ClaimRMI(ctx, "RMI-ACME-001", "session-2", "", time.Hour)
	if err != nil {
		t.Fatalf("re-claim: %v", err)
	}
	done, err := svc.CompleteWork(ctx, claim2.Assignment.ID, nil)
	if err != nil || done.Status != "completed" {
		t.Fatalf("CompleteWork = %+v, %v", done, err)
	}
	if srv, _ := fake.Service.GetRMI(ctx, "RMI-ACME-001"); srv.Status != "completed" || srv.CompletedAt == nil {
		t.Fatalf("server-side RMI after completion = %+v", srv)
	}
	all, err := rs.ListAllAssignments(ctx)
	if err != nil || len(all) != 2 {
		t.Fatalf("ListAllAssignments = %+v, %v", all, err)
	}
	if a, err := rs.GetActiveAssignment(ctx, "RMI-ACME-001"); err != nil || a != nil {
		t.Fatalf("GetActiveAssignment after completion = %+v, %v", a, err)
	}
}

func TestClaimConflictAcrossSessions(t *testing.T) {
	fake := newFake(t)
	ctx := context.Background()
	seedWork(t, fake)

	// Two sessions each read the RMI as claimable before either writes:
	// the server's lease rule decides, and the loser gets ErrConflict.
	one := newStore(t, fake)
	two := newStore(t, fake)
	if _, err := service.New(one).ClaimRMI(ctx, "RMI-ACME-002", "session-1", "", time.Hour); err != nil {
		t.Fatalf("first claim: %v", err)
	}
	now := time.Now()
	err := two.CreateAssignment(ctx, &store.Assignment{
		ID: "assign-RMI-ACME-002-2", RMIID: "RMI-ACME-002", Worker: "session-2", Status: "active",
		LeaseExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now,
	})
	if !errors.Is(err, remote.ErrConflict) || !strings.Contains(err.Error(), "LEASE_CONFLICT") {
		t.Fatalf("second claim err = %v, want ErrConflict with LEASE_CONFLICT", err)
	}
	// Through the service the second claim fails too (RMI already in progress).
	if _, err := service.New(two).ClaimRMI(ctx, "RMI-ACME-002", "session-2", "", time.Hour); err == nil {
		t.Fatal("second service claim succeeded")
	}
}

func TestEvidence(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	svc := service.New(rs)
	ctx := context.Background()
	seedWork(t, fake)

	ev, err := svc.AddEvidence(ctx, "RMI-ACME-001", "commit", "abc123")
	if err != nil {
		t.Fatalf("AddEvidence: %v", err)
	}
	if ev.CreatedAt.IsZero() {
		t.Fatalf("evidence = %+v", ev)
	}
	if _, err := svc.AddEvidence(ctx, "RMI-ACME-002", "pr", "https://example.com/pr/1"); err != nil {
		t.Fatal(err)
	}
	if err := rs.CreateEvidence(ctx, &store.DeliveryEvidence{ID: "ev-x", RMIID: "RMI-NOPE-001", EvidenceType: "commit", Reference: "r"}); !errors.Is(err, remote.ErrInvalidReference) {
		t.Fatalf("CreateEvidence unknown RMI err = %v", err)
	}
	if err := rs.CreateEvidence(ctx, &store.DeliveryEvidence{ID: ev.ID, RMIID: "RMI-ACME-001", EvidenceType: "commit", Reference: "r"}); !errors.Is(err, remote.ErrConflict) {
		t.Fatalf("duplicate CreateEvidence err = %v", err)
	}

	byRMI, err := rs.ListEvidenceByRMI(ctx, "RMI-ACME-001")
	if err != nil || len(byRMI) != 1 || byRMI[0].Reference != "abc123" {
		t.Fatalf("ListEvidenceByRMI = %+v, %v", byRMI, err)
	}
	byInit, err := rs.ListEvidenceByInitiative(ctx, "INIT-ACME-001")
	if err != nil || len(byInit) != 2 {
		t.Fatalf("ListEvidenceByInitiative = %+v, %v", byInit, err)
	}
	all, err := rs.ListAllEvidence(ctx)
	if err != nil || len(all) != 2 {
		t.Fatalf("ListAllEvidence = %+v, %v", all, err)
	}
}

func TestReleasesAndRepositoriesReadOnly(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	ctx := context.Background()
	seedWork(t, fake)

	for _, repo := range []*store.Repository{
		{ID: "github.com/acme/app", Organization: "acme", RepositoryName: "app", Status: "active"},
		{ID: "github.com/other/lib", Organization: "other", RepositoryName: "lib", Status: "active"},
	} {
		if err := fake.Service.Store.CreateRepository(ctx, repo); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC().Truncate(time.Second)
	rel := &store.Release{ID: "github.com/acme/app@v1.0.0", RepositoryID: "github.com/acme/app", Tag: "v1.0.0", ReleasedAt: now, InitiativeIDs: []string{"INIT-ACME-001"}, CreatedAt: now, UpdatedAt: now}
	if err := fake.Service.Store.CreateRelease(ctx, rel); err != nil {
		t.Fatal(err)
	}

	repo, err := rs.GetRepository(ctx, "github.com/acme/app")
	if err != nil || repo.RepositoryName != "app" {
		t.Fatalf("GetRepository = %+v, %v", repo, err)
	}
	if _, err := rs.GetRepository(ctx, "github.com/acme/nope"); !errors.Is(err, remote.ErrNotFound) {
		t.Fatalf("GetRepository missing err = %v", err)
	}
	repos, err := rs.ListRepositories(ctx)
	if err != nil || len(repos) != 2 {
		t.Fatalf("ListRepositories = %+v, %v", repos, err)
	}
	byOrg, err := rs.ListRepositoriesByOrg(ctx, "acme")
	if err != nil || len(byOrg) != 1 || byOrg[0].ID != "github.com/acme/app" {
		t.Fatalf("ListRepositoriesByOrg = %+v, %v", byOrg, err)
	}

	got, err := rs.GetRelease(ctx, rel.ID)
	if err != nil || got.Tag != "v1.0.0" {
		t.Fatalf("GetRelease = %+v, %v", got, err)
	}
	if r := lastRequest(t, fake); r.RawPath != "/o/acme/api/v1/releases/github.com%2Facme%2Fapp@v1.0.0" {
		t.Fatalf("GetRelease raw path = %q", r.RawPath)
	}
	for name, fn := range map[string]func() ([]*store.Release, error){
		"ListReleases":             func() ([]*store.Release, error) { return rs.ListReleases(ctx) },
		"ListReleasesByRepo":       func() ([]*store.Release, error) { return rs.ListReleasesByRepo(ctx, "github.com/acme/app") },
		"ListReleasesByInitiative": func() ([]*store.Release, error) { return rs.ListReleasesByInitiative(ctx, "INIT-ACME-001") },
	} {
		rels, err := fn()
		if err != nil || len(rels) != 1 {
			t.Fatalf("%s = %+v, %v", name, rels, err)
		}
	}
	none, err := rs.ListReleasesByRepo(ctx, "github.com/other/lib")
	if err != nil || len(none) != 0 {
		t.Fatalf("ListReleasesByRepo(other) = %+v, %v", none, err)
	}

	// Writes stay unsupported, before any request is sent.
	before := len(fake.Requests())
	if err := rs.CreateRelease(ctx, rel); !errors.Is(err, remote.ErrNotSupported) || !strings.Contains(err.Error(), "release writes") {
		t.Fatalf("CreateRelease err = %v", err)
	}
	if err := rs.CreateRepository(ctx, repo); !errors.Is(err, remote.ErrNotSupported) {
		t.Fatalf("CreateRepository err = %v", err)
	}
	if len(fake.Requests()) != before {
		t.Fatal("unsupported writes reached the server")
	}
}

func TestWorkflowsSpecDocumentsAndJudgeResults(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	svc := service.New(rs)
	ctx := context.Background()
	seedWork(t, fake)
	if err := fake.Service.Store.CreateSpecWorkflow(ctx, &store.SpecWorkflow{ID: "pbhq-lite", Name: "PBHQ Lite", SpecsRequired: []string{"prd"}}); err != nil {
		t.Fatal(err)
	}

	wfs, err := rs.ListSpecWorkflows(ctx)
	if err != nil || len(wfs) != 1 {
		t.Fatalf("ListSpecWorkflows = %+v, %v", wfs, err)
	}
	wf, err := rs.GetSpecWorkflow(ctx, "pbhq-lite")
	if err != nil || wf.Name != "PBHQ Lite" {
		t.Fatalf("GetSpecWorkflow = %+v, %v", wf, err)
	}

	iw, err := rs.GetWorkflowForInitiative(ctx, "INIT-ACME-001")
	if err != nil || iw != nil {
		t.Fatalf("GetWorkflowForInitiative(none) = %+v, %v; want nil, nil", iw, err)
	}
	if err := svc.SelectWorkflow(ctx, "INIT-ACME-001", "pbhq-lite"); err != nil {
		t.Fatalf("SelectWorkflow: %v", err)
	}
	iw, err = rs.GetWorkflowForInitiative(ctx, "INIT-ACME-001")
	if err != nil || iw == nil || iw.WorkflowID != "pbhq-lite" {
		t.Fatalf("GetWorkflowForInitiative = %+v, %v", iw, err)
	}
	if err := rs.SelectWorkflowForInitiative(ctx, "INIT-ACME-001", "nope"); !errors.Is(err, remote.ErrInvalidReference) {
		t.Fatalf("SelectWorkflowForInitiative unknown workflow err = %v", err)
	}
	if err := rs.SelectWorkflowForInitiative(ctx, "INIT-NOPE-001", "pbhq-lite"); !errors.Is(err, remote.ErrNotFound) {
		t.Fatalf("SelectWorkflowForInitiative unknown initiative err = %v", err)
	}

	doc := &store.SpecDocument{ID: "INIT-ACME-001/prd", InitiativeID: "INIT-ACME-001", RepositoryID: "github.com/acme/app", SpecType: "prd", FilePath: "docs/specs/PRD.md"}
	if err := rs.CreateSpecDocument(ctx, doc); err != nil {
		t.Fatalf("CreateSpecDocument: %v", err)
	}
	got, err := rs.GetSpecDocument(ctx, "INIT-ACME-001/prd")
	if err != nil || got.FilePath != "docs/specs/PRD.md" {
		t.Fatalf("GetSpecDocument = %+v, %v", got, err)
	}
	got.Title = "PRD"
	if err := rs.UpdateSpecDocument(ctx, got); err != nil {
		t.Fatalf("UpdateSpecDocument: %v", err)
	}
	for name, fn := range map[string]func() ([]*store.SpecDocument, error){
		"ListSpecDocuments":             func() ([]*store.SpecDocument, error) { return rs.ListSpecDocuments(ctx) },
		"ListSpecDocumentsByInitiative": func() ([]*store.SpecDocument, error) { return rs.ListSpecDocumentsByInitiative(ctx, "INIT-ACME-001") },
		"ListSpecDocumentsByRepo":       func() ([]*store.SpecDocument, error) { return rs.ListSpecDocumentsByRepo(ctx, "github.com/acme/app") },
	} {
		docs, err := fn()
		if err != nil || len(docs) != 1 || docs[0].Title != "PRD" {
			t.Fatalf("%s = %+v, %v", name, docs, err)
		}
	}
	if err := rs.DeleteSpecDocument(ctx, "INIT-ACME-001/prd"); err != nil {
		t.Fatalf("DeleteSpecDocument: %v", err)
	}
	if _, err := rs.GetSpecDocument(ctx, "INIT-ACME-001/prd"); !errors.Is(err, remote.ErrNotFound) {
		t.Fatalf("GetSpecDocument after delete err = %v", err)
	}
	if err := rs.DeleteSpecDocument(ctx, "INIT-ACME-001/prd"); !errors.Is(err, remote.ErrNotFound) {
		t.Fatalf("DeleteSpecDocument twice err = %v", err)
	}

	jr := &store.JudgeResult{ID: "jr-1", InitiativeID: "INIT-ACME-001", SpecPath: "docs/specs/PRD.md", SpecType: "prd", RubricID: "r", EvaluatedAt: time.Now().UTC()}
	if err := rs.CreateJudgeResult(ctx, jr); err != nil {
		t.Fatalf("CreateJudgeResult: %v", err)
	}
	if err := rs.CreateJudgeResult(ctx, &store.JudgeResult{ID: "jr-2", InitiativeID: "INIT-NOPE-001"}); !errors.Is(err, remote.ErrInvalidReference) {
		t.Fatalf("CreateJudgeResult unknown initiative err = %v", err)
	}
	results, err := rs.ListJudgeResults(ctx, "INIT-ACME-001")
	if err != nil || len(results) != 1 || results[0].ID != "jr-1" {
		t.Fatalf("ListJudgeResults = %+v, %v", results, err)
	}

	// Spec workflow writes stay unsupported.
	if err := rs.CreateSpecWorkflow(ctx, &store.SpecWorkflow{ID: "x"}); !errors.Is(err, remote.ErrNotSupported) {
		t.Fatalf("CreateSpecWorkflow err = %v", err)
	}
}

func TestIdentity(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	ctx := context.Background()

	if err := rs.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	if r := lastRequest(t, fake); r.Path != "/o/acme/api/v1/me" {
		t.Fatalf("Ping path = %q", r.Path)
	}
	tm, err := rs.TenantMe(ctx)
	if err != nil || tm.Tenant != testTenant || tm.Role != "owner" || tm.Principal.ID != remotetest.PrincipalID {
		t.Fatalf("TenantMe = %+v, %v", tm, err)
	}
	id, err := rs.Me(ctx)
	if err != nil || id.Principal.Email != remotetest.PrincipalEmail || len(id.Memberships) != 1 || id.Memberships[0].Tenant != testTenant {
		t.Fatalf("Me = %+v, %v", id, err)
	}
	if r := lastRequest(t, fake); r.Path != "/api/v1/me" {
		t.Fatalf("Me path = %q", r.Path)
	}

	other, err := remote.New(fake.URL, "other", remote.WithToken(testToken))
	if err != nil {
		t.Fatal(err)
	}
	if err := other.Ping(ctx); !errors.Is(err, remote.ErrForbidden) {
		t.Fatalf("Ping non-member err = %v, want ErrForbidden", err)
	}
	if _, err := other.Me(ctx); err != nil {
		t.Fatalf("Me works without tenant membership: %v", err)
	}
	bad, err := remote.New(fake.URL, testTenant, remote.WithToken("wrong"))
	if err != nil {
		t.Fatal(err)
	}
	var apiErr *remote.APIError
	if _, err := bad.Me(ctx); !errors.Is(err, remote.ErrUnauthorized) || !errors.As(err, &apiErr) || apiErr.Code != remote.CodeUnauthenticated {
		t.Fatalf("Me bad token err = %v", err)
	}
}

func TestEmptyListsDecodeEmpty(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	ctx := context.Background()
	as, err := rs.ListAllAssignments(ctx)
	if err != nil || as == nil || len(as) != 0 {
		t.Fatalf("ListAllAssignments = %#v, %v; want empty non-nil", as, err)
	}
	ev, err := rs.ListAllEvidence(ctx)
	if err != nil || ev == nil || len(ev) != 0 {
		t.Fatalf("ListAllEvidence = %#v, %v; want empty non-nil", ev, err)
	}
}
