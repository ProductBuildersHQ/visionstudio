package service

import (
	"context"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func (s *Service) CreateProgram(ctx context.Context, id, name, org, description string) (*store.Program, error) {
	now := time.Now()
	prog := &store.Program{
		ID:           id,
		Name:         name,
		Organization: org,
		Description:  description,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.Store.CreateProgram(ctx, prog); err != nil {
		return nil, err
	}
	return prog, nil
}

func (s *Service) GetProgram(ctx context.Context, id string) (*store.Program, error) {
	return s.Store.GetProgram(ctx, id)
}

func (s *Service) ListPrograms(ctx context.Context) ([]*store.Program, error) {
	return s.Store.ListPrograms(ctx)
}

func (s *Service) UpdateProgram(ctx context.Context, prog *store.Program) error {
	prog.UpdatedAt = time.Now()
	return s.Store.UpdateProgram(ctx, prog)
}
