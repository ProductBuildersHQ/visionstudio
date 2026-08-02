package tokens

import (
	"testing"
	"time"
)

func TestAttributor_SessionAttribution(t *testing.T) {
	now := time.Now()
	assignments := []AssignmentInfo{
		{
			ID:         "asn-1",
			Worker:     "session-abc",
			RMI:        "RMI-TEST-001",
			Phase:      "phase-1",
			Initiative: "INIT-TEST-001",
			Program:    "PROG-TEST",
			Repository: "github.com/test/repo",
			CreatedAt:  now.Add(-1 * time.Hour),
			CompletedAt: func() *time.Time {
				t := now.Add(1 * time.Hour)
				return &t
			}(),
		},
	}

	repositories := []RepositoryInfo{
		{ID: "github.com/test/repo", LocalPath: "/home/user/repo"},
	}

	attr := NewAttributor(assignments, repositories)

	t.Run("matches session and time window", func(t *testing.T) {
		event := Event{
			ID:           "evt-1",
			SessionID:    "session-abc",
			Workspace:    "/home/user/repo",
			Timestamp:    now,
			Model:        "claude-opus-4-8",
			InputTokens:  100,
			OutputTokens: 500,
		}

		result := attr.Attribute(event)

		if result.Attribution.Bucket != BucketAssigned {
			t.Errorf("Bucket = %q, want %q", result.Attribution.Bucket, BucketAssigned)
		}
		if result.Attribution.RMI != "RMI-TEST-001" {
			t.Errorf("RMI = %q, want %q", result.Attribution.RMI, "RMI-TEST-001")
		}
		if result.Attribution.Initiative != "INIT-TEST-001" {
			t.Errorf("Initiative = %q, want %q", result.Attribution.Initiative, "INIT-TEST-001")
		}
		if result.Attribution.AssignmentID != "asn-1" {
			t.Errorf("AssignmentID = %q, want %q", result.Attribution.AssignmentID, "asn-1")
		}
	})

	t.Run("event before assignment window falls back to repository", func(t *testing.T) {
		event := Event{
			ID:           "evt-2",
			SessionID:    "session-abc",
			Workspace:    "/home/user/repo",
			Timestamp:    now.Add(-2 * time.Hour), // Before assignment created
			Model:        "claude-opus-4-8",
			InputTokens:  100,
			OutputTokens: 500,
		}

		result := attr.Attribute(event)

		if result.Attribution.Bucket != BucketRepository {
			t.Errorf("Bucket = %q, want %q", result.Attribution.Bucket, BucketRepository)
		}
		if result.Attribution.Repository != "github.com/test/repo" {
			t.Errorf("Repository = %q, want %q", result.Attribution.Repository, "github.com/test/repo")
		}
	})

	t.Run("event after assignment window falls back to repository", func(t *testing.T) {
		event := Event{
			ID:           "evt-3",
			SessionID:    "session-abc",
			Workspace:    "/home/user/repo",
			Timestamp:    now.Add(2 * time.Hour), // After assignment completed
			Model:        "claude-opus-4-8",
			InputTokens:  100,
			OutputTokens: 500,
		}

		result := attr.Attribute(event)

		if result.Attribution.Bucket != BucketRepository {
			t.Errorf("Bucket = %q, want %q", result.Attribution.Bucket, BucketRepository)
		}
	})
}

func TestAttributor_OpenEndedAssignment(t *testing.T) {
	now := time.Now()
	assignments := []AssignmentInfo{
		{
			ID:          "asn-active",
			Worker:      "session-xyz",
			RMI:         "RMI-ACTIVE-001",
			Initiative:  "INIT-ACTIVE",
			CreatedAt:   now.Add(-1 * time.Hour),
			CompletedAt: nil, // Active/open-ended
		},
	}

	attr := NewAttributor(assignments, nil)

	event := Event{
		ID:           "evt-future",
		SessionID:    "session-xyz",
		Timestamp:    now.Add(10 * time.Hour), // Far in the future
		Model:        "claude-opus-4-8",
		InputTokens:  100,
		OutputTokens: 500,
	}

	result := attr.Attribute(event)

	if result.Attribution.Bucket != BucketAssigned {
		t.Errorf("Bucket = %q, want %q (open-ended assignment)", result.Attribution.Bucket, BucketAssigned)
	}
	if result.Attribution.RMI != "RMI-ACTIVE-001" {
		t.Errorf("RMI = %q, want %q", result.Attribution.RMI, "RMI-ACTIVE-001")
	}
}

