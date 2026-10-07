package cliconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialsRoundTripAndMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", CredentialsFileName)

	empty, err := LoadCredentialsFrom(path)
	if err != nil {
		t.Fatalf("load missing file: %v", err)
	}
	if len(empty.Cloud) != 0 {
		t.Fatalf("missing file yielded %+v", empty)
	}

	c := &Credentials{}
	c.Set("https://cloud.example.com", CloudCredential{Token: "vsc_live_x_y", Tenant: "acme"}) // #nosec G101 -- fake credential
	if err := c.SaveTo(path); err != nil {
		t.Fatal(err)
	}
	assertMode(t, path, 0o600)

	// An existing file with a looser mode is tightened on save.
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.SaveTo(path); err != nil {
		t.Fatal(err)
	}
	assertMode(t, path, 0o600)

	got, err := LoadCredentialsFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	cred, ok := got.Get("https://cloud.example.com")
	if !ok || cred.Token != "vsc_live_x_y" || cred.Tenant != "acme" {
		t.Fatalf("Get = %+v, %v", cred, ok)
	}
	if !got.Delete("https://cloud.example.com") || got.Delete("https://cloud.example.com") {
		t.Fatal("Delete should report true once, then false")
	}
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("mode = %o, want %o", got, want)
	}
}
