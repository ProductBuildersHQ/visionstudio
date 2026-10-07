package remote_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/remote"
	"github.com/ProductBuildersHQ/visionstudio/pkg/remote/remotetest"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

const (
	testToken  = "vsc_live_test_secret" // #nosec G101 -- fake credential for the in-process test server
	testTenant = "acme"
)

func newFake(t *testing.T) *remotetest.Server {
	t.Helper()
	fake := remotetest.NewServer(testToken, testTenant)
	t.Cleanup(fake.Close)
	return fake
}

func newStore(t *testing.T, fake *remotetest.Server, opts ...remote.Option) *remote.Store {
	t.Helper()
	opts = append([]remote.Option{remote.WithToken(testToken)}, opts...)
	rs, err := remote.New(fake.URL, testTenant, opts...)
	if err != nil {
		t.Fatalf("remote.New: %v", err)
	}
	return rs
}

func TestInitiativeCreateGetList(t *testing.T) {
	fake := newFake(t)
	svc := service.New(newStore(t, fake))
	ctx := context.Background()

	created, err := svc.CreateInitiative(ctx, "INIT-ACME-001", "acme", "Launch", "desc", "high", "", "pbhq-lite")
	if err != nil {
		t.Fatalf("CreateInitiative: %v", err)
	}
	if created.Status != "proposed" || created.InitType != "feature" || created.CreatedAt.IsZero() {
		t.Fatalf("created = %+v, want proposed/feature with server timestamps", created)
	}

	// It landed in the tenant's store, not anywhere local.
	if _, err := fake.Service.GetInitiative(ctx, "INIT-ACME-001"); err != nil {
		t.Fatalf("server-side GetInitiative: %v", err)
	}

	got, err := svc.GetInitiative(ctx, "INIT-ACME-001")
	if err != nil {
		t.Fatalf("GetInitiative: %v", err)
	}
	if got.Title != "Launch" || got.WorkflowID != "pbhq-lite" {
		t.Fatalf("got = %+v", got)
	}

	list, err := svc.ListInitiatives(ctx)
	if err != nil {
		t.Fatalf("ListInitiatives: %v", err)
	}
	if len(list) != 1 || list[0].ID != "INIT-ACME-001" {
		t.Fatalf("list = %+v", list)
	}
}

