package store

import (
	"context"
	"fmt"
	"sync"
)

// MemStore is an in-memory Store implementation for unit testing.
// It is safe for concurrent use.
type MemStore struct {
	mu                  sync.RWMutex
	programs            map[string]*Program
	initiatives         map[string]*Initiative
	phases              map[string]*Phase
	rmis                map[string]*RoadmapItem
	deps                []*RMIDependency
	initDeps            []*InitiativeDependency
	assignments         map[string]*Assignment
	evidence            map[string]*DeliveryEvidence
	repositories        map[string]*Repository
	repoDeps            []*RepositoryDependency
	workflows           map[string]*SpecWorkflow
	judgeRubrics        map[string]*JudgeRubric
	judgeResults        map[string]*JudgeResult
	capabilityModels    map[string]*CapabilityModel
	maturityAssessments map[string]*MaturityAssessment
}

// NewMemStore creates a new in-memory store.
func NewMemStore() *MemStore {
	return &MemStore{
		programs:            make(map[string]*Program),
		initiatives:         make(map[string]*Initiative),
		phases:              make(map[string]*Phase),
		rmis:                make(map[string]*RoadmapItem),
		assignments:         make(map[string]*Assignment),
		evidence:            make(map[string]*DeliveryEvidence),
		repositories:        make(map[string]*Repository),
		workflows:           make(map[string]*SpecWorkflow),
		judgeRubrics:        make(map[string]*JudgeRubric),
		judgeResults:        make(map[string]*JudgeResult),
		capabilityModels:    make(map[string]*CapabilityModel),
		maturityAssessments: make(map[string]*MaturityAssessment),
	}
}

func (m *MemStore) CreateProgram(_ context.Context, prog *Program) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.programs[prog.ID]; exists {
		return fmt.Errorf("program %s already exists", prog.ID)
	}
	m.programs[prog.ID] = prog
	return nil
}

func (m *MemStore) GetProgram(_ context.Context, id string) (*Program, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	prog, ok := m.programs[id]
	if !ok {
		return nil, fmt.Errorf("program %s not found", id)
	}
	return prog, nil
}

func (m *MemStore) ListPrograms(_ context.Context) ([]*Program, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Program, 0, len(m.programs))
	for _, v := range m.programs {
		result = append(result, v)
	}
	return result, nil
}

func (m *MemStore) UpdateProgram(_ context.Context, prog *Program) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.programs[prog.ID]; !exists {
		return fmt.Errorf("program %s not found", prog.ID)
	}
	m.programs[prog.ID] = prog
	return nil
}

func (m *MemStore) CreateInitiative(_ context.Context, init *Initiative) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.initiatives[init.ID]; exists {
		return fmt.Errorf("initiative %s already exists", init.ID)
	}
	m.initiatives[init.ID] = init
	return nil
}

func (m *MemStore) GetInitiative(_ context.Context, id string) (*Initiative, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	init, ok := m.initiatives[id]
	if !ok {
		return nil, fmt.Errorf("initiative %s not found", id)
	}
	return init, nil
}

func (m *MemStore) ListInitiatives(_ context.Context) ([]*Initiative, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Initiative, 0, len(m.initiatives))
	for _, v := range m.initiatives {
		result = append(result, v)
	}
	return result, nil
}

func (m *MemStore) UpdateInitiative(_ context.Context, init *Initiative) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.initiatives[init.ID]; !exists {
		return fmt.Errorf("initiative %s not found", init.ID)
	}
	m.initiatives[init.ID] = init
	return nil
}

func (m *MemStore) CreateInitiativeDependency(_ context.Context, dep *InitiativeDependency) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.initDeps {
		if d.SourceInitiativeID == dep.SourceInitiativeID &&
			d.TargetInitiativeID == dep.TargetInitiativeID &&
			d.Relationship == dep.Relationship {
			return nil
		}
	}
	m.initDeps = append(m.initDeps, dep)
	return nil
}

func (m *MemStore) ListInitiativeDependencies(_ context.Context, initiativeID string) ([]*InitiativeDependency, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*InitiativeDependency
	for _, d := range m.initDeps {
		if d.SourceInitiativeID == initiativeID || d.TargetInitiativeID == initiativeID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *MemStore) ListAllInitiativeDependencies(_ context.Context) ([]*InitiativeDependency, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*InitiativeDependency, len(m.initDeps))
	copy(result, m.initDeps)
	return result, nil
}

