package maturity_test

import (
	"context"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/maturity"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestBuiltInModels(t *testing.T) {
	models := maturity.BuiltInModels()
	if len(models) != 4 {
		t.Fatalf("expected 4 built-in models, got %d", len(models))
	}
	wantIDs := map[string]bool{
		"big-tech-essentials":  false,
		"big-tech-full":        false,
		"continuous-discovery": false,
		"api-first":            false,
	}
	for _, m := range models {
		if _, ok := wantIDs[m.ID]; !ok {
			t.Errorf("unexpected model ID %q", m.ID)
		}
		wantIDs[m.ID] = true
		if len(m.Dimensions) == 0 {
			t.Errorf("model %q has no dimensions", m.ID)
		}
		if m.MaxLevel == 0 {
			t.Errorf("model %q has zero max level", m.ID)
		}
	}
	for id, found := range wantIDs {
		if !found {
			t.Errorf("expected built-in model %q not found", id)
		}
	}
}

func TestSeedBuiltIn(t *testing.T) {
	ctx := context.Background()
	s := store.NewMemStore()

	created, err := maturity.SeedBuiltIn(ctx, s)
	if err != nil {
		t.Fatalf("SeedBuiltIn: %v", err)
	}
	if created != 4 {
		t.Errorf("expected 4 created, got %d", created)
	}

	models, err := s.ListCapabilityModels(ctx)
	if err != nil {
		t.Fatalf("ListCapabilityModels: %v", err)
	}
	if len(models) != 4 {
		t.Errorf("expected 4 stored models, got %d", len(models))
	}

	created, err = maturity.SeedBuiltIn(ctx, s)
	if err != nil {
		t.Fatalf("SeedBuiltIn (second run): %v", err)
	}
	if created != 0 {
		t.Errorf("expected 0 created on re-seed, got %d", created)
	}
}

func TestBigTechEssentialsDimensions(t *testing.T) {
	ctx := context.Background()
	s := store.NewMemStore()

	if _, err := maturity.SeedBuiltIn(ctx, s); err != nil {
		t.Fatalf("SeedBuiltIn: %v", err)
	}

	model, err := s.GetCapabilityModel(ctx, "big-tech-essentials")
	if err != nil {
		t.Fatalf("GetCapabilityModel: %v", err)
	}

	wantDims := []string{
		"customer-obsession",
		"okr-quality",
		"api-first",
		"explicit-tradeoffs",
		"documentation",
	}
	if len(model.Dimensions) != len(wantDims) {
		t.Fatalf("expected %d dimensions, got %d", len(wantDims), len(model.Dimensions))
	}

	for i, dim := range model.Dimensions {
		if dim.Key != wantDims[i] {
			t.Errorf("dimension %d: expected key %q, got %q", i, wantDims[i], dim.Key)
		}
		if len(dim.Levels) != 5 {
			t.Errorf("dimension %q: expected 5 levels, got %d", dim.Key, len(dim.Levels))
		}
	}
}
