package roadmap

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// GenerateInput holds the data needed to generate a ROADMAP.md.
type GenerateInput struct {
	Initiative *store.Initiative
	Phases     []GeneratePhase
	Deps       []*store.RMIDependency
}

// GeneratePhase is a phase with its RMIs for generation.
type GeneratePhase struct {
	Phase *store.Phase
	RMIs  []*store.RoadmapItem
}

// Generate writes a ROADMAP.md to w from the given input.
func Generate(w io.Writer, input *GenerateInput) error {
	init := input.Initiative
	title := init.Title
	if title == "" {
		title = init.ID
	}

	fmt.Fprintf(w, "# %s — Roadmap\n\n", title)
	fmt.Fprintf(w, "**Initiative:** `%s`\n", init.ID)
	if init.HomeRepo != "" {
		fmt.Fprintf(w, "**Repository:** `%s`\n", init.HomeRepo)
	}

	statusSummary := generateStatusSummary(input.Phases)
	fmt.Fprintf(w, "**Status:** %s\n", statusSummary)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "> RMI IDs are stable and permanent. Commits implementing an item carry the trailer `Refs: RMI-<REPOSLUG>-<NNN>`. Phase status is derived from member RMIs — a phase is complete only when all its required RMIs are complete.")

	depsBySource := map[string][]string{}
	for _, d := range input.Deps {
		if d.Relationship == "requires" {
			depsBySource[d.SourceRMIID] = append(depsBySource[d.SourceRMIID], d.TargetRMIID)
		}
	}

	for _, gp := range input.Phases {
		p := gp.Phase
		rmis := gp.RMIs
		sort.Slice(rmis, func(i, j int) bool {
			return rmis[i].SequenceNumber < rmis[j].SequenceNumber
		})

		fmt.Fprintln(w)
		fmt.Fprintf(w, "## Phase %d — %s\n\n", p.SequenceNumber, p.Title)
		if p.Theme != "" {
			fmt.Fprintf(w, "**Theme:** %s\n", p.Theme)
		}

		completed := 0
		for _, r := range rmis {
			if r.Status == "completed" || r.Status == "cancelled" {
				completed++
			}
		}
		phaseStatus := derivePhaseLabel(rmis)
		fmt.Fprintf(w, "**Status:** %s — %d of %d items completed\n\n", phaseStatus, completed, len(rmis))

		for _, r := range rmis {
			check := " "
			if r.Status == "completed" {
				check = "x"
			}
			fmt.Fprintf(w, "- [%s] `%s` %s\n", check, r.ID, r.Title)

			if deps, ok := depsBySource[r.ID]; ok && len(deps) > 0 {
				sort.Strings(deps)
				var backticked []string
				for _, d := range deps {
					backticked = append(backticked, "`"+d+"`")
				}
				fmt.Fprintf(w, "  - Depends on: %s\n", strings.Join(backticked, ", "))
			}

			if r.Description != "" {
				lines := strings.Split(r.Description, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" {
						fmt.Fprintf(w, "  - %s\n", line)
					}
				}
			}
		}
	}
	return nil
}

func generateStatusSummary(phases []GeneratePhase) string {
	if len(phases) == 0 {
		return "No phases"
	}
	allComplete := true
	lastComplete := 0
	firstIncomplete := 0
	for _, gp := range phases {
		done := true
		for _, r := range gp.RMIs {
			if r.Status != "completed" && r.Status != "cancelled" {
				done = false
				break
			}
		}
		if done {
			lastComplete = gp.Phase.SequenceNumber
		} else {
			allComplete = false
			if firstIncomplete == 0 {
				firstIncomplete = gp.Phase.SequenceNumber
			}
		}
	}
	if allComplete {
		return "All phases completed"
	}
	var parts []string
	if lastComplete > 0 {
		if lastComplete == 1 {
			parts = append(parts, "Phase 1 completed")
		} else {
			parts = append(parts, fmt.Sprintf("Phases 1–%d completed", lastComplete))
		}
	}
	if firstIncomplete > 0 {
		parts = append(parts, fmt.Sprintf("Phase %d in progress", firstIncomplete))
	}
	if len(parts) == 0 {
		return "Planned"
	}
	return strings.Join(parts, ", ")
}

func derivePhaseLabel(rmis []*store.RoadmapItem) string {
	if len(rmis) == 0 {
		return "Empty"
	}
	allDone := true
	anyStarted := false
	for _, r := range rmis {
		if r.Status != "completed" && r.Status != "cancelled" {
			allDone = false
		}
		if r.Status == "in_progress" || r.Status == "completed" || r.Status == "cancelled" {
			anyStarted = true
		}
	}
	if allDone {
		return "Completed"
	}
	if anyStarted {
		return "In progress"
	}
	return "Planned"
}