func TestRMICreateGetList(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	svc := service.New(rs)
	ctx := context.Background()

	if _, err := fake.Service.CreateInitiative(ctx, "INIT-ACME-001", "acme", "Launch", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := fake.Service.CreatePhase(ctx, "INIT-ACME-001/phase-1", "INIT-ACME-001", 1, "Foundation", ""); err != nil {
		t.Fatal(err)
	}

	rmi, err := svc.CreateRMI(ctx, "RMI-ACME-001", "github.com/acme/app", "INIT-ACME-001", "INIT-ACME-001/phase-1",
		"Build it", "d", "capability", "high", true, 1, []string{"works"})
	if err != nil {
		t.Fatalf("CreateRMI: %v", err)
	}
	if rmi.Status != "proposed" || rmi.CreatedAt.IsZero() {
		t.Fatalf("rmi = %+v", rmi)
	}
	if _, err := svc.CreateRMI(ctx, "RMI-OTHER-001", "github.com/acme/other", "", "", "Loose", "", "task", "", false, 0, nil); err != nil {
		t.Fatalf("CreateRMI (no initiative): %v", err)
	}

	got, err := svc.GetRMI(ctx, "RMI-ACME-001")
	if err != nil {
		t.Fatalf("GetRMI: %v", err)
	}
	if got.PhaseID != "INIT-ACME-001/phase-1" || len(got.AcceptanceCriteria) != 1 || !got.Required {
		t.Fatalf("got = %+v", got)
	}

	assertIDs(t, "ListRMIs", mustRMIs(t)(svc.ListRMIs(ctx, "INIT-ACME-001")), "RMI-ACME-001")
	assertIDs(t, "ListRMIs(empty)", mustRMIs(t)(rs.ListRMIs(ctx, "")), "RMI-OTHER-001")
	assertIDs(t, "ListAllRMIs", mustRMIs(t)(svc.ListAllRMIs(ctx)), "RMI-ACME-001", "RMI-OTHER-001")
	assertIDs(t, "ListRMIsByRepo", mustRMIs(t)(svc.ListRMIsByRepo(ctx, "github.com/acme/other")), "RMI-OTHER-001")
	assertIDs(t, "ListRMIsByStatus", mustRMIs(t)(rs.ListRMIsByStatus(ctx, "proposed")), "RMI-ACME-001", "RMI-OTHER-001")
	assertIDs(t, "ListRMIsByStatus(none)", mustRMIs(t)(rs.ListRMIsByStatus(ctx, "completed")))

	// Auto-numbering reads through the remote list.
	next, err := svc.NextRMIID(ctx, "github.com/acme/acme")
	if err != nil || next != "RMI-ACME-002" {
		t.Fatalf("NextRMIID = %q, %v", next, err)
	}

	phases, err := rs.ListPhases(ctx, "INIT-ACME-001")
	if err != nil || len(phases) != 1 || phases[0].Title != "Foundation" {
		t.Fatalf("ListPhases = %+v, %v", phases, err)
	}
}

func TestProgramsReadOnly(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	ctx := context.Background()
	if _, err := fake.Service.CreateProgram(ctx, "PROG-X", "X", "acme", ""); err != nil {
		t.Fatal(err)
	}
	p, err := rs.GetProgram(ctx, "PROG-X")
	if err != nil || p.Name != "X" {
		t.Fatalf("GetProgram = %+v, %v", p, err)
	}
	if _, err := rs.GetProgram(ctx, "PROG-NOPE"); !errors.Is(err, remote.ErrNotFound) {
		t.Fatalf("GetProgram missing err = %v, want ErrNotFound", err)
	}
	if err := rs.CreateProgram(ctx, &store.Program{ID: "PROG-Y"}); !errors.Is(err, remote.ErrNotSupported) {
		t.Fatalf("CreateProgram err = %v, want ErrNotSupported", err)
	}
}

func TestRequestHeaders(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake, remote.WithUserAgent("test-agent/1"))
	if _, err := rs.ListInitiatives(context.Background()); err != nil {
		t.Fatal(err)
	}
	reqs := fake.Requests()
	if len(reqs) != 1 {
		t.Fatalf("requests = %d", len(reqs))
	}
	r := reqs[0]
	if r.Path != "/t/acme/api/v1/initiatives" {
		t.Errorf("path = %q", r.Path)
	}
	if got := r.Header.Get("Authorization"); got != "Bearer "+testToken {
		t.Errorf("Authorization = %q", got)
	}
	if got := r.Header.Get("User-Agent"); got != "test-agent/1" {
		t.Errorf("User-Agent = %q", got)
	}
	if got := r.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q", got)
	}
}

