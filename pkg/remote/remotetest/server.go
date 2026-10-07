// Package remotetest provides an in-process fake of the VisionStudio Cloud
// tenant CRUD API for testing code that uses pkg/remote. It mirrors the
// cloud API's contract — routes under /t/{tenant}/api/v1, bearer/X-API-Key
// authentication, tenant membership (401 vs 403), 404 on missing entities,
// 400 on rejected writes, strict request decoding — and serves requests
// through the real pkg/service over an in-memory store, as the cloud does
// over a tenant database.
package remotetest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// Server is a running fake cloud API.
type Server struct {
	*httptest.Server

	// Service is the tenant's backing service; seed or inspect it directly.
	Service *service.Service

	mu       sync.Mutex
	token    string
	tenants  map[string]bool
	requests []Request
}

// Request records what the fake received (for header assertions).
type Request struct {
	Method string
	Path   string
	Query  string
	Header http.Header
}

// NewServer starts a fake whose only valid credential is token and whose
// principal is a member of tenants. Close it when done.
func NewServer(token string, tenants ...string) *Server {
	s := &Server{
		Service: service.New(store.NewMemStore()),
		token:   token,
		tenants: map[string]bool{},
	}
	for _, t := range tenants {
		s.tenants[t] = true
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /t/{tenant}/api/v1/initiatives", s.listInitiatives)
	mux.HandleFunc("POST /t/{tenant}/api/v1/initiatives", s.createInitiative)
	mux.HandleFunc("GET /t/{tenant}/api/v1/initiatives/{id}", s.getInitiative)
	mux.HandleFunc("GET /t/{tenant}/api/v1/rmis", s.listRMIs)
	mux.HandleFunc("POST /t/{tenant}/api/v1/rmis", s.createRMI)
	mux.HandleFunc("GET /t/{tenant}/api/v1/rmis/{id}", s.getRMI)
	mux.HandleFunc("GET /t/{tenant}/api/v1/programs", s.listPrograms)
	mux.HandleFunc("GET /t/{tenant}/api/v1/phases", s.listPhases)
	s.Server = httptest.NewServer(s.record(mux))
	return s
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
		s.requests = append(s.requests, Request{Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery, Header: r.Header.Clone()})
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

// withService mirrors the cloud API's authorize → resolve → delegate flow.
func (s *Server) withService(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, svc *service.Service) (any, int, error)) {
	if tok := s.credential(r); tok == "" || tok != s.token {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if !s.tenants[r.PathValue("tenant")] {
		writeError(w, http.StatusForbidden, "not a member of this tenant")
		return
	}
	body, status, err := fn(r.Context(), s.Service)
	if err != nil {
		if status == 0 {
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}
	if status == 0 {
		status = http.StatusOK
	}
	writeJSON(w, status, body)
}

func (s *Server) listInitiatives(w http.ResponseWriter, r *http.Request) {
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		inits, err := svc.ListInitiatives(ctx)
		if err != nil {
			return nil, 0, err
		}
		if inits == nil {
			inits = []*store.Initiative{}
		}
		return map[string]any{"initiatives": inits}, http.StatusOK, nil
	})
}

func (s *Server) getInitiative(w http.ResponseWriter, r *http.Request) {
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		in, err := svc.GetInitiative(ctx, r.PathValue("id"))
		if err != nil {
			return nil, http.StatusNotFound, err
		}
		return in, http.StatusOK, nil
	})
}

func (s *Server) createInitiative(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID           string `json:"id"`
		Organization string `json:"organization"`
		Title        string `json:"title"`
		Description  string `json:"description,omitempty"`
		Priority     string `json:"priority,omitempty"`
		InitType     string `json:"init_type,omitempty"`
		WorkflowID   string `json:"workflow_id,omitempty"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		in, err := svc.CreateInitiative(ctx, req.ID, req.Organization, req.Title, req.Description, req.Priority, req.InitType, req.WorkflowID)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		return in, http.StatusCreated, nil
	})
}

func (s *Server) listRMIs(w http.ResponseWriter, r *http.Request) {
	initiativeID := r.URL.Query().Get("initiative")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		var (
			rmis []*store.RoadmapItem
			err  error
		)
		if initiativeID != "" {
			rmis, err = svc.ListRMIs(ctx, initiativeID)
		} else {
			rmis, err = svc.ListAllRMIs(ctx)
		}
		if err != nil {
			return nil, 0, err
		}
		if rmis == nil {
			rmis = []*store.RoadmapItem{}
		}
		return map[string]any{"rmis": rmis}, http.StatusOK, nil
	})
}

func (s *Server) getRMI(w http.ResponseWriter, r *http.Request) {
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		rmi, err := svc.GetRMI(ctx, r.PathValue("id"))
		if err != nil {
			return nil, http.StatusNotFound, err
		}
		return rmi, http.StatusOK, nil
	})
}

func (s *Server) createRMI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID                 string   `json:"id"`
		RepositoryID       string   `json:"repository_id"`
		InitiativeID       string   `json:"initiative_id,omitempty"`
		PhaseID            string   `json:"phase_id,omitempty"`
		Title              string   `json:"title"`
		Description        string   `json:"description,omitempty"`
		ItemType           string   `json:"item_type"`
		Priority           string   `json:"priority,omitempty"`
		Required           bool     `json:"required"`
		SequenceNumber     int      `json:"sequence_number,omitempty"`
		AcceptanceCriteria []string `json:"acceptance_criteria,omitempty"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		rmi, err := svc.CreateRMI(ctx, req.ID, req.RepositoryID, req.InitiativeID, req.PhaseID, req.Title, req.Description, req.ItemType, req.Priority, req.Required, req.SequenceNumber, req.AcceptanceCriteria)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		return rmi, http.StatusCreated, nil
	})
}

func (s *Server) listPrograms(w http.ResponseWriter, r *http.Request) {
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		progs, err := svc.ListPrograms(ctx)
		if err != nil {
			return nil, 0, err
		}
		if progs == nil {
			progs = []*store.Program{}
		}
		return map[string]any{"programs": progs}, http.StatusOK, nil
	})
}

func (s *Server) listPhases(w http.ResponseWriter, r *http.Request) {
	initiativeID := r.URL.Query().Get("initiative")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		var phases []*store.Phase
		if initiativeID != "" {
			ph, err := svc.ListPhases(ctx, initiativeID)
			if err != nil {
				return nil, 0, err
			}
			phases = ph
		} else {
			inits, err := svc.ListInitiatives(ctx)
			if err != nil {
				return nil, 0, err
			}
			for _, in := range inits {
				ph, err := svc.ListPhases(ctx, in.ID)
				if err != nil {
					return nil, 0, err
				}
				phases = append(phases, ph...)
			}
		}
		if phases == nil {
			phases = []*store.Phase{}
		}
		return map[string]any{"phases": phases}, http.StatusOK, nil
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// A write failure means the client went away; nothing to report to.
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
