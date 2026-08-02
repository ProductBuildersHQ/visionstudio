package tokens

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/plexusone/omnidevx-core/report"
)

// Attribution bucket constants.
const (
	BucketAssigned   = "assigned"   // Attributed via session → assignment
	BucketRepository = "repository" // Attributed via workspace → repository (fallback)
	BucketUnmanaged  = "unmanaged"  // No attribution (coverage gap)
)

// AttributionResult describes how a token event was attributed.
type AttributionResult struct {
	// Bucket indicates the attribution method used.
	Bucket string

	// Planning graph attribution (only set for BucketAssigned).
	RMI        string
	Phase      string
	Initiative string
	Program    string

	// Repository attribution (set for BucketAssigned and BucketRepository).
	Repository string

	// AssignmentID is the matching assignment (only set for BucketAssigned).
	AssignmentID string
}

// AttributedEvent pairs a token event with its attribution and cost.
type AttributedEvent struct {
	Event       Event
	Attribution AttributionResult
	CostUSD     float64
}

// AssignmentInfo holds the subset of assignment data needed for attribution.
// This is designed to be populated from Ent queries.
type AssignmentInfo struct {
	ID          string
	Worker      string // Session ID (Claude Code UUID)
	RMI         string
	Phase       string
	Initiative  string
	Program     string
	Repository  string
	CreatedAt   time.Time
	CompletedAt *time.Time // nil for active/open-ended assignments
}

// RepositoryInfo holds the subset of repository data needed for attribution.
// This is designed to be populated from Ent queries.
type RepositoryInfo struct {
	ID        string
	LocalPath string // Absolute path on disk
}

// Attributor performs token event attribution using assignment and repository data.
type Attributor struct {
	// assignmentsByWorker maps session ID → list of assignments (may have multiple
	// if the same worker was assigned different RMIs over time).
	assignmentsByWorker map[string][]AssignmentInfo

	// repositoriesByPath maps normalized local paths → repository info.
	// Paths are stored with trailing slashes for prefix matching.
	repositoriesByPath map[string]RepositoryInfo
}

// NewAttributor creates an Attributor from assignment and repository data.
func NewAttributor(assignments []AssignmentInfo, repositories []RepositoryInfo) *Attributor {
	a := &Attributor{
		assignmentsByWorker: make(map[string][]AssignmentInfo),
		repositoriesByPath:  make(map[string]RepositoryInfo),
	}

	for _, asn := range assignments {
		a.assignmentsByWorker[asn.Worker] = append(a.assignmentsByWorker[asn.Worker], asn)
	}

	for _, repo := range repositories {
		if repo.LocalPath != "" {
			// Normalize: clean path and ensure trailing slash for prefix matching
			path := filepath.Clean(repo.LocalPath)
			if !strings.HasSuffix(path, string(filepath.Separator)) {
				path += string(filepath.Separator)
			}
			a.repositoriesByPath[path] = repo
		}
	}

	return a
}

// Attribute determines the attribution for a token event and computes its cost.
func (a *Attributor) Attribute(event Event) AttributedEvent {
	result := AttributedEvent{
		Event:   event,
		CostUSD: a.computeCost(event),
	}

	// Try session-based attribution first (highest precedence).
	if attr, ok := a.attributeBySession(event); ok {
		result.Attribution = attr
		return result
	}

	// Fall back to repository-based attribution.
	if attr, ok := a.attributeByRepository(event); ok {
		result.Attribution = attr
		return result
	}

	// No attribution found — unmanaged.
	result.Attribution = AttributionResult{
		Bucket: BucketUnmanaged,
	}
	return result
}

// attributeBySession tries to attribute an event via session ID → assignment.
func (a *Attributor) attributeBySession(event Event) (AttributionResult, bool) {
	assignments, ok := a.assignmentsByWorker[event.SessionID]
	if !ok || len(assignments) == 0 {
		return AttributionResult{}, false
	}

	// Find an assignment whose time window contains the event timestamp.
	for _, asn := range assignments {
		if a.isInTimeWindow(event.Timestamp, asn) {
			return AttributionResult{
				Bucket:       BucketAssigned,
				RMI:          asn.RMI,
				Phase:        asn.Phase,
				Initiative:   asn.Initiative,
				Program:      asn.Program,
				Repository:   asn.Repository,
				AssignmentID: asn.ID,
			}, true
		}
	}

	return AttributionResult{}, false
}

