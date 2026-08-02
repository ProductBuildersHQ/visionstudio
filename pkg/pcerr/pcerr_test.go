package pcerr

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorFormat(t *testing.T) {
	e := New(StateWrongStatus, "RMI-X-001 is completed", "nothing to claim")
	got := e.Error()
	want := "[STATE_WRONG_STATUS] RMI-X-001 is completed — nothing to claim"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestErrorFormatWithCause(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	e := Wrap(InternalStore, "failed to list RMIs", "check that the Dolt database is running", cause)
	got := e.Error()
	if got == "" {
		t.Fatal("empty error string")
	}
	if !errors.Is(e, cause) {
		t.Fatal("Unwrap should return the cause")
	}
}

func TestCodeExtraction(t *testing.T) {
	e := New(NotFound, "RMI not found", "check ID")
	if Code(e) != NotFound {
		t.Fatalf("got %q", Code(e))
	}

	wrapped := fmt.Errorf("outer: %w", e)
	if Code(wrapped) != NotFound {
		t.Fatalf("wrapped: got %q", Code(wrapped))
	}

	if Code(fmt.Errorf("plain error")) != "" {
		t.Fatal("plain error should return empty code")
	}
}

func TestHasCategory(t *testing.T) {
	tests := []struct {
		code string
		cat  Category
		want bool
	}{
		{StateWrongStatus, CatState, true},
		{StateConflict, CatState, true},
		{NotFound, CatNotFound, true},
		{InputMissing, CatInput, true},
		{InternalStore, CatInternal, true},
		{BlockedDependency, CatBlocked, true},
		{NotFound, CatState, false},
		{InputMissing, CatInternal, false},
	}
	for _, tt := range tests {
		e := New(tt.code, "msg", "hint")
		got := HasCategory(e, tt.cat)
		if got != tt.want {
			t.Errorf("HasCategory(%s, %s) = %v, want %v", tt.code, tt.cat, got, tt.want)
		}
	}
}

func TestConvenienceChecks(t *testing.T) {
	if !IsNotFound(New(NotFound, "", "")) {
		t.Error("IsNotFound failed")
	}
	if !IsState(New(StateWrongStatus, "", "")) {
		t.Error("IsState failed")
	}
	if !IsBlocked(New(BlockedDependency, "", "")) {
		t.Error("IsBlocked failed")
	}
	if !IsInput(New(InputMissing, "", "")) {
		t.Error("IsInput failed")
	}
	if !IsInternal(New(InternalStore, "", "")) {
		t.Error("IsInternal failed")
	}
	if IsNotFound(fmt.Errorf("plain")) {
		t.Error("plain error should not be NotFound")
	}
}
