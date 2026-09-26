package service

import (
	"context"
	"testing"
)

func TestNextRMIID(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	// Empty store: first ID for a slug is 001.
	id, err := svc.NextRMIID(ctx, "github.com/plexusone/systemforge")
	if err != nil {
		t.Fatal(err)
	}
	if id != "RMI-SYSTEMFORGE-001" {
		t.Fatalf("empty store: got %q, want RMI-SYSTEMFORGE-001", id)
	}

	// Seed a few RMIs for the slug, plus one for a different slug and one with
	// a non-conforming ID, none of which should affect the SYSTEMFORGE count.
	seed := []struct{ id, repo string }{
		{"RMI-SYSTEMFORGE-011", "github.com/plexusone/systemforge"},
		{"RMI-SYSTEMFORGE-058", "github.com/plexusone/systemforge"},
		{"RMI-DASHFORGE-017", "github.com/plexusone/dashforge"},
		{"LEGACY-SYSTEMFORGE-999", "github.com/plexusone/systemforge"},
	}
	for _, s := range seed {
		if _, err := svc.CreateRMI(ctx, s.id, s.repo, "", "", "seed", "", "capability", "", true, 0, nil); err != nil {
			t.Fatalf("seed %s: %v", s.id, err)
		}
	}

	// Next is max(011,058)+1 = 059; the DASHFORGE and malformed IDs are ignored.
	id, err = svc.NextRMIID(ctx, "github.com/plexusone/systemforge")
	if err != nil {
		t.Fatal(err)
	}
	if id != "RMI-SYSTEMFORGE-059" {
		t.Fatalf("got %q, want RMI-SYSTEMFORGE-059", id)
	}

	// A different slug is numbered independently.
	id, err = svc.NextRMIID(ctx, "github.com/plexusone/dashforge")
	if err != nil {
		t.Fatal(err)
	}
	if id != "RMI-DASHFORGE-018" {
		t.Fatalf("got %q, want RMI-DASHFORGE-018", id)
	}

	// An allocated ID actually creates without collision.
	next, err := svc.NextRMIID(ctx, "github.com/plexusone/systemforge")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRMI(ctx, next, "github.com/plexusone/systemforge", "", "", "real", "", "capability", "", true, 0, nil); err != nil {
		t.Fatalf("create with allocated id %s: %v", next, err)
	}
}

func TestNextRMIIDErrors(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	if _, err := svc.NextRMIID(ctx, "/"); err == nil {
		t.Fatal("expected error for repo with empty slug")
	}
}

func TestNextInitiativeID(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	id, err := svc.NextInitiativeID(ctx, "systemforge")
	if err != nil {
		t.Fatal(err)
	}
	if id != "INIT-SYSTEMFORGE-001" {
		t.Fatalf("empty store: got %q, want INIT-SYSTEMFORGE-001", id)
	}

	for _, iid := range []string{"INIT-SYSTEMFORGE-001", "INIT-SYSTEMFORGE-003", "INIT-OTHER-009"} {
		if _, err := svc.CreateInitiative(ctx, iid, "org", "seed", "", "", "", ""); err != nil {
			t.Fatalf("seed %s: %v", iid, err)
		}
	}

	// max(001,003)+1 = 004; the OTHER slug does not count.
	id, err = svc.NextInitiativeID(ctx, "systemforge")
	if err != nil {
		t.Fatal(err)
	}
	if id != "INIT-SYSTEMFORGE-004" {
		t.Fatalf("got %q, want INIT-SYSTEMFORGE-004", id)
	}

	// A repository ID resolves to its final path segment as the slug.
	id, err = svc.NextInitiativeID(ctx, "github.com/plexusone/systemforge")
	if err != nil {
		t.Fatal(err)
	}
	if id != "INIT-SYSTEMFORGE-004" {
		t.Fatalf("repo-derived slug: got %q, want INIT-SYSTEMFORGE-004", id)
	}

	if _, err := svc.NextInitiativeID(ctx, ""); err == nil {
		t.Fatal("expected error for empty slug")
	}
}
