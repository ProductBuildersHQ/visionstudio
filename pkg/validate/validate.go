// Package validate implements consistency checks for a PRISM Control store (TRD §9).
package validate

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

var rmiIDPattern = regexp.MustCompile(`^RMI-[A-Z0-9]+-\d{3,}$`)

// Finding is a single validation problem.
type Finding struct {
	Check   string `json:"check"`
	Level   string `json:"level"` // "error" or "warning"
	Message string `json:"message"`
}

// Result holds all findings from a validation run.
type Result struct {
	Findings []Finding `json:"findings"`
}

func (r *Result) add(check, level, msg string) {
	r.Findings = append(r.Findings, Finding{Check: check, Level: level, Message: msg})
}

func (r *Result) errorf(check, format string, args ...any) {
	r.add(check, "error", fmt.Sprintf(format, args...))
}

func (r *Result) warnf(check, format string, args ...any) {
	r.add(check, "warning", fmt.Sprintf(format, args...))
}

// Errors returns only error-level findings.
func (r *Result) Errors() []Finding {
	var out []Finding
	for _, f := range r.Findings {
		if f.Level == "error" {
			out = append(out, f)
		}
	}
	return out
}

// OK returns true if there are no error-level findings.
func (r *Result) OK() bool {
	return len(r.Errors()) == 0
}

// Run executes all validation checks against the store.
func Run(ctx context.Context, s store.Store, now time.Time) (*Result, error) {
	rmis, err := s.ListAllRMIs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list RMIs: %w", err)
	}
	deps, err := s.ListAllDependencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list dependencies: %w", err)
	}
	assignments, err := s.ListAllAssignments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list assignments: %w", err)
	}
	evidence, err := s.ListAllEvidence(ctx)
	if err != nil {
		return nil, fmt.Errorf("list evidence: %w", err)
	}
	initiatives, err := s.ListInitiatives(ctx)
	if err != nil {
		return nil, fmt.Errorf("list initiatives: %w", err)
	}

	rmiByID := make(map[string]*store.RoadmapItem, len(rmis))
	for _, rmi := range rmis {
		rmiByID[rmi.ID] = rmi
	}

	repos, err := s.ListRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}
	repoByID := make(map[string]*store.Repository, len(repos))
	for _, repo := range repos {
		repoByID[repo.ID] = repo
	}

	r := &Result{}
	checkIDFormat(r, rmis)
	checkDuplicateIDs(r, rmis)
	checkEvidenceRefs(r, evidence, rmiByID)
	checkDependencies(r, deps, rmiByID)
	checkExpiredLeases(r, assignments, now)
	checkStatusCoherence(r, initiatives, rmis, evidence)
	checkContextSpecs(r, rmis, repoByID)
	return r, nil
}

// Check 1: ID format violations.
func checkIDFormat(r *Result, rmis []*store.RoadmapItem) {
	for _, rmi := range rmis {
		if !rmiIDPattern.MatchString(rmi.ID) {
			r.errorf("id_format", "RMI %q does not match expected pattern RMI-<REPOSLUG>-<NNN>", rmi.ID)
		}
	}
}

// Check 2: Duplicate RMI IDs (the in-memory store may allow them).
func checkDuplicateIDs(r *Result, rmis []*store.RoadmapItem) {
	seen := make(map[string]bool, len(rmis))
	for _, rmi := range rmis {
		if seen[rmi.ID] {
			r.errorf("duplicate_id", "duplicate RMI ID: %s", rmi.ID)
		}
		seen[rmi.ID] = true
	}
}

// Check 3: Evidence references non-existent RMI, or RMI repo mismatch.
func checkEvidenceRefs(r *Result, evidence []*store.DeliveryEvidence, rmiByID map[string]*store.RoadmapItem) {
	for _, ev := range evidence {
		if _, ok := rmiByID[ev.RMIID]; !ok {
			r.errorf("evidence_ref", "evidence %s references non-existent RMI %s", ev.ID, ev.RMIID)
		}
	}
}

