

package doltstore

import (
	"context"
	"fmt"

	"github.com/ProductBuildersHQ/visionstudio/ent"
	"github.com/ProductBuildersHQ/visionstudio/ent/prismgoal"
	"github.com/ProductBuildersHQ/visionstudio/ent/prismroadmap"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func entPRISMRoadmapToStore(e *ent.PRISMRoadmap) *store.PRISMRoadmap {
	return &store.PRISMRoadmap{
		ID:           e.ID,
		Organization: e.Organization,
		RepositoryID: e.RepositoryID,
		Name:         e.Name,
		Phases:       e.Phases,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func entPRISMGoalToStore(e *ent.PRISMGoal) *store.PRISMGoal {
	return &store.PRISMGoal{
		ID:           e.ID,
		Organization: e.Organization,
		RepositoryID: e.RepositoryID,
		GoalType:     e.GoalType,
		Document:     e.Document,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func (d *DoltStore) CreatePRISMRoadmap(ctx context.Context, roadmap *store.PRISMRoadmap) error {
	b := d.client.PRISMRoadmap.Create().
		SetID(roadmap.ID).
		SetRepositoryID(roadmap.RepositoryID).
		SetCreatedAt(roadmap.CreatedAt).
		SetUpdatedAt(roadmap.UpdatedAt)
	if roadmap.Organization != "" {
		b.SetOrganization(roadmap.Organization)
	}
	if roadmap.Name != "" {
		b.SetName(roadmap.Name)
	}
	if roadmap.Phases != nil {
		b.SetPhases(roadmap.Phases)
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("create prism roadmap: %w", err)
	}
	return nil
}

func (d *DoltStore) GetPRISMRoadmap(ctx context.Context, id string) (*store.PRISMRoadmap, error) {
	e, err := d.client.PRISMRoadmap.Query().
		Where(prismroadmap.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get prism roadmap %s: %w", id, err)
	}
	return entPRISMRoadmapToStore(e), nil
}

func (d *DoltStore) ListPRISMRoadmaps(ctx context.Context) ([]*store.PRISMRoadmap, error) {
	rows, err := d.client.PRISMRoadmap.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list prism roadmaps: %w", err)
	}
	result := make([]*store.PRISMRoadmap, len(rows))
	for i, e := range rows {
		result[i] = entPRISMRoadmapToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListPRISMRoadmapsByRepo(ctx context.Context, repoID string) ([]*store.PRISMRoadmap, error) {
	rows, err := d.client.PRISMRoadmap.Query().
		Where(prismroadmap.RepositoryIDEQ(repoID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list prism roadmaps by repo: %w", err)
	}
	result := make([]*store.PRISMRoadmap, len(rows))
	for i, e := range rows {
		result[i] = entPRISMRoadmapToStore(e)
	}
	return result, nil
}

func (d *DoltStore) UpdatePRISMRoadmap(ctx context.Context, roadmap *store.PRISMRoadmap) error {
	b := d.client.PRISMRoadmap.UpdateOneID(roadmap.ID).
		SetRepositoryID(roadmap.RepositoryID).
		SetUpdatedAt(roadmap.UpdatedAt)
	if roadmap.Organization != "" {
		b.SetOrganization(roadmap.Organization)
	} else {
		b.ClearOrganization()
	}
	if roadmap.Name != "" {
		b.SetName(roadmap.Name)
	} else {
		b.ClearName()
	}
	if roadmap.Phases != nil {
		b.SetPhases(roadmap.Phases)
	} else {
		b.ClearPhases()
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("update prism roadmap %s: %w", roadmap.ID, err)
	}
	return nil
}

func (d *DoltStore) CreatePRISMGoal(ctx context.Context, goal *store.PRISMGoal) error {
	b := d.client.PRISMGoal.Create().
		SetID(goal.ID).
		SetRepositoryID(goal.RepositoryID).
		SetGoalType(goal.GoalType).
		SetCreatedAt(goal.CreatedAt).
		SetUpdatedAt(goal.UpdatedAt)
	if goal.Organization != "" {
		b.SetOrganization(goal.Organization)
	}
	if goal.Document != nil {
		b.SetDocument(goal.Document)
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("create prism goal: %w", err)
	}
	return nil
}

func (d *DoltStore) GetPRISMGoal(ctx context.Context, id string) (*store.PRISMGoal, error) {
	e, err := d.client.PRISMGoal.Query().
		Where(prismgoal.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get prism goal %s: %w", id, err)
	}
	return entPRISMGoalToStore(e), nil
}

func (d *DoltStore) ListPRISMGoals(ctx context.Context, repoID string) ([]*store.PRISMGoal, error) {
	rows, err := d.client.PRISMGoal.Query().
		Where(prismgoal.RepositoryIDEQ(repoID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list prism goals: %w", err)
	}
	result := make([]*store.PRISMGoal, len(rows))
	for i, e := range rows {
		result[i] = entPRISMGoalToStore(e)
	}
	return result, nil
}

func (d *DoltStore) UpdatePRISMGoal(ctx context.Context, goal *store.PRISMGoal) error {
	b := d.client.PRISMGoal.UpdateOneID(goal.ID).
		SetRepositoryID(goal.RepositoryID).
		SetGoalType(goal.GoalType).
		SetUpdatedAt(goal.UpdatedAt)
	if goal.Organization != "" {
		b.SetOrganization(goal.Organization)
	} else {
		b.ClearOrganization()
	}
	if goal.Document != nil {
		b.SetDocument(goal.Document)
	} else {
		b.ClearDocument()
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("update prism goal %s: %w", goal.ID, err)
	}
	return nil
}
