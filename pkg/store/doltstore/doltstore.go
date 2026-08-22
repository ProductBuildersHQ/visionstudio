// Package doltstore implements the store.Store interface backed by
// Ent and Dolt. Every mutating operation runs inside a unit-of-work
// that wraps an Ent transaction followed by a DOLT_COMMIT.
package doltstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/grokify/godolt"
	"github.com/plexusone/structured-evaluation/rubric"

	"github.com/ProductBuildersHQ/visionstudio/ent"
	"github.com/ProductBuildersHQ/visionstudio/ent/assignment"
	"github.com/ProductBuildersHQ/visionstudio/ent/deliveryevidence"
	initiativeEnt "github.com/ProductBuildersHQ/visionstudio/ent/initiative"
	"github.com/ProductBuildersHQ/visionstudio/ent/initiativedependency"
	"github.com/ProductBuildersHQ/visionstudio/ent/judgeresult"
	"github.com/ProductBuildersHQ/visionstudio/ent/maturityassessment"
	"github.com/ProductBuildersHQ/visionstudio/ent/phase"
	"github.com/ProductBuildersHQ/visionstudio/ent/repository"
	"github.com/ProductBuildersHQ/visionstudio/ent/repositorydependency"
	"github.com/ProductBuildersHQ/visionstudio/ent/rmidependency"
	"github.com/ProductBuildersHQ/visionstudio/ent/roadmapitem"
	"github.com/ProductBuildersHQ/visionstudio/ent/schema"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// DoltStore wraps an Ent client connected to a Dolt MySQL server.
type DoltStore struct {
	client *ent.Client
	db     *sql.DB
	dolt   *godolt.Client
}

// New creates a DoltStore from a MySQL-compatible DSN.
// It ensures parseTime=true is set so time.Time columns scan correctly.
func New(dsn string) (*DoltStore, error) {
	dsn = godolt.EnsureParseTime(dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))
	return &DoltStore{client: client, db: db, dolt: godolt.New(db)}, nil
}

// Close closes the underlying database connection.
func (d *DoltStore) Close() error {
	return d.client.Close()
}

// Ping verifies that the Dolt server is reachable. Because sql.Open is lazy,
// New never dials the server; Ping is the cheapest way to confirm the
// connection is live before issuing real queries.
func (d *DoltStore) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Client returns the Ent client for direct query access.
func (d *DoltStore) Client() *ent.Client {
	return d.client
}

// ExecSQL executes a raw SQL statement against the underlying database.
func (d *DoltStore) ExecSQL(ctx context.Context, query string) error {
	_, err := d.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("exec sql: %w", err)
	}
	return nil
}

// DB returns the underlying *sql.DB for raw queries.
func (d *DoltStore) DB() *sql.DB {
	return d.db
}

// Migrate runs Ent auto-migration against the Dolt database.
func (d *DoltStore) Migrate(ctx context.Context) error {
	return d.client.Schema.Create(ctx)
}

// Commit stages all changes and creates a Dolt commit with the given message.
// This is useful for explicit commits outside of UnitOfWork, or for
// committing accumulated changes from multiple operations.
func (d *DoltStore) Commit(ctx context.Context, message string) error {
	if err := d.dolt.AddAll(ctx); err != nil {
		return fmt.Errorf("dolt add: %w", err)
	}
	if _, err := d.db.ExecContext(ctx, "CALL DOLT_COMMIT('-m', ?, '--allow-empty')", message); err != nil {
		// Ignore "nothing to commit" errors
		if !strings.Contains(err.Error(), "nothing to commit") {
			return fmt.Errorf("dolt commit: %w", err)
		}
	}
	return nil
}

// HasUncommittedChanges returns true if there are uncommitted changes in the working set.
func (d *DoltStore) HasUncommittedChanges(ctx context.Context) (bool, error) {
	dirty, err := d.dolt.HasUncommittedChanges(ctx)
	if err != nil {
		return false, fmt.Errorf("check dolt status: %w", err)
	}
	return dirty, nil
}

// CommitIfDirty commits only if there are uncommitted changes.
func (d *DoltStore) CommitIfDirty(ctx context.Context, message string) (bool, error) {
	dirty, err := d.HasUncommittedChanges(ctx)
	if err != nil {
		return false, err
	}
	if !dirty {
		return false, nil
	}
	if err := d.Commit(ctx, message); err != nil {
		return false, err
	}
	return true, nil
}

// DoltUnitOfWork implements store.UnitOfWork with Ent transactions
// followed by a Dolt commit.
type DoltUnitOfWork struct {
	store *DoltStore
	actor string
}

// NewUnitOfWork creates a unit-of-work that attributes Dolt commits to actor.
func NewUnitOfWork(s *DoltStore, actor string) *DoltUnitOfWork {
	return &DoltUnitOfWork{store: s, actor: actor}
}

