package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestNormalizeVisibility(t *testing.T) {
	cases := map[string]string{
		"PUBLIC":   "public",
		"public":   "public",
		"PRIVATE":  "private",
		"INTERNAL": "private", // internal is not public
		"":         "unknown",
		"garbage":  "unknown",
	}
	for in, want := range cases {
		if got := NormalizeVisibility(in); got != want {
			t.Errorf("NormalizeVisibility(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRefreshVisibility(t *testing.T) {
	s := store.NewMemStore()
	svc := New(s)
	ctx := context.Background()

	seed := []*store.Repository{
		{ID: "github.com/plexusone/omnidevx", Organization: "plexusone", RepositoryName: "omnidevx", Status: "active"},
		{ID: "github.com/plexusone/secret", Organization: "plexusone", RepositoryName: "secret", Status: "active"},
		{ID: "github.com/plexusone/flaky", Organization: "plexusone", RepositoryName: "flaky", Status: "active", Visibility: "unknown"},
	}
	for _, r := range seed {
		if err := s.CreateRepository(ctx, r); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	lookup := func(_ context.Context, org, name string) (string, error) {
		switch name {
		case "omnidevx":
			return "PUBLIC", nil
		case "secret":
			return "PRIVATE", nil
		default:
			return "", fmt.Errorf("api error")
		}
	}

	res, err := svc.RefreshVisibility(ctx, lookup, "")
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if res.Updated != 2 {
		t.Fatalf("updated = %d, want 2", res.Updated)
	}
	if len(res.Errors) != 1 {
		t.Fatalf("errors = %v, want 1", res.Errors)
	}

	pub, _ := s.GetRepository(ctx, "github.com/plexusone/omnidevx")
	if pub.Visibility != "public" {
		t.Fatalf("visibility = %q, want public", pub.Visibility)
	}
	priv, _ := s.GetRepository(ctx, "github.com/plexusone/secret")
	if priv.Visibility != "private" {
		t.Fatalf("visibility = %q, want private", priv.Visibility)
	}
	// Lookup failure leaves unknown untouched — never upgraded to public.
	flaky, _ := s.GetRepository(ctx, "github.com/plexusone/flaky")
	if flaky.Visibility != "unknown" {
		t.Fatalf("visibility = %q, want unknown preserved", flaky.Visibility)
	}

	// Single-repo scope.
	res2, err := svc.RefreshVisibility(ctx, lookup, "github.com/plexusone/omnidevx")
	if err != nil {
		t.Fatalf("single refresh: %v", err)
	}
	if res2.Updated != 0 || res2.Unchanged != 1 {
		t.Fatalf("second run: %+v, want unchanged", res2)
	}
}
