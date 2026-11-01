package service

import (
	"context"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func renameService(t *testing.T) *Service {
	t.Helper()
	s := New(store.NewMemStore())
	ctx := context.Background()
	if _, err := s.RegisterRepository(ctx, "acme", "old-name", "main", "/src/old-name", ""); err != nil {
		t.Fatal(err)
	}
	old, err := s.GetRepository(ctx, "github.com/acme/old-name")
	if err != nil {
		t.Fatal(err)
	}
	old.GoModule = "github.com/acme/old-name/v2"
	if err := s.Store.UpdateRepository(ctx, old); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestRenameRepositoryFillsInTheNewRecord(t *testing.T) {
	s := renameService(t)
	to, res, err := s.RenameRepository(context.Background(), "github.com/acme/old-name", "github.com/acme/new-name", RenameRepositoryRequest{LocalPath: "/src/new-name"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Created {
		t.Error("the new record should have been created")
	}
	if to.RepositoryName != "new-name" || to.Organization != "acme" || to.LocalPath != "/src/new-name" {
		t.Errorf("new record = %+v", to)
	}
	// A module path under the old ID follows the rename.
	if to.GoModule != "github.com/acme/new-name/v2" {
		t.Errorf("go module = %q", to.GoModule)
	}
}

func TestRenameRepositoryToAnotherOrganization(t *testing.T) {
	s := renameService(t)
	to, _, err := s.RenameRepository(context.Background(), "github.com/acme/old-name", "github.com/otherorg/old-name", RenameRepositoryRequest{GoModule: "example.com/mod"})
	if err != nil {
		t.Fatal(err)
	}
	if to.Organization != "otherorg" || to.OrganizationID == "" || to.GoModule != "example.com/mod" {
		t.Errorf("new record = %+v", to)
	}
	if to.LocalPath != "/src/old-name" {
		t.Errorf("an unchanged path should be kept, got %q", to.LocalPath)
	}
}

func TestRenameRepositoryRefuses(t *testing.T) {
	ctx := context.Background()
	for name, tc := range map[string]struct {
		setup  func(t *testing.T, s *Service)
		newID  string
		reason string
	}{
		"a new ID that is not a repository ID": {nil, "acme/new-name", "github.com/<org>/<name>"},
		"the same ID":                          {nil, "github.com/acme/old-name", "same as the old one"},
		"a repository superseded by another": {func(t *testing.T, s *Service) {
			r, err := s.GetRepository(ctx, "github.com/acme/old-name")
			if err != nil {
				t.Fatal(err)
			}
			r.Status, r.SupersededBy = "archived", "github.com/acme/elsewhere"
			if err := s.Store.UpdateRepository(ctx, r); err != nil {
				t.Fatal(err)
			}
		}, "github.com/acme/new-name", "already superseded by github.com/acme/elsewhere"},
		"an archived target": {func(t *testing.T, s *Service) {
			r, err := s.RegisterRepository(ctx, "acme", "new-name", "main", "", "")
			if err != nil {
				t.Fatal(err)
			}
			r.Status = "archived"
			if err := s.Store.UpdateRepository(ctx, r); err != nil {
				t.Fatal(err)
			}
		}, "github.com/acme/new-name", "archived and cannot receive"},
	} {
		t.Run(name, func(t *testing.T) {
			s := renameService(t)
			if tc.setup != nil {
				tc.setup(t, s)
			}
			_, _, err := s.RenameRepository(ctx, "github.com/acme/old-name", tc.newID, RenameRepositoryRequest{})
			if err == nil || !strings.Contains(err.Error(), tc.reason) {
				t.Errorf("err = %v, want it to mention %q", err, tc.reason)
			}
		})
	}
	s := renameService(t)
	if _, _, err := s.RenameRepository(ctx, "github.com/acme/missing", "github.com/acme/new-name", RenameRepositoryRequest{}); err == nil {
		t.Error("renaming a repository that is not registered was accepted")
	}
}

// A dry run creates no organization, even when the rename crosses into a new
// one.
func TestRenameRepositoryDryRunChangesNothing(t *testing.T) {
	s := renameService(t)
	ctx := context.Background()
	before, err := s.Store.ListOrganizations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.RenameRepository(ctx, "github.com/acme/old-name", "github.com/neworg/old-name", RenameRepositoryRequest{DryRun: true}); err != nil {
		t.Fatal(err)
	}
	after, err := s.Store.ListOrganizations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Errorf("a dry run created an organization: %d -> %d", len(before), len(after))
	}
	if r, _ := s.GetRepository(ctx, "github.com/acme/old-name"); r == nil || r.Status != "active" {
		t.Errorf("a dry run archived the old record: %+v", r)
	}
}