func TestAttributor_RepositoryFallback(t *testing.T) {
	repositories := []RepositoryInfo{
		{ID: "github.com/org/project", LocalPath: "/home/user/projects/project"},
		{ID: "github.com/org/subproject", LocalPath: "/home/user/projects/project/packages/sub"},
	}

	attr := NewAttributor(nil, repositories)

	t.Run("matches repository by workspace prefix", func(t *testing.T) {
		event := Event{
			ID:           "evt-repo",
			SessionID:    "session-unknown",
			Workspace:    "/home/user/projects/project/src",
			Timestamp:    time.Now(),
			Model:        "claude-opus-4-8",
			InputTokens:  100,
			OutputTokens: 500,
		}

		result := attr.Attribute(event)

		if result.Attribution.Bucket != BucketRepository {
			t.Errorf("Bucket = %q, want %q", result.Attribution.Bucket, BucketRepository)
		}
		if result.Attribution.Repository != "github.com/org/project" {
			t.Errorf("Repository = %q, want %q", result.Attribution.Repository, "github.com/org/project")
		}
	})

	t.Run("matches most specific repository", func(t *testing.T) {
		event := Event{
			ID:           "evt-sub",
			SessionID:    "session-unknown",
			Workspace:    "/home/user/projects/project/packages/sub/lib",
			Timestamp:    time.Now(),
			Model:        "claude-opus-4-8",
			InputTokens:  100,
			OutputTokens: 500,
		}

		result := attr.Attribute(event)

		if result.Attribution.Bucket != BucketRepository {
			t.Errorf("Bucket = %q, want %q", result.Attribution.Bucket, BucketRepository)
		}
		// Should match the more specific subproject path
		if result.Attribution.Repository != "github.com/org/subproject" {
			t.Errorf("Repository = %q, want %q", result.Attribution.Repository, "github.com/org/subproject")
		}
	})
}

func TestAttributor_UnmanagedWorkspace(t *testing.T) {
	repositories := []RepositoryInfo{
		{ID: "github.com/org/project", LocalPath: "/home/user/projects/project"},
	}

	attr := NewAttributor(nil, repositories)

	event := Event{
		ID:           "evt-unmanaged",
		SessionID:    "session-unknown",
		Workspace:    "/home/user/personal/hobby-project",
		Timestamp:    time.Now(),
		Model:        "claude-opus-4-8",
		InputTokens:  100,
		OutputTokens: 500,
	}

	result := attr.Attribute(event)

	if result.Attribution.Bucket != BucketUnmanaged {
		t.Errorf("Bucket = %q, want %q", result.Attribution.Bucket, BucketUnmanaged)
	}
	if result.Attribution.Repository != "" {
		t.Errorf("Repository = %q, want empty", result.Attribution.Repository)
	}
}

func TestAttributor_CostCalculation(t *testing.T) {
	attr := NewAttributor(nil, nil)

	event := Event{
		ID:                  "evt-cost",
		Model:               "claude-opus-4-8",
		InputTokens:         1000,
		OutputTokens:        500,
		CacheReadTokens:     2000,
		CacheCreationTokens: 100,
	}

	result := attr.Attribute(event)

	// Cost should be non-zero for a known model
	if result.CostUSD == 0 {
		t.Error("CostUSD = 0, want non-zero for claude-opus-4-8")
	}
}

func TestAttributor_UnknownModelCost(t *testing.T) {
	attr := NewAttributor(nil, nil)

	event := Event{
		ID:           "evt-unknown",
		Model:        "unknown-model-xyz",
		InputTokens:  1000,
		OutputTokens: 500,
	}

	result := attr.Attribute(event)

	// Cost should be zero for unknown model
	if result.CostUSD != 0 {
		t.Errorf("CostUSD = %f, want 0 for unknown model", result.CostUSD)
	}
}

