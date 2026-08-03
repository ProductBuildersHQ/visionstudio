

package doltstore

import (
	"context"
	"fmt"

	"github.com/ProductBuildersHQ/visionstudio/ent"
	"github.com/ProductBuildersHQ/visionstudio/ent/devxperiodreport"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func entDevXPeriodReportToStore(e *ent.DevXPeriodReport) *store.DevXPeriodReport {
	r := &store.DevXPeriodReport{
		ID:            e.ID,
		Organization:  e.Organization,
		RepositoryID:  e.RepositoryID,
		PersonID:      e.PersonID,
		PeriodType:    e.PeriodType,
		PeriodLabel:   e.PeriodLabel,
		PeriodStart:   e.PeriodStart,
		PeriodEnd:     e.PeriodEnd,
		Metrics:       e.Metrics,
		ByModel:       e.ByModel,
		CoverageScore: e.CoverageScore,
		CreatedAt:     e.CreatedAt,
	}
	return r
}

func (d *DoltStore) CreateDevXPeriodReport(ctx context.Context, report *store.DevXPeriodReport) error {
	b := d.client.DevXPeriodReport.Create().
		SetID(report.ID).
		SetPersonID(report.PersonID).
		SetPeriodType(report.PeriodType).
		SetPeriodLabel(report.PeriodLabel).
		SetPeriodStart(report.PeriodStart).
		SetPeriodEnd(report.PeriodEnd).
		SetCreatedAt(report.CreatedAt)
	if report.Organization != "" {
		b.SetOrganization(report.Organization)
	}
	if report.RepositoryID != "" {
		b.SetRepositoryID(report.RepositoryID)
	}
	if report.Metrics != nil {
		b.SetMetrics(report.Metrics)
	}
	if report.ByModel != nil {
		b.SetByModel(report.ByModel)
	}
	if report.CoverageScore > 0 {
		b.SetCoverageScore(report.CoverageScore)
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("create devx period report: %w", err)
	}
	return nil
}

func (d *DoltStore) GetDevXPeriodReport(ctx context.Context, id string) (*store.DevXPeriodReport, error) {
	e, err := d.client.DevXPeriodReport.Query().
		Where(devxperiodreport.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get devx period report %s: %w", id, err)
	}
	return entDevXPeriodReportToStore(e), nil
}

func (d *DoltStore) ListDevXPeriodReports(ctx context.Context, personID string) ([]*store.DevXPeriodReport, error) {
	rows, err := d.client.DevXPeriodReport.Query().
		Where(devxperiodreport.PersonIDEQ(personID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list devx period reports: %w", err)
	}
	result := make([]*store.DevXPeriodReport, len(rows))
	for i, e := range rows {
		result[i] = entDevXPeriodReportToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListDevXPeriodReportsByRepo(ctx context.Context, repoID string) ([]*store.DevXPeriodReport, error) {
	rows, err := d.client.DevXPeriodReport.Query().
		Where(devxperiodreport.RepositoryIDEQ(repoID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list devx period reports by repo: %w", err)
	}
	result := make([]*store.DevXPeriodReport, len(rows))
	for i, e := range rows {
		result[i] = entDevXPeriodReportToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListDevXPeriodReportsByOrg(ctx context.Context, org string) ([]*store.DevXPeriodReport, error) {
	rows, err := d.client.DevXPeriodReport.Query().
		Where(devxperiodreport.OrganizationEQ(org)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list devx period reports by org: %w", err)
	}
	result := make([]*store.DevXPeriodReport, len(rows))
	for i, e := range rows {
		result[i] = entDevXPeriodReportToStore(e)
	}
	return result, nil
}