// Execute runs fn inside an Ent transaction. On success, it stages
// all changes and issues a DOLT_COMMIT.
func (u *DoltUnitOfWork) Execute(ctx context.Context, fn func(ctx context.Context, s store.Store) error) error {
	tx, err := u.store.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(ctx, u.store); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("rollback failed: %v (original: %w)", rerr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	if _, err := u.store.db.ExecContext(ctx, "CALL DOLT_ADD('.')"); err != nil {
		return fmt.Errorf("dolt add: %w", err)
	}
	msg := fmt.Sprintf("visionstudio: %s", u.actor)
	if _, err := u.store.db.ExecContext(ctx, "CALL DOLT_COMMIT('-m', ?)", msg); err != nil {
		return fmt.Errorf("dolt commit: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Initiative CRUD
// ---------------------------------------------------------------------------

func entInitiativeToStore(e *ent.Initiative) *store.Initiative {
	si := &store.Initiative{
		ID:                 e.ID,
		Organization:       e.Organization,
		Title:              e.Title,
		Description:        e.Description,
		Status:             e.Status,
		InitType:           e.InitType,
		Priority:           e.Priority,
		HomeRepo:           e.HomeRepo,
		Workspace:          e.Workspace,
		Hidden:             e.Hidden,
		Visibility:         e.Visibility,
		Specs:              e.Specs,
		CreatedAt:          e.CreatedAt,
		PlannedAt:          e.PlannedAt,
		ExecutingAt:        e.ExecutingAt,
		DeliveryCompleteAt: e.DeliveryCompleteAt,
		ReleasedAt:         e.ReleasedAt,
		ClosedAt:           e.ClosedAt,
		UpdatedAt:          e.UpdatedAt,
	}
	if prog, err := e.Edges.ProgramOrErr(); err == nil {
		si.ProgramID = prog.ID
	}
	if wf, err := e.Edges.WorkflowOrErr(); err == nil {
		si.WorkflowID = wf.ID
	}
	return si
}

func (d *DoltStore) CreateInitiative(ctx context.Context, init *store.Initiative) error {
	b := d.client.Initiative.Create().
		SetID(init.ID).
		SetOrganization(init.Organization).
		SetTitle(init.Title).
		SetStatus(init.Status).
		SetHidden(init.Hidden).
		SetCreatedAt(init.CreatedAt).
		SetUpdatedAt(init.UpdatedAt)
	if init.Visibility != "" {
		b.SetVisibility(init.Visibility)
	}
	if init.InitType != "" {
		b.SetInitType(init.InitType)
	}
	if init.WorkflowID != "" {
		b.SetWorkflowID(init.WorkflowID)
	}
	if init.Description != "" {
		b.SetDescription(init.Description)
	}
	if init.Priority != "" {
		b.SetPriority(init.Priority)
	}
	if init.HomeRepo != "" {
		b.SetHomeRepo(init.HomeRepo)
	}
	if init.Workspace != "" {
		b.SetWorkspace(init.Workspace)
	}
	if init.ProgramID != "" {
		b.SetProgramID(init.ProgramID)
	}
	if len(init.Specs) > 0 {
		b.SetSpecs(init.Specs)
	}
	b.SetNillablePlannedAt(init.PlannedAt)
	b.SetNillableExecutingAt(init.ExecutingAt)
	b.SetNillableDeliveryCompleteAt(init.DeliveryCompleteAt)
	b.SetNillableReleasedAt(init.ReleasedAt)
	b.SetNillableClosedAt(init.ClosedAt)
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("create initiative: %w", err)
	}
	return nil
}

func (d *DoltStore) GetInitiative(ctx context.Context, id string) (*store.Initiative, error) {
	e, err := d.client.Initiative.Query().
		Where(initiativeEnt.IDEQ(id)).
		WithProgram().
		WithWorkflow().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get initiative %s: %w", id, err)
	}
	return entInitiativeToStore(e), nil
}

func (d *DoltStore) ListInitiatives(ctx context.Context) ([]*store.Initiative, error) {
	rows, err := d.client.Initiative.Query().WithProgram().WithWorkflow().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list initiatives: %w", err)
	}
	result := make([]*store.Initiative, len(rows))
	for i, e := range rows {
		result[i] = entInitiativeToStore(e)
	}
	return result, nil
}

func (d *DoltStore) UpdateInitiative(ctx context.Context, init *store.Initiative) error {
	b := d.client.Initiative.UpdateOneID(init.ID).
		SetOrganization(init.Organization).
		SetTitle(init.Title).
		SetStatus(init.Status).
		SetHidden(init.Hidden).
		SetUpdatedAt(init.UpdatedAt)
	if init.Visibility != "" {
		b.SetVisibility(init.Visibility)
	} else {
		b.SetVisibility("internal")
	}
	if init.InitType != "" {
		b.SetInitType(init.InitType)
	}
	if init.WorkflowID != "" {
		b.SetWorkflowID(init.WorkflowID)
	} else {
		b.ClearWorkflow()
	}
	if init.Description != "" {
		b.SetDescription(init.Description)
	} else {
		b.ClearDescription()
	}
	if init.Priority != "" {
		b.SetPriority(init.Priority)
	} else {
		b.ClearPriority()
	}
	if init.HomeRepo != "" {
		b.SetHomeRepo(init.HomeRepo)
	} else {
		b.ClearHomeRepo()
	}
	if init.Workspace != "" {
		b.SetWorkspace(init.Workspace)
	} else {
		b.ClearWorkspace()
	}
	if init.ProgramID != "" {
		b.SetProgramID(init.ProgramID)
	} else {
		b.ClearProgram()
	}
	if len(init.Specs) > 0 {
		b.SetSpecs(init.Specs)
	} else {
		b.ClearSpecs()
	}
	// SetNillable*(nil) is a no-op in Ent; a backwards lifecycle transition
	// clears stamps, so nil must explicitly clear the column.
	if init.PlannedAt != nil {
		b.SetPlannedAt(*init.PlannedAt)
	} else {
		b.ClearPlannedAt()
	}
	if init.ExecutingAt != nil {
		b.SetExecutingAt(*init.ExecutingAt)
	} else {
		b.ClearExecutingAt()
	}
	if init.DeliveryCompleteAt != nil {
		b.SetDeliveryCompleteAt(*init.DeliveryCompleteAt)
	} else {
		b.ClearDeliveryCompleteAt()
	}
	if init.ReleasedAt != nil {
		b.SetReleasedAt(*init.ReleasedAt)
	} else {
		b.ClearReleasedAt()
	}
	if init.ClosedAt != nil {
		b.SetClosedAt(*init.ClosedAt)
	} else {
		b.ClearClosedAt()
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("update initiative %s: %w", init.ID, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Initiative Dependency CRUD
// ---------------------------------------------------------------------------

func (d *DoltStore) CreateInitiativeDependency(ctx context.Context, dep *store.InitiativeDependency) error {
	_, err := d.client.InitiativeDependency.Create().
		SetSourceInitiativeID(dep.SourceInitiativeID).
		SetTargetInitiativeID(dep.TargetInitiativeID).
		SetRelationship(dep.Relationship).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create initiative dependency: %w", err)
	}
	return nil
}

func (d *DoltStore) ListInitiativeDependencies(ctx context.Context, initiativeID string) ([]*store.InitiativeDependency, error) {
	rows, err := d.client.InitiativeDependency.Query().
		Where(
			initiativedependency.Or(
				initiativedependency.SourceInitiativeID(initiativeID),
				initiativedependency.TargetInitiativeID(initiativeID),
			),
		).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list initiative dependencies: %w", err)
	}
	result := make([]*store.InitiativeDependency, len(rows))
	for i, e := range rows {
		result[i] = &store.InitiativeDependency{
			SourceInitiativeID: e.SourceInitiativeID,
			TargetInitiativeID: e.TargetInitiativeID,
			Relationship:       e.Relationship,
		}
	}
	return result, nil
}

func (d *DoltStore) ListAllInitiativeDependencies(ctx context.Context) ([]*store.InitiativeDependency, error) {
	rows, err := d.client.InitiativeDependency.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all initiative dependencies: %w", err)
	}
	result := make([]*store.InitiativeDependency, len(rows))
	for i, e := range rows {
		result[i] = &store.InitiativeDependency{
			SourceInitiativeID: e.SourceInitiativeID,
			TargetInitiativeID: e.TargetInitiativeID,
			Relationship:       e.Relationship,
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Phase CRUD
// ---------------------------------------------------------------------------

func entPhaseToStore(e *ent.Phase, initiativeID string) *store.Phase {
	return &store.Phase{
		ID:             e.ID,
		InitiativeID:   initiativeID,
		SequenceNumber: e.SequenceNumber,
		Title:          e.Title,
		Theme:          e.Theme,
	}
}

func (d *DoltStore) CreatePhase(ctx context.Context, p *store.Phase) error {
	b := d.client.Phase.Create().
		SetID(p.ID).
		SetSequenceNumber(p.SequenceNumber).
		SetTitle(p.Title).
		SetInitiativeID(p.InitiativeID)
	if p.Theme != "" {
		b.SetTheme(p.Theme)
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("create phase: %w", err)
	}
	return nil
}

func (d *DoltStore) ListPhases(ctx context.Context, initiativeID string) ([]*store.Phase, error) {
	rows, err := d.client.Phase.Query().
		Where(phase.HasInitiativeWith(initiativeEnt.IDEQ(initiativeID))).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list phases for %s: %w", initiativeID, err)
	}
	result := make([]*store.Phase, len(rows))
	for i, e := range rows {
		result[i] = entPhaseToStore(e, initiativeID)
	}
	return result, nil
}

func (d *DoltStore) DeletePhase(ctx context.Context, id string) error {
	if err := d.client.Phase.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete phase %s: %w", id, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// RMI CRUD
// ---------------------------------------------------------------------------

func entRMIToStore(e *ent.RoadmapItem) *store.RoadmapItem {
	r := &store.RoadmapItem{
		ID:                 e.ID,
		Title:              e.Title,
		Description:        e.Description,
		ItemType:           e.ItemType,
		Status:             e.Status,
		Priority:           e.Priority,
		Required:           e.Required,
		Origin:             e.Origin,
		SequenceNumber:     e.SequenceNumber,
		AcceptanceCriteria: e.AcceptanceCriteria,
		CreatedAt:          e.CreatedAt,
		CompletedAt:        e.CompletedAt,
		UpdatedAt:          e.UpdatedAt,
	}
	if init, err := e.Edges.InitiativeOrErr(); err == nil {
		r.InitiativeID = init.ID
	}
	if ph, err := e.Edges.PhaseOrErr(); err == nil {
		r.PhaseID = ph.ID
	}
	if repo, err := e.Edges.RepositoryOrErr(); err == nil {
		r.RepositoryID = repo.ID
	}
	return r
}

func (d *DoltStore) rmiQuery() *ent.RoadmapItemQuery {
	return d.client.RoadmapItem.Query().
		WithInitiative().
		WithPhase().
		WithRepository()
}

func (d *DoltStore) CreateRMI(ctx context.Context, rmi *store.RoadmapItem) error {
	b := d.client.RoadmapItem.Create().
		SetID(rmi.ID).
		SetTitle(rmi.Title).
		SetItemType(rmi.ItemType).
		SetStatus(rmi.Status).
		SetRequired(rmi.Required).
		SetCreatedAt(rmi.CreatedAt).
		SetUpdatedAt(rmi.UpdatedAt).
		SetRepositoryID(rmi.RepositoryID)
	if rmi.Description != "" {
		b.SetDescription(rmi.Description)
	}
	if rmi.Priority != "" {
		b.SetPriority(rmi.Priority)
	}
	// Leave unset (falls through to the schema's Default("spec")) when the
	// caller didn't specify an origin -- an explicit empty string should
	// mean "use the default," not literally clear the column.
	if rmi.Origin != "" {
		b.SetOrigin(rmi.Origin)
	}
	if rmi.SequenceNumber != 0 {
		b.SetSequenceNumber(rmi.SequenceNumber)
	}
	if len(rmi.AcceptanceCriteria) > 0 {
		b.SetAcceptanceCriteria(rmi.AcceptanceCriteria)
	}
	b.SetNillableCompletedAt(rmi.CompletedAt)
	if rmi.InitiativeID != "" {
		b.SetInitiativeID(rmi.InitiativeID)
	}
	if rmi.PhaseID != "" {
		b.SetPhaseID(rmi.PhaseID)
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("create RMI: %w", err)
	}
	return nil
}

func (d *DoltStore) GetRMI(ctx context.Context, id string) (*store.RoadmapItem, error) {
	e, err := d.rmiQuery().Where(roadmapitem.IDEQ(id)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get RMI %s: %w", id, err)
	}
	return entRMIToStore(e), nil
}

func (d *DoltStore) ListAllRMIs(ctx context.Context) ([]*store.RoadmapItem, error) {
	rows, err := d.rmiQuery().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all RMIs: %w", err)
	}
	result := make([]*store.RoadmapItem, len(rows))
	for i, e := range rows {
		result[i] = entRMIToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListRMIs(ctx context.Context, initiativeID string) ([]*store.RoadmapItem, error) {
	rows, err := d.rmiQuery().
		Where(roadmapitem.HasInitiativeWith(initiativeEnt.IDEQ(initiativeID))).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list RMIs for initiative %s: %w", initiativeID, err)
	}
	result := make([]*store.RoadmapItem, len(rows))
	for i, e := range rows {
		result[i] = entRMIToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListRMIsByStatus(ctx context.Context, status string) ([]*store.RoadmapItem, error) {
	rows, err := d.rmiQuery().
		Where(roadmapitem.StatusEQ(status)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list RMIs by status %s: %w", status, err)
	}
	result := make([]*store.RoadmapItem, len(rows))
	for i, e := range rows {
		result[i] = entRMIToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListRMIsByRepo(ctx context.Context, repoID string) ([]*store.RoadmapItem, error) {
	rows, err := d.rmiQuery().
		Where(roadmapitem.HasRepositoryWith(repository.IDEQ(repoID))).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list RMIs for repo %s: %w", repoID, err)
	}
	result := make([]*store.RoadmapItem, len(rows))
	for i, e := range rows {
		result[i] = entRMIToStore(e)
	}
	return result, nil
}

func (d *DoltStore) UpdateRMI(ctx context.Context, rmi *store.RoadmapItem) error {
	b := d.client.RoadmapItem.UpdateOneID(rmi.ID).
		SetTitle(rmi.Title).
		SetItemType(rmi.ItemType).
		SetStatus(rmi.Status).
		SetRequired(rmi.Required).
		SetUpdatedAt(rmi.UpdatedAt).
		SetRepositoryID(rmi.RepositoryID)
	if rmi.Description != "" {
		b.SetDescription(rmi.Description)
	} else {
		b.ClearDescription()
	}
	if rmi.Priority != "" {
		b.SetPriority(rmi.Priority)
	} else {
		b.ClearPriority()
	}
	if rmi.Origin != "" {
		b.SetOrigin(rmi.Origin)
	}
	if rmi.SequenceNumber != 0 {
		b.SetSequenceNumber(rmi.SequenceNumber)
	}
	if len(rmi.AcceptanceCriteria) > 0 {
		b.SetAcceptanceCriteria(rmi.AcceptanceCriteria)
	} else {
		b.ClearAcceptanceCriteria()
	}
	b.SetNillableCompletedAt(rmi.CompletedAt)
	if rmi.InitiativeID != "" {
		b.SetInitiativeID(rmi.InitiativeID)
	} else {
		b.ClearInitiative()
	}
	if rmi.PhaseID != "" {
		b.SetPhaseID(rmi.PhaseID)
	} else {
		b.ClearPhase()
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("update RMI %s: %w", rmi.ID, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// RMI Dependency CRUD
// ---------------------------------------------------------------------------

func (d *DoltStore) CreateDependency(ctx context.Context, dep *store.RMIDependency) error {
	_, err := d.client.RMIDependency.Create().
		SetSourceRmiID(dep.SourceRMIID).
		SetTargetRmiID(dep.TargetRMIID).
		SetRelationship(dep.Relationship).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create RMI dependency: %w", err)
	}
	return nil
}

func (d *DoltStore) ListDependencies(ctx context.Context, rmiID string) ([]*store.RMIDependency, error) {
	rows, err := d.client.RMIDependency.Query().
		Where(
			rmidependency.Or(
				rmidependency.SourceRmiIDEQ(rmiID),
				rmidependency.TargetRmiIDEQ(rmiID),
			),
		).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list RMI dependencies for %s: %w", rmiID, err)
	}
	result := make([]*store.RMIDependency, len(rows))
	for i, r := range rows {
		result[i] = &store.RMIDependency{
			SourceRMIID:  r.SourceRmiID,
			TargetRMIID:  r.TargetRmiID,
			Relationship: r.Relationship,
		}
	}
	return result, nil
}

func (d *DoltStore) ListAllDependencies(ctx context.Context) ([]*store.RMIDependency, error) {
	rows, err := d.client.RMIDependency.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all RMI dependencies: %w", err)
	}
	result := make([]*store.RMIDependency, len(rows))
	for i, r := range rows {
		result[i] = &store.RMIDependency{
			SourceRMIID:  r.SourceRmiID,
			TargetRMIID:  r.TargetRmiID,
			Relationship: r.Relationship,
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Assignment CRUD
// ---------------------------------------------------------------------------

func handoffToMap(h *store.Handoff) map[string]any {
	if h == nil {
		return nil
	}
	b, err := json.Marshal(h)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}

func mapToHandoff(m map[string]any) *store.Handoff {
	if m == nil {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	var h store.Handoff
	if err := json.Unmarshal(b, &h); err != nil {
		return nil
	}
	return &h
}

func entAssignmentToStore(e *ent.Assignment) *store.Assignment {
	a := &store.Assignment{
		ID:             e.ID,
		Worker:         e.Worker,
		Status:         e.Status,
		LeaseExpiresAt: e.LeaseExpiresAt,
		Workspace:      e.Workspace,
		Handoff:        mapToHandoff(e.Handoff),
		CreatedAt:      e.CreatedAt,
		CompletedAt:    e.CompletedAt,
		UpdatedAt:      e.UpdatedAt,
	}
	if rmi, err := e.Edges.RoadmapItemOrErr(); err == nil {
		a.RMIID = rmi.ID
	}
	return a
}

func (d *DoltStore) assignmentQuery() *ent.AssignmentQuery {
	return d.client.Assignment.Query().WithRoadmapItem()
}

func (d *DoltStore) CreateAssignment(ctx context.Context, a *store.Assignment) error {
	b := d.client.Assignment.Create().
		SetID(a.ID).
		SetWorker(a.Worker).
		SetStatus(a.Status).
		SetLeaseExpiresAt(a.LeaseExpiresAt).
		SetNillableCompletedAt(a.CompletedAt).
		SetCreatedAt(a.CreatedAt).
		SetUpdatedAt(a.UpdatedAt).
		SetRoadmapItemID(a.RMIID)
	if a.Workspace != "" {
		b.SetWorkspace(a.Workspace)
	}
	if h := handoffToMap(a.Handoff); h != nil {
		b.SetHandoff(h)
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("create assignment: %w", err)
	}
	return nil
}

func (d *DoltStore) GetAssignment(ctx context.Context, id string) (*store.Assignment, error) {
	e, err := d.assignmentQuery().Where(assignment.IDEQ(id)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get assignment %s: %w", id, err)
	}
	return entAssignmentToStore(e), nil
}

func (d *DoltStore) GetActiveAssignment(ctx context.Context, rmiID string) (*store.Assignment, error) {
	e, err := d.assignmentQuery().
		Where(
			assignment.StatusEQ("active"),
			assignment.HasRoadmapItemWith(roadmapitem.IDEQ(rmiID)),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get active assignment for %s: %w", rmiID, err)
	}
	return entAssignmentToStore(e), nil
}

func (d *DoltStore) ListActiveAssignments(ctx context.Context) ([]*store.Assignment, error) {
	rows, err := d.assignmentQuery().
		Where(assignment.StatusEQ("active")).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active assignments: %w", err)
	}
	result := make([]*store.Assignment, len(rows))
	for i, e := range rows {
		result[i] = entAssignmentToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListAllAssignments(ctx context.Context) ([]*store.Assignment, error) {
	rows, err := d.assignmentQuery().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all assignments: %w", err)
	}
	result := make([]*store.Assignment, len(rows))
	for i, e := range rows {
		result[i] = entAssignmentToStore(e)
	}
	return result, nil
}

func (d *DoltStore) UpdateAssignment(ctx context.Context, a *store.Assignment) error {
	b := d.client.Assignment.UpdateOneID(a.ID).
		SetWorker(a.Worker).
		SetStatus(a.Status).
		SetLeaseExpiresAt(a.LeaseExpiresAt).
		SetNillableCompletedAt(a.CompletedAt).
		SetUpdatedAt(a.UpdatedAt).
		SetRoadmapItemID(a.RMIID)
	if a.Workspace != "" {
		b.SetWorkspace(a.Workspace)
	} else {
		b.ClearWorkspace()
	}
	if h := handoffToMap(a.Handoff); h != nil {
		b.SetHandoff(h)
	} else {
		b.ClearHandoff()
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("update assignment %s: %w", a.ID, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Evidence CRUD
// ---------------------------------------------------------------------------

func entEvidenceToStore(e *ent.DeliveryEvidence) *store.DeliveryEvidence {
	ev := &store.DeliveryEvidence{
		ID:           e.ID,
		EvidenceType: e.EvidenceType,
		Reference:    e.Reference,
		CommitType:   e.CommitType,
		CommitScope:  e.CommitScope,
		OccurredAt:   e.OccurredAt,
		CreatedAt:    e.CreatedAt,
	}
	if rmi, err := e.Edges.RoadmapItemOrErr(); err == nil {
		ev.RMIID = rmi.ID
	}
	return ev
}

func (d *DoltStore) evidenceQuery() *ent.DeliveryEvidenceQuery {
	return d.client.DeliveryEvidence.Query().WithRoadmapItem()
}

func (d *DoltStore) CreateEvidence(ctx context.Context, ev *store.DeliveryEvidence) error {
	b := d.client.DeliveryEvidence.Create().
		SetID(ev.ID).
		SetEvidenceType(ev.EvidenceType).
		SetReference(ev.Reference).
		SetCreatedAt(ev.CreatedAt).
		SetRoadmapItemID(ev.RMIID)
	if ev.CommitType != "" {
		b.SetCommitType(ev.CommitType)
	}
	if ev.CommitScope != "" {
		b.SetCommitScope(ev.CommitScope)
	}
	b.SetNillableOccurredAt(ev.OccurredAt)
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("create evidence: %w", err)
	}
	return nil
}

func (d *DoltStore) ListEvidenceByRMI(ctx context.Context, rmiID string) ([]*store.DeliveryEvidence, error) {
	rows, err := d.evidenceQuery().
		Where(deliveryevidence.HasRoadmapItemWith(roadmapitem.IDEQ(rmiID))).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list evidence for RMI %s: %w", rmiID, err)
	}
	result := make([]*store.DeliveryEvidence, len(rows))
	for i, e := range rows {
		result[i] = entEvidenceToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListEvidenceByInitiative(ctx context.Context, initiativeID string) ([]*store.DeliveryEvidence, error) {
	rows, err := d.evidenceQuery().
		Where(deliveryevidence.HasRoadmapItemWith(
			roadmapitem.HasInitiativeWith(initiativeEnt.IDEQ(initiativeID)),
		)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list evidence for initiative %s: %w", initiativeID, err)
	}
	result := make([]*store.DeliveryEvidence, len(rows))
	for i, e := range rows {
		result[i] = entEvidenceToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListAllEvidence(ctx context.Context) ([]*store.DeliveryEvidence, error) {
	rows, err := d.evidenceQuery().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all evidence: %w", err)
	}
	result := make([]*store.DeliveryEvidence, len(rows))
	for i, e := range rows {
		result[i] = entEvidenceToStore(e)
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Program CRUD
// ---------------------------------------------------------------------------

func entProgramToStore(e *ent.Program) *store.Program {
	return &store.Program{
		ID:           e.ID,
		Name:         e.Name,
		Organization: e.Organization,
		Description:  e.Description,
		Hidden:       e.Hidden,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func (d *DoltStore) CreateProgram(ctx context.Context, prog *store.Program) error {
	b := d.client.Program.Create().
		SetID(prog.ID).
		SetName(prog.Name).
		SetOrganization(prog.Organization).
		SetHidden(prog.Hidden).
		SetCreatedAt(prog.CreatedAt).
		SetUpdatedAt(prog.UpdatedAt)
	if prog.Description != "" {
		b.SetDescription(prog.Description)
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("create program: %w", err)
	}
	return nil
}

func (d *DoltStore) GetProgram(ctx context.Context, id string) (*store.Program, error) {
	e, err := d.client.Program.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get program %s: %w", id, err)
	}
	return entProgramToStore(e), nil
}

func (d *DoltStore) ListPrograms(ctx context.Context) ([]*store.Program, error) {
	rows, err := d.client.Program.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list programs: %w", err)
	}
	result := make([]*store.Program, len(rows))
	for i, e := range rows {
		result[i] = entProgramToStore(e)
	}
	return result, nil
}

func (d *DoltStore) UpdateProgram(ctx context.Context, prog *store.Program) error {
	b := d.client.Program.UpdateOneID(prog.ID).
		SetName(prog.Name).
		SetOrganization(prog.Organization).
		SetHidden(prog.Hidden).
		SetUpdatedAt(prog.UpdatedAt)
	if prog.Description != "" {
		b.SetDescription(prog.Description)
	} else {
		b.ClearDescription()
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("update program %s: %w", prog.ID, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Repository CRUD
// ---------------------------------------------------------------------------

func entRepoToStore(r *ent.Repository) *store.Repository {
	return &store.Repository{
		ID:              r.ID,
		Organization:    r.Organization,
		RepositoryName:  r.RepositoryName,
		DefaultBranch:   r.DefaultBranch,
		LocalPath:       r.LocalPath,
		GoModule:        r.GoModule,
		Domain:          r.Domain,
		Status:          r.Status,
		IngestHighWater: r.IngestHighWater,
		OrganizationID:  r.OrganizationID,
		Visibility:      r.Visibility,
		SupersededBy:    r.SupersededBy,
	}
}

func (d *DoltStore) CreateRepository(ctx context.Context, repo *store.Repository) error {
	b := d.client.Repository.Create().
		SetID(repo.ID).
		SetOrganization(repo.Organization).
		SetRepositoryName(repo.RepositoryName).
		SetDefaultBranch(repo.DefaultBranch).
		SetStatus(repo.Status)
	if repo.LocalPath != "" {
		b.SetLocalPath(repo.LocalPath)
	}
	if repo.GoModule != "" {
		b.SetGoModule(repo.GoModule)
	}
	if repo.Domain != "" {
		b.SetDomain(repo.Domain)
	}
	if repo.IngestHighWater != "" {
		b.SetIngestHighWater(repo.IngestHighWater)
	}
	if repo.OrganizationID != "" {
		b.SetOrganizationID(repo.OrganizationID)
	}
	if repo.Visibility != "" {
		b.SetVisibility(repo.Visibility)
	}
	if repo.SupersededBy != "" {
		b.SetSupersededBy(repo.SupersededBy)
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("create repository: %w", err)
	}
	return nil
}

func (d *DoltStore) GetRepository(ctx context.Context, id string) (*store.Repository, error) {
	r, err := d.client.Repository.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get repository %s: %w", id, err)
	}
	return entRepoToStore(r), nil
}

func (d *DoltStore) ListRepositories(ctx context.Context) ([]*store.Repository, error) {
	rows, err := d.client.Repository.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}
	result := make([]*store.Repository, len(rows))
	for i, r := range rows {
		result[i] = entRepoToStore(r)
	}
	return result, nil
}

func (d *DoltStore) ListRepositoriesByOrg(ctx context.Context, org string) ([]*store.Repository, error) {
	rows, err := d.client.Repository.Query().
		Where(repository.OrganizationEQ(org)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repositories by org %s: %w", org, err)
	}
	result := make([]*store.Repository, len(rows))
	for i, r := range rows {
		result[i] = entRepoToStore(r)
	}
	return result, nil
}

func (d *DoltStore) UpdateRepository(ctx context.Context, repo *store.Repository) error {
	b := d.client.Repository.UpdateOneID(repo.ID).
		SetOrganization(repo.Organization).
		SetRepositoryName(repo.RepositoryName).
		SetDefaultBranch(repo.DefaultBranch).
		SetStatus(repo.Status)
	if repo.LocalPath != "" {
		b.SetLocalPath(repo.LocalPath)
	} else {
		b.ClearLocalPath()
	}
	if repo.GoModule != "" {
		b.SetGoModule(repo.GoModule)
	} else {
		b.ClearGoModule()
	}
	if repo.Domain != "" {
		b.SetDomain(repo.Domain)
	} else {
		b.ClearDomain()
	}
	if repo.IngestHighWater != "" {
		b.SetIngestHighWater(repo.IngestHighWater)
	} else {
		b.ClearIngestHighWater()
	}
	if repo.OrganizationID != "" {
		b.SetOrganizationID(repo.OrganizationID)
	} else {
		b.ClearOrganizationID()
	}
	if repo.Visibility != "" {
		b.SetVisibility(repo.Visibility)
	} else {
		b.SetVisibility("unknown")
	}
	if repo.SupersededBy != "" {
		b.SetSupersededBy(repo.SupersededBy)
	} else {
		b.ClearSupersededBy()
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("update repository %s: %w", repo.ID, err)
	}
	return nil
}

func (d *DoltStore) DeleteRepository(ctx context.Context, id string) error {
	if err := d.client.Repository.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete repository %s: %w", id, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Repository Dependency CRUD
// ---------------------------------------------------------------------------

func (d *DoltStore) CreateRepoDependency(ctx context.Context, dep *store.RepositoryDependency) error {
	_, err := d.client.RepositoryDependency.Create().
		SetSourceRepositoryID(dep.SourceRepositoryID).
		SetTargetRepositoryID(dep.TargetRepositoryID).
		SetDependencyType(dep.DependencyType).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create repo dependency: %w", err)
	}
	return nil
}

func (d *DoltStore) ListRepoDependencies(ctx context.Context, repoID string) ([]*store.RepositoryDependency, error) {
	rows, err := d.client.RepositoryDependency.Query().
		Where(
			repositorydependency.Or(
				repositorydependency.SourceRepositoryIDEQ(repoID),
				repositorydependency.TargetRepositoryIDEQ(repoID),
			),
		).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repo dependencies for %s: %w", repoID, err)
	}
	result := make([]*store.RepositoryDependency, len(rows))
	for i, r := range rows {
		result[i] = &store.RepositoryDependency{
			SourceRepositoryID: r.SourceRepositoryID,
			TargetRepositoryID: r.TargetRepositoryID,
			DependencyType:     r.DependencyType,
		}
	}
	return result, nil
}

func (d *DoltStore) ListAllRepoDependencies(ctx context.Context) ([]*store.RepositoryDependency, error) {
	rows, err := d.client.RepositoryDependency.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all repo dependencies: %w", err)
	}
	result := make([]*store.RepositoryDependency, len(rows))
	for i, r := range rows {
		result[i] = &store.RepositoryDependency{
			SourceRepositoryID: r.SourceRepositoryID,
			TargetRepositoryID: r.TargetRepositoryID,
			DependencyType:     r.DependencyType,
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// SpecWorkflow CRUD
// ---------------------------------------------------------------------------

func (d *DoltStore) CreateSpecWorkflow(ctx context.Context, wf *store.SpecWorkflow) error {
	_, err := d.client.SpecWorkflow.Create().
		SetID(wf.ID).
		SetName(wf.Name).
		SetDescription(wf.Description).
		SetSpecsRequired(wf.SpecsRequired).
		SetSpecsOptional(wf.SpecsOptional).
		SetInitTypes(wf.InitTypes).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create spec workflow: %w", err)
	}
	return nil
}

func (d *DoltStore) GetSpecWorkflow(ctx context.Context, id string) (*store.SpecWorkflow, error) {
	row, err := d.client.SpecWorkflow.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get spec workflow %s: %w", id, err)
	}
	return &store.SpecWorkflow{
		ID:            row.ID,
		Name:          row.Name,
		Description:   row.Description,
		SpecsRequired: row.SpecsRequired,
		SpecsOptional: row.SpecsOptional,
		InitTypes:     row.InitTypes,
	}, nil
}

func (d *DoltStore) ListSpecWorkflows(ctx context.Context) ([]*store.SpecWorkflow, error) {
	rows, err := d.client.SpecWorkflow.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list spec workflows: %w", err)
	}
	result := make([]*store.SpecWorkflow, len(rows))
	for i, r := range rows {
		result[i] = &store.SpecWorkflow{
			ID:            r.ID,
			Name:          r.Name,
			Description:   r.Description,
			SpecsRequired: r.SpecsRequired,
			SpecsOptional: r.SpecsOptional,
			InitTypes:     r.InitTypes,
		}
	}
	return result, nil
}

func (d *DoltStore) UpdateSpecWorkflow(ctx context.Context, wf *store.SpecWorkflow) error {
	_, err := d.client.SpecWorkflow.UpdateOneID(wf.ID).
		SetName(wf.Name).
		SetDescription(wf.Description).
		SetSpecsRequired(wf.SpecsRequired).
		SetSpecsOptional(wf.SpecsOptional).
		SetInitTypes(wf.InitTypes).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update spec workflow %s: %w", wf.ID, err)
	}
	return nil
}

func (d *DoltStore) DeleteSpecWorkflow(ctx context.Context, id string) error {
	if err := d.client.SpecWorkflow.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete spec workflow %s: %w", id, err)
	}
	return nil
}

func (d *DoltStore) SelectWorkflowForInitiative(ctx context.Context, initiativeID, workflowID string) error {
	now := time.Now()
	_, err := d.client.InitiativeWorkflow.Get(ctx, initiativeID)
	if ent.IsNotFound(err) {
		_, err = d.client.InitiativeWorkflow.Create().
			SetID(initiativeID).
			SetWorkflowID(workflowID).
			SetSelectedAt(now).
			Save(ctx)
	} else if err == nil {
		_, err = d.client.InitiativeWorkflow.UpdateOneID(initiativeID).
			SetWorkflowID(workflowID).
			SetSelectedAt(now).
			Save(ctx)
	}
	if err != nil {
		return fmt.Errorf("select workflow for initiative %s: %w", initiativeID, err)
	}
	return nil
}

func (d *DoltStore) GetWorkflowForInitiative(ctx context.Context, initiativeID string) (*store.InitiativeWorkflow, error) {
	row, err := d.client.InitiativeWorkflow.Get(ctx, initiativeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get workflow for initiative %s: %w", initiativeID, err)
	}
	return &store.InitiativeWorkflow{
		InitiativeID: row.ID,
		WorkflowID:   row.WorkflowID,
		SelectedAt:   row.SelectedAt,
	}, nil
}

// ---------------------------------------------------------------------------
// JudgeResult CRUD
// ---------------------------------------------------------------------------

func (d *DoltStore) CreateJudgeResult(ctx context.Context, result *store.JudgeResult) error {
	builder := d.client.JudgeResult.Create().
		SetID(result.ID).
		SetInitiativeID(result.InitiativeID).
		SetSpecPath(result.SpecPath).
		SetEvaluatedAt(result.EvaluatedAt)

	if result.SpecType != "" {
		builder = builder.SetSpecType(result.SpecType)
	}
	if result.RubricID != "" {
		builder = builder.SetRubricID(result.RubricID)
	}

	// Store the full report as JSON and extract indexed fields
	if result.Report != nil {
		reportMap := rubricToMap(result.Report)
		builder = builder.SetReport(reportMap)
		builder = builder.SetIntScore(int(result.Report.IntScore))
		builder = builder.SetPass(result.Report.Pass)
		if result.Report.Judge != nil {
			builder = builder.SetModel(result.Report.Judge.Model)
		}
	}

	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create judge result: %w", err)
	}
	return nil
}

func (d *DoltStore) ListJudgeResults(ctx context.Context, initiativeID string) ([]*store.JudgeResult, error) {
	rows, err := d.client.JudgeResult.Query().
		Where(judgeresult.InitiativeIDEQ(initiativeID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list judge results for %s: %w", initiativeID, err)
	}
	result := make([]*store.JudgeResult, len(rows))
	for i, r := range rows {
		var report *rubric.Rubric
		if r.Report != nil {
			report = mapToRubric(r.Report)
		}

		result[i] = &store.JudgeResult{
			ID:           r.ID,
			InitiativeID: r.InitiativeID,
			SpecPath:     r.SpecPath,
			SpecType:     r.SpecType,
			RubricID:     r.RubricID,
			EvaluatedAt:  r.EvaluatedAt,
			Report:       report,
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// CapabilityModel CRUD
// ---------------------------------------------------------------------------

func entDimensionsToStore(dims []schema.Dimension) []store.Dimension {
	result := make([]store.Dimension, len(dims))
	for i, d := range dims {
		levels := make([]store.Level, len(d.Levels))
		for j, l := range d.Levels {
			levels[j] = store.Level{
				Level:       l.Level,
				Name:        l.Name,
				Description: l.Description,
			}
		}
		result[i] = store.Dimension{
			Key:         d.Key,
			Name:        d.Name,
			Description: d.Description,
			Levels:      levels,
			Sources:     d.Sources,
		}
	}
	return result
}

func storeDimensionsToEnt(dims []store.Dimension) []schema.Dimension {
	result := make([]schema.Dimension, len(dims))
	for i, d := range dims {
		levels := make([]schema.Level, len(d.Levels))
		for j, l := range d.Levels {
			levels[j] = schema.Level{
				Level:       l.Level,
				Name:        l.Name,
				Description: l.Description,
			}
		}
		result[i] = schema.Dimension{
			Key:         d.Key,
			Name:        d.Name,
			Description: d.Description,
			Levels:      levels,
			Sources:     d.Sources,
		}
	}
	return result
}

func (d *DoltStore) CreateCapabilityModel(ctx context.Context, model *store.CapabilityModel) error {
	_, err := d.client.CapabilityModel.Create().
		SetID(model.ID).
		SetName(model.Name).
		SetDescription(model.Description).
		SetDimensions(storeDimensionsToEnt(model.Dimensions)).
		SetMaxLevel(model.MaxLevel).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create capability model: %w", err)
	}
	return nil
}

func (d *DoltStore) GetCapabilityModel(ctx context.Context, id string) (*store.CapabilityModel, error) {
	row, err := d.client.CapabilityModel.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get capability model %s: %w", id, err)
	}
	return &store.CapabilityModel{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Dimensions:  entDimensionsToStore(row.Dimensions),
		MaxLevel:    row.MaxLevel,
	}, nil
}

func (d *DoltStore) ListCapabilityModels(ctx context.Context) ([]*store.CapabilityModel, error) {
	rows, err := d.client.CapabilityModel.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list capability models: %w", err)
	}
	result := make([]*store.CapabilityModel, len(rows))
	for i, r := range rows {
		result[i] = &store.CapabilityModel{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Dimensions:  entDimensionsToStore(r.Dimensions),
			MaxLevel:    r.MaxLevel,
		}
	}
	return result, nil
}

func (d *DoltStore) UpdateCapabilityModel(ctx context.Context, model *store.CapabilityModel) error {
	_, err := d.client.CapabilityModel.UpdateOneID(model.ID).
		SetName(model.Name).
		SetDescription(model.Description).
		SetDimensions(storeDimensionsToEnt(model.Dimensions)).
		SetMaxLevel(model.MaxLevel).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update capability model %s: %w", model.ID, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// MaturityAssessment CRUD
// ---------------------------------------------------------------------------

func entScoresToStore(scores map[string]schema.DimensionScore) map[string]store.DimensionScore {
	result := make(map[string]store.DimensionScore, len(scores))
	for k, v := range scores {
		result[k] = store.DimensionScore{
			Level:     v.Level,
			Rationale: v.Rationale,
			Evidence:  v.Evidence,
		}
	}
	return result
}

func storeScoresToEnt(scores map[string]store.DimensionScore) map[string]schema.DimensionScore {
	result := make(map[string]schema.DimensionScore, len(scores))
	for k, v := range scores {
		result[k] = schema.DimensionScore{
			Level:     v.Level,
			Rationale: v.Rationale,
			Evidence:  v.Evidence,
		}
	}
	return result
}

func (d *DoltStore) CreateMaturityAssessment(ctx context.Context, a *store.MaturityAssessment) error {
	builder := d.client.MaturityAssessment.Create().
		SetID(a.ID).
		SetScores(storeScoresToEnt(a.Scores)).
		SetSummary(a.Summary).
		SetAssessedAt(a.AssessedAt)
	if a.InitiativeID != "" {
		builder = builder.SetInitiativeID(a.InitiativeID)
	}
	if a.Organization != "" {
		builder = builder.SetOrganization(a.Organization)
	}
	if a.OverallScore != nil {
		builder = builder.SetOverallScore(*a.OverallScore)
	}
	if a.AssessedBy != "" {
		builder = builder.SetAssessedBy(a.AssessedBy)
	}
	if a.Model != "" {
		builder = builder.SetModel(a.Model)
	}
	if a.ModelID != "" {
		builder = builder.SetCapabilityModelID(a.ModelID)
	}
	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create maturity assessment: %w", err)
	}
	return nil
}

func (d *DoltStore) GetMaturityAssessment(ctx context.Context, id string) (*store.MaturityAssessment, error) {
	row, err := d.client.MaturityAssessment.Query().
		Where(maturityassessment.IDEQ(id)).
		WithCapabilityModel().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get maturity assessment %s: %w", id, err)
	}
	modelID := ""
	if cm, err := row.Edges.CapabilityModelOrErr(); err == nil {
		modelID = cm.ID
	}
	return &store.MaturityAssessment{
		ID:           row.ID,
		ModelID:      modelID,
		InitiativeID: row.InitiativeID,
		Organization: row.Organization,
		Scores:       entScoresToStore(row.Scores),
		OverallScore: row.OverallScore,
		Summary:      row.Summary,
		AssessedBy:   row.AssessedBy,
		Model:        row.Model,
		AssessedAt:   row.AssessedAt,
	}, nil
}

func (d *DoltStore) ListMaturityAssessments(ctx context.Context, initiativeID string) ([]*store.MaturityAssessment, error) {
	rows, err := d.client.MaturityAssessment.Query().
		Where(maturityassessment.InitiativeIDEQ(initiativeID)).
		WithCapabilityModel().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list maturity assessments for %s: %w", initiativeID, err)
	}
	return mapMaturityAssessments(rows), nil
}

func (d *DoltStore) ListMaturityAssessmentsByOrg(ctx context.Context, org string) ([]*store.MaturityAssessment, error) {
	rows, err := d.client.MaturityAssessment.Query().
		Where(maturityassessment.OrganizationEQ(org)).
		WithCapabilityModel().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list maturity assessments for org %s: %w", org, err)
	}
	return mapMaturityAssessments(rows), nil
}

// mapMaturityAssessments converts ent rows (with the capability-model edge
// loaded) into store.MaturityAssessment values.
func mapMaturityAssessments(rows []*ent.MaturityAssessment) []*store.MaturityAssessment {
	result := make([]*store.MaturityAssessment, len(rows))
	for i, r := range rows {
		modelID := ""
		if cm, err := r.Edges.CapabilityModelOrErr(); err == nil {
			modelID = cm.ID
		}
		result[i] = &store.MaturityAssessment{
			ID:           r.ID,
			ModelID:      modelID,
			InitiativeID: r.InitiativeID,
			Organization: r.Organization,
			Scores:       entScoresToStore(r.Scores),
			OverallScore: r.OverallScore,
			Summary:      r.Summary,
			AssessedBy:   r.AssessedBy,
			Model:        r.Model,
			AssessedAt:   r.AssessedAt,
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Rubric conversion helpers
// ---------------------------------------------------------------------------

// rubricToMap converts a rubric.Rubric to a map for JSON storage.
func rubricToMap(r *rubric.Rubric) map[string]any {
	if r == nil {
		return nil
	}
	data, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	return m
}

// mapToRubric converts a stored map back to a rubric.Rubric.
func mapToRubric(m map[string]any) *rubric.Rubric {
	if m == nil {
		return nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	var r rubric.Rubric
	if err := json.Unmarshal(data, &r); err != nil {
		return nil
	}
	return &r
}