func TestAttributeAll(t *testing.T) {
	now := time.Now()
	assignments := []AssignmentInfo{
		{
			ID:          "asn-1",
			Worker:      "session-1",
			RMI:         "RMI-001",
			Initiative:  "INIT-001",
			CreatedAt:   now.Add(-1 * time.Hour),
			CompletedAt: nil,
		},
	}

	attr := NewAttributor(assignments, nil)

	events := []Event{
		{ID: "evt-1", SessionID: "session-1", Timestamp: now, Model: "claude-opus-4-8", InputTokens: 100},
		{ID: "evt-2", SessionID: "session-1", Timestamp: now, Model: "claude-opus-4-8", InputTokens: 200},
		{ID: "evt-3", SessionID: "session-other", Timestamp: now, Model: "claude-opus-4-8", InputTokens: 300},
	}

	results := attr.AttributeAll(events)

	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}

	// First two should be assigned
	if results[0].Attribution.Bucket != BucketAssigned {
		t.Errorf("results[0].Bucket = %q, want %q", results[0].Attribution.Bucket, BucketAssigned)
	}
	if results[1].Attribution.Bucket != BucketAssigned {
		t.Errorf("results[1].Bucket = %q, want %q", results[1].Attribution.Bucket, BucketAssigned)
	}

	// Third should be unmanaged
	if results[2].Attribution.Bucket != BucketUnmanaged {
		t.Errorf("results[2].Bucket = %q, want %q", results[2].Attribution.Bucket, BucketUnmanaged)
	}
}

func TestSummarize(t *testing.T) {
	events := []AttributedEvent{
		{
			Event:       Event{InputTokens: 100, OutputTokens: 50},
			Attribution: AttributionResult{Bucket: BucketAssigned, Initiative: "INIT-A", Repository: "repo-1"},
			CostUSD:     0.01,
		},
		{
			Event:       Event{InputTokens: 200, OutputTokens: 100},
			Attribution: AttributionResult{Bucket: BucketAssigned, Initiative: "INIT-A", Repository: "repo-1"},
			CostUSD:     0.02,
		},
		{
			Event:       Event{InputTokens: 150, OutputTokens: 75},
			Attribution: AttributionResult{Bucket: BucketRepository, Repository: "repo-2"},
			CostUSD:     0.015,
		},
		{
			Event:       Event{InputTokens: 50, OutputTokens: 25},
			Attribution: AttributionResult{Bucket: BucketUnmanaged},
			CostUSD:     0.005,
		},
	}

	summary := Summarize(events)

	if summary.TotalEvents != 4 {
		t.Errorf("TotalEvents = %d, want 4", summary.TotalEvents)
	}

	if summary.ByBucket[BucketAssigned].EventCount != 2 {
		t.Errorf("Assigned events = %d, want 2", summary.ByBucket[BucketAssigned].EventCount)
	}

	if summary.ByBucket[BucketRepository].EventCount != 1 {
		t.Errorf("Repository events = %d, want 1", summary.ByBucket[BucketRepository].EventCount)
	}

	if summary.ByBucket[BucketUnmanaged].EventCount != 1 {
		t.Errorf("Unmanaged events = %d, want 1", summary.ByBucket[BucketUnmanaged].EventCount)
	}

	if summary.ByInitiative["INIT-A"].EventCount != 2 {
		t.Errorf("INIT-A events = %d, want 2", summary.ByInitiative["INIT-A"].EventCount)
	}

	if summary.ByRepository["repo-1"].EventCount != 2 {
		t.Errorf("repo-1 events = %d, want 2", summary.ByRepository["repo-1"].EventCount)
	}

	if summary.ByRepository["repo-2"].EventCount != 1 {
		t.Errorf("repo-2 events = %d, want 1", summary.ByRepository["repo-2"].EventCount)
	}
}
