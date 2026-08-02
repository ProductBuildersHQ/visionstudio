package release

import (
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestPlanNoDeps(t *testing.T) {
	rmis := []*store.RoadmapItem{
		{ID: "RMI-A-001", RepositoryID: "repo-a", Status: "completed"},
		{ID: "RMI-B-001", RepositoryID: "repo-b", Status: "completed"},
	}
	rs, err := Plan("INIT-1", rmis, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rs.Components) != 2 {
		t.Fatalf("components = %d, want 2", len(rs.Components))
	}
	for _, c := range rs.Components {
		if c.Stage != 0 {
			t.Errorf("%s stage = %d, want 0", c.RepositoryID, c.Stage)
		}
	}
	stages := rs.Stages()
	if len(stages) != 1 {
		t.Errorf("stages = %d, want 1", len(stages))
	}
}

func TestPlanWithDeps(t *testing.T) {
	rmis := []*store.RoadmapItem{
		{ID: "RMI-LIB-001", RepositoryID: "lib", Status: "completed"},
		{ID: "RMI-APP-001", RepositoryID: "app", Status: "completed"},
		{ID: "RMI-WEB-001", RepositoryID: "web", Status: "completed"},
	}
	deps := []*store.RepositoryDependency{
		{SourceRepositoryID: "app", TargetRepositoryID: "lib", DependencyType: "go_module"},
		{SourceRepositoryID: "web", TargetRepositoryID: "app", DependencyType: "go_module"},
	}

	rs, err := Plan("INIT-1", rmis, deps)
	if err != nil {
		t.Fatal(err)
	}

	stageOf := map[string]int{}
	for _, c := range rs.Components {
		stageOf[c.RepositoryID] = c.Stage
	}

	if stageOf["lib"] != 0 {
		t.Errorf("lib stage = %d, want 0", stageOf["lib"])
	}
	if stageOf["app"] != 1 {
		t.Errorf("app stage = %d, want 1", stageOf["app"])
	}
	if stageOf["web"] != 2 {
		t.Errorf("web stage = %d, want 2", stageOf["web"])
	}

	stages := rs.Stages()
	if len(stages) != 3 {
		t.Fatalf("stages = %d, want 3", len(stages))
	}
	if stages[0].Number != 0 || stages[0].Repos[0].RepositoryID != "lib" {
		t.Errorf("stage 0: %+v", stages[0])
	}
}

func TestPlanCycleDetection(t *testing.T) {
	rmis := []*store.RoadmapItem{
		{ID: "RMI-A-001", RepositoryID: "a", Status: "completed"},
		{ID: "RMI-B-001", RepositoryID: "b", Status: "completed"},
	}
	deps := []*store.RepositoryDependency{
		{SourceRepositoryID: "a", TargetRepositoryID: "b", DependencyType: "go_module"},
		{SourceRepositoryID: "b", TargetRepositoryID: "a", DependencyType: "go_module"},
	}

	_, err := Plan("INIT-1", rmis, deps)
	if err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestPlanSkipsNonCompleted(t *testing.T) {
	rmis := []*store.RoadmapItem{
		{ID: "RMI-A-001", RepositoryID: "repo-a", Status: "completed"},
		{ID: "RMI-B-001", RepositoryID: "repo-b", Status: "in_progress"},
	}
	rs, err := Plan("INIT-1", rmis, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rs.Components) != 1 {
		t.Fatalf("components = %d, want 1 (only completed)", len(rs.Components))
	}
}

func TestPlanExternalDepsIgnored(t *testing.T) {
	rmis := []*store.RoadmapItem{
		{ID: "RMI-A-001", RepositoryID: "repo-a", Status: "completed"},
	}
	deps := []*store.RepositoryDependency{
		{SourceRepositoryID: "repo-a", TargetRepositoryID: "external-lib", DependencyType: "go_module"},
	}
	rs, err := Plan("INIT-1", rmis, deps)
	if err != nil {
		t.Fatal(err)
	}
	if rs.Components[0].Stage != 0 {
		t.Errorf("stage = %d, want 0 (external dep ignored)", rs.Components[0].Stage)
	}
}