// isInTimeWindow checks if a timestamp falls within an assignment's active period.
func (a *Attributor) isInTimeWindow(ts time.Time, asn AssignmentInfo) bool {
	// Event must be at or after assignment creation.
	if ts.Before(asn.CreatedAt) {
		return false
	}

	// If assignment is still active (no completion time), event is in window.
	if asn.CompletedAt == nil {
		return true
	}

	// Event must be at or before completion.
	return !ts.After(*asn.CompletedAt)
}

// attributeByRepository tries to attribute an event via workspace → repository.
func (a *Attributor) attributeByRepository(event Event) (AttributionResult, bool) {
	if event.Workspace == "" {
		return AttributionResult{}, false
	}

	// Normalize workspace path.
	workspace := filepath.Clean(event.Workspace)
	if !strings.HasSuffix(workspace, string(filepath.Separator)) {
		workspace += string(filepath.Separator)
	}

	// Find the longest matching repository path (most specific match).
	var bestMatch RepositoryInfo
	var bestLen int

	for path, repo := range a.repositoriesByPath {
		if strings.HasPrefix(workspace, path) && len(path) > bestLen {
			bestMatch = repo
			bestLen = len(path)
		}
	}

	if bestLen > 0 {
		return AttributionResult{
			Bucket:     BucketRepository,
			Repository: bestMatch.ID,
		}, true
	}

	return AttributionResult{}, false
}

// computeCost calculates the USD cost for an event using omnidevx-core pricing.
func (a *Attributor) computeCost(event Event) float64 {
	pricing, ok := report.LookupPricing(event.Model)
	if !ok {
		return 0
	}

	return report.EstimateCost(
		pricing,
		event.InputTokens,
		event.OutputTokens,
		event.CacheReadTokens,
		event.CacheCreationTokens,
	)
}

// AttributeAll attributes a slice of events and returns the results.
func (a *Attributor) AttributeAll(events []Event) []AttributedEvent {
	results := make([]AttributedEvent, len(events))
	for i, event := range events {
		results[i] = a.Attribute(event)
	}
	return results
}

// Summary aggregates attribution results by bucket.
type Summary struct {
	// ByBucket aggregates events by attribution bucket.
	ByBucket map[string]BucketSummary

	// ByInitiative aggregates assigned events by initiative.
	ByInitiative map[string]BucketSummary

	// ByRepository aggregates events by repository (both assigned and fallback).
	ByRepository map[string]BucketSummary

	// TotalEvents is the total number of events processed.
	TotalEvents int

	// TotalCostUSD is the total cost across all events.
	TotalCostUSD float64
}

// BucketSummary holds aggregated metrics for a group of events.
type BucketSummary struct {
	EventCount          int
	TotalTokens         int64
	InputTokens         int64
	OutputTokens        int64
	CacheReadTokens     int64
	CacheCreationTokens int64
	CostUSD             float64
}

// Summarize computes aggregate statistics from attributed events.
func Summarize(events []AttributedEvent) Summary {
	s := Summary{
		ByBucket:     make(map[string]BucketSummary),
		ByInitiative: make(map[string]BucketSummary),
		ByRepository: make(map[string]BucketSummary),
	}

	for _, ae := range events {
		s.TotalEvents++
		s.TotalCostUSD += ae.CostUSD

		// Aggregate by bucket.
		bucket := s.ByBucket[ae.Attribution.Bucket]
		addToSummary(&bucket, ae)
		s.ByBucket[ae.Attribution.Bucket] = bucket

		// Aggregate by initiative (assigned only).
		if ae.Attribution.Bucket == BucketAssigned && ae.Attribution.Initiative != "" {
			init := s.ByInitiative[ae.Attribution.Initiative]
			addToSummary(&init, ae)
			s.ByInitiative[ae.Attribution.Initiative] = init
		}

		// Aggregate by repository (assigned and repository buckets).
		if ae.Attribution.Repository != "" {
			repo := s.ByRepository[ae.Attribution.Repository]
			addToSummary(&repo, ae)
			s.ByRepository[ae.Attribution.Repository] = repo
		}
	}

	return s
}

func addToSummary(s *BucketSummary, ae AttributedEvent) {
	s.EventCount++
	s.InputTokens += ae.Event.InputTokens
	s.OutputTokens += ae.Event.OutputTokens
	s.CacheReadTokens += ae.Event.CacheReadTokens
	s.CacheCreationTokens += ae.Event.CacheCreationTokens
	s.TotalTokens += ae.Event.TotalTokens()
	s.CostUSD += ae.CostUSD
}
