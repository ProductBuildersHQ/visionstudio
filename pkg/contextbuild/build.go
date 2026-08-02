package contextbuild

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/initiative"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// Builder assembles context packages from store data.
type Builder struct {
	store      store.Store
	doltCommit string
}

// NewBuilder creates a Builder backed by the given store.
// doltCommit is the current Dolt HEAD revision for provenance tracking.
func NewBuilder(s store.Store, doltCommit string) *Builder {
	return &Builder{
		store:      s,
		doltCommit: doltCommit,
	}
}

// BuildForPhase assembles a context package for an entire phase.
func (b *Builder) BuildForPhase(ctx context.Context, phaseID string) (*ContextPackage, error) {
	parts := strings.Split(phaseID, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid phase ID format: %s", phaseID)
	}
	initiativeID := parts[0]

	init, err := b.store.GetInitiative(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("get initiative: %w", err)
	}

	phases, err := b.store.ListPhases(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("list phases: %w", err)
	}

	var targetPhase *store.Phase
	var priorPhases []*store.Phase
	for _, p := range phases {
		if p.ID == phaseID {
			targetPhase = p
		} else if targetPhase == nil {
			priorPhases = append(priorPhases, p)
		}
	}
	if targetPhase == nil {
		return nil, fmt.Errorf("phase not found: %s", phaseID)
	}

	rmis, err := b.store.ListRMIs(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("list RMIs: %w", err)
	}

	allDeps, err := b.store.ListAllDependencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list dependencies: %w", err)
	}

	var program *store.Program
	if init.ProgramID != "" {
		program, err = b.store.GetProgram(ctx, init.ProgramID)
		if err != nil {
			return nil, fmt.Errorf("get program: %w", err)
		}
	}

	pkg := &ContextPackage{
		Version:        SchemaVersion,
		TargetType:     "phase",
		TargetID:       phaseID,
		BuildTimestamp: time.Now().UTC().Truncate(time.Second),
	}

	prov := Provenance{Source: "dolt", Revision: b.doltCommit}

	if program != nil {
		pkg.Sections.Program = &ProgramSection{
			Stability:   Stable,
			Provenance:  prov,
			ID:          program.ID,
			Name:        program.Name,
			Description: program.Description,
		}
	}

	pkg.Sections.Initiative = InitiativeSection{
		Stability:   Stable,
		Provenance:  prov,
		ID:          init.ID,
		Title:       init.Title,
		Description: init.Description,
		Status:      init.Status,
		Priority:    init.Priority,
		HomeRepo:    init.HomeRepo,
		Specs:       init.Specs,
	}

	phaseRMIs := filterRMIsByPhase(rmis, phaseID)
	phaseDeps := filterDependenciesByRMIs(allDeps, phaseRMIs)

	pkg.Sections.Phase = PhaseSection{
		Stability:       PhaseStable,
		Provenance:      prov,
		ID:              targetPhase.ID,
		Title:           targetPhase.Title,
		Theme:           targetPhase.Theme,
		SequenceNumber:  targetPhase.SequenceNumber,
		DerivedStatus:   initiative.DerivePhaseStatus(phaseRMIs),
		MemberRMIs:      toRMISummaries(phaseRMIs),
		DependencyEdges: toDependencyEdges(phaseDeps),
	}

	handoffs, err := b.buildPrerequisiteHandoffs(ctx, priorPhases, rmis, prov)
	if err != nil {
		return nil, fmt.Errorf("build prerequisite handoffs: %w", err)
	}
	pkg.Sections.PrerequisiteHandoffs = handoffs

	pkg.Sections.SpecReferences = b.buildSpecReferences(ctx, init, nil)

	pkg.DerivedRepos = b.deriveRepoSet(ctx, phaseRMIs, allDeps)

	return pkg, nil
}

