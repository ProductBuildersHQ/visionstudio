// Package roadmap parses and generates ROADMAP.md files that mirror
// the PRISM Control database state. The canonical format uses:
//
//	## Phase N — Title
//	**Theme:** ...
//	- [x] `RMI-REPO-NNN` Title
//	  - Depends on: `RMI-REPO-001`, `RMI-REPO-002`
package roadmap

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Roadmap is the parsed representation of a ROADMAP.md file.
type Roadmap struct {
	InitiativeID string
	Repository   string
	Title        string
	Phases       []Phase
}

// Phase is a themed grouping of RMIs.
type Phase struct {
	Number int
	Title  string
	Theme  string
	Items  []Item
}

// Item is a parsed RMI line.
type Item struct {
	ID          string
	Title       string
	Completed   bool
	DependsOn   []string
	Description string // sub-bullet text (Delivered, Acceptance, etc.)
}

var (
	reInitiative = regexp.MustCompile(`(?i)\*\*Initiative:\*\*\s*` + "`([^`]+)`")
	reRepository = regexp.MustCompile(`(?i)\*\*Repository:\*\*\s*` + "`?" + `([^\x60\s]+)` + "`?")
	rePhase      = regexp.MustCompile(`^##\s+Phase\s+(\d+)\s*[—–-]\s*(.+)`)
	reTheme      = regexp.MustCompile(`^\*\*Theme:\*\*\s*(.+)`)
	reRMI        = regexp.MustCompile(`^- \[([ xX])\]\s*` + "`(RMI-[A-Z0-9-]+)`" + `\s*(.*)`)
	reDepends    = regexp.MustCompile(`(?i)^\s+-\s+Depends on:\s*(.*)`)
	reRMIRef     = regexp.MustCompile("`(RMI-[A-Z0-9-]+)`")
	reTitle      = regexp.MustCompile(`^#\s+(.+)`)
)

// Parse reads a ROADMAP.md from r and returns its structured representation.
func Parse(r io.Reader) (*Roadmap, error) {
	scanner := bufio.NewScanner(r)
	rm := &Roadmap{}
	var currentPhase *Phase
	var currentItem *Item

	flushItem := func() {
		if currentItem != nil && currentPhase != nil {
			currentPhase.Items = append(currentPhase.Items, *currentItem)
			currentItem = nil
		}
	}
	flushPhase := func() {
		flushItem()
		if currentPhase != nil {
			rm.Phases = append(rm.Phases, *currentPhase)
			currentPhase = nil
		}
	}

	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if m := reTitle.FindStringSubmatch(line); m != nil && rm.Title == "" {
			rm.Title = strings.TrimSpace(m[1])
			continue
		}

		if m := reInitiative.FindStringSubmatch(line); m != nil {
			rm.InitiativeID = m[1]
			continue
		}

		if m := reRepository.FindStringSubmatch(line); m != nil {
			rm.Repository = strings.TrimSpace(m[1])
			continue
		}

		if m := rePhase.FindStringSubmatch(line); m != nil {
			flushPhase()
			num := 0
			if _, err := fmt.Sscanf(m[1], "%d", &num); err != nil {
				return nil, fmt.Errorf("line %d: invalid phase number %q", lineNum, m[1])
			}
			currentPhase = &Phase{
				Number: num,
				Title:  strings.TrimSpace(m[2]),
			}
			continue
		}

		if currentPhase != nil {
			if m := reTheme.FindStringSubmatch(line); m != nil {
				currentPhase.Theme = strings.TrimSpace(m[1])
				continue
			}
		}

		if m := reRMI.FindStringSubmatch(line); m != nil && currentPhase != nil {
			flushItem()
			currentItem = &Item{
				ID:        m[2],
				Title:     strings.TrimSpace(m[3]),
				Completed: strings.ToLower(m[1]) == "x",
			}
			continue
		}

		if currentItem != nil {
			if m := reDepends.FindStringSubmatch(line); m != nil {
				refs := reRMIRef.FindAllStringSubmatch(m[1], -1)
				for _, ref := range refs {
					currentItem.DependsOn = append(currentItem.DependsOn, ref[1])
				}
				continue
			}
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(line, "  - ") || strings.HasPrefix(line, "  ") && trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "**") {
				if currentItem.Description != "" {
					currentItem.Description += "\n"
				}
				currentItem.Description += trimmed
			}
		}
	}
	flushPhase()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	return rm, nil
}
