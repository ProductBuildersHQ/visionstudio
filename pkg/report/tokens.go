// Package report computes initiative reports from store data.
// Token reports map token spend to the planning graph (TRD §16).
package report

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
	"github.com/ProductBuildersHQ/visionstudio/pkg/tokens"
)

// TokenReport is the token attribution report for an initiative or period.
type TokenReport struct {
	// Mode is "initiative" or "period"
	Mode string `json:"mode"`

	// Initiative is set when Mode == "initiative"
	InitiativeID    string `json:"initiative_id,omitempty"`
	InitiativeTitle string `json:"initiative_title,omitempty"`

	// Period defines the time range
	Period TokenPeriod `json:"period"`

	// Totals across all attributed events
	Totals TokenTotals `json:"totals"`

	// ByInitiative groups spend by initiative (period mode)
	ByInitiative []InitiativeTokens `json:"by_initiative,omitempty"`

	// ByRMI groups spend by RMI (initiative mode)
	ByRMI []RMITokens `json:"by_rmi,omitempty"`

	// ByModel groups spend by model
	ByModel []ModelTokens `json:"by_model"`

	// Residual is spend attributed to repository but not RMI
	Residual TokenTotals `json:"residual"`

	// Unmanaged is spend from workspaces not in the registry
	Unmanaged TokenTotals `json:"unmanaged"`

	// Coverage is managed spend / total spend
	Coverage float64 `json:"coverage"`

	// SessionCount is the number of unique sessions
	SessionCount int `json:"session_count"`

	// EventCount is the number of token events processed
	EventCount int `json:"event_count"`
}

// TokenPeriod defines the report time range.
type TokenPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Label string    `json:"label,omitempty"` // e.g., "2026-Q3"
}

// TokenTotals aggregates token counts and cost.
type TokenTotals struct {
	InputTokens         int64   `json:"input_tokens"`
	OutputTokens        int64   `json:"output_tokens"`
	CacheReadTokens     int64   `json:"cache_read_tokens"`
	CacheCreationTokens int64   `json:"cache_creation_tokens"`
	TotalTokens         int64   `json:"total_tokens"`
	CostUSD             float64 `json:"cost_usd"`
}

// InitiativeTokens is spend for one initiative.
type InitiativeTokens struct {
	InitiativeID string      `json:"initiative_id"`
	Title        string      `json:"title"`
	Totals       TokenTotals `json:"totals"`
	RMIs         []RMITokens `json:"rmis,omitempty"`
}

// RMITokens is spend for one RMI.
type RMITokens struct {
	RMIID   string      `json:"rmi_id"`
	Title   string      `json:"title"`
	PhaseID string      `json:"phase_id,omitempty"`
	Totals  TokenTotals `json:"totals"`
}

// ModelTokens is spend for one model.
type ModelTokens struct {
	Model  string      `json:"model"`
	Totals TokenTotals `json:"totals"`
}

// GenerateInitiativeTokenReport computes token spend for a single initiative.
func GenerateInitiativeTokenReport(
	ctx context.Context,
	s store.Store,
	source tokens.Source,
	initiativeID string,
) (*TokenReport, error) {
	// Get initiative
	init, err := s.GetInitiative(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("get initiative: %w", err)
	}

	// Get RMIs for this initiative
	rmis, err := s.ListRMIs(ctx, initiativeID)
	if err != nil {
		return nil, fmt.Errorf("list RMIs: %w", err)
	}

	// Determine period from initiative lifecycle
	period := determineInitiativePeriod(init)

	// Build attributor
	attr, err := buildAttributor(ctx, s)
	if err != nil {
		return nil, err
	}

	// Read events
	result, err := source.Read(ctx, tokens.Query{
		Period: tokens.Period{Start: period.Start, End: period.End},
	})
	if err != nil {
		return nil, fmt.Errorf("read events: %w", err)
	}

	// Attribute events
	attributed := attr.AttributeAll(result.Events)

	// Filter to this initiative and build report
	return buildInitiativeReport(init, rmis, attributed, period), nil
}

// GeneratePeriodTokenReport computes token spend for a time period.
func GeneratePeriodTokenReport(
	ctx context.Context,
	s store.Store,
	source tokens.Source,
	period TokenPeriod,
) (*TokenReport, error) {
	// Get all initiatives
	initiatives, err := s.ListInitiatives(ctx)
	if err != nil {
		return nil, fmt.Errorf("list initiatives: %w", err)
	}

	// Get all RMIs
	allRMIs, err := s.ListAllRMIs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all RMIs: %w", err)
	}

	// Build attributor
	attr, err := buildAttributor(ctx, s)
	if err != nil {
		return nil, err
	}

	// Read events
	result, err := source.Read(ctx, tokens.Query{
		Period: tokens.Period{Start: period.Start, End: period.End},
	})
	if err != nil {
		return nil, fmt.Errorf("read events: %w", err)
	}

	// Attribute events
	attributed := attr.AttributeAll(result.Events)

	return buildPeriodReport(initiatives, allRMIs, attributed, period), nil
}