// BuildForRMI assembles a context package for a specific RMI.
func (b *Builder) BuildForRMI(ctx context.Context, rmiID string) (*ContextPackage, error) {
	rmi, err := b.store.GetRMI(ctx, rmiID)
	if err != nil {
		return nil, fmt.Errorf("get RMI: %w", err)
	}

	if rmi.InitiativeID == "" {
		return nil, fmt.Errorf("RMI %s has no initiative", rmiID)
	}

	init, err := b.store.GetInitiative(ctx, rmi.InitiativeID)
	if err != nil {
		return nil, fmt.Errorf("get initiative: %w", err)
	}

	phases, err := b.store.ListPhases(ctx, rmi.InitiativeID)
	if err != nil {
		return nil, fmt.Errorf("list phases: %w", err)
	}

	var targetPhase *store.Phase
	var priorPhases []*store.Phase
	for _, p := range phases {
		if p.ID == rmi.PhaseID {
			targetPhase = p
		} else if targetPhase == nil {
			priorPhases = append(priorPhases, p)
		}
	}
	if targetPhase == nil && rmi.PhaseID != "" {
		return nil, fmt.Errorf("phase not found: %s", rmi.PhaseID)
	}

	rmis, err := b.store.ListRMIs(ctx, rmi.InitiativeID)
	if err != nil {
		return nil, fmt.Errorf("list RMIs: %w", err)
	}

	allDeps, err := b.store.ListAllDependencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list dependencies: %w", err)
	}

	var program *store.Program
	if init.ProgramID != "" {
		program, err = b.store.GetProgram(ctx, init.ProgramID)
		if err != nil {
			return nil, fmt.Errorf("get program: %w", err)
		}
	}

	pkg := &ContextPackage{
		Version:        SchemaVersion,
		TargetType:     "rmi",
		TargetID:       rmiID,
		BuildTimestamp: time.Now().UTC().Truncate(time.Second),
	}

	prov := Provenance{Source: "dolt", Revision: b.doltCommit}

	if program != nil {
		pkg.Sections.Program = &ProgramSection{
			Stability:   Stable,
			Provenance:  prov,
			ID:          program.ID,
			Name:        program.Name,
			Description: program.Description,
		}
	}

	pkg.Sections.Initiative = InitiativeSection{
		Stability:   Stable,
		Provenance:  prov,
		ID:          init.ID,
		Title:       init.Title,
		Description: init.Description,
		Status:      init.Status,
		Priority:    init.Priority,
		HomeRepo:    init.HomeRepo,
		Specs:       init.Specs,
	}

	if targetPhase != nil {
		phaseRMIs := filterRMIsByPhase(rmis, targetPhase.ID)
		phaseDeps := filterDependenciesByRMIs(allDeps, phaseRMIs)

		pkg.Sections.Phase = PhaseSection{
			Stability:       PhaseStable,
			Provenance:      prov,
			ID:              targetPhase.ID,
			Title:           targetPhase.Title,
			Theme:           targetPhase.Theme,
			SequenceNumber:  targetPhase.SequenceNumber,
			DerivedStatus:   initiative.DerivePhaseStatus(phaseRMIs),
			MemberRMIs:      toRMISummaries(phaseRMIs),
			DependencyEdges: toDependencyEdges(phaseDeps),
		}

		handoffs, err := b.buildPrerequisiteHandoffs(ctx, priorPhases, rmis, prov)
		if err != nil {
			return nil, fmt.Errorf("build prerequisite handoffs: %w", err)
		}
		pkg.Sections.PrerequisiteHandoffs = handoffs
	}

	rmiDeps := filterDependenciesForRMI(allDeps, rmiID)
	var deps, dependents []DependencyEdge
	for _, d := range rmiDeps {
		if d.SourceRMIID == rmiID {
			deps = append(deps, DependencyEdge{
				SourceID:     d.SourceRMIID,
				TargetID:     d.TargetRMIID,
				Relationship: d.Relationship,
			})
		} else {
			dependents = append(dependents, DependencyEdge{
				SourceID:     d.SourceRMIID,
				TargetID:     d.TargetRMIID,
				Relationship: d.Relationship,
			})
		}
	}

	pkg.Sections.CurrentRMI = &RMISection{
		Stability:          RMIStable,
		Provenance:         prov,
		ID:                 rmi.ID,
		Title:              rmi.Title,
		Description:        rmi.Description,
		ItemType:           rmi.ItemType,
		Status:             rmi.Status,
		Priority:           rmi.Priority,
		Required:           rmi.Required,
		SequenceNumber:     rmi.SequenceNumber,
		AcceptanceCriteria: rmi.AcceptanceCriteria,
		RepositoryID:       rmi.RepositoryID,
		Dependencies:       deps,
		Dependents:         dependents,
	}

	pkg.Sections.SpecReferences = b.buildSpecReferences(ctx, init, rmi.ContextSpec)

	assignment, err := b.store.GetActiveAssignment(ctx, rmiID)
	if err == nil && assignment != nil {
		evidence, err := b.store.ListEvidenceByRMI(ctx, rmiID)
		if err != nil {
			return nil, fmt.Errorf("list evidence: %w", err)
		}

		pkg.Sections.Assignment = &AssignmentSection{
			Stability:      Volatile,
			Provenance:     prov,
			ID:             assignment.ID,
			RMIID:          assignment.RMIID,
			Worker:         assignment.Worker,
			Status:         assignment.Status,
			LeaseExpiresAt: assignment.LeaseExpiresAt,
			Workspace:      assignment.Workspace,
			CreatedAt:      assignment.CreatedAt,
			CompletedAt:    assignment.CompletedAt,
			Evidence:       toEvidenceRefs(evidence),
		}

		if assignment.Handoff != nil {
			pkg.Sections.Assignment.Handoff = &HandoffData{
				Completed:  assignment.Handoff.Completed,
				Remaining:  assignment.Handoff.Remaining,
				Decisions:  assignment.Handoff.Decisions,
				NextAction: assignment.Handoff.NextAction,
			}
		}
	}

	pkg.DerivedRepos = b.deriveRepoSetForRMI(ctx, rmi, allDeps)

	return pkg, nil
}

