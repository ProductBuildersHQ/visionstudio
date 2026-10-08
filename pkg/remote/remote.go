// Package remote is an HTTP-backed implementation of store.Store that talks
// to a VisionStudio Cloud tenant's CRUD API (/t/{tenant}/api/v1/...).
//
// Because pkg/service and every adapter above it (the CLI commands and
// pkg/mcpserver) depend only on the store.Store interface, wrapping a remote
// Store in service.New gives the unmodified CLI and MCP server a cloud
// backend — functional parity by construction, not a second implementation:
//
//	rs, err := remote.New("https://cloud.example.com", "acme",
//		remote.WithToken(os.Getenv("VISIONSTUDIO_REMOTE_TOKEN")))
//	if err != nil { ... }
//	svc := service.New(rs)
//	err = mcpserver.Run(ctx, svc)
//
// The cloud tenant API covers the agent-work surface: programs, initiatives,
// phases, RMIs and their dependencies, assignments (leases), delivery
// evidence, spec workflows and documents, judge results, and read access to
// releases and repositories. Store methods with no endpoint (initiative
// dependencies, repository and release writes, organizations, people, spec
// workflow writes, maturity, DevX and PRISM data) return an error wrapping
// ErrNotSupported rather than faking a result.
//
// Authentication is pluggable. WithToken sends "Authorization: Bearer
// <token>" (a VisionStudio Cloud API key or a session JWT). Callers that sign
// requests some other way — e.g. an agent launcher that supplies an
// AAuth-signing http.RoundTripper — pass it via WithTransport or
// WithHTTPClient and may omit the token entirely.
package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
	"github.com/ProductBuildersHQ/visionstudio/pkg/version"
)

// Sentinel errors. Use errors.Is against an error returned by any Store
// method; *APIError and *NotSupportedError implement the mapping.
var (
	// ErrUnauthorized: the server rejected or did not receive a credential (401).
	ErrUnauthorized = errors.New("remote: unauthorized")
	// ErrForbidden: the credential is valid but not a member of the tenant (403).
	ErrForbidden = errors.New("remote: forbidden")
	// ErrNotFound: the entity or route does not exist (404).
	ErrNotFound = errors.New("remote: not found")
	// ErrConflict: the write conflicts with existing state (409), including
	// a lease conflict.
	ErrConflict = errors.New("remote: conflict")
	// ErrLeaseConflict: a 409 with code LEASE_CONFLICT — another session
	// holds a live lease on the RMI, or the assignment is no longer active.
	// Errors matching it also match ErrConflict.
	ErrLeaseConflict = errors.New("remote: lease conflict")
	// ErrInvalid: the server rejected the request as invalid (400).
	ErrInvalid = errors.New("remote: invalid request")
	// ErrInvalidReference: a 400 with code INVALID_REFERENCE — the body
	// references an entity that does not exist. Errors matching it also
	// match ErrInvalid.
	ErrInvalidReference = errors.New("remote: invalid reference")
	// ErrNotSupported: the operation has no cloud API endpoint yet.
	ErrNotSupported = errors.New("remote: not supported in remote mode")
)

// Error codes the cloud API sends in an error response's "code" field. The
// HTTP status is authoritative; the code refines it. Service errors may also
// carry a pcerr code (e.g. STATE_CONFLICT) verbatim.
const (
	CodeUnauthenticated  = "UNAUTHENTICATED"
	CodeForbidden        = "FORBIDDEN"
	CodeInvalid          = "INVALID"
	CodeInvalidReference = "INVALID_REFERENCE"
	CodeNotFound         = "NOT_FOUND"
	CodeConflict         = "CONFLICT"
	CodeLeaseConflict    = "LEASE_CONFLICT"
	CodeInternal         = "INTERNAL"
)

// APIError is a non-2xx response from the cloud API.
type APIError struct {
	Method     string
	Path       string
	StatusCode int
	// Code is the server's machine-readable error code (see the Code*
	// constants), or empty when the response carried none.
	Code string
	// Message is the server's {"error": "..."} text, or the raw body when
	// the response was not the JSON error shape.
	Message string
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	status := fmt.Sprint(e.StatusCode)
	if e.Code != "" {
		status += " " + e.Code
	}
	s := fmt.Sprintf("remote %s %s: %s: %s", e.Method, e.Path, status, msg)
	switch e.StatusCode {
	case http.StatusUnauthorized:
		s += " (missing or invalid credential; run 'visionstudio cloud login' or set VISIONSTUDIO_REMOTE_TOKEN)"
	case http.StatusForbidden:
		s += " (credential is not a member of this tenant)"
	case http.StatusConflict:
		if e.Code == CodeLeaseConflict {
			s += " (another session holds the lease, or the assignment is no longer active)"
		}
	}
	return s
}

