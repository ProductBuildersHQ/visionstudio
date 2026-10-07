package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	config "github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
)

// verifyTimeout bounds the credential check in login/status.
const verifyTimeout = 15 * time.Second

func cloudLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store a VisionStudio Cloud API key for remote mode",
		Long: `Store a VisionStudio Cloud credential so CLI commands and 'visionstudio mcp'
can run against a cloud tenant with --remote.

The cloud URL comes from --remote (or $VISIONSTUDIO_REMOTE_URL, or the URL of a
previous login). The credential is a VisionStudio Cloud API key (vsc_live_...),
created in the cloud web app or via POST /api/v1/me/api-keys; a session JWT
also works. It is read, in order, from:

  --with-token        standard input (e.g. piped from a password manager)
  $VISIONSTUDIO_REMOTE_TOKEN
  an interactive prompt (input hidden)

When a tenant is known (--remote-tenant, $VISIONSTUDIO_REMOTE_TENANT, or a URL
ending in /t/<tenant>), the credential is verified against that tenant before it
is saved, and the tenant is remembered as the default for this URL.

The credential is stored in ~/.productbuildershq/visionstudio/credentials.json
with owner-only (0600) permissions, separate from config.json. A browser-based
device login flow is planned; until then, paste an API key.

Logging in never switches commands to remote mode by itself: pass --remote <url>
(or set $VISIONSTUDIO_REMOTE_URL) per invocation.`,
		Example: `  visionstudio cloud login --remote https://cloud.example.com --remote-tenant acme
  op read op://vault/visionstudio/key | visionstudio cloud login --remote https://cloud.example.com/t/acme --with-token`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			rs, ok, err := resolveRemote(cmd, cfg.Cloud.BaseURL)
			if err != nil {
				return err
			}
			if !ok {
				return errors.New("no cloud URL: pass --remote <url> or set $" + envRemoteURL)
			}

			withToken, _ := cmd.Flags().GetBool("with-token")
			token, err := readLoginToken(cmd, withToken)
			if err != nil {
				return err
			}
			rs.Token = token

			noVerify, _ := cmd.Flags().GetBool("no-verify")
			switch {
			case noVerify:
				cmd.PrintErrln("Skipping verification (--no-verify).")
			case rs.Tenant == "":
				cmd.PrintErrln("No tenant given; saving without verification. Pass --remote-tenant to verify.")
			default:
				if err := verifyRemote(cmd.Context(), rs); err != nil {
					return fmt.Errorf("credential not saved: %w", err)
				}
				cmd.Printf("Verified access to tenant %s.\n", rs.Tenant)
			}

			creds, err := config.LoadCredentials()
			if err != nil {
				return err
			}
			creds.Set(rs.BaseURL, config.CloudCredential{Token: token, Tenant: rs.Tenant, SavedAt: time.Now().UTC()})
			if err := creds.Save(); err != nil {
				return err
			}
			cfg.Cloud.BaseURL = rs.BaseURL
			if err := cfg.Save(); err != nil {
				return err
			}
			path, err := config.CredentialsPath()
			if err != nil {
				return err
			}
			cmd.Printf("Saved credential for %s to %s\n", rs.BaseURL, path)
			return nil
		},
	}
	cmd.Flags().Bool("with-token", false, "Read the API key from standard input")
	cmd.Flags().Bool("no-verify", false, "Save without checking the credential against the tenant")
	return cmd
}

// readLoginToken reads the credential from stdin (--with-token), the
// environment, or a hidden interactive prompt.
func readLoginToken(cmd *cobra.Command, withToken bool) (string, error) {
	if withToken {
		data, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), 64<<10))
		if err != nil {
			return "", fmt.Errorf("read token from stdin: %w", err)
		}
		return nonEmptyToken(string(data))
	}
	if env := os.Getenv(envRemoteToken); strings.TrimSpace(env) != "" {
		return nonEmptyToken(env)
	}
	in := cmd.InOrStdin()
	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) { // #nosec G115 -- file descriptors fit in int
		cmd.PrintErr("Paste your VisionStudio Cloud API key: ")
		b, err := term.ReadPassword(int(f.Fd())) // #nosec G115 -- file descriptors fit in int
		cmd.PrintErrln()
		if err != nil {
			return "", fmt.Errorf("read token: %w", err)
		}
		return nonEmptyToken(string(b))
	}
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read token: %w", err)
	}
	return nonEmptyToken(line)
}