func TestDefaultUserAgent(t *testing.T) {
	fake := newFake(t)
	if _, err := newStore(t, fake).ListInitiatives(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ua := fake.Requests()[0].Header.Get("User-Agent"); !strings.HasPrefix(ua, "visionstudio/") {
		t.Errorf("User-Agent = %q, want visionstudio/ prefix", ua)
	}
}

// signingTransport stands in for an injected request signer (e.g. an
// AAuth launcher): it authenticates every request itself, so the store is
// configured with no token at all.
type signingTransport struct {
	calls atomic.Int32
	key   string
	next  http.RoundTripper
}

func (s *signingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	s.calls.Add(1)
	r = r.Clone(r.Context())
	r.Header.Set("X-API-Key", s.key)
	return s.next.RoundTrip(r)
}

func TestInjectedTransportIsUsed(t *testing.T) {
	fake := newFake(t)
	rt := &signingTransport{key: testToken, next: http.DefaultTransport}
	rs, err := remote.New(fake.URL, testTenant, remote.WithTransport(rt))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rs.ListInitiatives(context.Background()); err != nil {
		t.Fatalf("ListInitiatives via signing transport: %v", err)
	}
	if rt.calls.Load() != 1 {
		t.Fatalf("transport calls = %d, want 1", rt.calls.Load())
	}
	if got := fake.Requests()[0].Header.Get("Authorization"); got != "" {
		t.Errorf("Authorization = %q, want none when no token is configured", got)
	}
}

func TestInjectedHTTPClientIsUsed(t *testing.T) {
	fake := newFake(t)
	rt := &signingTransport{key: testToken, next: http.DefaultTransport}
	hc := &http.Client{Transport: rt, Timeout: 5 * time.Second}
	rs, err := remote.New(fake.URL, testTenant, remote.WithHTTPClient(hc))
	if err != nil {
		t.Fatal(err)
	}
	if err := rs.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	if rt.calls.Load() != 1 {
		t.Fatalf("client transport calls = %d, want 1", rt.calls.Load())
	}
}

func TestErrorMapping(t *testing.T) {
	fake := newFake(t)
	ctx := context.Background()

	bad, err := remote.New(fake.URL, testTenant, remote.WithToken("wrong"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = bad.ListInitiatives(ctx)
	if !errors.Is(err, remote.ErrUnauthorized) {
		t.Fatalf("bad token err = %v, want ErrUnauthorized", err)
	}
	if !strings.Contains(err.Error(), "cloud login") {
		t.Errorf("401 error lacks login guidance: %v", err)
	}

	other, err := remote.New(fake.URL, "other", remote.WithToken(testToken))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.ListInitiatives(ctx); !errors.Is(err, remote.ErrForbidden) {
		t.Fatalf("non-member err = %v, want ErrForbidden", err)
	}

	rs := newStore(t, fake)
	_, err = rs.GetInitiative(ctx, "INIT-NOPE-001")
	if !errors.Is(err, remote.ErrNotFound) {
		t.Fatalf("missing initiative err = %v, want ErrNotFound", err)
	}
	var apiErr *remote.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound || !strings.Contains(apiErr.Message, "not found") {
		t.Fatalf("APIError = %+v", apiErr)
	}
	before := len(fake.Requests())
	if _, err := rs.GetRMI(ctx, ""); !errors.Is(err, remote.ErrNotFound) || len(fake.Requests()) != before {
		t.Fatalf("empty RMI ID err = %v (requests sent: %d)", err, len(fake.Requests())-before)
	}
	if _, err := rs.GetRMI(ctx, "RMI-NOPE-001"); !errors.Is(err, remote.ErrNotFound) {
		t.Fatalf("missing RMI err = %v, want ErrNotFound", err)
	}

	// Service-layer validation surfaces as a 400 with the server's message.
	err = rs.CreateRMI(ctx, &store.RoadmapItem{ID: "RMI-X-001"})
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest || !strings.Contains(apiErr.Message, "required") {
		t.Fatalf("invalid create err = %v", err)
	}
}

func TestStatusMappingFromRawServer(t *testing.T) {
	cases := []struct {
		status int
		want   error
	}{
		{http.StatusUnauthorized, remote.ErrUnauthorized},
		{http.StatusForbidden, remote.ErrForbidden},
		{http.StatusNotFound, remote.ErrNotFound},
		{http.StatusConflict, remote.ErrConflict},
	}
	for _, tc := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(`{"error":"boom"}`))
		}))
		rs, err := remote.New(srv.URL, testTenant)
		if err != nil {
			t.Fatal(err)
		}
		_, err = rs.ListInitiatives(context.Background())
		srv.Close()
		if !errors.Is(err, tc.want) {
			t.Errorf("status %d: err = %v, want %v", tc.status, err, tc.want)
		}
		for _, other := range cases {
			if other.want != tc.want && errors.Is(err, other.want) {
				t.Errorf("status %d also matched %v", tc.status, other.want)
			}
		}
	}

	// Non-JSON error bodies are surfaced verbatim; 5xx matches no sentinel.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream exploded", http.StatusBadGateway)
	}))
	defer srv.Close()
	rs, err := remote.New(srv.URL, testTenant)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rs.ListInitiatives(context.Background())
	var apiErr *remote.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadGateway || apiErr.Message != "upstream exploded" {
		t.Fatalf("502 err = %v", err)
	}
}

