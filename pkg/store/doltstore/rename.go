package doltstore

import (
	"context"
	"fmt"
	"slices"

	"github.com/ProductBuildersHQ/visionstudio/ent"
	"github.com/ProductBuildersHQ/visionstudio/ent/devxperiodreport"
	"github.com/ProductBuildersHQ/visionstudio/ent/initiative"
	"github.com/ProductBuildersHQ/visionstudio/ent/prismdocument"
	"github.com/ProductBuildersHQ/visionstudio/ent/prismgoal"
	"github.com/ProductBuildersHQ/visionstudio/ent/prismroadmap"
	"github.com/ProductBuildersHQ/visionstudio/ent/release"
	"github.com/ProductBuildersHQ/visionstudio/ent/repository"
	"github.com/ProductBuildersHQ/visionstudio/ent/repositorydependency"
	"github.com/ProductBuildersHQ/visionstudio/ent/roadmapitem"
	"github.com/ProductBuildersHQ/visionstudio/ent/specdocument"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// RenameRepository moves a repository to a new ID inside one transaction. A
// dry run does the same work and rolls it back, so its counts are exact. On
// success the change is recorded as one Dolt commit, so it can be reverted as
// one step.
func (d *DoltStore) RenameRepository(ctx context.Context, from, to *store.Repository, opts store.RenameRepositoryOptions) (*store.RenameRepositoryResult, error) {
	if from == nil || to == nil || from.ID == "" || to.ID == "" || from.ID == to.ID {
		return nil, fmt.Errorf("rename repository: need two different repository IDs")
	}
	tx, err := d.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	res, err := renameRepository(ctx, tx.Client(), from, to, opts)
	if err != nil || opts.DryRun {
		if rerr := tx.Rollback(); rerr != nil {
			if err != nil {
				return nil, fmt.Errorf("rollback failed: %v (original: %w)", rerr, err)
			}
			return nil, fmt.Errorf("roll back dry run: %w", rerr)
		}
		return res, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	if err := d.Commit(ctx, fmt.Sprintf("visionstudio: rename repository %s to %s", from.ID, to.ID)); err != nil {
		return nil, err
	}
	return res, nil
}

func renameRepository(ctx context.Context, c *ent.Client, from, to *store.Repository, opts store.RenameRepositoryOptions) (*store.RenameRepositoryResult, error) {
	if _, err := c.Repository.Get(ctx, from.ID); err != nil {
		return nil, fmt.Errorf("repository %s not found: %w", from.ID, err)
	}
	res := &store.RenameRepositoryResult{}
	exists, err := c.Repository.Query().Where(repository.ID(to.ID)).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("look up %s: %w", to.ID, err)
	}
	if !exists {
		// The new record must exist before rows can point at it.
		if err := createRepository(ctx, c, to); err != nil {
			return nil, err
		}
		res.Created = true
	}

	moves := []struct {
		count *int
		name  string
		run   func() (int, error)
	}{
		{&res.RMIs, "RMIs", func() (int, error) {
			return c.RoadmapItem.Update().Where(roadmapitem.HasRepositoryWith(repository.ID(from.ID))).SetRepositoryID(to.ID).Save(ctx)
		}},
		{&res.SpecDocuments, "spec documents", func() (int, error) {
			return c.SpecDocument.Update().Where(specdocument.RepositoryID(from.ID)).SetRepositoryID(to.ID).Save(ctx)
		}},
		{&res.Initiatives, "initiative home repositories", func() (int, error) {
			return c.Initiative.Update().Where(initiative.HomeRepo(from.ID)).SetHomeRepo(to.ID).Save(ctx)
		}},
		{&res.PRISMRoadmaps, "PRISM roadmaps", func() (int, error) {
			return c.PRISMRoadmap.Update().Where(prismroadmap.RepositoryID(from.ID)).SetRepositoryID(to.ID).Save(ctx)
		}},
		{&res.PRISMGoals, "PRISM goals", func() (int, error) {
			return c.PRISMGoal.Update().Where(prismgoal.RepositoryID(from.ID)).SetRepositoryID(to.ID).Save(ctx)
		}},
		{&res.PRISMDocuments, "PRISM documents", func() (int, error) {
			return c.PRISMDocument.Update().Where(prismdocument.RepositoryID(from.ID)).SetRepositoryID(to.ID).Save(ctx)
		}},
		{&res.DevXReports, "DevX reports", func() (int, error) {
			return c.DevXPeriodReport.Update().Where(devxperiodreport.RepositoryID(from.ID)).SetRepositoryID(to.ID).Save(ctx)
		}},
	}
	if opts.MoveReleases {
		moves = append(moves, struct {
			count *int
			name  string
			run   func() (int, error)
		}{&res.Releases, "releases", func() (int, error) {
			return c.Release.Update().Where(release.RepositoryID(from.ID)).SetRepositoryID(to.ID).Save(ctx)
		}})
	}
	for _, mv := range moves {
		n, err := mv.run()
		if err != nil {
			return nil, fmt.Errorf("move %s: %w", mv.name, err)
		}
		*mv.count = n
	}

	if err := moveDependencies(ctx, c, from.ID, to.ID, res); err != nil {
		return nil, err
	}

	if _, err := c.Repository.UpdateOneID(from.ID).SetStatus("archived").SetSupersededBy(to.ID).Save(ctx); err != nil {
		return nil, fmt.Errorf("archive %s: %w", from.ID, err)
	}
	return res, nil
}

// moveDependencies repoints every dependency edge of from to to. An edge that
// would then point a repository at itself, or repeat an edge already there, is
// deleted instead.
func moveDependencies(ctx context.Context, c *ent.Client, from, to string, res *store.RenameRepositoryResult) error {
	edges, err := c.RepositoryDependency.Query().
		Where(repositorydependency.Or(
			repositorydependency.SourceRepositoryID(from),
			repositorydependency.TargetRepositoryID(from),
		)).All(ctx)
	if err != nil {
		return fmt.Errorf("list dependencies: %w", err)
	}
	type triple struct{ src, dst, typ string }
	var placed []triple
	for _, e := range edges {
		src, dst := e.SourceRepositoryID, e.TargetRepositoryID
		if src == from {
			src = to
		}
		if dst == from {
			dst = to
		}
		t := triple{src, dst, e.DependencyType}
		dup := slices.Contains(placed, t)
		if !dup && src != dst {
			dup, err = c.RepositoryDependency.Query().Where(
				repositorydependency.SourceRepositoryID(src),
				repositorydependency.TargetRepositoryID(dst),
				repositorydependency.DependencyType(e.DependencyType),
				repositorydependency.IDNEQ(e.ID),
			).Exist(ctx)
			if err != nil {
				return fmt.Errorf("check dependency %s -> %s: %w", src, dst, err)
			}
		}
		if src == dst || dup {
			if err := c.RepositoryDependency.DeleteOneID(e.ID).Exec(ctx); err != nil {
				return fmt.Errorf("drop dependency %s -> %s: %w", src, dst, err)
			}
			res.DependenciesDropped++
			continue
		}
		if _, err := c.RepositoryDependency.UpdateOneID(e.ID).SetSourceRepositoryID(src).SetTargetRepositoryID(dst).Save(ctx); err != nil {
			return fmt.Errorf("move dependency %s -> %s: %w", src, dst, err)
		}
		placed = append(placed, t)
		res.Dependencies++
	}
	return nil
}
