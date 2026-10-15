package roadmap

import (
	"context"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

const phasesRoadmap = `# Roadmap

**Initiative:** ` + "`INIT-TEST-001`" + `
**Repository:** ` + "`github.com/example/project`" + `

## Phase 1 — First product

**Theme:** Shipped by the host.
`

func phaseImporter(t *testing.T, update func(context.Context, *store.Phase) error) *Importer {
	t.Helper()
	s := store.NewMemStore()
	if err := s.CreatePhase(context.Background(), &store.Phase{
		ID: "INIT-TEST-001/phase-1", InitiativeID: "INIT-TEST-001", SequenceNumber: 1, Title: "Adoption", Theme: "Old theme",
	}); err != nil {
		t.Fatal(err)
	}
	return &Importer{Store: s, UpdatePhase: update}
}

func parsePhases(t *testing.T) *Roadmap {
	t.Helper()
	parsed, err := Parse(strings.NewReader(phasesRoadmap))
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func TestImportSavesAChangedPhase(t *testing.T) {
	var saved *store.Phase
	imp := phaseImporter(t, func(_ context.Context, p *store.Phase) error {
		c := *p
		saved = &c
		return nil
	})
	actions, err := imp.Import(context.Background(), parsePhases(t), "", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if saved == nil || saved.Title != "First product" || saved.Theme != "Shipped by the host." {
		t.Fatalf("saved = %+v, want the new title and theme", saved)
	}
	if len(actions) != 1 || actions[0].Action != "updated" {
		t.Fatalf("actions = %+v", actions)
	}
}

func TestImportDoesNotSaveInADryRun(t *testing.T) {
	called := false
	imp := phaseImporter(t, func(context.Context, *store.Phase) error { called = true; return nil })
	if _, err := imp.Import(context.Background(), parsePhases(t), "", "", true); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("a dry run must not save")
	}
}

func TestImportRefusesAPhaseChangeItCannotSave(t *testing.T) {
	imp := phaseImporter(t, nil)
	_, err := imp.Import(context.Background(), parsePhases(t), "", "", false)
	if err == nil || !strings.Contains(err.Error(), "cannot update phases") {
		t.Fatalf("err = %v, want a refusal rather than a silent skip", err)
	}
}
