package doltstore

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store/storetest"
)

// newTestStore boots a throwaway dolt sql-server in a temp directory on a
// random port, migrates the schema, and returns a connected DoltStore.
// Skips when the dolt binary is unavailable. This exists because MemStore
// cannot catch ent edge-semantics bugs (an O2M-vs-M2M defect shipped past
// the mem-backed suites and was only caught by live data).
func newTestStore(t *testing.T) *DoltStore {
	t.Helper()
	if _, err := exec.LookPath("dolt"); err != nil {
		t.Skip("dolt binary not available")
	}

	root := t.TempDir()
	dbDir := filepath.Join(root, "visionstudio")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	initCmd := exec.Command("dolt", "init", "--name", "test", "--email", "test@local")
	initCmd.Dir = dbDir
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("dolt init: %v\n%s", err, out)
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	_ = lis.Close()

	// #nosec G204 -- test harness; port comes from net.Listen on loopback.
	srv := exec.Command("dolt", "sql-server", "--host", "127.0.0.1", "--port", fmt.Sprintf("%d", port))
	srv.Dir = root
	if err := srv.Start(); err != nil {
		t.Fatalf("start dolt sql-server: %v", err)
	}
	t.Cleanup(func() {
		_ = srv.Process.Kill()
		_, _ = srv.Process.Wait()
	})

	dsn := fmt.Sprintf("root:@tcp(127.0.0.1:%d)/visionstudio", port)
	var ds *DoltStore
	deadline := time.Now().Add(15 * time.Second)
	for {
		ds, err = New(dsn)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			pingErr := ds.Ping(ctx)
			cancel()
			if pingErr == nil {
				break
			}
			_ = ds.Close()
			err = pingErr
		}
		if time.Now().After(deadline) {
			t.Fatalf("dolt sql-server not ready: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Cleanup(func() { _ = ds.Close() })

	if err := ds.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return ds
}

func TestDoltStoreRoundTrips(t *testing.T) {
	ds := newTestStore(t)
	ctx := context.Background()
	now := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)

	// Fixtures shared by the subtests.
	org := &store.Organization{ID: "github.com/testorg", Login: "testorg", Kind: "organization", CreatedAt: now, UpdatedAt: now}
	if err := ds.CreateOrganization(ctx, org); err != nil {
		t.Fatalf("create org: %v", err)
	}
	repo := &store.Repository{
		ID: "github.com/testorg/app", Organization: "testorg", RepositoryName: "app",
		DefaultBranch: "main", Status: "active", OrganizationID: org.ID, Visibility: "public",
	}
	if err := ds.CreateRepository(ctx, repo); err != nil {
		t.Fatalf("create repo: %v", err)
	}
	inits := []*store.Initiative{
		{ID: "INIT-APP-001", Organization: "default", Title: "One", Status: "executing", Visibility: "public", CreatedAt: now, UpdatedAt: now},
		{ID: "INIT-APP-002", Organization: "default", Title: "Two", Status: "executing", CreatedAt: now, UpdatedAt: now},
	}
	for _, in := range inits {
		if err := ds.CreateInitiative(ctx, in); err != nil {
			t.Fatalf("create initiative: %v", err)
		}
	}
	rmi := &store.RoadmapItem{
		ID: "RMI-APP-001", RepositoryID: repo.ID, InitiativeID: "INIT-APP-001",
		Title: "x", ItemType: "capability", Status: "completed", CreatedAt: now, UpdatedAt: now,
	}
	if err := ds.CreateRMI(ctx, rmi); err != nil {
		t.Fatalf("create rmi: %v", err)
	}

	t.Run("phase update saves title and theme", func(t *testing.T) {
		ph := &store.Phase{ID: "INIT-APP-001/phase-1", InitiativeID: "INIT-APP-001", SequenceNumber: 1, Title: "Adoption", Theme: "Old"}
		if err := ds.CreatePhase(ctx, ph); err != nil {
			t.Fatal(err)
		}
		if err := ds.UpdatePhase(ctx, &store.Phase{ID: ph.ID, Title: "First product"}); err != nil {
			t.Fatal(err)
		}
		phases, err := ds.ListPhases(ctx, "INIT-APP-001")
		if err != nil {
			t.Fatal(err)
		}
		if len(phases) != 1 || phases[0].Title != "First product" || phases[0].Theme != "" {
			t.Fatalf("phases = %+v, want the new title and no theme", phases)
		}
		if err := ds.UpdatePhase(ctx, &store.Phase{ID: "INIT-APP-001/phase-9", Title: "x"}); err == nil {
			t.Fatal("updating a missing phase must fail")
		}
	})

	t.Run("dependency delete removes only the named edge", func(t *testing.T) {
		other := &store.RoadmapItem{
			ID: "RMI-APP-002", RepositoryID: repo.ID, InitiativeID: "INIT-APP-001",
			Title: "y", ItemType: "capability", Status: "proposed", CreatedAt: now, UpdatedAt: now,
		}
		if err := ds.CreateRMI(ctx, other); err != nil {
			t.Fatal(err)
		}
		for _, d := range []*store.RMIDependency{
			{SourceRMIID: "RMI-APP-002", TargetRMIID: "RMI-APP-001", Relationship: "requires"},
			{SourceRMIID: "RMI-APP-001", TargetRMIID: "RMI-APP-002", Relationship: "relates"},
		} {
			if err := ds.CreateDependency(ctx, d); err != nil {
				t.Fatal(err)
			}
		}
		n, err := ds.DeleteDependency(ctx, "RMI-APP-002", "RMI-APP-001")
		if err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatalf("removed %d edges, want 1", n)
		}
		deps, err := ds.ListDependencies(ctx, "RMI-APP-002")
		if err != nil {
			t.Fatal(err)
		}
		if len(deps) != 1 || deps[0].SourceRMIID != "RMI-APP-001" {
			t.Fatalf("remaining edges = %+v, want only RMI-APP-001 -> RMI-APP-002", deps)
		}
	})

	t.Run("repository fields round-trip", func(t *testing.T) {
		got, err := ds.GetRepository(ctx, repo.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.OrganizationID != org.ID || got.Visibility != "public" {
			t.Fatalf("round-trip lost fields: %+v", got)
		}
	})

	t.Run("initiative visibility round-trip and default", func(t *testing.T) {
		got, err := ds.GetInitiative(ctx, "INIT-APP-001")
		if err != nil {
			t.Fatal(err)
		}
		if got.Visibility != "public" {
			t.Fatalf("visibility = %q", got.Visibility)
		}
		def, err := ds.GetInitiative(ctx, "INIT-APP-002")
		if err != nil {
			t.Fatal(err)
		}
		if def.Visibility != "internal" {
			t.Fatalf("default visibility = %q, want internal", def.Visibility)
		}
	})

	t.Run("person org affiliation M2M", func(t *testing.T) {
		p := &store.Person{
			ID: "person:tester", GitHubLogin: "tester",
			EmailIdentities: []string{"t@local"},
			OrgIDs:          []string{org.ID},
			CreatedAt:       now, UpdatedAt: now,
		}
		if err := ds.CreatePerson(ctx, p); err != nil {
			t.Fatal(err)
		}
		got, err := ds.GetPerson(ctx, p.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.OrgIDs) != 1 || got.OrgIDs[0] != org.ID {
			t.Fatalf("org IDs = %v", got.OrgIDs)
		}
		if len(got.EmailIdentities) != 1 {
			t.Fatalf("emails = %v", got.EmailIdentities)
		}
	})

	t.Run("release M2M regression: one initiative across releases", func(t *testing.T) {
		// The O2M defect allowed an initiative to join only ONE release;
		// two releases sharing INIT-APP-001 is the regression case.
		for i, tag := range []string{"v0.1.0", "v0.2.0"} {
			rel := &store.Release{
				ID: repo.ID + "@" + tag, RepositoryID: repo.ID, Tag: tag,
				ReleasedAt:    now.AddDate(0, 0, i),
				InitiativeIDs: []string{"INIT-APP-001"},
				RMIIDs:        []string{"RMI-APP-001"},
				Body:          "Release notes for " + tag, // round-trip check below
				CreatedAt:     now, UpdatedAt: now,
			}
			if err := ds.CreateRelease(ctx, rel); err != nil {
				t.Fatalf("create release %s (M2M regression?): %v", tag, err)
			}
		}
		byInit, err := ds.ListReleasesByInitiative(ctx, "INIT-APP-001")
		if err != nil {
			t.Fatal(err)
		}
		if len(byInit) != 2 {
			t.Fatalf("releases for initiative = %d, want 2", len(byInit))
		}
		// Newest first.
		if byInit[0].Tag != "v0.2.0" {
			t.Fatalf("order: first = %s, want v0.2.0", byInit[0].Tag)
		}
		// Associations and Body survive the round-trip on both.
		for _, r := range byInit {
			if len(r.InitiativeIDs) != 1 || len(r.RMIIDs) != 1 {
				t.Fatalf("associations lost on %s: %+v", r.Tag, r)
			}
			if r.Body != "Release notes for "+r.Tag {
				t.Fatalf("body lost on %s: %q", r.Tag, r.Body)
			}
		}
	})

	t.Run("release update replaces associations", func(t *testing.T) {
		id := repo.ID + "@v0.1.0"
		rel, err := ds.GetRelease(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		rel.InitiativeIDs = []string{"INIT-APP-002"}
		rel.RMIIDs = nil
		rel.UpdatedAt = now
		if err := ds.UpdateRelease(ctx, rel); err != nil {
			t.Fatal(err)
		}
		got, err := ds.GetRelease(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.InitiativeIDs) != 1 || got.InitiativeIDs[0] != "INIT-APP-002" || len(got.RMIIDs) != 0 {
			t.Fatalf("update result: %+v", got)
		}
	})

	t.Run("release delete", func(t *testing.T) {
		id := repo.ID + "@v0.1.0"
		if err := ds.DeleteRelease(ctx, id); err != nil {
			t.Fatal(err)
		}
		if _, err := ds.GetRelease(ctx, id); err == nil {
			t.Fatal("expected not-found after delete")
		}
	})
}

func TestDoltStoreRenameRepository(t *testing.T) {
	ds := newTestStore(t)
	storetest.RenameRepository(t, ds)

	// The rename is recorded as one Dolt commit, so it can be reverted as one.
	ctx := context.Background()
	if dirty, err := ds.HasUncommittedChanges(ctx); err != nil || dirty {
		t.Errorf("working set after rename: dirty=%v err=%v", dirty, err)
	}
	var msg string
	if err := ds.db.QueryRowContext(ctx, "SELECT message FROM dolt_log ORDER BY date DESC LIMIT 1").Scan(&msg); err != nil {
		t.Fatal(err)
	}
	if want := "visionstudio: rename repository github.com/acme/old-name to github.com/acme/new-name"; msg != want {
		t.Errorf("last commit = %q, want %q", msg, want)
	}
}

func TestDoltStoreRenameRepositoryFinishesAPartialRename(t *testing.T) {
	storetest.RenameRepositoryFinishesAPartialRename(t, newTestStore(t))
}

// The store a unit of work hands its function is bound to the transaction, so
// an error rolls back every write, and success commits them as one Dolt commit.
func TestDoltUnitOfWorkIsATransaction(t *testing.T) {
	ds := newTestStore(t)
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	initiative := func(id string) *store.Initiative {
		return &store.Initiative{ID: id, Organization: "default", Title: id, Status: "draft", CreatedAt: now, UpdatedAt: now}
	}
	failure := errors.New("stop here")

	t.Run("an error rolls back every write", func(t *testing.T) {
		err := NewUnitOfWork(ds, "test").Execute(ctx, func(ctx context.Context, s store.Store) error {
			if err := s.CreateInitiative(ctx, initiative("INIT-UOW-001")); err != nil {
				return err
			}
			if err := s.CreateInitiative(ctx, initiative("INIT-UOW-002")); err != nil {
				return err
			}
			return failure
		})
		if !errors.Is(err, failure) {
			t.Fatalf("err = %v, want the function's error", err)
		}
		for _, id := range []string{"INIT-UOW-001", "INIT-UOW-002"} {
			if _, err := ds.GetInitiative(ctx, id); err == nil {
				t.Errorf("%s was written although the unit of work failed", id)
			}
		}
	})

	t.Run("success commits once", func(t *testing.T) {
		err := NewUnitOfWork(ds, "alice").Execute(ctx, func(ctx context.Context, s store.Store) error {
			return s.CreateInitiative(ctx, initiative("INIT-UOW-003"))
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ds.GetInitiative(ctx, "INIT-UOW-003"); err != nil {
			t.Fatalf("the write did not land: %v", err)
		}
		var msg string
		if err := ds.db.QueryRowContext(ctx, "SELECT message FROM dolt_log ORDER BY date DESC LIMIT 1").Scan(&msg); err != nil {
			t.Fatal(err)
		}
		if msg != "visionstudio: alice" {
			t.Errorf("last commit = %q", msg)
		}
	})

	t.Run("a rename joins the transaction and rolls back with it", func(t *testing.T) {
		for _, id := range []string{"github.com/uow/old", "github.com/uow/other"} {
			if err := ds.CreateRepository(ctx, &store.Repository{ID: id, Organization: "uow", RepositoryName: id[len("github.com/uow/"):], DefaultBranch: "main", Status: "active"}); err != nil {
				t.Fatal(err)
			}
		}
		from, err := ds.GetRepository(ctx, "github.com/uow/old")
		if err != nil {
			t.Fatal(err)
		}
		to := *from
		to.ID, to.RepositoryName = "github.com/uow/new", "new"
		err = NewUnitOfWork(ds, "test").Execute(ctx, func(ctx context.Context, s store.Store) error {
			if _, err := s.RenameRepository(ctx, from, &to, store.RenameRepositoryOptions{}); err != nil {
				return err
			}
			return failure
		})
		if !errors.Is(err, failure) {
			t.Fatalf("err = %v", err)
		}
		if _, err := ds.GetRepository(ctx, "github.com/uow/new"); err == nil {
			t.Error("the renamed record survived the rollback")
		}
		if r, err := ds.GetRepository(ctx, "github.com/uow/old"); err != nil || r.Status != "active" {
			t.Errorf("the old record = %+v (%v), want it untouched", r, err)
		}
	})

	t.Run("what cannot run inside a unit of work is refused", func(t *testing.T) {
		err := NewUnitOfWork(ds, "test").Execute(ctx, func(ctx context.Context, s store.Store) error {
			inner, ok := s.(*DoltStore)
			if !ok {
				t.Fatalf("store is %T", s)
			}
			if err := inner.Commit(ctx, "mid-transaction"); err == nil {
				t.Error("a Dolt commit inside the transaction was accepted")
			}
			if err := NewUnitOfWork(inner, "nested").Execute(ctx, func(context.Context, store.Store) error { return nil }); err == nil {
				t.Error("a nested unit of work was accepted")
			}
			if _, err := s.RenameRepository(ctx, &store.Repository{ID: "github.com/uow/old"}, &store.Repository{ID: "github.com/uow/x"}, store.RenameRepositoryOptions{DryRun: true}); err == nil {
				t.Error("a dry-run rename inside the transaction was accepted")
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}
