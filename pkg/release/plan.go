// Package release implements release sets and dependency-ordered
// release planning. A release set groups component releases across
// repositories in an initiative; topological sorting ensures
// libraries are released before their consumers.
package release

import (
	"fmt"
	"sort"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// ReleaseSet groups component releases for one initiative.
type ReleaseSet struct {
	InitiativeID string
	Components   []ComponentRelease
}

// ComponentRelease is one repo's release within a set.
type ComponentRelease struct {
	RepositoryID string
	Version      string
	Stage        int // topological stage (0 = no deps, 1 = depends on stage 0, etc.)
	RMIs         []string
}

// Stage is a group of repos that can be released in parallel.
type Stage struct {
	Number int
	Repos  []ComponentRelease
}

// Plan computes a dependency-ordered release plan for an initiative.
// It uses repository dependencies to determine release order:
// - Stage 0: repos with no in-initiative dependencies (libraries)
// - Stage 1: repos that depend only on stage-0 repos
// - Stage N: repos that depend on stage N-1 or earlier
func Plan(initiative string, rmis []*store.RoadmapItem, repoDeps []*store.RepositoryDependency) (*ReleaseSet, error) {
	repoRMIs := map[string][]string{}
	for _, r := range rmis {
		if r.Status == "completed" || r.Status == "cancelled" {
			repoRMIs[r.RepositoryID] = append(repoRMIs[r.RepositoryID], r.ID)
		}
	}

	if len(repoRMIs) == 0 {
		return &ReleaseSet{InitiativeID: initiative}, nil
	}

	inInitiative := map[string]bool{}
	for repo := range repoRMIs {
		inInitiative[repo] = true
	}

	deps := map[string][]string{}
	for _, d := range repoDeps {
		if inInitiative[d.SourceRepositoryID] && inInitiative[d.TargetRepositoryID] {
			deps[d.SourceRepositoryID] = append(deps[d.SourceRepositoryID], d.TargetRepositoryID)
		}
	}

	stages, err := topoStages(repoRMIs, deps)
	if err != nil {
		return nil, err
	}

	var components []ComponentRelease
	for _, repo := range sortedKeys(repoRMIs) {
		stage := stages[repo]
		components = append(components, ComponentRelease{
			RepositoryID: repo,
			Stage:        stage,
			RMIs:         repoRMIs[repo],
		})
	}

	sort.Slice(components, func(i, j int) bool {
		if components[i].Stage != components[j].Stage {
			return components[i].Stage < components[j].Stage
		}
		return components[i].RepositoryID < components[j].RepositoryID
	})

	return &ReleaseSet{
		InitiativeID: initiative,
		Components:   components,
	}, nil
}

// Stages groups components into parallel-releasable stages.
func (rs *ReleaseSet) Stages() []Stage {
	byStage := map[int][]ComponentRelease{}
	for _, c := range rs.Components {
		byStage[c.Stage] = append(byStage[c.Stage], c)
	}
	var stages []Stage
	for _, n := range sortedIntKeys(byStage) {
		stages = append(stages, Stage{Number: n, Repos: byStage[n]})
	}
	return stages
}

func topoStages(repos map[string][]string, deps map[string][]string) (map[string]int, error) {
	stages := map[string]int{}

	for repo := range repos {
		if _, err := visit(repo, deps, stages, map[string]bool{}); err != nil {
			return nil, err
		}
	}
	return stages, nil
}

func visit(repo string, deps map[string][]string, stages map[string]int, visiting map[string]bool) (int, error) {
	if s, ok := stages[repo]; ok {
		return s, nil
	}
	if visiting[repo] {
		return 0, fmt.Errorf("dependency cycle involving %s", repo)
	}
	visiting[repo] = true

	maxDep := -1
	for _, dep := range deps[repo] {
		s, err := visit(dep, deps, stages, visiting)
		if err != nil {
			return 0, err
		}
		if s > maxDep {
			maxDep = s
		}
	}

	stage := maxDep + 1
	stages[repo] = stage
	delete(visiting, repo)
	return stage, nil
}

func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedIntKeys(m map[int][]ComponentRelease) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