// buildAttributor creates an Attributor from store data.
func buildAttributor(ctx context.Context, s store.Store) (*tokens.Attributor, error) {
	// Load all assignments
	assignments, err := s.ListAllAssignments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list assignments: %w", err)
	}

	// Load all repositories
	repos, err := s.ListRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}

	// Load all RMIs to get phase/initiative mapping
	allRMIs, err := s.ListAllRMIs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all RMIs: %w", err)
	}

	// Load all phases to get initiative mapping
	initiatives, err := s.ListInitiatives(ctx)
	if err != nil {
		return nil, fmt.Errorf("list initiatives: %w", err)
	}

	// Build RMI lookup
	rmiLookup := make(map[string]*store.RoadmapItem)
	for _, rmi := range allRMIs {
		rmiLookup[rmi.ID] = rmi
	}

	// Build initiative lookup for program
	initLookup := make(map[string]*store.Initiative)
	for _, init := range initiatives {
		initLookup[init.ID] = init
	}

	// Convert assignments
	assignmentInfos := make([]tokens.AssignmentInfo, 0, len(assignments))
	for _, a := range assignments {
		rmi := rmiLookup[a.RMIID]
		if rmi == nil {
			continue
		}
		init := initLookup[rmi.InitiativeID]
		programID := ""
		if init != nil {
			programID = init.ProgramID
		}

		info := tokens.AssignmentInfo{
			ID:          a.ID,
			Worker:      a.Worker,
			RMI:         a.RMIID,
			Phase:       rmi.PhaseID,
			Initiative:  rmi.InitiativeID,
			Program:     programID,
			CreatedAt:   a.CreatedAt,
			CompletedAt: a.CompletedAt,
		}
		assignmentInfos = append(assignmentInfos, info)
	}

	// Convert repositories
	repoInfos := make([]tokens.RepositoryInfo, 0, len(repos))
	for _, r := range repos {
		if r.LocalPath == "" {
			continue
		}
		repoInfos = append(repoInfos, tokens.RepositoryInfo{
			ID:        r.ID,
			LocalPath: r.LocalPath,
		})
	}

	return tokens.NewAttributor(assignmentInfos, repoInfos), nil
}

// determineInitiativePeriod returns the period for an initiative's token report.
func determineInitiativePeriod(init *store.Initiative) TokenPeriod {
	start := init.CreatedAt
	if init.ExecutingAt != nil {
		start = *init.ExecutingAt
	}

	end := time.Now()
	if init.ClosedAt != nil {
		end = *init.ClosedAt
	} else if init.DeliveryCompleteAt != nil {
		end = *init.DeliveryCompleteAt
	}

	return TokenPeriod{Start: start, End: end}
}

// buildInitiativeReport constructs the report for initiative mode.
func buildInitiativeReport(
	init *store.Initiative,
	rmis []*store.RoadmapItem,
	events []tokens.AttributedEvent,
	period TokenPeriod,
) *TokenReport {
	r := &TokenReport{
		Mode:            "initiative",
		InitiativeID:    init.ID,
		InitiativeTitle: init.Title,
		Period:          period,
	}

	// Build RMI lookup
	rmiLookup := make(map[string]*store.RoadmapItem)
	for _, rmi := range rmis {
		rmiLookup[rmi.ID] = rmi
	}

	// Aggregate by RMI and model
	byRMI := make(map[string]*TokenTotals)
	byModel := make(map[string]*TokenTotals)
	sessions := make(map[string]bool)

	for _, ev := range events {
		sessions[ev.Event.SessionID] = true
		r.EventCount++

		// Check if this event belongs to this initiative
		if ev.Attribution.Initiative == init.ID {
			// Add to totals
			addToTotals(&r.Totals, ev)

			// Add to RMI breakdown
			if ev.Attribution.RMI != "" {
				if byRMI[ev.Attribution.RMI] == nil {
					byRMI[ev.Attribution.RMI] = &TokenTotals{}
				}
				addToTotals(byRMI[ev.Attribution.RMI], ev)
			}

			// Add to model breakdown
			model := ev.Event.Model
			if model == "" {
				model = "unknown"
			}
			if byModel[model] == nil {
				byModel[model] = &TokenTotals{}
			}
			addToTotals(byModel[model], ev)
		} else if ev.Attribution.Bucket == tokens.BucketRepository {
			addToTotals(&r.Residual, ev)
		} else if ev.Attribution.Bucket == tokens.BucketUnmanaged {
			addToTotals(&r.Unmanaged, ev)
		}
	}

	// Build RMI list
	for rmiID, totals := range byRMI {
		rmi := rmiLookup[rmiID]
		title := rmiID
		phaseID := ""
		if rmi != nil {
			title = rmi.Title
			phaseID = rmi.PhaseID
		}
		r.ByRMI = append(r.ByRMI, RMITokens{
			RMIID:   rmiID,
			Title:   title,
			PhaseID: phaseID,
			Totals:  *totals,
		})
	}

	// Sort by cost descending
	sort.Slice(r.ByRMI, func(i, j int) bool {
		return r.ByRMI[i].Totals.CostUSD > r.ByRMI[j].Totals.CostUSD
	})

	// Build model list
	for model, totals := range byModel {
		r.ByModel = append(r.ByModel, ModelTokens{
			Model:  model,
			Totals: *totals,
		})
	}

	// Sort by cost descending
	sort.Slice(r.ByModel, func(i, j int) bool {
		return r.ByModel[i].Totals.CostUSD > r.ByModel[j].Totals.CostUSD
	})

	r.SessionCount = len(sessions)

	// Calculate coverage
	totalSpend := r.Totals.CostUSD + r.Residual.CostUSD + r.Unmanaged.CostUSD
	if totalSpend > 0 {
		r.Coverage = (r.Totals.CostUSD + r.Residual.CostUSD) / totalSpend
	}

	return r
}

