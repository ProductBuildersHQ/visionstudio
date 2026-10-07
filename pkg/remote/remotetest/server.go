// Package remotetest provides an in-process fake of the VisionStudio Cloud
// tenant API for testing code that uses pkg/remote. It mirrors the cloud
// API's documented contract — routes under /t/{tenant}/api/v1 plus
// GET /api/v1/me, bearer/X-API-Key authentication, tenant membership (401 vs
// 403), the {"error", "code"} error body with 400/404/409 classification,
// strict request decoding, [] for empty lists, the {"assignment": null} and
// {"workflow": null} wrappers, and the server-enforced lease rules — and
// serves requests through the real pkg/service over an in-memory store, as
// the cloud does over a tenant database.
package remotetest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/ProductBuildersHQ/visionstudio/pkg/pcerr"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// Principal identity the fake reports from its "me" endpoints.
const (
	PrincipalID    = "test-principal"
	PrincipalEmail = "dev@example.com"
	AuthMethod     = "api_key"
)

// Error codes in the "code" field of an error response (the cloud's).
const (
	CodeUnauthenticated = "UNAUTHENTICATED"
	CodeForbidden       = "FORBIDDEN"
	CodeInvalid         = "INVALID"
	CodeInvalidRef      = "INVALID_REFERENCE"
	CodeNotFound        = "NOT_FOUND"
	CodeConflict        = "CONFLICT"
	CodeLeaseConflict   = "LEASE_CONFLICT"
	CodeInternal        = "INTERNAL"
)

// maxRequestBytes caps a request body, like the cloud.
const maxRequestBytes = 8 << 20

// Server is a running fake cloud API.
type Server struct {
	*httptest.Server

	// Service is the tenant's backing service; seed or inspect it directly.
	Service *service.Service

	mu       sync.Mutex
	token    string
	tenants  []string
	members  map[string]bool
	requests []Request

	// leaseMu serializes assignment writes, standing in for the cloud's
	// per-tenant, per-RMI advisory lock.
	leaseMu sync.Mutex
}

// Request records what the fake received (for header assertions).
type Request struct {
	Method string
	Path   string
	// RawPath is the path as sent, with any percent-escapes intact.
	RawPath string
	Query   string
	Header  http.Header
}

// NewServer starts a fake whose only valid credential is token and whose
// principal is a member (owner) of tenants. Close it when done.
func NewServer(token string, tenants ...string) *Server {
	s := &Server{
		Service: service.New(store.NewMemStore()),
		token:   token,
		tenants: append([]string(nil), tenants...),
		members: map[string]bool{},
	}
	for _, t := range tenants {
		s.members[t] = true
	}
	mux := http.NewServeMux()
	s.routes(mux)
	s.Server = httptest.NewServer(s.record(mux))
	return s
}

