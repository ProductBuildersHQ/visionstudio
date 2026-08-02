package service

import (
	"context"
	"testing"
)

func TestCreateProgram(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	prog, err := svc.CreateProgram(ctx, "PROG-TEST", "Test Program", "default", "A test program")
	if err != nil {
		t.Fatal(err)
	}
	if prog.Name != "Test Program" {
		t.Fatalf("expected name 'Test Program', got %s", prog.Name)
	}
	if prog.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}

	_, err = svc.CreateProgram(ctx, "PROG-TEST", "Dup", "default", "")
	if err == nil {
		t.Fatal("expected error on duplicate")
	}
}

func TestListPrograms(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateProgram(ctx, "PROG-A", "A", "default", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateProgram(ctx, "PROG-B", "B", "default", ""); err != nil {
		t.Fatal(err)
	}

	progs, err := svc.ListPrograms(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(progs) != 2 {
		t.Fatalf("expected 2 programs, got %d", len(progs))
	}
}

func TestUpdateProgram(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	prog, err := svc.CreateProgram(ctx, "PROG-UP", "Original", "default", "")
	if err != nil {
		t.Fatal(err)
	}

	prog.Name = "Updated"
	if err := svc.UpdateProgram(ctx, prog); err != nil {
		t.Fatal(err)
	}

	got, err := svc.GetProgram(ctx, "PROG-UP")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Updated" {
		t.Fatalf("expected 'Updated', got %s", got.Name)
	}
}
