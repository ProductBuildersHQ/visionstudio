package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// RenameRepositoryRequest says how a renamed repository's record is filled in.
type RenameRepositoryRequest struct {
	// LocalPath is the new working-tree path. Empty keeps the old one.
	LocalPath string
	// GoModule is the new Go module path. Empty keeps the old one, except that
	// a module path that was the old repository ID, or under it, follows the
	// rename.
	GoModule string
	// MoveReleases also moves the releases, which otherwise stay on the old ID.
	MoveReleases bool
	// DryRun reports what would move and changes nothing.
	DryRun bool
}

var repoIDPattern = regexp.MustCompile(`^github\.com/([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+)$`)

// RenameRepository moves the repository oldID to newID, with everything that
// references it, and archives oldID as superseded by newID. Running it again
// after a partial manual rename, where newID was already registered or some
// references were already moved, finishes the job.
func (s *Service) RenameRepository(ctx context.Context, oldID, newID string, req RenameRepositoryRequest) (*store.Repository, *store.RenameRepositoryResult, error) {
	m := repoIDPattern.FindStringSubmatch(newID)
	if m == nil {
		return nil, nil, fmt.Errorf("new repository ID %q is not of the form github.com/<org>/<name>", newID)
	}
	if oldID == newID {
		return nil, nil, fmt.Errorf("the new repository ID is the same as the old one")
	}
	old, err := s.Store.GetRepository(ctx, oldID)
	if err != nil {
		return nil, nil, err
	}
	if old.SupersededBy != "" && old.SupersededBy != newID {
		return nil, nil, fmt.Errorf("repository %s is already superseded by %s", oldID, old.SupersededBy)
	}
	if existing, err := s.Store.GetRepository(ctx, newID); err == nil && existing.Status == "archived" {
		return nil, nil, fmt.Errorf("repository %s is archived and cannot receive a rename", newID)
	}

	to := *old
	to.ID = newID
	to.Organization = m[1]
	to.RepositoryName = m[2]
	to.Status = "active"
	to.SupersededBy = ""
	if req.LocalPath != "" {
		to.LocalPath = req.LocalPath
	}
	switch {
	case req.GoModule != "":
		to.GoModule = req.GoModule
	case old.GoModule == oldID || strings.HasPrefix(old.GoModule, oldID+"/"):
		to.GoModule = newID + strings.TrimPrefix(old.GoModule, oldID)
	}
	if to.Organization != old.Organization {
		to.OrganizationID = ""
		if !req.DryRun {
			orgRow, _, err := s.EnsureOrganization(ctx, to.Organization, "")
			if err != nil {
				return nil, nil, fmt.Errorf("ensure organization %s: %w", to.Organization, err)
			}
			to.OrganizationID = orgRow.ID
		}
	}

	res, err := s.Store.RenameRepository(ctx, old, &to, store.RenameRepositoryOptions{MoveReleases: req.MoveReleases, DryRun: req.DryRun})
	if err != nil {
		return nil, nil, err
	}
	return &to, res, nil
}
