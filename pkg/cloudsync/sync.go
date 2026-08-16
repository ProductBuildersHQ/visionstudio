// Package cloudsync drives Dolt-native sync from the local app to a
// VisionStudio Cloud tenant remote via godolt.
//
// Per the RMI-VISIONSTUDIO-205 spike finding, this is currently
// whole-database push — dogfood-only. The local Dolt database holds every
// registered repository's data in one set of tables (no per-tenant
// partitioning locally), so a full push carries fields the sync contract
// excludes (e.g. local_path) unfiltered. A synced-entity projection
// database is required before this ships to any external tenant; every
// caller of Push must surface DogfoodOnly to the operator.
package cloudsync

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/grokify/godolt"
)

// remoteName is the Dolt remote name used for a tenant sync. Namespaced
// so it never collides with a developer's own git/dolt remotes.
func remoteName(tenantSlug string) string {
	return "cloud-" + tenantSlug
}

// PushResult reports the outcome of a sync push.
type PushResult struct {
	TenantSlug  string
	RemoteName  string
	RemoteURL   string
	Message     string
	DogfoodOnly bool // always true until the projection DB ships (T5a)
}

// Push ensures a Dolt remote named for the tenant exists (adding it if
// necessary) and pushes the active branch to it.
func Push(ctx context.Context, db *sql.DB, branch, tenantSlug, remoteURL string) (*PushResult, error) {
	if tenantSlug == "" || remoteURL == "" {
		return nil, fmt.Errorf("cloudsync: tenant slug and remote URL are required")
	}
	c := godolt.New(db)
	name := remoteName(tenantSlug)

	existing, err := c.Remotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudsync: list remotes: %w", err)
	}
	found := false
	for _, r := range existing {
		if r.Name == name {
			found = true
			break
		}
	}
	if !found {
		if err := c.RemoteAdd(ctx, name, remoteURL); err != nil {
			return nil, fmt.Errorf("cloudsync: add remote for tenant %s: %w", tenantSlug, err)
		}
	}

	if branch == "" {
		branch, err = c.ActiveBranch(ctx)
		if err != nil {
			return nil, fmt.Errorf("cloudsync: active branch: %w", err)
		}
	}

	msg, err := c.Push(ctx, name, branch)
	if err != nil {
		return nil, fmt.Errorf("cloudsync: push to tenant %s: %w", tenantSlug, err)
	}
	return &PushResult{
		TenantSlug:  tenantSlug,
		RemoteName:  name,
		RemoteURL:   remoteURL,
		Message:     msg,
		DogfoodOnly: true,
	}, nil
}
