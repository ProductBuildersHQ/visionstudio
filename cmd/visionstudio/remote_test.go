package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	config "github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
	"github.com/ProductBuildersHQ/visionstudio/pkg/remote/remotetest"
)

const (
	cliTestToken  = "vsc_live_abc123_supersecretvalue" // #nosec G101 -- fake credential for the in-process test server
	cliTestTenant = "acme"
)

// isolateRemoteEnv points HOME at a temp dir and clears every variable that
// could select a backend, so tests never touch the developer's real
// config, credentials, or database.
func isolateRemoteEnv(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, k := range []string{envRemoteURL, envRemoteTenant, envRemoteToken, "VISIONSTUDIO_DSN", "VISIONSTUDIO_DATA"} {
		t.Setenv(k, "")
	}
	return home
}

func runCLI(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	root := rootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return out.String(), err
}

func newFakeCloud(t *testing.T) *remotetest.Server {
	t.Helper()
	fake := remotetest.NewServer(cliTestToken, cliTestTenant)
	t.Cleanup(fake.Close)
	return fake
}

func TestRemoteModeInitiativeCreate(t *testing.T) {
	isolateRemoteEnv(t)
	fake := newFakeCloud(t)
	t.Setenv(envRemoteToken, cliTestToken)

	out, err := runCLI(t, "", "initiative", "create",
		"--id", "INIT-ACME-001", "--title", "Launch", "--org", "acme", "--workflow", "pbhq-lite",
		"--remote", fake.URL, "--remote-tenant", cliTestTenant)
	if err != nil {
		t.Fatalf("initiative create: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Created initiative: INIT-ACME-001 (proposed)") {
		t.Fatalf("output = %q", out)
	}
	if _, err := fake.Service.GetInitiative(context.Background(), "INIT-ACME-001"); err != nil {
		t.Fatalf("initiative not created in the tenant: %v", err)
	}
}

func TestRemoteModeFromEnvAndURLTenant(t *testing.T) {
	isolateRemoteEnv(t)
	fake := newFakeCloud(t)
	t.Setenv(envRemoteToken, cliTestToken)
	t.Setenv(envRemoteURL, fake.URL+"/t/"+cliTestTenant)

	out, err := runCLI(t, "", "initiative", "create", "--id", "INIT-ACME-002", "--title", "Env", "--workflow", "pbhq-lite")
	if err != nil {
		t.Fatalf("initiative create: %v\n%s", err, out)
	}
	if _, err := fake.Service.GetInitiative(context.Background(), "INIT-ACME-002"); err != nil {
		t.Fatalf("initiative not created in the tenant: %v", err)
	}
}

func TestRemoteModeUnsupportedIsClear(t *testing.T) {
	isolateRemoteEnv(t)
	fake := newFakeCloud(t)
	t.Setenv(envRemoteToken, cliTestToken)

	// `rmi create` validates --repo against the registry, which has no
	// cloud endpoint yet: the error must say so, not "not found".
	_, err := runCLI(t, "", "rmi", "create", "--repo", "github.com/acme/app", "--title", "t", "--type", "task",
		"--remote", fake.URL, "--remote-tenant", cliTestTenant)
	if err == nil || !strings.Contains(err.Error(), "not supported in remote mode") {
		t.Fatalf("rmi create err = %v, want not-supported", err)
	}
}

func TestRemoteModeFlagErrors(t *testing.T) {
	isolateRemoteEnv(t)
	fake := newFakeCloud(t)

	_, err := runCLI(t, "", "initiative", "list", "--remote", fake.URL, "--remote-tenant", cliTestTenant, "--dsn", "root:@tcp(127.0.0.1:1)/x")
	if err == nil || !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("--remote with --dsn err = %v", err)
	}

	_, err = runCLI(t, "", "initiative", "list", "--remote", fake.URL)
	if err == nil || !strings.Contains(err.Error(), "--remote-tenant") {
		t.Fatalf("missing tenant err = %v", err)
	}

	// No credential at all: the server's 401 surfaces with login guidance.
	_, err = runCLI(t, "", "initiative", "list", "--remote", fake.URL, "--remote-tenant", cliTestTenant)
	if err == nil || !strings.Contains(err.Error(), "cloud login") {
		t.Fatalf("no credential err = %v", err)
	}
}