func (s *Server) routes(mux *http.ServeMux) {
	const t = "/t/{tenant}/api/v1"

	mux.HandleFunc("GET /api/v1/me", s.getMe)
	mux.HandleFunc("GET "+t+"/me", s.getTenantMe)

	mux.HandleFunc("GET "+t+"/programs", s.listPrograms)
	mux.HandleFunc("POST "+t+"/programs", s.createProgram)
	mux.HandleFunc("GET "+t+"/programs/{id}", s.getProgram)
	mux.HandleFunc("PUT "+t+"/programs/{id}", s.updateProgram)
	mux.HandleFunc("PATCH "+t+"/programs/{id}", s.updateProgram)

	mux.HandleFunc("GET "+t+"/initiatives", s.listInitiatives)
	mux.HandleFunc("POST "+t+"/initiatives", s.createInitiative)
	mux.HandleFunc("GET "+t+"/initiatives/{id}", s.getInitiative)
	mux.HandleFunc("PUT "+t+"/initiatives/{id}", s.updateInitiative)
	mux.HandleFunc("PATCH "+t+"/initiatives/{id}", s.updateInitiative)
	mux.HandleFunc("POST "+t+"/initiatives/{id}/transition", s.transitionInitiative)
	mux.HandleFunc("GET "+t+"/initiatives/{id}/workflow", s.getInitiativeWorkflow)
	mux.HandleFunc("PUT "+t+"/initiatives/{id}/workflow", s.selectInitiativeWorkflow)
	mux.HandleFunc("GET "+t+"/initiatives/{id}/judge-results", s.listJudgeResults)

	mux.HandleFunc("GET "+t+"/phases", s.listPhases)
	mux.HandleFunc("POST "+t+"/phases", s.createPhase)
	mux.HandleFunc("DELETE "+t+"/phases/{id...}", s.deletePhase)

	mux.HandleFunc("GET "+t+"/rmis", s.listRMIs)
	mux.HandleFunc("POST "+t+"/rmis", s.createRMI)
	mux.HandleFunc("GET "+t+"/rmis/{id}", s.getRMI)
	mux.HandleFunc("PUT "+t+"/rmis/{id}", s.updateRMI)
	mux.HandleFunc("PATCH "+t+"/rmis/{id}", s.updateRMI)
	mux.HandleFunc("POST "+t+"/rmis/{id}/status", s.updateRMIStatus)
	mux.HandleFunc("POST "+t+"/rmis/{id}/move", s.moveRMI)
	mux.HandleFunc("GET "+t+"/rmis/{id}/dependencies", s.listRMIDependencies)
	mux.HandleFunc("GET "+t+"/rmis/{id}/active-assignment", s.getActiveAssignment)

	mux.HandleFunc("GET "+t+"/dependencies", s.listDependencies)
	mux.HandleFunc("POST "+t+"/dependencies", s.createDependency)

	mux.HandleFunc("GET "+t+"/assignments", s.listAssignments)
	mux.HandleFunc("POST "+t+"/assignments", s.createAssignment)
	mux.HandleFunc("GET "+t+"/assignments/{id}", s.getAssignment)
	mux.HandleFunc("PUT "+t+"/assignments/{id}", s.updateAssignment)

	mux.HandleFunc("GET "+t+"/evidence", s.listEvidence)
	mux.HandleFunc("POST "+t+"/evidence", s.createEvidence)

	mux.HandleFunc("GET "+t+"/releases", s.listReleases)
	mux.HandleFunc("GET "+t+"/releases/{id...}", s.getRelease)

	mux.HandleFunc("GET "+t+"/repositories", s.listRepositories)
	mux.HandleFunc("GET "+t+"/repositories/{id...}", s.getRepository)

	mux.HandleFunc("GET "+t+"/workflows", s.listWorkflows)
	mux.HandleFunc("GET "+t+"/workflows/{id}", s.getWorkflow)

	mux.HandleFunc("GET "+t+"/spec-documents", s.listSpecDocuments)
	mux.HandleFunc("POST "+t+"/spec-documents", s.createSpecDocument)
	mux.HandleFunc("GET "+t+"/spec-documents/{id...}", s.getSpecDocument)
	mux.HandleFunc("PUT "+t+"/spec-documents/{id...}", s.replaceSpecDocument)
	mux.HandleFunc("DELETE "+t+"/spec-documents/{id...}", s.deleteSpecDocument)

	mux.HandleFunc("POST "+t+"/judge-results", s.createJudgeResult)
}

// Requests returns a copy of every request received so far.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Request(nil), s.requests...)
}

func (s *Server) record(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.requests = append(s.requests, Request{
			Method: r.Method, Path: r.URL.Path, RawPath: r.URL.EscapedPath(),
			Query: r.URL.RawQuery, Header: r.Header.Clone(),
		})
		s.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) credential(r *http.Request) string {
	const prefix = "Bearer "
	if h := r.Header.Get("Authorization"); len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	return strings.TrimSpace(r.Header.Get("X-API-Key"))
}

func (s *Server) authenticated(r *http.Request) bool {
	tok := s.credential(r)
	return tok != "" && tok == s.token
}

// authorize mirrors the cloud's 401 (credential) then 403 (membership).
func (s *Server) authorize(w http.ResponseWriter, r *http.Request) bool {
	if !s.authenticated(r) {
		writeError(w, http.StatusUnauthorized, CodeUnauthenticated, "authentication required")
		return false
	}
	if !s.members[r.PathValue("tenant")] {
		writeError(w, http.StatusForbidden, CodeForbidden, "not a member of this tenant")
		return false
	}
	return true
}

// serviceFunc is a handler body: it returns the response body and success
// status (0 → 200), or an error classify maps to a status and code.
type serviceFunc func(ctx context.Context, svc *service.Service) (any, int, error)

// withService mirrors the cloud API's authorize → delegate → encode flow.
func (s *Server) withService(w http.ResponseWriter, r *http.Request, fn serviceFunc) {
	if !s.authorize(w, r) {
		return
	}
	body, status, err := fn(r.Context(), s.Service)
	if err != nil {
		st, code := classify(err)
		msg := err.Error()
		if st >= http.StatusInternalServerError {
			msg = "internal error"
		}
		writeError(w, st, code, msg)
		return
	}
	if status == http.StatusNoContent {
		w.WriteHeader(status)
		return
	}
	if status == 0 {
		status = http.StatusOK
	}
	writeJSON(w, status, body)
}

// --- identity ---

type mePrincipal struct {
	ID         string `json:"id"`
	Email      string `json:"email,omitempty"`
	AuthMethod string `json:"auth_method,omitempty"`
}

type membership struct {
	Tenant string `json:"tenant"`
	Name   string `json:"name,omitempty"`
	Role   string `json:"role,omitempty"`
}