// Is maps the HTTP status to the package sentinels.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrUnauthorized:
		return e.StatusCode == http.StatusUnauthorized
	case ErrForbidden:
		return e.StatusCode == http.StatusForbidden
	case ErrNotFound:
		return e.StatusCode == http.StatusNotFound
	case ErrConflict:
		return e.StatusCode == http.StatusConflict
	case ErrLeaseConflict:
		return e.StatusCode == http.StatusConflict && e.Code == CodeLeaseConflict
	case ErrInvalid:
		return e.StatusCode == http.StatusBadRequest
	case ErrInvalidReference:
		return e.StatusCode == http.StatusBadRequest && e.Code == CodeInvalidReference
	}
	return false
}

// NotSupportedError reports a store operation that remote mode cannot
// perform yet.
type NotSupportedError struct {
	Op     string
	Reason string
}

func (e *NotSupportedError) Error() string {
	reason := e.Reason
	if reason == "" {
		reason = "the VisionStudio Cloud API has no endpoint for it yet"
	}
	return fmt.Sprintf("%s is not supported in remote mode: %s", e.Op, reason)
}

// Unwrap lets errors.Is(err, ErrNotSupported) match.
func (e *NotSupportedError) Unwrap() error { return ErrNotSupported }

// notServed reports an operation the cloud API deliberately does not serve
// to remote stores, naming the area so the message says why.
func notServed(op, area string) error {
	return &NotSupportedError{Op: op, Reason: "the VisionStudio Cloud API does not serve " + area + " to remote clients"}
}

// Option configures a Store.
type Option func(*Store)

// WithToken sets a bearer credential sent as "Authorization: Bearer <token>".
// An empty token sends no Authorization header (e.g. when the transport
// signs requests itself).
func WithToken(token string) Option {
	return func(s *Store) { s.token = strings.TrimSpace(token) }
}

// WithHTTPClient sets the http.Client used for every request. Its
// Transport may add or replace authentication (e.g. request signing).
func WithHTTPClient(hc *http.Client) Option {
	return func(s *Store) {
		if hc != nil {
			s.hc = hc
		}
	}
}

// WithTransport sets the http.RoundTripper used for every request, keeping
// the default client timeout. Applied after WithHTTPClient it replaces that
// client's transport on a copy, never mutating the caller's client.
func WithTransport(rt http.RoundTripper) Option {
	return func(s *Store) {
		if rt == nil {
			return
		}
		c := *s.hc
		c.Transport = rt
		s.hc = &c
	}
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(s *Store) {
		if ua != "" {
			s.userAgent = ua
		}
	}
}

// DefaultTimeout bounds each request when no custom client is supplied.
const DefaultTimeout = 30 * time.Second

// maxResponseBytes caps a response body read into memory.
const maxResponseBytes = 64 << 20

// Store implements store.Store against a VisionStudio Cloud tenant.
type Store struct {
	base      *url.URL // scheme://host[/prefix], no trailing slash, no /t/{tenant}
	tenant    string
	token     string
	hc        *http.Client
	userAgent string
}

var _ store.Store = (*Store)(nil)

// New returns a Store for the tenant at baseURL. baseURL may be the cloud
// root ("https://cloud.example.com") or already include the tenant segment
// ("https://cloud.example.com/t/acme"); in the latter case tenant may be
// empty, and must match if given.
func New(baseURL, tenant string, opts ...Option) (*Store, error) {
	base, urlTenant, err := ParseBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	tenant = strings.TrimSpace(tenant)
	switch {
	case tenant == "" && urlTenant == "":
		return nil, errors.New("remote: a tenant slug is required (pass one, or use a URL ending in /t/<tenant>)")
	case tenant == "":
		tenant = urlTenant
	case urlTenant != "" && urlTenant != tenant:
		return nil, fmt.Errorf("remote: tenant %q conflicts with tenant %q in URL %s", tenant, urlTenant, baseURL)
	}
	if strings.ContainsAny(tenant, "/?#") {
		return nil, fmt.Errorf("remote: invalid tenant slug %q", tenant)
	}
	s := &Store{
		base:      base,
		tenant:    tenant,
		hc:        &http.Client{Timeout: DefaultTimeout},
		userAgent: defaultUserAgent(),
	}
	for _, o := range opts {
		o(s)
	}
	return s, nil
}

// ParseBaseURL validates a cloud URL and splits off an optional trailing
// /t/{tenant} (and anything after it, e.g. /api/v1). It returns the cloud
// root without a trailing slash and the tenant slug, if present.
func ParseBaseURL(raw string) (*url.URL, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, "", errors.New("remote: base URL is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, "", fmt.Errorf("remote: invalid base URL %q: %w", raw, err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, "", fmt.Errorf("remote: base URL %q must use http or https", raw)
	}
	if u.Host == "" {
		return nil, "", fmt.Errorf("remote: base URL %q has no host", raw)
	}
	if u.User != nil {
		return nil, "", fmt.Errorf("remote: base URL must not embed credentials")
	}
	u.RawQuery, u.Fragment, u.RawFragment = "", "", ""
	path := strings.TrimRight(u.Path, "/")
	var tenant string
	if i := strings.Index(path+"/", "/t/"); i >= 0 {
		rest := strings.TrimPrefix(path[i:], "/t/")
		tenant, _, _ = strings.Cut(rest, "/")
		path = path[:i]
	}
	u.Path, u.RawPath = path, ""
	return u, tenant, nil
}

