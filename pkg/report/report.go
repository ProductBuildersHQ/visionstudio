// Package report computes initiative reports from store data.
// All computation is SQL-portable (no Dolt system tables).
package report

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/initiative"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// Report is the full initiative report (TRD §10).
type Report struct {
	InitiativeID string        `json:"initiative_id"`
	Title        string        `json:"title"`
	Status       string        `json:"status"`
	Duration     Duration      `json:"duration"`
	Phases       []PhaseReport `json:"phases"`
	Repos        RepoSummary   `json:"repos"`
	RMIs         RMISummary    `json:"rmis"`
	Commits      CommitSummary `json:"commits"`
	Releases     []Release     `json:"releases"`
}

// Duration tracks initiative lifecycle timing.
type Duration struct {
	Created          *time.Time `json:"created,omitempty"`
	Executing        *time.Time `json:"executing,omitempty"`
	DeliveryComplete *time.Time `json:"delivery_complete,omitempty"`
	DaysExecuting    int        `json:"days_executing"`
}

// PhaseReport is one phase's summary.
type PhaseReport struct {
	PhaseID      string `json:"phase_id"`
	Title        string `json:"title"`
	Theme        string `json:"theme,omitempty"`
	Status       string `json:"status"`
	RMIsTotal    int    `json:"rmis_total"`
	RMIsComplete int    `json:"rmis_completed"`
}

// RepoSummary counts participating repositories.
type RepoSummary struct {
	Count int      `json:"count"`
	List  []string `json:"list"`
}

// RMISummary counts RMIs by status.
type RMISummary struct {
	Total             int `json:"total"`
	Completed         int `json:"completed"`
	RequiredCompleted int `json:"required_completed"`
	InProgress        int `json:"in_progress"`
	Ready             int `json:"ready"`
	Blocked           int `json:"blocked"`
}

// CommitSummary aggregates evidence of type "commit".
type CommitSummary struct {
	Total                int            `json:"total"`
	ByType               map[string]int `json:"by_type"`
	ByRepo               map[string]int `json:"by_repo"`
	UnattributedInWindow int            `json:"unattributed_in_window"`
}

// Release is a release evidence entry.
type Release struct {
	Repo    string `json:"repo"`
	Version string `json:"version"`
}

// Generate computes a full initiative report from the store.
func Generate(ctx context.Context, s store.Store, initiativeID string) (*Report, error) {
	init, err := s.GetInitiative(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("get initiative: %w", err)
	}

	phases, err := s.ListPhases(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("list phases: %w", err)
	}
	// ListPhases order is store-dependent (MemStore iterates a map), so sort by
	// sequence to present phases deterministically in the report.
	sort.Slice(phases, func(i, j int) bool {
		return phases[i].SequenceNumber < phases[j].SequenceNumber
	})

	rmis, err := s.ListRMIs(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("list RMIs: %w", err)
	}

	evidence, err := s.ListEvidenceByInitiative(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("list evidence: %w", err)
	}

	return build(init, phases, rmis, evidence), nil
}