func nonEmptyToken(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("no API key provided (use --with-token, $" + envRemoteToken + ", or the interactive prompt)")
	}
	if strings.ContainsAny(s, " \t\r\n") {
		return "", errors.New("the API key must be a single token without whitespace")
	}
	return s, nil
}

// verifyRemote performs one authenticated read against the tenant.
func verifyRemote(ctx context.Context, rs *remoteSettings) error {
	st, err := newRemoteStore(rs)
	if err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, verifyTimeout)
	defer cancel()
	return st.Ping(ctx)
}

func cloudLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove a stored VisionStudio Cloud credential",
		Long: `Remove the stored credential for the cloud URL given by --remote (or
$VISIONSTUDIO_REMOTE_URL, or the URL of the last login). --all removes every
stored cloud credential. This only deletes the local copy; revoke the API key
in VisionStudio Cloud to invalidate it server-side.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := config.LoadCredentials()
			if err != nil {
				return err
			}
			if all, _ := cmd.Flags().GetBool("all"); all {
				n := len(creds.Cloud)
				creds.Cloud = nil
				if err := creds.Save(); err != nil {
					return err
				}
				cmd.Printf("Removed %d stored cloud credential(s).\n", n)
				return nil
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			rs, ok, err := resolveRemote(cmd, cfg.Cloud.BaseURL)
			if err != nil {
				return err
			}
			if !ok {
				return errors.New("no cloud URL: pass --remote <url> (or --all)")
			}
			if !creds.Delete(rs.BaseURL) {
				cmd.Printf("No stored credential for %s.\n", rs.BaseURL)
				return nil
			}
			if err := creds.Save(); err != nil {
				return err
			}
			cmd.Printf("Removed stored credential for %s.\n", rs.BaseURL)
			return nil
		},
	}
	cmd.Flags().Bool("all", false, "Remove every stored cloud credential")
	return cmd
}

func cloudStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the resolved remote-mode settings and verify the credential",
		Long: `Show which cloud URL, tenant, and credential remote mode would use, where
each came from, and whether the credential can access the tenant. Exits non-zero
when verification fails. Without --remote/$VISIONSTUDIO_REMOTE_URL it reports on
the URL of the last 'cloud login' (commands still run locally unless --remote
is given).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			rs, ok, err := resolveRemote(cmd, cfg.Cloud.BaseURL)
			if err != nil {
				return err
			}
			if !ok {
				cmd.Println("Remote mode: not configured (no --remote, $" + envRemoteURL + ", or prior 'cloud login').")
				return nil
			}
			if _, active := remoteURLFlag(cmd); active == "" {
				cmd.Println("Remote mode: inactive for this invocation (commands run locally unless --remote is given)")
			} else {
				cmd.Println("Remote mode: active")
			}
			cmd.Printf("  URL:        %s  (from %s)\n", rs.BaseURL, rs.URLSource)
			if rs.Tenant != "" {
				cmd.Printf("  Tenant:     %s  (from %s)\n", rs.Tenant, rs.TenantSource)
			} else {
				cmd.Println("  Tenant:     (none — set --remote-tenant or $" + envRemoteTenant + ")")
			}
			if rs.Token != "" {
				cmd.Printf("  Credential: %s  (from %s)\n", maskToken(rs.Token), rs.TokenSource)
			} else {
				cmd.Println("  Credential: (none — run 'visionstudio cloud login' or set $" + envRemoteToken + ")")
			}
			if rs.Tenant == "" || rs.Token == "" {
				return nil
			}
			if err := verifyRemote(cmd.Context(), rs); err != nil {
				cmd.Println("  Access:     FAILED")
				return err
			}
			cmd.Println("  Access:     ok")
			return nil
		},
	}
}
