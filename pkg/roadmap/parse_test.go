package roadmap

import (
	"strings"
	"testing"
)

const testRoadmap = `# My Project — Roadmap

**Initiative:** ` + "`INIT-TEST-001`" + `
**Repository:** ` + "`github.com/example/project`" + `
**Status:** Phase 1 completed

## Phase 1 — Foundation

**Theme:** Build the base.
**Status:** Completed — 2 of 2 items completed

- [x] ` + "`RMI-TEST-001`" + ` First item
  - Depends on: ` + "`RMI-TEST-000`" + `
  - Acceptance: it works
- [x] ` + "`RMI-TEST-002`" + ` Second item
  - Depends on: ` + "`RMI-TEST-001`" + `

## Phase 2 — Features

**Theme:** Add things.
**Status:** In progress — 1 of 2 items completed

- [x] ` + "`RMI-TEST-003`" + ` Done feature
- [ ] ` + "`RMI-TEST-004`" + ` Pending feature
  - Depends on: ` + "`RMI-TEST-002`" + `, ` + "`RMI-TEST-003`" + `
`

func TestParse(t *testing.T) {
	rm, err := Parse(strings.NewReader(testRoadmap))
	if err != nil {
		t.Fatal(err)
	}

	if rm.InitiativeID != "INIT-TEST-001" {
		t.Errorf("initiative = %q, want INIT-TEST-001", rm.InitiativeID)
	}
	if rm.Repository != "github.com/example/project" {
		t.Errorf("repository = %q, want github.com/example/project", rm.Repository)
	}
	if rm.Title != "My Project — Roadmap" {
		t.Errorf("title = %q, want 'My Project — Roadmap'", rm.Title)
	}
	if len(rm.Phases) != 2 {
		t.Fatalf("phases = %d, want 2", len(rm.Phases))
	}

	p1 := rm.Phases[0]
	if p1.Number != 1 || p1.Title != "Foundation" {
		t.Errorf("phase 1: num=%d title=%q", p1.Number, p1.Title)
	}
	if p1.Theme != "Build the base." {
		t.Errorf("phase 1 theme = %q", p1.Theme)
	}
	if len(p1.Items) != 2 {
		t.Fatalf("phase 1 items = %d, want 2", len(p1.Items))
	}
	if p1.Items[0].ID != "RMI-TEST-001" || !p1.Items[0].Completed {
		t.Errorf("item 0: id=%q completed=%v", p1.Items[0].ID, p1.Items[0].Completed)
	}
	if len(p1.Items[0].DependsOn) != 1 || p1.Items[0].DependsOn[0] != "RMI-TEST-000" {
		t.Errorf("item 0 deps = %v", p1.Items[0].DependsOn)
	}

	p2 := rm.Phases[1]
	if len(p2.Items) != 2 {
		t.Fatalf("phase 2 items = %d, want 2", len(p2.Items))
	}
	if !p2.Items[0].Completed {
		t.Error("phase 2 item 0 should be completed")
	}
	if p2.Items[1].Completed {
		t.Error("phase 2 item 1 should not be completed")
	}
	if len(p2.Items[1].DependsOn) != 2 {
		t.Errorf("phase 2 item 1 deps = %v, want 2", p2.Items[1].DependsOn)
	}
}

func TestParseEmpty(t *testing.T) {
	rm, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(rm.Phases) != 0 {
		t.Errorf("phases = %d, want 0", len(rm.Phases))
	}
}