func (b *Builder) buildPrerequisiteHandoffs(ctx context.Context, priorPhases []*store.Phase, rmis []*store.RoadmapItem, prov Provenance) ([]PhaseHandoff, error) {
	var handoffs []PhaseHandoff

	for _, phase := range priorPhases {
		phaseRMIs := filterRMIsByPhase(rmis, phase.ID)
		status := initiative.DerivePhaseStatus(phaseRMIs)

		if status != "completed" && status != "partial" {
			continue
		}

		var completed []string
		var decisions []string

		for _, rmi := range phaseRMIs {
			if rmi.Status == "completed" {
				completed = append(completed, rmi.Title)
			}
		}

		assignments, err := b.store.ListActiveAssignments(ctx)
		if err != nil {
			return nil, err
		}

		for _, a := range assignments {
			for _, rmi := range phaseRMIs {
				if a.RMIID == rmi.ID && a.Handoff != nil {
					decisions = append(decisions, a.Handoff.Decisions...)
				}
			}
		}

		handoffs = append(handoffs, PhaseHandoff{
			Stability:     PhaseStable,
			Provenance:    prov,
			PhaseID:       phase.ID,
			PhaseTitle:    phase.Title,
			DerivedStatus: status,
			Completed:     completed,
			Decisions:     decisions,
		})
	}

	return handoffs, nil
}

func (b *Builder) buildSpecReferences(ctx context.Context, init *store.Initiative, contextSpec *store.ContextSpec) []SpecReference {
	var refs []SpecReference

	if init.HomeRepo == "" {
		return refs
	}

	repo, err := b.store.GetRepository(ctx, init.HomeRepo)
	if err != nil {
		return refs
	}

	if repo.LocalPath == "" {
		return refs
	}

	excludeSet := make(map[string]bool)
	if contextSpec != nil {
		for _, path := range contextSpec.ExcludeSpecs {
			excludeSet[path] = true
		}
	}

	for _, specPath := range init.Specs {
		if specPath == "" {
			continue
		}

		if excludeSet[specPath] {
			continue
		}

		fullPath := filepath.Join(repo.LocalPath, specPath)
		gitRev := getGitRevision(repo.LocalPath, specPath)

		prov := Provenance{Source: "git", Revision: gitRev}

		refs = append(refs, SpecReference{
			Stability:  RMIStable,
			Provenance: prov,
			Path:       specPath,
			RepoID:     repo.ID,
			LocalPath:  fullPath,
		})
	}

	if contextSpec != nil {
		for _, specPath := range contextSpec.IncludeSpecs {
			if excludeSet[specPath] {
				continue
			}

			alreadyIncluded := false
			for _, ref := range refs {
				if ref.Path == specPath {
					alreadyIncluded = true
					break
				}
			}
			if alreadyIncluded {
				continue
			}

			fullPath := filepath.Join(repo.LocalPath, specPath)
			gitRev := getGitRevision(repo.LocalPath, specPath)

			prov := Provenance{Source: "git", Revision: gitRev}

			refs = append(refs, SpecReference{
				Stability:  RMIStable,
				Provenance: prov,
				Path:       specPath,
				RepoID:     repo.ID,
				LocalPath:  fullPath,
			})
		}
	}

	sort.Slice(refs, func(i, j int) bool {
		return refs[i].Path < refs[j].Path
	})

	return refs
}