// buildPeriodReport constructs the report for period mode.
func buildPeriodReport(
	initiatives []*store.Initiative,
	allRMIs []*store.RoadmapItem,
	events []tokens.AttributedEvent,
	period TokenPeriod,
) *TokenReport {
	r := &TokenReport{
		Mode:   "period",
		Period: period,
	}

	// Build lookups
	initLookup := make(map[string]*store.Initiative)
	for _, init := range initiatives {
		initLookup[init.ID] = init
	}

	rmiLookup := make(map[string]*store.RoadmapItem)
	for _, rmi := range allRMIs {
		rmiLookup[rmi.ID] = rmi
	}

	// Aggregate by initiative, RMI, and model
	byInit := make(map[string]*InitiativeTokens)
	byModel := make(map[string]*TokenTotals)
	sessions := make(map[string]bool)

	for _, ev := range events {
		sessions[ev.Event.SessionID] = true
		r.EventCount++

		// Add to model breakdown (all events)
		model := ev.Event.Model
		if model == "" {
			model = "unknown"
		}
		if byModel[model] == nil {
			byModel[model] = &TokenTotals{}
		}
		addToTotals(byModel[model], ev)

		switch ev.Attribution.Bucket {
		case tokens.BucketAssigned:
			addToTotals(&r.Totals, ev)

			// Add to initiative
			initID := ev.Attribution.Initiative
			if byInit[initID] == nil {
				init := initLookup[initID]
				title := initID
				if init != nil {
					title = init.Title
				}
				byInit[initID] = &InitiativeTokens{
					InitiativeID: initID,
					Title:        title,
				}
			}
			addToTotals(&byInit[initID].Totals, ev)

		case tokens.BucketRepository:
			addToTotals(&r.Residual, ev)

		case tokens.BucketUnmanaged:
			addToTotals(&r.Unmanaged, ev)
		}
	}

	// Build initiative list
	for _, initTokens := range byInit {
		r.ByInitiative = append(r.ByInitiative, *initTokens)
	}

	// Sort by cost descending
	sort.Slice(r.ByInitiative, func(i, j int) bool {
		return r.ByInitiative[i].Totals.CostUSD > r.ByInitiative[j].Totals.CostUSD
	})

	// Build model list
	for model, totals := range byModel {
		r.ByModel = append(r.ByModel, ModelTokens{
			Model:  model,
			Totals: *totals,
		})
	}

	// Sort by cost descending
	sort.Slice(r.ByModel, func(i, j int) bool {
		return r.ByModel[i].Totals.CostUSD > r.ByModel[j].Totals.CostUSD
	})

	r.SessionCount = len(sessions)

	// Calculate coverage
	totalSpend := r.Totals.CostUSD + r.Residual.CostUSD + r.Unmanaged.CostUSD
	if totalSpend > 0 {
		r.Coverage = (r.Totals.CostUSD + r.Residual.CostUSD) / totalSpend
	}

	return r
}

// addToTotals accumulates an event into totals.
func addToTotals(t *TokenTotals, ev tokens.AttributedEvent) {
	t.InputTokens += ev.Event.InputTokens
	t.OutputTokens += ev.Event.OutputTokens
	t.CacheReadTokens += ev.Event.CacheReadTokens
	t.CacheCreationTokens += ev.Event.CacheCreationTokens
	t.TotalTokens += ev.Event.TotalTokens()
	t.CostUSD += ev.CostUSD
}

