package doltstore

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
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
