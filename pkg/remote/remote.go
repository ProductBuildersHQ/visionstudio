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
// The cloud API's first slice covers initiatives and RMIs (plus read-only
// phases and programs). Every other store method returns an error wrapping
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
	// ErrConflict: the write conflicts with existing state (409).
	ErrConflict = errors.New("remote: conflict")
	// ErrNotSupported: the operation has no cloud API endpoint yet.
	ErrNotSupported = errors.New("remote: not supported in remote mode")
)

// APIError is a non-2xx response from the cloud API.
type APIError struct {
	Method     string
	Path       string
	StatusCode int
	// Message is the server's {"error": "..."} text, or the raw body when
	// the response was not the JSON error shape.
	Message string
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	s := fmt.Sprintf("remote %s %s: %d: %s", e.Method, e.Path, e.StatusCode, msg)
	switch e.StatusCode {
	case http.StatusUnauthorized:
		s += " (missing or invalid credential; run 'visionstudio cloud login' or set VISIONSTUDIO_REMOTE_TOKEN)"
	case http.StatusForbidden:
		s += " (credential is not a member of this tenant)"
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

func notSupported(op string) error { return &NotSupportedError{Op: op} }

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

// Ping verifies the URL, credential, and tenant membership with one cheap
// authenticated read. A nil error means the store is usable.
func (s *Store) Ping(ctx context.Context) error {
	var resp struct {
		Programs []*store.Program `json:"programs"`
	}
	return s.do(ctx, http.MethodGet, "/programs", nil, nil, &resp)
}

func defaultUserAgent() string {
	v, _, _ := strings.Cut(version.String(), " ")
	return "visionstudio/" + v + " (remote-store)"
}

// endpoint builds base + /t/{tenant}/api/v1 + rel (+ query).
func (s *Store) endpoint(rel string, q url.Values) string {
	u := *s.base
	u.Path = u.Path + "/t/" + url.PathEscape(s.tenant) + "/api/v1" + rel
	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// do performs one request. body (if non-nil) is JSON-encoded; out (if
// non-nil) receives the decoded 2xx response.
func (s *Store) do(ctx context.Context, method, rel string, q url.Values, body, out any) error {
	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("remote: encode request: %w", err)
		}
		rdr = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, s.endpoint(rel, q), rdr)
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
		return &APIError{Method: method, Path: rel, StatusCode: resp.StatusCode, Message: errorMessage(data)}
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("remote %s %s: decode response: %w", method, rel, err)
	}
	return nil
}

// errorMessage extracts {"error": "..."} or falls back to the trimmed body.
func errorMessage(data []byte) string {
	var e struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &e); err == nil && e.Error != "" {
		return e.Error
	}
	msg := strings.TrimSpace(string(data))
	if len(msg) > 512 {
		msg = msg[:512] + "..."
	}
	return msg
}
