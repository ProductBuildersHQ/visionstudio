package contextbuild

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestBuildForPhase_ByteIdentical(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")
	assertByteIdentical(t, func() (*ContextPackage, error) {
		return builder.BuildForPhase(ctx, "INIT-TEST-001/phase-1")
	})
}

func TestBuildForRMI_ByteIdentical(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")
	assertByteIdentical(t, func() (*ContextPackage, error) {
		return builder.BuildForRMI(ctx, "RMI-TESTREPO-001")
	})
}

func assertByteIdentical(t *testing.T, buildFn func() (*ContextPackage, error)) {
	t.Helper()

	pkg1, err := buildFn()
	if err != nil {
		t.Fatalf("first build: %v", err)
	}

	pkg2, err := buildFn()
	if err != nil {
		t.Fatalf("second build: %v", err)
	}

	pkg1.BuildTimestamp = time.Time{}
	pkg2.BuildTimestamp = time.Time{}

	json1, err := json.MarshalIndent(pkg1, "", "  ")
	if err != nil {
		t.Fatalf("marshal pkg1: %v", err)
	}

	json2, err := json.MarshalIndent(pkg2, "", "  ")
	if err != nil {
		t.Fatalf("marshal pkg2: %v", err)
	}

	if string(json1) != string(json2) {
		t.Errorf("outputs not byte-identical:\n--- pkg1 ---\n%s\n--- pkg2 ---\n%s", json1, json2)
	}
}

func TestBuildForPhase_Sections(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	pkg, err := builder.BuildForPhase(ctx, "INIT-TEST-001/phase-1")
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if pkg.Version != SchemaVersion {
		t.Errorf("version = %q, want %q", pkg.Version, SchemaVersion)
	}

	if pkg.TargetType != "phase" {
		t.Errorf("target_type = %q, want %q", pkg.TargetType, "phase")
	}

	if pkg.TargetID != "INIT-TEST-001/phase-1" {
		t.Errorf("target_id = %q, want %q", pkg.TargetID, "INIT-TEST-001/phase-1")
	}

	if pkg.Sections.Program == nil {
		t.Error("program section is nil")
	} else {
		if pkg.Sections.Program.Stability != Stable {
			t.Errorf("program stability = %q, want %q", pkg.Sections.Program.Stability, Stable)
		}
		if pkg.Sections.Program.ID != "PROG-TEST" {
			t.Errorf("program ID = %q, want %q", pkg.Sections.Program.ID, "PROG-TEST")
		}
	}

	if pkg.Sections.Initiative.Stability != Stable {
		t.Errorf("initiative stability = %q, want %q", pkg.Sections.Initiative.Stability, Stable)
	}
	if pkg.Sections.Initiative.ID != "INIT-TEST-001" {
		t.Errorf("initiative ID = %q, want %q", pkg.Sections.Initiative.ID, "INIT-TEST-001")
	}

	if pkg.Sections.Phase.Stability != PhaseStable {
		t.Errorf("phase stability = %q, want %q", pkg.Sections.Phase.Stability, PhaseStable)
	}
	if pkg.Sections.Phase.ID != "INIT-TEST-001/phase-1" {
		t.Errorf("phase ID = %q, want %q", pkg.Sections.Phase.ID, "INIT-TEST-001/phase-1")
	}

	if len(pkg.Sections.Phase.MemberRMIs) != 2 {
		t.Errorf("member_rmis count = %d, want 2", len(pkg.Sections.Phase.MemberRMIs))
	}
}

func TestBuildForRMI_Sections(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	pkg, err := builder.BuildForRMI(ctx, "RMI-TESTREPO-001")
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if pkg.TargetType != "rmi" {
		t.Errorf("target_type = %q, want %q", pkg.TargetType, "rmi")
	}

	if pkg.Sections.CurrentRMI == nil {
		t.Fatal("current_rmi section is nil")
	}

	if pkg.Sections.CurrentRMI.Stability != RMIStable {
		t.Errorf("current_rmi stability = %q, want %q", pkg.Sections.CurrentRMI.Stability, RMIStable)
	}

	if pkg.Sections.CurrentRMI.ID != "RMI-TESTREPO-001" {
		t.Errorf("current_rmi ID = %q, want %q", pkg.Sections.CurrentRMI.ID, "RMI-TESTREPO-001")
	}

	if len(pkg.Sections.CurrentRMI.AcceptanceCriteria) != 2 {
		t.Errorf("acceptance_criteria count = %d, want 2", len(pkg.Sections.CurrentRMI.AcceptanceCriteria))
	}
}

