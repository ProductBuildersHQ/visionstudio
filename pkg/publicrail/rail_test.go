package publicrail

import (
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func repo(vis string) *store.Repository {
	return &store.Repository{ID: "github.com/x/" + vis, Visibility: vis}
}

func TestRepoPublic(t *testing.T) {
	if !RepoPublic(repo("public")) {
		t.Fatal("public repo must pass")
	}
	for _, vis := range []string{"private", "unknown", "", "PUBLIC", "internal"} {
		if RepoPublic(repo(vis)) {
			t.Fatalf("visibility %q must not pass — only exact 'public'", vis)
		}
	}
	if RepoPublic(nil) {
		t.Fatal("nil repo must not pass")
	}
}

func TestInitiativePublic(t *testing.T) {
	if !InitiativePublic(&store.Initiative{ID: "i", Visibility: "public"}) {
		t.Fatal("public initiative must pass")
	}
	cases := []*store.Initiative{
		{ID: "internal", Visibility: "internal"},
		{ID: "default-empty"},
		{ID: "hidden-public", Visibility: "public", Hidden: true},
	}
	for _, i := range cases {
		if InitiativePublic(i) {
			t.Fatalf("initiative %s must not pass", i.ID)
		}
	}
	if InitiativePublic(nil) {
		t.Fatal("nil initiative must not pass")
	}
}

func TestFilterRepos(t *testing.T) {
	in := []*store.Repository{repo("public"), repo("private"), repo("unknown"), repo("public")}
	out := FilterRepos(in)
	if len(out) != 2 {
		t.Fatalf("filtered = %d, want 2", len(out))
	}
	for _, r := range out {
		if r.Visibility != "public" {
			t.Fatalf("leaked %q", r.Visibility)
		}
	}
}

func TestInitiativeAllowed(t *testing.T) {
	pub := &store.Initiative{ID: "i", Visibility: "public"}
	internal := &store.Initiative{ID: "j"}

	// Public initiative, no repo references: allowed (nothing repo-identifying).
	if !InitiativeAllowed(pub, nil) {
		t.Fatal("public initiative without repos must pass")
	}
	// Public initiative, all repos private/unknown: exports nothing.
	if InitiativeAllowed(pub, []*store.Repository{repo("private"), repo("unknown")}) {
		t.Fatal("public initiative with no public repos must NOT pass")
	}
	// Public initiative, mixed repos: allowed (payload names public subset only).
	if !InitiativeAllowed(pub, []*store.Repository{repo("private"), repo("public")}) {
		t.Fatal("public initiative with one public repo must pass")
	}
	// Internal initiative never passes regardless of repos.
	if InitiativeAllowed(internal, []*store.Repository{repo("public")}) {
		t.Fatal("internal initiative must never pass")
	}
}