// MarkdownTokenReport renders the token report as Markdown.
func (r *TokenReport) MarkdownTokenReport() string {
	var b strings.Builder

	if r.Mode == "initiative" {
		fmt.Fprintf(&b, "# Token Report: %s\n\n", r.InitiativeID)
		fmt.Fprintf(&b, "**Title:** %s\n\n", r.InitiativeTitle)
	} else {
		fmt.Fprintf(&b, "# Token Report: %s\n\n", r.Period.Label)
	}

	fmt.Fprintf(&b, "**Period:** %s to %s\n\n",
		r.Period.Start.Format("2006-01-02"),
		r.Period.End.Format("2006-01-02"))

	fmt.Fprintf(&b, "## Summary\n\n")
	fmt.Fprintf(&b, "| Metric | Value |\n")
	fmt.Fprintf(&b, "|--------|-------|\n")
	fmt.Fprintf(&b, "| Total Cost | $%.2f |\n", r.Totals.CostUSD)
	fmt.Fprintf(&b, "| Total Tokens | %s |\n", formatTokens(r.Totals.TotalTokens))
	fmt.Fprintf(&b, "| Sessions | %d |\n", r.SessionCount)
	fmt.Fprintf(&b, "| Events | %d |\n", r.EventCount)
	fmt.Fprintf(&b, "| Coverage | %.1f%% |\n", r.Coverage*100)

	if r.Mode == "initiative" && len(r.ByRMI) > 0 {
		fmt.Fprintf(&b, "\n## By RMI\n\n")
		fmt.Fprintf(&b, "| RMI | Cost | Tokens |\n")
		fmt.Fprintf(&b, "|-----|------|--------|\n")
		for _, rmi := range r.ByRMI {
			fmt.Fprintf(&b, "| %s | $%.2f | %s |\n",
				rmi.RMIID, rmi.Totals.CostUSD, formatTokens(rmi.Totals.TotalTokens))
		}
	}

	if r.Mode == "period" && len(r.ByInitiative) > 0 {
		fmt.Fprintf(&b, "\n## By Initiative\n\n")
		fmt.Fprintf(&b, "| Initiative | Cost | Tokens |\n")
		fmt.Fprintf(&b, "|------------|------|--------|\n")
		for _, init := range r.ByInitiative {
			fmt.Fprintf(&b, "| %s | $%.2f | %s |\n",
				init.InitiativeID, init.Totals.CostUSD, formatTokens(init.Totals.TotalTokens))
		}
	}

	if len(r.ByModel) > 0 {
		fmt.Fprintf(&b, "\n## By Model\n\n")
		fmt.Fprintf(&b, "| Model | Cost | Tokens |\n")
		fmt.Fprintf(&b, "|-------|------|--------|\n")
		for _, m := range r.ByModel {
			fmt.Fprintf(&b, "| %s | $%.2f | %s |\n",
				m.Model, m.Totals.CostUSD, formatTokens(m.Totals.TotalTokens))
		}
	}

	if r.Residual.TotalTokens > 0 {
		fmt.Fprintf(&b, "\n## Residual (Repository-level)\n\n")
		fmt.Fprintf(&b, "- Cost: $%.2f\n", r.Residual.CostUSD)
		fmt.Fprintf(&b, "- Tokens: %s\n", formatTokens(r.Residual.TotalTokens))
	}

	if r.Unmanaged.TotalTokens > 0 {
		fmt.Fprintf(&b, "\n## Unmanaged\n\n")
		fmt.Fprintf(&b, "- Cost: $%.2f\n", r.Unmanaged.CostUSD)
		fmt.Fprintf(&b, "- Tokens: %s\n", formatTokens(r.Unmanaged.TotalTokens))
	}

	return b.String()
}

// formatTokens formats a token count with K/M suffixes.
func formatTokens(n int64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

// ParseQuarter parses a quarter string (e.g., "2026-Q3") into a period.
func ParseQuarter(q string) (TokenPeriod, error) {
	var year, quarter int
	if _, err := fmt.Sscanf(q, "%d-Q%d", &year, &quarter); err != nil {
		return TokenPeriod{}, fmt.Errorf("invalid quarter format %q (expected YYYY-QN)", q)
	}
	if quarter < 1 || quarter > 4 {
		return TokenPeriod{}, fmt.Errorf("invalid quarter %d (expected 1-4)", quarter)
	}

	startMonth := time.Month((quarter-1)*3 + 1)
	start := time.Date(year, startMonth, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 3, 0).Add(-time.Nanosecond)

	return TokenPeriod{
		Start: start,
		End:   end,
		Label: q,
	}, nil
}