func TestBuildForPhase_DerivedRepos(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	pkg, err := builder.BuildForPhase(ctx, "INIT-TEST-001/phase-1")
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if len(pkg.DerivedRepos) == 0 {
		t.Error("derived_repos is empty")
	}

	var primaryFound bool
	for _, dr := range pkg.DerivedRepos {
		if dr.Role == "primary" {
			primaryFound = true
			if dr.ID != "github.com/test/repo" {
				t.Errorf("primary repo ID = %q, want %q", dr.ID, "github.com/test/repo")
			}
		}
	}

	if !primaryFound {
		t.Error("no primary repo found in derived_repos")
	}

	for i := 1; i < len(pkg.DerivedRepos); i++ {
		prev := pkg.DerivedRepos[i-1]
		curr := pkg.DerivedRepos[i]
		roleOrder := map[string]int{"primary": 0, "dependency_rmi": 1, "repo_dependency": 2}
		if roleOrder[prev.Role] > roleOrder[curr.Role] {
			t.Errorf("derived_repos not sorted by role: %s before %s", prev.Role, curr.Role)
		}
	}
}

func TestBuildForPhase_SectionOrdering(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	pkg, err := builder.BuildForPhase(ctx, "INIT-TEST-001/phase-1")
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	sections := []struct {
		name      string
		stability StabilityClass
	}{
		{"program", pkg.Sections.Program.Stability},
		{"initiative", pkg.Sections.Initiative.Stability},
		{"phase", pkg.Sections.Phase.Stability},
	}

	stabilityOrder := map[StabilityClass]int{
		Stable:      0,
		PhaseStable: 1,
		RMIStable:   2,
		Volatile:    3,
	}

	for i := 1; i < len(sections); i++ {
		prev := sections[i-1]
		curr := sections[i]
		if stabilityOrder[prev.stability] > stabilityOrder[curr.stability] {
			t.Errorf("sections not ordered stable→volatile: %s (%s) before %s (%s)",
				prev.name, prev.stability, curr.name, curr.stability)
		}
	}
}

func TestBuildForRMI_Dependencies(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupTestData(t, ctx, ms)

	builder := NewBuilder(ms, "test-dolt-commit-abc123")

	pkg, err := builder.BuildForRMI(ctx, "RMI-TESTREPO-002")
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if pkg.Sections.CurrentRMI == nil {
		t.Fatal("current_rmi section is nil")
	}

	if len(pkg.Sections.CurrentRMI.Dependencies) != 1 {
		t.Errorf("dependencies count = %d, want 1", len(pkg.Sections.CurrentRMI.Dependencies))
	}

	if len(pkg.Sections.CurrentRMI.Dependencies) > 0 {
		dep := pkg.Sections.CurrentRMI.Dependencies[0]
		if dep.TargetID != "RMI-TESTREPO-001" {
			t.Errorf("dependency target = %q, want %q", dep.TargetID, "RMI-TESTREPO-001")
		}
	}
}