func TestUnsupportedOperationsFailClearly(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	svc := service.New(rs)
	ctx := context.Background()

	if _, err := fake.Service.CreateInitiative(ctx, "INIT-ACME-001", "acme", "Launch", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	_, err := svc.TransitionInitiative(ctx, "INIT-ACME-001", "planned")
	if !errors.Is(err, remote.ErrNotSupported) || !strings.Contains(err.Error(), "UpdateInitiative is not supported in remote mode") {
		t.Fatalf("TransitionInitiative err = %v", err)
	}
	if _, err := svc.ListRepositories(ctx); !errors.Is(err, remote.ErrNotSupported) {
		t.Fatalf("ListRepositories err = %v", err)
	}
	if _, err := fake.Service.CreateRMI(ctx, "RMI-ACME-001", "github.com/acme/app", "INIT-ACME-001", "", "t", "", "task", "", true, 1, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ClaimRMI(ctx, "RMI-ACME-001", "w", "", time.Hour); !errors.Is(err, remote.ErrNotSupported) {
		t.Fatalf("ClaimRMI err = %v", err)
	}

	// Create-time fields the cloud create endpoint cannot carry are
	// rejected rather than silently dropped, before any request is sent.
	before := len(fake.Requests())
	err = rs.CreateInitiative(ctx, &store.Initiative{ID: "INIT-ACME-002", Title: "t", HomeRepo: "github.com/acme/app"})
	if !errors.Is(err, remote.ErrNotSupported) || !strings.Contains(err.Error(), "home_repo") {
		t.Fatalf("CreateInitiative with home_repo err = %v", err)
	}
	err = rs.CreateRMI(ctx, &store.RoadmapItem{ID: "RMI-ACME-009", RepositoryID: "r", Title: "t", ItemType: "task", Status: "completed"})
	if !errors.Is(err, remote.ErrNotSupported) {
		t.Fatalf("CreateRMI completed err = %v", err)
	}
	if len(fake.Requests()) != before {
		t.Fatalf("unsupported operations reached the server: %+v", fake.Requests())
	}
}

func TestContextCancellation(t *testing.T) {
	fake := newFake(t)
	rs := newStore(t, fake)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := rs.ListInitiatives(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestNewURLAndTenant(t *testing.T) {
	cases := []struct {
		url, tenant       string
		wantBase, wantTen string
		wantErr           bool
	}{
		{url: "https://cloud.example.com", tenant: "acme", wantBase: "https://cloud.example.com", wantTen: "acme"},
		{url: "https://cloud.example.com/", tenant: "acme", wantBase: "https://cloud.example.com", wantTen: "acme"},
		{url: "https://cloud.example.com/t/acme", wantBase: "https://cloud.example.com", wantTen: "acme"},
		{url: "https://cloud.example.com/t/acme/api/v1/", tenant: "acme", wantBase: "https://cloud.example.com", wantTen: "acme"},
		{url: "https://example.com/vs/t/acme", wantBase: "https://example.com/vs", wantTen: "acme"},
		{url: "https://cloud.example.com/t/acme", tenant: "other", wantErr: true},
		{url: "https://cloud.example.com", wantErr: true},
		{url: "cloud.example.com", tenant: "acme", wantErr: true},
		{url: "ftp://cloud.example.com", tenant: "acme", wantErr: true},
		{url: "https://user:pw@cloud.example.com", tenant: "acme", wantErr: true}, // #nosec G101 -- asserts embedded credentials are rejected
		{url: "https://cloud.example.com", tenant: "a/b", wantErr: true},
		{url: "", tenant: "acme", wantErr: true},
	}
	for _, tc := range cases {
		rs, err := remote.New(tc.url, tc.tenant)
		if tc.wantErr {
			if err == nil {
				t.Errorf("New(%q, %q) succeeded, want error", tc.url, tc.tenant)
			}
			continue
		}
		if err != nil {
			t.Errorf("New(%q, %q): %v", tc.url, tc.tenant, err)
			continue
		}
		if rs.BaseURL() != tc.wantBase || rs.Tenant() != tc.wantTen {
			t.Errorf("New(%q, %q) = %q, %q; want %q, %q", tc.url, tc.tenant, rs.BaseURL(), rs.Tenant(), tc.wantBase, tc.wantTen)
		}
	}
}

func TestBasePathPrefixPreserved(t *testing.T) {
	fake := newFake(t)
	mux := http.NewServeMux()
	mux.Handle("/prefix/", http.StripPrefix("/prefix", fake.Config.Handler))
	front := httptest.NewServer(mux)
	defer front.Close()

	rs, err := remote.New(front.URL+"/prefix/t/"+testTenant, "", remote.WithToken(testToken))
	if err != nil {
		t.Fatal(err)
	}
	if err := rs.Ping(context.Background()); err != nil {
		t.Fatalf("Ping through prefix: %v", err)
	}
}

func mustRMIs(t *testing.T) func([]*store.RoadmapItem, error) []*store.RoadmapItem {
	return func(r []*store.RoadmapItem, err error) []*store.RoadmapItem {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
}

func assertIDs(t *testing.T, label string, rmis []*store.RoadmapItem, want ...string) {
	t.Helper()
	got := map[string]bool{}
	for _, r := range rmis {
		got[r.ID] = true
	}
	if len(got) != len(want) {
		t.Fatalf("%s: got %v, want %v", label, keys(got), want)
	}
	for _, w := range want {
		if !got[w] {
			t.Fatalf("%s: got %v, want %v", label, keys(got), want)
		}
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