func principal() mePrincipal {
	return mePrincipal{ID: PrincipalID, Email: PrincipalEmail, AuthMethod: AuthMethod}
}

// getMe serves GET /api/v1/me: principal and every membership.
func (s *Server) getMe(w http.ResponseWriter, r *http.Request) {
	if !s.authenticated(r) {
		writeError(w, http.StatusUnauthorized, CodeUnauthenticated, "authentication required")
		return
	}
	ms := make([]membership, 0, len(s.tenants))
	for _, t := range s.tenants {
		ms = append(ms, membership{Tenant: t, Name: t, Role: "owner"})
	}
	writeJSON(w, http.StatusOK, map[string]any{"principal": principal(), "memberships": ms})
}

// getTenantMe serves GET /t/{tenant}/api/v1/me.
func (s *Server) getTenantMe(w http.ResponseWriter, r *http.Request) {
	if !s.authorize(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"principal": principal(), "tenant": r.PathValue("tenant"), "role": "owner"})
}

// --- errors ---

// apiError is an error a handler raises with an explicit status and code.
type apiError struct {
	status int
	code   string
	msg    string
}

func (e *apiError) Error() string { return e.msg }

func badRequest(msg string) error {
	return &apiError{status: http.StatusBadRequest, code: CodeInvalid, msg: msg}
}

func conflict(code, msg string) error {
	return &apiError{status: http.StatusConflict, code: code, msg: msg}
}

func errMissing(field string) error { return badRequest(field + " is required") }

// classify maps an error to a status and code the way the cloud does for
// its in-memory store: explicit apiErrors, pcerr categories, then the
// store's "not found" / "already exists" message forms; anything else the
// service rejected is a 400.
func classify(err error) (int, string) {
	var ae *apiError
	if errors.As(err, &ae) {
		return ae.status, ae.code
	}
	var pe *pcerr.Error
	if errors.As(err, &pe) {
		switch {
		case pcerr.IsNotFound(err):
			return http.StatusNotFound, pe.Code
		case pcerr.IsInput(err):
			return http.StatusBadRequest, pe.Code
		case pcerr.IsInternal(err):
			return http.StatusInternalServerError, pe.Code
		default:
			return http.StatusConflict, pe.Code
		}
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return http.StatusInternalServerError, CodeInternal
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return http.StatusNotFound, CodeNotFound
	case strings.Contains(msg, "already exists"), strings.Contains(msg, "duplicate"):
		return http.StatusConflict, CodeConflict
	}
	return http.StatusBadRequest, CodeInvalid
}

// asReference re-labels a not-found error from a create or move as a 400:
// the missing entity is one the body references, not the URL's.
func asReference(err error) error {
	if err == nil {
		return nil
	}
	if status, _ := classify(err); status == http.StatusNotFound {
		return &apiError{status: http.StatusBadRequest, code: CodeInvalidRef, msg: err.Error()}
	}
	return err
}

// --- encoding ---

// decodeJSON strictly decodes the request body (unknown fields and
// trailing data are rejected) and writes a 400 on failure.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	data, ok := readBody(w, r)
	if !ok {
		return false
	}
	if err := decodeInto(data, dst); err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalid, err.Error())
		return false
	}
	return true
}

func readBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBytes))
	if err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalid, "invalid request body: "+err.Error())
		return nil, false
	}
	return data, true
}

func decodeInto(data []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return badRequest("invalid request body: " + err.Error())
	}
	if dec.More() {
		return badRequest("invalid request body: unexpected data after the JSON value")
	}
	return nil
}

// buildUpdate returns the entity a PUT (replace) or PATCH (merge onto the
// stored entity) asks to store; a body id must equal the URL's.
func buildUpdate[T any](ctx context.Context, method string, data []byte, pathID string,
	load func(ctx context.Context, id string) (*T, error), idOf func(*T) *string,
) (*T, error) {
	out := new(T)
	if method == http.MethodPatch {
		cur, err := load(ctx, pathID)
		if err != nil {
			return nil, err
		}
		raw, err := json.Marshal(cur)
		if err != nil {
			return nil, fmt.Errorf("copy stored entity: %w", err)
		}
		if err := json.Unmarshal(raw, out); err != nil {
			return nil, fmt.Errorf("copy stored entity: %w", err)
		}
	}
	if err := decodeInto(data, out); err != nil {
		return nil, err
	}
	id := idOf(out)
	switch {
	case *id == "":
		*id = pathID
	case *id != pathID:
		return nil, badRequest(fmt.Sprintf("body id %q does not match URL id %q", *id, pathID))
	}
	return out, nil
}

// nonNil makes empty lists encode as [] rather than null.
func nonNil[T any](v []T) []T {
	if v == nil {
		return []T{}
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// A write failure means the client went away; nothing to report to.
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]string{"error": msg, "code": code})
}