func (m *MemStore) CreatePhase(_ context.Context, phase *Phase) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.phases[phase.ID]; exists {
		return fmt.Errorf("phase %s already exists", phase.ID)
	}
	m.phases[phase.ID] = phase
	return nil
}

func (m *MemStore) ListPhases(_ context.Context, initiativeID string) ([]*Phase, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Phase
	for _, p := range m.phases {
		if p.InitiativeID == initiativeID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MemStore) DeletePhase(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.phases[id]; !exists {
		return fmt.Errorf("phase %s not found", id)
	}
	delete(m.phases, id)
	return nil
}

func (m *MemStore) CreateRMI(_ context.Context, rmi *RoadmapItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.rmis[rmi.ID]; exists {
		return fmt.Errorf("RMI %s already exists", rmi.ID)
	}
	m.rmis[rmi.ID] = rmi
	return nil
}

func (m *MemStore) GetRMI(_ context.Context, id string) (*RoadmapItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rmi, ok := m.rmis[id]
	if !ok {
		return nil, fmt.Errorf("RMI %s not found", id)
	}
	return rmi, nil
}

func (m *MemStore) ListRMIs(_ context.Context, initiativeID string) ([]*RoadmapItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*RoadmapItem
	for _, r := range m.rmis {
		if r.InitiativeID == initiativeID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MemStore) ListAllRMIs(_ context.Context) ([]*RoadmapItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*RoadmapItem, 0, len(m.rmis))
	for _, r := range m.rmis {
		result = append(result, r)
	}
	return result, nil
}

func (m *MemStore) ListRMIsByStatus(_ context.Context, status string) ([]*RoadmapItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*RoadmapItem
	for _, r := range m.rmis {
		if r.Status == status {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MemStore) ListRMIsByRepo(_ context.Context, repoID string) ([]*RoadmapItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*RoadmapItem
	for _, r := range m.rmis {
		if r.RepositoryID == repoID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MemStore) UpdateRMI(_ context.Context, rmi *RoadmapItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.rmis[rmi.ID]; !exists {
		return fmt.Errorf("RMI %s not found", rmi.ID)
	}
	m.rmis[rmi.ID] = rmi
	return nil
}

func (m *MemStore) ListAllDependencies(_ context.Context) ([]*RMIDependency, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*RMIDependency, len(m.deps))
	copy(result, m.deps)
	return result, nil
}

func (m *MemStore) CreateDependency(_ context.Context, dep *RMIDependency) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.deps {
		if d.SourceRMIID == dep.SourceRMIID && d.TargetRMIID == dep.TargetRMIID && d.Relationship == dep.Relationship {
			return fmt.Errorf("dependency already exists")
		}
	}
	m.deps = append(m.deps, dep)
	return nil
}

func (m *MemStore) ListDependencies(_ context.Context, rmiID string) ([]*RMIDependency, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*RMIDependency
	for _, d := range m.deps {
		if d.SourceRMIID == rmiID || d.TargetRMIID == rmiID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *MemStore) CreateAssignment(_ context.Context, a *Assignment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.assignments[a.ID]; exists {
		return fmt.Errorf("assignment %s already exists", a.ID)
	}
	m.assignments[a.ID] = a
	return nil
}

func (m *MemStore) GetAssignment(_ context.Context, id string) (*Assignment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.assignments[id]
	if !ok {
		return nil, fmt.Errorf("assignment %s not found", id)
	}
	return a, nil
}

func (m *MemStore) GetActiveAssignment(_ context.Context, rmiID string) (*Assignment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, a := range m.assignments {
		if a.RMIID == rmiID && a.Status == "active" {
			return a, nil
		}
	}
	return nil, nil
}

func (m *MemStore) ListAllAssignments(_ context.Context) ([]*Assignment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Assignment, 0, len(m.assignments))
	for _, a := range m.assignments {
		result = append(result, a)
	}
	return result, nil
}

func (m *MemStore) ListActiveAssignments(_ context.Context) ([]*Assignment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Assignment
	for _, a := range m.assignments {
		if a.Status == "active" {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *MemStore) UpdateAssignment(_ context.Context, a *Assignment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.assignments[a.ID]; !exists {
		return fmt.Errorf("assignment %s not found", a.ID)
	}
	m.assignments[a.ID] = a
	return nil
}

func (m *MemStore) CreateEvidence(_ context.Context, ev *DeliveryEvidence) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.evidence[ev.ID]; exists {
		return fmt.Errorf("evidence %s already exists", ev.ID)
	}
	m.evidence[ev.ID] = ev
	return nil
}

func (m *MemStore) ListEvidenceByRMI(_ context.Context, rmiID string) ([]*DeliveryEvidence, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*DeliveryEvidence
	for _, e := range m.evidence {
		if e.RMIID == rmiID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *MemStore) ListAllEvidence(_ context.Context) ([]*DeliveryEvidence, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*DeliveryEvidence, 0, len(m.evidence))
	for _, e := range m.evidence {
		result = append(result, e)
	}
	return result, nil
}

func (m *MemStore) ListEvidenceByInitiative(ctx context.Context, initiativeID string) ([]*DeliveryEvidence, error) {
	rmis, err := m.ListRMIs(ctx, initiativeID)
	if err != nil {
		return nil, err
	}
	rmiSet := make(map[string]bool, len(rmis))
	for _, r := range rmis {
		rmiSet[r.ID] = true
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*DeliveryEvidence
	for _, e := range m.evidence {
		if rmiSet[e.RMIID] {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *MemStore) CreateRepository(_ context.Context, repo *Repository) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.repositories[repo.ID]; exists {
		return fmt.Errorf("repository %s already exists", repo.ID)
	}
	m.repositories[repo.ID] = repo
	return nil
}

func (m *MemStore) GetRepository(_ context.Context, id string) (*Repository, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	repo, ok := m.repositories[id]
	if !ok {
		return nil, fmt.Errorf("repository %s not found", id)
	}
	return repo, nil
}

func (m *MemStore) ListRepositories(_ context.Context) ([]*Repository, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Repository, 0, len(m.repositories))
	for _, v := range m.repositories {
		result = append(result, v)
	}
	return result, nil
}

func (m *MemStore) UpdateRepository(_ context.Context, repo *Repository) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.repositories[repo.ID]; !exists {
		return fmt.Errorf("repository %s not found", repo.ID)
	}
	m.repositories[repo.ID] = repo
	return nil
}

func (m *MemStore) ListRepositoriesByOrg(_ context.Context, org string) ([]*Repository, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Repository
	for _, v := range m.repositories {
		if v.Organization == org {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *MemStore) CreateRepoDependency(_ context.Context, dep *RepositoryDependency) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.repoDeps {
		if d.SourceRepositoryID == dep.SourceRepositoryID && d.TargetRepositoryID == dep.TargetRepositoryID && d.DependencyType == dep.DependencyType {
			return fmt.Errorf("repository dependency already exists")
		}
	}
	m.repoDeps = append(m.repoDeps, dep)
	return nil
}

func (m *MemStore) ListRepoDependencies(_ context.Context, repoID string) ([]*RepositoryDependency, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*RepositoryDependency
	for _, d := range m.repoDeps {
		if d.SourceRepositoryID == repoID || d.TargetRepositoryID == repoID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *MemStore) ListAllRepoDependencies(_ context.Context) ([]*RepositoryDependency, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*RepositoryDependency, len(m.repoDeps))
	copy(result, m.repoDeps)
	return result, nil
}

func (m *MemStore) CreateSpecWorkflow(_ context.Context, wf *SpecWorkflow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.workflows[wf.ID]; exists {
		return fmt.Errorf("spec workflow %s already exists", wf.ID)
	}
	m.workflows[wf.ID] = wf
	return nil
}

func (m *MemStore) GetSpecWorkflow(_ context.Context, id string) (*SpecWorkflow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	wf, ok := m.workflows[id]
	if !ok {
		return nil, fmt.Errorf("spec workflow %s not found", id)
	}
	return wf, nil
}

func (m *MemStore) ListSpecWorkflows(_ context.Context) ([]*SpecWorkflow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*SpecWorkflow, 0, len(m.workflows))
	for _, v := range m.workflows {
		result = append(result, v)
	}
	return result, nil
}

func (m *MemStore) UpdateSpecWorkflow(_ context.Context, wf *SpecWorkflow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.workflows[wf.ID]; !exists {
		return fmt.Errorf("spec workflow %s not found", wf.ID)
	}
	m.workflows[wf.ID] = wf
	return nil
}

func (m *MemStore) CreateJudgeRubric(_ context.Context, rubric *JudgeRubric) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.judgeRubrics[rubric.ID]; exists {
		return fmt.Errorf("judge rubric %s already exists", rubric.ID)
	}
	m.judgeRubrics[rubric.ID] = rubric
	return nil
}

func (m *MemStore) GetJudgeRubric(_ context.Context, id string) (*JudgeRubric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rubric, ok := m.judgeRubrics[id]
	if !ok {
		return nil, fmt.Errorf("judge rubric %s not found", id)
	}
	return rubric, nil
}

func (m *MemStore) ListJudgeRubrics(_ context.Context, workflowID string) ([]*JudgeRubric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*JudgeRubric
	for _, r := range m.judgeRubrics {
		if workflowID == "" || r.WorkflowID == workflowID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MemStore) CreateJudgeResult(_ context.Context, result *JudgeResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.judgeResults[result.ID]; exists {
		return fmt.Errorf("judge result %s already exists", result.ID)
	}
	m.judgeResults[result.ID] = result
	return nil
}

func (m *MemStore) ListJudgeResults(_ context.Context, initiativeID string) ([]*JudgeResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*JudgeResult
	for _, r := range m.judgeResults {
		if r.InitiativeID == initiativeID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MemStore) CreateCapabilityModel(_ context.Context, model *CapabilityModel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.capabilityModels[model.ID]; exists {
		return fmt.Errorf("capability model %s already exists", model.ID)
	}
	m.capabilityModels[model.ID] = model
	return nil
}

func (m *MemStore) GetCapabilityModel(_ context.Context, id string) (*CapabilityModel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	model, ok := m.capabilityModels[id]
	if !ok {
		return nil, fmt.Errorf("capability model %s not found", id)
	}
	return model, nil
}

func (m *MemStore) ListCapabilityModels(_ context.Context) ([]*CapabilityModel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*CapabilityModel, 0, len(m.capabilityModels))
	for _, v := range m.capabilityModels {
		result = append(result, v)
	}
	return result, nil
}

func (m *MemStore) UpdateCapabilityModel(_ context.Context, model *CapabilityModel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.capabilityModels[model.ID]; !exists {
		return fmt.Errorf("capability model %s not found", model.ID)
	}
	m.capabilityModels[model.ID] = model
	return nil
}

func (m *MemStore) CreateMaturityAssessment(_ context.Context, assessment *MaturityAssessment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.maturityAssessments[assessment.ID]; exists {
		return fmt.Errorf("maturity assessment %s already exists", assessment.ID)
	}
	m.maturityAssessments[assessment.ID] = assessment
	return nil
}

func (m *MemStore) GetMaturityAssessment(_ context.Context, id string) (*MaturityAssessment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	assessment, ok := m.maturityAssessments[id]
	if !ok {
		return nil, fmt.Errorf("maturity assessment %s not found", id)
	}
	return assessment, nil
}

func (m *MemStore) ListMaturityAssessments(_ context.Context, initiativeID string) ([]*MaturityAssessment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*MaturityAssessment
	for _, a := range m.maturityAssessments {
		if a.InitiativeID == initiativeID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *MemStore) ListMaturityAssessmentsByOrg(_ context.Context, org string) ([]*MaturityAssessment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*MaturityAssessment
	for _, a := range m.maturityAssessments {
		if a.Organization == org {
			result = append(result, a)
		}
	}
	return result, nil
}

// MemUnitOfWork is a no-op UnitOfWork for the in-memory store.
type MemUnitOfWork struct {
	Store *MemStore
}

func (u *MemUnitOfWork) Execute(ctx context.Context, fn func(ctx context.Context, s Store) error) error {
	return fn(ctx, u.Store)
}