func TestCloudLoginStatusLogout(t *testing.T) {
	home := isolateRemoteEnv(t)
	fake := newFakeCloud(t)

	// A bad key is rejected and nothing is saved.
	out, err := runCLI(t, "wrong-key\n", "cloud", "login", "--with-token", "--remote", fake.URL, "--remote-tenant", cliTestTenant)
	if err == nil || !strings.Contains(err.Error(), "credential not saved") {
		t.Fatalf("bad login err = %v\n%s", err, out)
	}
	credPath := filepath.Join(home, config.Dir, config.CredentialsFileName)
	if _, statErr := os.Stat(credPath); !os.IsNotExist(statErr) {
		t.Fatalf("credentials written after failed login: %v", statErr)
	}

	out, err = runCLI(t, cliTestToken+"\n", "cloud", "login", "--with-token", "--remote", fake.URL, "--remote-tenant", cliTestTenant)
	if err != nil {
		t.Fatalf("login: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Verified access to tenant acme") {
		t.Fatalf("login output = %q", out)
	}
	info, err := os.Stat(credPath)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("credentials mode = %o, want 600", perm)
	}
	creds, err := config.LoadCredentialsFrom(credPath)
	if err != nil {
		t.Fatal(err)
	}
	if c, ok := creds.Get(fake.URL); !ok || c.Token != cliTestToken || c.Tenant != cliTestTenant {
		t.Fatalf("stored credential = %+v, %v", c, ok)
	}

	// The stored credential and tenant are used by remote mode with no env.
	out, err = runCLI(t, "", "initiative", "create", "--id", "INIT-ACME-003", "--title", "Stored", "--workflow", "pbhq-lite", "--remote", fake.URL)
	if err != nil {
		t.Fatalf("create with stored credential: %v\n%s", err, out)
	}

	// Status (no --remote: reports on the logged-in URL) masks the key.
	out, err = runCLI(t, "", "cloud", "status")
	if err != nil {
		t.Fatalf("status: %v\n%s", err, out)
	}
	if strings.Contains(out, cliTestToken) {
		t.Fatalf("status leaked the credential: %q", out)
	}
	for _, want := range []string{"inactive for this invocation", fake.URL, "acme", "vsc_live_…alue", "Access:     ok"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status output missing %q:\n%s", want, out)
		}
	}

	// Logging in never switches plain commands to remote mode: without
	// --remote, `initiative list` would use the local backend. Resolution
	// alone shows that.
	cmd := rootCmd()
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := resolveRemote(cmd, ""); err != nil || ok {
		t.Fatalf("resolveRemote without --remote = ok %v, err %v; want local mode", ok, err)
	}

	out, err = runCLI(t, "", "cloud", "logout")
	if err != nil || !strings.Contains(out, "Removed stored credential") {
		t.Fatalf("logout: %v\n%s", err, out)
	}
	creds, err = config.LoadCredentialsFrom(credPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := creds.Get(fake.URL); ok {
		t.Fatal("credential still present after logout")
	}
}

func TestMaskToken(t *testing.T) {
	cases := map[string]string{
		"":                             "(none)",
		"short":                        "****",
		"vsc_live_id_0123456789abcdef": "vsc_live_…cdef",
		"eyJhbGciOiJIUzI1NiJ9.payload": "eyJh…load",
	}
	for in, want := range cases {
		if got := maskToken(in); got != want {
			t.Errorf("maskToken(%q) = %q, want %q", in, got, want)
		}
	}
}