func build(init *store.Initiative, phases []*store.Phase, rmis []*store.RoadmapItem, evidence []*store.DeliveryEvidence) *Report {
	r := &Report{
		InitiativeID: init.ID,
		Title:        init.Title,
		Status:       init.Status,
	}

	// Duration
	r.Duration.Created = &init.CreatedAt
	r.Duration.Executing = init.ExecutingAt
	r.Duration.DeliveryComplete = init.DeliveryCompleteAt
	if init.ExecutingAt != nil {
		end := time.Now()
		if init.DeliveryCompleteAt != nil {
			end = *init.DeliveryCompleteAt
		}
		r.Duration.DaysExecuting = int(math.Ceil(end.Sub(*init.ExecutingAt).Hours() / 24))
	}

	// Group RMIs by phase
	rmisByPhase := make(map[string][]*store.RoadmapItem)
	for _, rmi := range rmis {
		rmisByPhase[rmi.PhaseID] = append(rmisByPhase[rmi.PhaseID], rmi)
	}

	for _, p := range phases {
		phaseRMIs := rmisByPhase[p.ID]
		pr := PhaseReport{
			PhaseID:   p.ID,
			Title:     p.Title,
			Theme:     p.Theme,
			Status:    initiative.DerivePhaseStatus(phaseRMIs),
			RMIsTotal: len(phaseRMIs),
		}
		for _, rmi := range phaseRMIs {
			if rmi.Status == "completed" {
				pr.RMIsComplete++
			}
		}
		r.Phases = append(r.Phases, pr)
	}

	// RMI summary
	repos := make(map[string]bool)
	for _, rmi := range rmis {
		r.RMIs.Total++
		repos[rmi.RepositoryID] = true
		switch rmi.Status {
		case "completed":
			r.RMIs.Completed++
			if rmi.Required {
				r.RMIs.RequiredCompleted++
			}
		case "in_progress":
			r.RMIs.InProgress++
		case "ready":
			r.RMIs.Ready++
		case "blocked":
			r.RMIs.Blocked++
		}
	}

	// Repos
	for repoID := range repos {
		short := repoID
		if idx := strings.LastIndex(short, "/"); idx >= 0 {
			short = short[idx+1:]
		}
		r.Repos.List = append(r.Repos.List, short)
	}
	r.Repos.Count = len(r.Repos.List)

	// Evidence → commit summary + releases
	r.Commits.ByType = make(map[string]int)
	r.Commits.ByRepo = make(map[string]int)

	rmiToRepo := make(map[string]string)
	for _, rmi := range rmis {
		short := rmi.RepositoryID
		if idx := strings.LastIndex(short, "/"); idx >= 0 {
			short = short[idx+1:]
		}
		rmiToRepo[rmi.ID] = short
	}

	for _, ev := range evidence {
		switch ev.EvidenceType {
		case "commit":
			r.Commits.Total++
			if ev.CommitType != "" {
				r.Commits.ByType[ev.CommitType]++
			}
			if repo, ok := rmiToRepo[ev.RMIID]; ok {
				r.Commits.ByRepo[repo]++
			}
		case "release":
			if repo, ok := rmiToRepo[ev.RMIID]; ok {
				r.Releases = append(r.Releases, Release{
					Repo:    repo,
					Version: ev.Reference,
				})
			}
		}
	}

	return r
}

// Markdown renders the report as a Markdown string.
func (r *Report) Markdown() string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Initiative Report: %s\n\n", r.InitiativeID)
	fmt.Fprintf(&b, "**Title:** %s\n", r.Title)
	fmt.Fprintf(&b, "**Status:** %s\n\n", r.Status)

	if r.Duration.DaysExecuting > 0 {
		fmt.Fprintf(&b, "**Days executing:** %d\n\n", r.Duration.DaysExecuting)
	}

	fmt.Fprintf(&b, "## Phases\n\n")
	fmt.Fprintf(&b, "| Phase | Title | Status | RMIs |\n")
	fmt.Fprintf(&b, "|-------|-------|--------|------|\n")
	for _, p := range r.Phases {
		fmt.Fprintf(&b, "| %s | %s | %s | %d/%d |\n",
			p.PhaseID, p.Title, p.Status, p.RMIsComplete, p.RMIsTotal)
	}

	fmt.Fprintf(&b, "\n## RMI Summary\n\n")
	fmt.Fprintf(&b, "- Total: %d\n", r.RMIs.Total)
	fmt.Fprintf(&b, "- Completed: %d\n", r.RMIs.Completed)
	fmt.Fprintf(&b, "- In progress: %d\n", r.RMIs.InProgress)
	fmt.Fprintf(&b, "- Ready: %d\n", r.RMIs.Ready)
	fmt.Fprintf(&b, "- Blocked: %d\n", r.RMIs.Blocked)

	fmt.Fprintf(&b, "\n## Repositories (%d)\n\n", r.Repos.Count)
	for _, repo := range r.Repos.List {
		fmt.Fprintf(&b, "- %s\n", repo)
	}

	if r.Commits.Total > 0 {
		fmt.Fprintf(&b, "\n## Commits (%d)\n\n", r.Commits.Total)
		if len(r.Commits.ByType) > 0 {
			fmt.Fprintf(&b, "**By type:**\n\n")
			for t, n := range r.Commits.ByType {
				fmt.Fprintf(&b, "- %s: %d\n", t, n)
			}
		}
		if len(r.Commits.ByRepo) > 0 {
			fmt.Fprintf(&b, "\n**By repo:**\n\n")
			for repo, n := range r.Commits.ByRepo {
				fmt.Fprintf(&b, "- %s: %d\n", repo, n)
			}
		}
	}

	if len(r.Releases) > 0 {
		fmt.Fprintf(&b, "\n## Releases\n\n")
		for _, rel := range r.Releases {
			fmt.Fprintf(&b, "- %s %s\n", rel.Repo, rel.Version)
		}
	}

	return b.String()
}