func (b *Builder) deriveRepoSet(ctx context.Context, phaseRMIs []*store.RoadmapItem, allDeps []*store.RMIDependency) []DerivedRepo {
	repoSet := make(map[string]*DerivedRepo)

	for _, rmi := range phaseRMIs {
		if rmi.RepositoryID == "" {
			continue
		}
		if _, exists := repoSet[rmi.RepositoryID]; !exists {
			repo, err := b.store.GetRepository(ctx, rmi.RepositoryID)
			if err != nil {
				continue
			}
			repoSet[rmi.RepositoryID] = &DerivedRepo{
				ID:            repo.ID,
				Role:          "primary",
				LocalPath:     repo.LocalPath,
				DefaultBranch: repo.DefaultBranch,
			}
		}
	}

	rmiIDs := make(map[string]bool)
	for _, rmi := range phaseRMIs {
		rmiIDs[rmi.ID] = true
	}

	for _, dep := range allDeps {
		if !rmiIDs[dep.SourceRMIID] {
			continue
		}

		targetRMI, err := b.store.GetRMI(ctx, dep.TargetRMIID)
		if err != nil || targetRMI.RepositoryID == "" {
			continue
		}

		if _, exists := repoSet[targetRMI.RepositoryID]; exists {
			continue
		}

		repo, err := b.store.GetRepository(ctx, targetRMI.RepositoryID)
		if err != nil {
			continue
		}

		repoSet[targetRMI.RepositoryID] = &DerivedRepo{
			ID:            repo.ID,
			Role:          "dependency_rmi",
			LocalPath:     repo.LocalPath,
			DefaultBranch: repo.DefaultBranch,
			SourceRMI:     dep.TargetRMIID,
		}
	}

	b.addRepoDependencies(ctx, repoSet)

	return sortDerivedRepos(repoSet)
}

func (b *Builder) deriveRepoSetForRMI(ctx context.Context, rmi *store.RoadmapItem, allDeps []*store.RMIDependency) []DerivedRepo {
	repoSet := make(map[string]*DerivedRepo)

	if rmi.RepositoryID != "" {
		repo, err := b.store.GetRepository(ctx, rmi.RepositoryID)
		if err == nil {
			repoSet[rmi.RepositoryID] = &DerivedRepo{
				ID:            repo.ID,
				Role:          "primary",
				LocalPath:     repo.LocalPath,
				DefaultBranch: repo.DefaultBranch,
			}
		}
	}

	for _, dep := range allDeps {
		if dep.SourceRMIID != rmi.ID {
			continue
		}

		targetRMI, err := b.store.GetRMI(ctx, dep.TargetRMIID)
		if err != nil || targetRMI.RepositoryID == "" {
			continue
		}

		if _, exists := repoSet[targetRMI.RepositoryID]; exists {
			continue
		}

		repo, err := b.store.GetRepository(ctx, targetRMI.RepositoryID)
		if err != nil {
			continue
		}

		repoSet[targetRMI.RepositoryID] = &DerivedRepo{
			ID:            repo.ID,
			Role:          "dependency_rmi",
			LocalPath:     repo.LocalPath,
			DefaultBranch: repo.DefaultBranch,
			SourceRMI:     dep.TargetRMIID,
		}
	}

	b.addRepoDependencies(ctx, repoSet)

	b.addExplicitRepos(ctx, rmi, repoSet)

	return sortDerivedRepos(repoSet)
}

func (b *Builder) addExplicitRepos(ctx context.Context, rmi *store.RoadmapItem, repoSet map[string]*DerivedRepo) {
	if rmi.ContextSpec == nil || len(rmi.ContextSpec.ExtraRepos) == 0 {
		return
	}

	for _, repoID := range rmi.ContextSpec.ExtraRepos {
		if _, exists := repoSet[repoID]; exists {
			continue
		}

		repo, err := b.store.GetRepository(ctx, repoID)
		if err != nil {
			continue
		}

		repoSet[repoID] = &DerivedRepo{
			ID:            repo.ID,
			Role:          "explicit",
			LocalPath:     repo.LocalPath,
			DefaultBranch: repo.DefaultBranch,
		}
	}
}