func setupTestData(t *testing.T, ctx context.Context, ms *store.MemStore) {
	t.Helper()

	if err := ms.CreateProgram(ctx, &store.Program{
		ID:          "PROG-TEST",
		Name:        "Test Program",
		Description: "A test program for unit tests",
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateRepository(ctx, &store.Repository{
		ID:             "github.com/test/repo",
		Organization:   "test",
		RepositoryName: "repo",
		DefaultBranch:  "main",
		LocalPath:      "/tmp/test/repo",
		Status:         "active",
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateInitiative(ctx, &store.Initiative{
		ID:          "INIT-TEST-001",
		Title:       "Test Initiative",
		Description: "A test initiative for unit tests",
		Status:      "executing",
		Priority:    "high",
		ProgramID:   "PROG-TEST",
		HomeRepo:    "github.com/test/repo",
		Specs: map[string]string{
			"prd": "docs/specs/PRD.md",
			"trd": "docs/specs/TRD.md",
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreatePhase(ctx, &store.Phase{
		ID:             "INIT-TEST-001/phase-1",
		InitiativeID:   "INIT-TEST-001",
		SequenceNumber: 1,
		Title:          "Phase 1",
		Theme:          "Foundation",
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateRMI(ctx, &store.RoadmapItem{
		ID:                 "RMI-TESTREPO-001",
		RepositoryID:       "github.com/test/repo",
		InitiativeID:       "INIT-TEST-001",
		PhaseID:            "INIT-TEST-001/phase-1",
		Title:              "First RMI",
		Description:        "The first roadmap item",
		ItemType:           "capability",
		Status:             "proposed",
		Required:           true,
		SequenceNumber:     1,
		AcceptanceCriteria: []string{"Criterion 1", "Criterion 2"},
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateRMI(ctx, &store.RoadmapItem{
		ID:             "RMI-TESTREPO-002",
		RepositoryID:   "github.com/test/repo",
		InitiativeID:   "INIT-TEST-001",
		PhaseID:        "INIT-TEST-001/phase-1",
		Title:          "Second RMI",
		Description:    "The second roadmap item",
		ItemType:       "capability",
		Status:         "proposed",
		Required:       true,
		SequenceNumber: 2,
	}); err != nil {
		t.Fatal(err)
	}

	if err := ms.CreateDependency(ctx, &store.RMIDependency{
		SourceRMIID:  "RMI-TESTREPO-002",
		TargetRMIID:  "RMI-TESTREPO-001",
		Relationship: "requires",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestBuildForRMI_ContextSpec(t *testing.T) {
	ms := store.NewMemStore()
	ctx := context.Background()
	setupTestData(t, ctx, ms)

	// Create an additional repository for extra_repos testing
	if err := ms.CreateRepository(ctx, &store.Repository{
		ID:             "github.com/test/extra",
		Organization:   "test",
		RepositoryName: "extra",
		DefaultBranch:  "main",
		LocalPath:      "/tmp/test/extra",
		Status:         "active",
	}); err != nil {
		t.Fatal(err)
	}

	// Create an RMI with a ContextSpec
	if err := ms.CreateRMI(ctx, &store.RoadmapItem{
		ID:             "RMI-TESTREPO-003",
		RepositoryID:   "github.com/test/repo",
		InitiativeID:   "INIT-TEST-001",
		PhaseID:        "INIT-TEST-001/phase-1",
		Title:          "RMI with ContextSpec",
		ItemType:       "capability",
		Status:         "proposed",
		Required:       true,
		SequenceNumber: 3,
		ContextSpec: &store.ContextSpec{
			ExtraRepos:   []string{"github.com/test/extra"},
			IncludeSpecs: []string{"docs/additional.md"},
			ExcludeSpecs: []string{"docs/specs/TRD.md"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	builder := NewBuilder(ms, "test-dolt-commit-abc123")
	pkg, err := builder.BuildForRMI(ctx, "RMI-TESTREPO-003")
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	// Check that extra repo was added with role "explicit"
	var foundExplicit bool
	for _, repo := range pkg.DerivedRepos {
		if repo.ID == "github.com/test/extra" {
			foundExplicit = true
			if repo.Role != "explicit" {
				t.Errorf("extra repo role = %q, want %q", repo.Role, "explicit")
			}
			break
		}
	}
	if !foundExplicit {
		t.Error("extra repo github.com/test/extra not found in DerivedRepos")
	}

	// Check ordering: primary < explicit < dependency_rmi < repo_dependency
	primaryIdx := -1
	explicitIdx := -1
	for i, repo := range pkg.DerivedRepos {
		if repo.Role == "primary" && primaryIdx == -1 {
			primaryIdx = i
		}
		if repo.Role == "explicit" && explicitIdx == -1 {
			explicitIdx = i
		}
	}
	if primaryIdx >= 0 && explicitIdx >= 0 && primaryIdx > explicitIdx {
		t.Errorf("primary repos should come before explicit repos: primary at %d, explicit at %d", primaryIdx, explicitIdx)
	}

	// Check that included spec appears in spec references
	var foundIncluded bool
	for _, ref := range pkg.Sections.SpecReferences {
		if ref.Path == "docs/additional.md" {
			foundIncluded = true
			break
		}
	}
	if !foundIncluded {
		t.Error("included spec docs/additional.md not found in SpecReferences")
	}

	// Check that excluded spec does NOT appear
	for _, ref := range pkg.Sections.SpecReferences {
		if ref.Path == "docs/specs/TRD.md" {
			t.Error("excluded spec docs/specs/TRD.md should not appear in SpecReferences")
		}
	}
}
