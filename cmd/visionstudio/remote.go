package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	config "github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
	"github.com/ProductBuildersHQ/visionstudio/pkg/remote"
)

// Remote mode (RMI-VISIONSTUDIO-556): when a cloud URL is given, every
// command that opens the store through connectService — the CLI's entity
// commands and `visionstudio mcp` — runs against a VisionStudio Cloud tenant
// through pkg/remote instead of the local Dolt database. Remote mode is
// strictly opt-in per invocation (flag or environment); a URL stored by
// `cloud login` never switches commands to remote on its own.
const (
	envRemoteURL    = "VISIONSTUDIO_REMOTE_URL"
	envRemoteTenant = "VISIONSTUDIO_REMOTE_TENANT"
	envRemoteToken  = "VISIONSTUDIO_REMOTE_TOKEN" // #nosec G101 -- environment variable name, not a credential
)

// remoteSettings is a resolved remote-mode configuration. Source fields
// describe where each value came from, for `cloud status`.
type remoteSettings struct {
	RawURL       string // as given (may include /t/<tenant>)
	BaseURL      string // canonical cloud root, the credentials key
	URLSource    string
	Tenant       string
	TenantSource string
	Token        string
	TokenSource  string
}

// addRemoteFlags registers the remote-mode persistent flags on the root.
// The tenant flag is --remote-tenant (not --tenant) because `sync` and
// `pull` already define a local --tenant with different meaning.
func addRemoteFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("remote", "", "VisionStudio Cloud URL: run against a cloud tenant instead of the local database (default: $"+envRemoteURL+"; may end in /t/<tenant>)")
	cmd.PersistentFlags().String("remote-tenant", "", "Cloud tenant slug for --remote (default: $"+envRemoteTenant+", the URL's /t/<tenant>, or the tenant saved by 'cloud login')")
}

// remoteURLFlag returns the remote URL from --remote or the environment,
// and its source; "" when remote mode was not requested.
func remoteURLFlag(cmd *cobra.Command) (string, string) {
	if v, _ := cmd.Flags().GetString("remote"); strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v), "--remote flag"
	}
	if v := strings.TrimSpace(os.Getenv(envRemoteURL)); v != "" {
		return v, "$" + envRemoteURL
	}
	return "", ""
}

// resolveRemote resolves remote mode for a command. ok is false when no
// remote URL was requested (local mode). fallbackURL, if non-empty, is used
// when neither the flag nor the environment names a URL (the cloud
// subcommands pass the URL saved by `cloud login`).
func resolveRemote(cmd *cobra.Command, fallbackURL string) (*remoteSettings, bool, error) {
	raw, src := remoteURLFlag(cmd)
	if raw == "" && fallbackURL != "" {
		raw, src = fallbackURL, "config (cloud login)"
	}
	if raw == "" {
		return nil, false, nil
	}
	base, urlTenant, err := remote.ParseBaseURL(raw)
	if err != nil {
		return nil, false, err
	}
	rs := &remoteSettings{RawURL: raw, BaseURL: base.String(), URLSource: src}

	creds, err := config.LoadCredentials()
	if err != nil {
		return nil, false, fmt.Errorf("load cloud credentials: %w", err)
	}
	stored, hasStored := creds.Get(rs.BaseURL)

	switch {
	case flagString(cmd, "remote-tenant") != "":
		rs.Tenant, rs.TenantSource = flagString(cmd, "remote-tenant"), "--remote-tenant flag"
	case strings.TrimSpace(os.Getenv(envRemoteTenant)) != "":
		rs.Tenant, rs.TenantSource = strings.TrimSpace(os.Getenv(envRemoteTenant)), "$"+envRemoteTenant
	case urlTenant != "":
		rs.Tenant, rs.TenantSource = urlTenant, "URL"
	case hasStored && stored.Tenant != "":
		rs.Tenant, rs.TenantSource = stored.Tenant, "credentials (cloud login)"
	}

	switch {
	case strings.TrimSpace(os.Getenv(envRemoteToken)) != "":
		rs.Token, rs.TokenSource = strings.TrimSpace(os.Getenv(envRemoteToken)), "$"+envRemoteToken
	case hasStored && stored.Token != "":
		rs.Token, rs.TokenSource = stored.Token, "credentials (cloud login)"
	}
	return rs, true, nil
}

func flagString(cmd *cobra.Command, name string) string {
	v, _ := cmd.Flags().GetString(name)
	return strings.TrimSpace(v)
}

// newRemoteStore builds the HTTP store for resolved settings. A missing
// token is not an error here (a signing transport may authenticate
// instead); the server's 401 carries the login guidance.
func newRemoteStore(rs *remoteSettings, opts ...remote.Option) (*remote.Store, error) {
	opts = append([]remote.Option{remote.WithToken(rs.Token)}, opts...)
	st, err := remote.New(rs.RawURL, rs.Tenant, opts...)
	if err != nil {
		if rs.Tenant == "" {
			return nil, fmt.Errorf("%w (set --remote-tenant or $%s, or save one with 'visionstudio cloud login --remote-tenant <slug>')", err, envRemoteTenant)
		}
		return nil, err
	}
	return st, nil
}

// maskToken shows enough of a credential to identify it, never the secret.
func maskToken(tok string) string {
	if tok == "" {
		return "(none)"
	}
	if len(tok) < 16 {
		return "****"
	}
	prefix := tok[:4]
	if strings.HasPrefix(tok, "vsc_live_") {
		prefix = "vsc_live_"
	}
	return prefix + "…" + tok[len(tok)-4:]
}
