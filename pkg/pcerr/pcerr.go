// Package pcerr defines structured errors for PRISM Control.
//
// Every error carries a Code (machine-readable), Message (human-readable),
// and Recovery hint (actionable guidance for agents). Codes are grouped
// by category so callers can branch on the prefix:
//
//   - INPUT_*    — bad user input (comparable to HTTP 400)
//   - NOT_FOUND  — resource doesn't exist (HTTP 404)
//   - STATE_*    — wrong lifecycle state for the operation (HTTP 409)
//   - BLOCKED_*  — precondition not met, e.g. dependency (HTTP 422)
//   - INTEGRITY_ — data constraint violation (HTTP 409/422)
//   - INTERNAL_* — unexpected server-side error (HTTP 500)
package pcerr

import (
	"errors"
	"fmt"
	"strings"
)

// Category groups error codes for programmatic matching.
type Category string

const (
	CatInput     Category = "INPUT"
	CatNotFound  Category = "NOT_FOUND"
	CatState     Category = "STATE"
	CatBlocked   Category = "BLOCKED"
	CatIntegrity Category = "INTEGRITY"
	CatInternal  Category = "INTERNAL"
)

// Common error codes. Use these constants in service-layer code so
// CLI and MCP adapters can switch on them.
const (
	// Input errors — caller supplied bad data.
	InputMissing = "INPUT_MISSING" // required field/flag omitted
	InputInvalid = "INPUT_INVALID" // value present but malformed

	// Not-found errors — referenced resource doesn't exist.
	NotFound = "NOT_FOUND" // generic; entity type is in the message

	// State errors — resource exists but is in the wrong lifecycle state.
	StateWrongStatus = "STATE_WRONG_STATUS" // e.g. claim on a completed RMI
	StateAlreadyDone = "STATE_ALREADY_DONE" // operation is a no-op
	StateConflict    = "STATE_CONFLICT"     // e.g. double claim

	// Blocked errors — precondition prevents the operation.
	BlockedDependency = "BLOCKED_DEPENDENCY" // upstream RMI not completed
	BlockedEmpty      = "BLOCKED_EMPTY"      // bulk op matched zero items

	// Integrity errors — data invariant violated.
	IntegrityDuplicate  = "INTEGRITY_DUPLICATE"  // unique constraint
	IntegrityConstraint = "INTEGRITY_CONSTRAINT" // FK or check constraint

	// Internal errors — unexpected failure in the service or store.
	InternalStore   = "INTERNAL_STORE"   // database/store layer failure
	InternalUnknown = "INTERNAL_UNKNOWN" // catch-all
)

// Error is the structured error type returned by the service layer.
type Error struct {
	Code     string // machine-readable code (e.g. "STATE_WRONG_STATUS")
	Message  string // human-readable summary
	Recovery string // actionable hint for recovery (shown to agents)
	Cause    error  // wrapped underlying error, if any
}

func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s] %s", e.Code, e.Message)
	if e.Recovery != "" {
		fmt.Fprintf(&b, " — %s", e.Recovery)
	}
	if e.Cause != nil {
		fmt.Fprintf(&b, " (cause: %v)", e.Cause)
	}
	return b.String()
}

func (e *Error) Unwrap() error { return e.Cause }

// New creates a structured error with a code, message, and recovery hint.
func New(code, message, recovery string) *Error {
	return &Error{Code: code, Message: message, Recovery: recovery}
}

// Wrap creates a structured error wrapping an underlying cause.
func Wrap(code, message, recovery string, cause error) *Error {
	return &Error{Code: code, Message: message, Recovery: recovery, Cause: cause}
}

// Code extracts the error code from an error chain. Returns "" if the
// error is not a *Error.
func Code(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

// HasCategory returns true if the error's code starts with the given category.
func HasCategory(err error, cat Category) bool {
	c := Code(err)
	return c != "" && strings.HasPrefix(c, string(cat))
}

// IsNotFound is a convenience check for NOT_FOUND errors.
func IsNotFound(err error) bool { return HasCategory(err, CatNotFound) }

// IsState is a convenience check for STATE_* errors.
func IsState(err error) bool { return HasCategory(err, CatState) }

// IsBlocked is a convenience check for BLOCKED_* errors.
func IsBlocked(err error) bool { return HasCategory(err, CatBlocked) }

// IsInput is a convenience check for INPUT_* errors.
func IsInput(err error) bool { return HasCategory(err, CatInput) }

// IsInternal is a convenience check for INTERNAL_* errors.
func IsInternal(err error) bool { return HasCategory(err, CatInternal) }