func (b *Builder) addRepoDependencies(ctx context.Context, repoSet map[string]*DerivedRepo) {
	primaryRepos := make([]string, 0)
	for id, dr := range repoSet {
		if dr.Role == "primary" {
			primaryRepos = append(primaryRepos, id)
		}
	}

	for _, repoID := range primaryRepos {
		deps, err := b.store.ListRepoDependencies(ctx, repoID)
		if err != nil {
			continue
		}

		for _, dep := range deps {
			if _, exists := repoSet[dep.TargetRepositoryID]; exists {
				continue
			}

			repo, err := b.store.GetRepository(ctx, dep.TargetRepositoryID)
			if err != nil {
				continue
			}

			repoSet[dep.TargetRepositoryID] = &DerivedRepo{
				ID:            repo.ID,
				Role:          "repo_dependency",
				LocalPath:     repo.LocalPath,
				DefaultBranch: repo.DefaultBranch,
				SourceRepo:    repoID,
			}
		}
	}
}

func filterRMIsByPhase(rmis []*store.RoadmapItem, phaseID string) []*store.RoadmapItem {
	var result []*store.RoadmapItem
	for _, rmi := range rmis {
		if rmi.PhaseID == phaseID {
			result = append(result, rmi)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SequenceNumber != result[j].SequenceNumber {
			return result[i].SequenceNumber < result[j].SequenceNumber
		}
		return result[i].ID < result[j].ID
	})
	return result
}

func filterDependenciesByRMIs(deps []*store.RMIDependency, rmis []*store.RoadmapItem) []*store.RMIDependency {
	rmiIDs := make(map[string]bool)
	for _, rmi := range rmis {
		rmiIDs[rmi.ID] = true
	}

	var result []*store.RMIDependency
	for _, dep := range deps {
		if rmiIDs[dep.SourceRMIID] || rmiIDs[dep.TargetRMIID] {
			result = append(result, dep)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SourceRMIID != result[j].SourceRMIID {
			return result[i].SourceRMIID < result[j].SourceRMIID
		}
		return result[i].TargetRMIID < result[j].TargetRMIID
	})
	return result
}

func filterDependenciesForRMI(deps []*store.RMIDependency, rmiID string) []*store.RMIDependency {
	var result []*store.RMIDependency
	for _, dep := range deps {
		if dep.SourceRMIID == rmiID || dep.TargetRMIID == rmiID {
			result = append(result, dep)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SourceRMIID != result[j].SourceRMIID {
			return result[i].SourceRMIID < result[j].SourceRMIID
		}
		return result[i].TargetRMIID < result[j].TargetRMIID
	})
	return result
}

func toRMISummaries(rmis []*store.RoadmapItem) []RMISummary {
	result := make([]RMISummary, len(rmis))
	for i, rmi := range rmis {
		result[i] = RMISummary{
			ID:       rmi.ID,
			Title:    rmi.Title,
			Status:   rmi.Status,
			Required: rmi.Required,
			Sequence: rmi.SequenceNumber,
			RepoID:   rmi.RepositoryID,
		}
	}
	return result
}

func toDependencyEdges(deps []*store.RMIDependency) []DependencyEdge {
	result := make([]DependencyEdge, len(deps))
	for i, dep := range deps {
		result[i] = DependencyEdge{
			SourceID:     dep.SourceRMIID,
			TargetID:     dep.TargetRMIID,
			Relationship: dep.Relationship,
		}
	}
	return result
}

func toEvidenceRefs(evidence []*store.DeliveryEvidence) []EvidenceRef {
	result := make([]EvidenceRef, len(evidence))
	for i, ev := range evidence {
		result[i] = EvidenceRef{
			Type:       ev.EvidenceType,
			Reference:  ev.Reference,
			OccurredAt: ev.OccurredAt,
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Type != result[j].Type {
			return result[i].Type < result[j].Type
		}
		return result[i].Reference < result[j].Reference
	})
	return result
}

func sortDerivedRepos(repoSet map[string]*DerivedRepo) []DerivedRepo {
	result := make([]DerivedRepo, 0, len(repoSet))
	for _, dr := range repoSet {
		result = append(result, *dr)
	}

	roleOrder := map[string]int{"primary": 0, "explicit": 1, "dependency_rmi": 2, "repo_dependency": 3}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Role != result[j].Role {
			return roleOrder[result[i].Role] < roleOrder[result[j].Role]
		}
		return result[i].ID < result[j].ID
	})

	return result
}

func getGitRevision(repoPath, filePath string) string {
	cmd := exec.Command("git", "log", "-1", "--format=%H", "--", filePath)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
