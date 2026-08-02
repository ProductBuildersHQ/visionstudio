package roadmap

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// ImportAction describes what happened to one entity during import.
type ImportAction struct {
	Entity string // "phase", "rmi", "dependency"
	ID     string
	Action string // "created", "updated", "unchanged", "skipped"
	Detail string
}

// Importer creates or updates DB entities from a parsed Roadmap.
type Importer struct {
	Store       store.Store
	CreateRMI   func(ctx context.Context, id, repoID, initiativeID, phaseID, title, desc, itemType, priority string, required bool, seq int, acceptance []string) error
	CreatePhase func(ctx context.Context, id, initiativeID string, seq int, title, theme string) error
	CreateDep   func(ctx context.Context, sourceID, targetID, relationship string) error
	UpdateRMI   func(ctx context.Context, rmi *store.RoadmapItem) error
	UpdatePhase func(ctx context.Context, phase *store.Phase) error
}

// Import processes a parsed Roadmap, creating or updating entities.
// defaultRepo is the repository ID to use for new RMIs.
// defaultType is the item type for new RMIs (e.g. "capability").
func (imp *Importer) Import(ctx context.Context, parsed *Roadmap, defaultRepo, defaultType string, dryRun bool) ([]ImportAction, error) {
	if parsed.InitiativeID == "" {
		return nil, fmt.Errorf("roadmap has no initiative ID")
	}
	if defaultRepo == "" && parsed.Repository != "" {
		defaultRepo = parsed.Repository
	}
	if defaultRepo == "" {
		return nil, fmt.Errorf("no repository specified (use --repo or set **Repository:** in the file)")
	}
	if defaultType == "" {
		defaultType = "capability"
	}

	initID := parsed.InitiativeID
	var actions []ImportAction

	existingPhases, err := imp.Store.ListPhases(ctx, initID)
	if err != nil {
		return nil, fmt.Errorf("list phases: %w", err)
	}
	phasesBySeq := map[int]*store.Phase{}
	for _, p := range existingPhases {
		phasesBySeq[p.SequenceNumber] = p
	}

	existingRMIs, err := imp.Store.ListRMIs(ctx, initID)
	if err != nil {
		return nil, fmt.Errorf("list RMIs: %w", err)
	}
	rmisByID := map[string]*store.RoadmapItem{}
	for _, r := range existingRMIs {
		rmisByID[r.ID] = r
	}

	for _, fp := range parsed.Phases {
		phaseID := fmt.Sprintf("%s/phase-%d", initID, fp.Number)

		if existing, ok := phasesBySeq[fp.Number]; ok {
			changed := false
			if existing.Title != fp.Title {
				existing.Title = fp.Title
				changed = true
			}
			if existing.Theme != fp.Theme {
				existing.Theme = fp.Theme
				changed = true
			}
			if changed && !dryRun {
				if imp.UpdatePhase != nil {
					if err := imp.UpdatePhase(ctx, existing); err != nil {
						return actions, fmt.Errorf("update phase %s: %w", phaseID, err)
					}
				}
			}
			action := "unchanged"
			if changed {
				action = "updated"
			}
			actions = append(actions, ImportAction{Entity: "phase", ID: phaseID, Action: action})
		} else {
			if !dryRun {
				if err := imp.CreatePhase(ctx, phaseID, initID, fp.Number, fp.Title, fp.Theme); err != nil {
					return actions, fmt.Errorf("create phase %s: %w", phaseID, err)
				}
			}
			actions = append(actions, ImportAction{Entity: "phase", ID: phaseID, Action: "created"})
		}

		for i, item := range fp.Items {
			seq := i + 1
			status := statusFromCheckbox(item.Completed)

			if existing, ok := rmisByID[item.ID]; ok {
				var fields []string
				if existing.Title != item.Title {
					fields = append(fields, "title")
					existing.Title = item.Title
				}
				if existing.PhaseID != phaseID {
					fields = append(fields, "phase")
					existing.PhaseID = phaseID
				}
				if existing.SequenceNumber != seq {
					fields = append(fields, "sequence")
					existing.SequenceNumber = seq
				}
				if item.Completed && existing.Status != "completed" {
					fields = append(fields, "status→completed")
					existing.Status = "completed"
				}
				if len(fields) > 0 && !dryRun {
					if err := imp.UpdateRMI(ctx, existing); err != nil {
						return actions, fmt.Errorf("update RMI %s: %w", item.ID, err)
					}
				}
				action := "unchanged"
				detail := ""
				if len(fields) > 0 {
					action = "updated"
					detail = strings.Join(fields, ", ")
				}
				actions = append(actions, ImportAction{Entity: "rmi", ID: item.ID, Action: action, Detail: detail})
			} else {
				if !dryRun {
					if err := imp.CreateRMI(ctx, item.ID, defaultRepo, initID, phaseID, item.Title, item.Description, defaultType, "", true, seq, nil); err != nil {
						return actions, fmt.Errorf("create RMI %s: %w", item.ID, err)
					}
					if status == "completed" {
						r, err := imp.Store.GetRMI(ctx, item.ID)
						if err != nil {
							return actions, fmt.Errorf("get created RMI %s: %w", item.ID, err)
						}
						r.Status = "completed"
						if err := imp.UpdateRMI(ctx, r); err != nil {
							return actions, fmt.Errorf("mark RMI %s completed: %w", item.ID, err)
						}
					}
				}
				actions = append(actions, ImportAction{Entity: "rmi", ID: item.ID, Action: "created"})
			}

			if !dryRun {
				for _, depID := range item.DependsOn {
					if err := imp.CreateDep(ctx, item.ID, depID, "requires"); err != nil {
						return actions, fmt.Errorf("create dependency %s→%s: %w", item.ID, depID, err)
					}
				}
			}
			for _, depID := range item.DependsOn {
				actions = append(actions, ImportAction{Entity: "dependency", ID: fmt.Sprintf("%s→%s", item.ID, depID), Action: "created"})
			}
		}
	}

	return actions, nil
}

func statusFromCheckbox(completed bool) string {
	if completed {
		return "completed"
	}
	return "proposed"
}