// Check 4: Dangling dependency edges and dependency cycles.
func checkDependencies(r *Result, deps []*store.RMIDependency, rmiByID map[string]*store.RoadmapItem) {
	for _, dep := range deps {
		if _, ok := rmiByID[dep.SourceRMIID]; !ok {
			r.errorf("dangling_dep", "dependency source %s does not exist", dep.SourceRMIID)
		}
		if _, ok := rmiByID[dep.TargetRMIID]; !ok {
			r.errorf("dangling_dep", "dependency target %s does not exist", dep.TargetRMIID)
		}
	}

	// Cycle detection via DFS.
	adj := make(map[string][]string)
	for _, dep := range deps {
		if dep.Relationship == "requires" {
			adj[dep.SourceRMIID] = append(adj[dep.SourceRMIID], dep.TargetRMIID)
		}
	}
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[string]int)
	var path []string
	var dfs func(node string) bool
	dfs = func(node string) bool {
		color[node] = gray
		path = append(path, node)
		for _, next := range adj[node] {
			switch color[next] {
			case gray:
				cycleStart := -1
				for i, n := range path {
					if n == next {
						cycleStart = i
						break
					}
				}
				cycle := path[cycleStart:]
				r.errorf("dependency_cycle", "cycle detected: %v", cycle)
				return true
			case white:
				if dfs(next) {
					return true
				}
			}
		}
		path = path[:len(path)-1]
		color[node] = black
		return false
	}
	for id := range rmiByID {
		if color[id] == white {
			dfs(id)
		}
	}
}

// Check 5: Expired active leases.
func checkExpiredLeases(r *Result, assignments []*store.Assignment, now time.Time) {
	for _, a := range assignments {
		if a.Status == "active" && now.After(a.LeaseExpiresAt) {
			r.warnf("expired_lease", "assignment %s (RMI %s, worker %s) lease expired at %s",
				a.ID, a.RMIID, a.Worker, a.LeaseExpiresAt.Format(time.RFC3339))
		}
	}
}

// Check 6: Status coherence.
func checkStatusCoherence(r *Result, initiatives []*store.Initiative, rmis []*store.RoadmapItem, evidence []*store.DeliveryEvidence) {
	rmisByInit := make(map[string][]*store.RoadmapItem)
	for _, rmi := range rmis {
		rmisByInit[rmi.InitiativeID] = append(rmisByInit[rmi.InitiativeID], rmi)
	}

	// Initiative marked delivery_complete with open required RMIs.
	for _, init := range initiatives {
		if init.DeliveryCompleteAt == nil {
			continue
		}
		for _, rmi := range rmisByInit[init.ID] {
			if rmi.Required && rmi.Status != "completed" {
				r.errorf("status_coherence",
					"initiative %s is delivery_complete but required RMI %s has status %q",
					init.ID, rmi.ID, rmi.Status)
			}
		}
	}

	// Completed RMIs with no evidence.
	evidenceByRMI := make(map[string]bool)
	for _, ev := range evidence {
		evidenceByRMI[ev.RMIID] = true
	}
	for _, rmi := range rmis {
		if rmi.Status == "completed" && !evidenceByRMI[rmi.ID] {
			r.warnf("status_coherence", "RMI %s is completed but has no delivery evidence", rmi.ID)
		}
	}
}

// Check 7: ContextSpec validation — extra_repos must reference existing repositories.
func checkContextSpecs(r *Result, rmis []*store.RoadmapItem, repoByID map[string]*store.Repository) {
	for _, rmi := range rmis {
		if rmi.ContextSpec == nil {
			continue
		}

		for _, repoID := range rmi.ContextSpec.ExtraRepos {
			if _, ok := repoByID[repoID]; !ok {
				r.errorf("context_spec", "RMI %s context_spec.extra_repos references non-existent repository %q", rmi.ID, repoID)
			}
		}
	}
}