// BaseURL returns the cloud root URL (without the tenant segment).
func (s *Store) BaseURL() string { return s.base.String() }

// Tenant returns the tenant slug requests are scoped to.
func (s *Store) Tenant() string { return s.tenant }

// Principal is the authenticated caller as the cloud reports it.
type Principal struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
	// AuthMethod is session, access_token, api_key, or dev.
	AuthMethod string `json:"auth_method,omitempty"`
}

// Membership is one tenant the principal belongs to.
type Membership struct {
	Tenant string `json:"tenant"`
	Name   string `json:"name,omitempty"`
	Role   string `json:"role,omitempty"`
}

// Identity is the body of GET /api/v1/me: the principal and every tenant
// membership.
type Identity struct {
	Principal   Principal    `json:"principal"`
	Memberships []Membership `json:"memberships"`
}

// TenantIdentity is the body of GET /t/{tenant}/api/v1/me: the principal
// and its role in the store's tenant.
type TenantIdentity struct {
	Principal Principal `json:"principal"`
	Tenant    string    `json:"tenant"`
	Role      string    `json:"role,omitempty"`
}

// Ping verifies the URL, credential, and tenant membership with one cheap
// authenticated read (GET /t/{tenant}/api/v1/me). A nil error means the
// store is usable.
func (s *Store) Ping(ctx context.Context) error {
	_, err := s.TenantMe(ctx)
	return err
}

// TenantMe returns the caller's principal and role in the store's tenant.
func (s *Store) TenantMe(ctx context.Context) (*TenantIdentity, error) {
	var out TenantIdentity
	if err := s.do(ctx, http.MethodGet, "/me", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Me returns the caller's principal and every tenant membership
// (GET /api/v1/me, outside any tenant).
func (s *Store) Me(ctx context.Context) (*Identity, error) {
	var out Identity
	if err := s.request(ctx, http.MethodGet, "/api/v1/me", nil, nil, &out); err != nil {
		return nil, err
	}
	if out.Memberships == nil {
		out.Memberships = []Membership{}
	}
	return &out, nil
}

func defaultUserAgent() string {
	v, _, _ := strings.Cut(version.String(), " ")
	return "visionstudio/" + v + " (remote-store)"
}

// idPath joins a collection path and an entity ID, escaping the ID as one
// path segment (a "/" in it is sent as %2F, which the API accepts).
func idPath(collection, id string) string {
	return collection + "/" + url.PathEscape(id)
}

// endpoint builds base + rawRel (+ query). rawRel is an already-escaped
// path relative to the cloud root.
func (s *Store) endpoint(rawRel string, q url.Values) (string, error) {
	u := *s.base
	raw := s.base.EscapedPath() + rawRel
	p, err := url.PathUnescape(raw)
	if err != nil {
		return "", fmt.Errorf("remote: invalid request path %q: %w", rawRel, err)
	}
	u.Path, u.RawPath = p, raw
	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

// do performs one request against the tenant API: rel is an escaped path
// relative to /t/{tenant}/api/v1. body (if non-nil) is JSON-encoded; out (if
// non-nil) receives the decoded 2xx response.
func (s *Store) do(ctx context.Context, method, rel string, q url.Values, body, out any) error {
	return s.requestAs(ctx, method, "/t/"+url.PathEscape(s.tenant)+"/api/v1"+rel, rel, q, body, out)
}

// request performs one request at an escaped path relative to the cloud
// root (for endpoints outside a tenant).
func (s *Store) request(ctx context.Context, method, rawRel string, q url.Values, body, out any) error {
	return s.requestAs(ctx, method, rawRel, rawRel, q, body, out)
}

// requestAs performs the request; label is the path errors report.
func (s *Store) requestAs(ctx context.Context, method, rawRel, label string, q url.Values, body, out any) error {
	rel := label
	target, err := s.endpoint(rawRel, q)
	if err != nil {
		return err
	}
	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("remote: encode request: %w", err)
		}
		rdr = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, rdr)
	if err != nil {
		return fmt.Errorf("remote: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}

	resp, err := s.hc.Do(req)
	if err != nil {
		return fmt.Errorf("remote %s %s: %w", method, rel, err)
	}
	defer func() {
		// Drain so the connection can be reused; a failure here only
		// costs connection reuse, never correctness.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseBytes))
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("remote %s %s: read response: %w", method, rel, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg, code := errorMessage(data)
		return &APIError{Method: method, Path: rel, StatusCode: resp.StatusCode, Code: code, Message: msg}
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("remote %s %s: decode response: %w", method, rel, err)
	}
	return nil
}

// errorMessage extracts {"error": "...", "code": "..."} or falls back to the
// trimmed body and no code.
func errorMessage(data []byte) (msg, code string) {
	var e struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if err := json.Unmarshal(data, &e); err == nil && e.Error != "" {
		return e.Error, e.Code
	}
	msg = strings.TrimSpace(string(data))
	if len(msg) > 512 {
		msg = msg[:512] + "..."
	}
	return msg, ""
}
