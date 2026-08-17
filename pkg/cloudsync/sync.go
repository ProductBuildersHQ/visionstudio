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
	"errors"
	"fmt"
	"strings"

	"github.com/grokify/godolt"
)

// remoteName is the Dolt remote name used for a tenant sync. Namespaced
// so it never collides with a developer's own git/dolt remotes.
func remoteName(tenantSlug string) string {
	return "cloud-" + tenantSlug
}

// ErrNonFastForward is returned by Push when the tenant remote has
// commits the local branch doesn't have. Dolt's DOLT_PUSH already
// enforces fast-forward-only at the remote (verified against a real
// diverged-history push before this was written: the raw rejection
// contains "non-fast-forward"); this is the entire multi-user
// concurrency model (RMI-VISIONSTUDIO-535, cloud-security ADR-002's
// git model) — the cloud never merges divergent user work server-side,
// so the caller must Pull and resolve locally before pushing again.
var ErrNonFastForward = errors.New("cloudsync: push rejected: the tenant remote has commits you don't have — pull first")

// ensureRemote registers the tenant's Dolt remote if it isn't already
// configured. Shared by Push and Pull so both are idempotent regardless
// of which ran first.
func ensureRemote(ctx context.Context, c *godolt.Client, name, url string) error {
	existing, err := c.Remotes(ctx)
	if err != nil {
		return fmt.Errorf("cloudsync: list remotes: %w", err)
	}
	for _, r := range existing {
		if r.Name == name {
			return nil
		}
	}
	if err := c.RemoteAdd(ctx, name, url); err != nil {
		return fmt.Errorf("cloudsync: add remote: %w", err)
	}
	return nil
}

// resolveBranch returns branch unchanged if set, otherwise the
// connection's active branch.
func resolveBranch(ctx context.Context, c *godolt.Client, branch string) (string, error) {
	if branch != "" {
		return branch, nil
	}
	b, err := c.ActiveBranch(ctx)
	if err != nil {
		return "", fmt.Errorf("cloudsync: active branch: %w", err)
	}
	return b, nil
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
// necessary) and pushes the active branch to it. Returns ErrNonFastForward
// (wrapped) if the tenant remote has diverged — see that error's comment.
func Push(ctx context.Context, db *sql.DB, branch, tenantSlug, remoteURL string) (*PushResult, error) {
	if tenantSlug == "" || remoteURL == "" {
		return nil, fmt.Errorf("cloudsync: tenant slug and remote URL are required")
	}
	c := godolt.New(db)
	name := remoteName(tenantSlug)

	if err := ensureRemote(ctx, c, name, remoteURL); err != nil {
		return nil, err
	}
	branch, err := resolveBranch(ctx, c, branch)
	if err != nil {
		return nil, err
	}

	msg, err := c.Push(ctx, name, branch)
	if err != nil {
		if strings.Contains(err.Error(), "non-fast-forward") {
			return nil, fmt.Errorf("%w (tenant %s)", ErrNonFastForward, tenantSlug)
		}
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

// PullResult reports the outcome of a sync pull.
type PullResult struct {
	TenantSlug  string
	RemoteName  string
	RemoteURL   string
	Message     string
	FastForward bool
	// Conflicts is the number of tables DOLT_PULL left in a conflicted
	// state after merging. Non-zero is not an error: the merge completed
	// and the caller must resolve conflict rows in dolt_conflicts_<table>
	// locally before the branch is clean again — this is the local
	// resolution half of the git model (cloud-security ADR-002).
	Conflicts int
}

// Pull ensures a Dolt remote named for the tenant exists (adding it if
// necessary), fetches, and merges the tenant remote's branch into the
// active branch.
func Pull(ctx context.Context, db *sql.DB, branch, tenantSlug, remoteURL string) (*PullResult, error) {
	if tenantSlug == "" || remoteURL == "" {
		return nil, fmt.Errorf("cloudsync: tenant slug and remote URL are required")
	}
	c := godolt.New(db)
	name := remoteName(tenantSlug)

	if err := ensureRemote(ctx, c, name, remoteURL); err != nil {
		return nil, err
	}
	branch, err := resolveBranch(ctx, c, branch)
	if err != nil {
		return nil, err
	}

	res, err := c.Pull(ctx, name, branch)
	if err != nil {
		return nil, fmt.Errorf("cloudsync: pull from tenant %s: %w", tenantSlug, err)
	}
	return &PullResult{
		TenantSlug:  tenantSlug,
		RemoteName:  name,
		RemoteURL:   remoteURL,
		Message:     res.Message,
		FastForward: res.FastForward,
		Conflicts:   res.Conflicts,
	}, nil
}
