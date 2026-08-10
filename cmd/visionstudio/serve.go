package main

import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/web"
)

// resolveWebUI returns the filesystem to serve the SPA from, plus a short
// human-readable description of where it came from. Resolution order:
//  1. --web-dist flag or $VISIONSTUDIO_WEB_DIST (explicit override)
//  2. a web/dist directory found near the working dir (fresh dev builds win)
//  3. the SPA embedded in the binary (works from any directory)
//
// It errors only when none of these yields a usable UI.
func resolveWebUI(cmd *cobra.Command) (fs.FS, string, error) {
	override, _ := cmd.Flags().GetString("web-dist")
	if override == "" {
		override = os.Getenv("VISIONSTUDIO_WEB_DIST")
	}
	if override != "" {
		dir := expandHome(override)
		if !dirHasIndex(dir) {
			return nil, "", fmt.Errorf("--web-dist %q has no index.html; run 'npm run build' in web/ first", dir)
		}
		return os.DirFS(dir), "disk: " + dir, nil
	}

	if p := findWebDistPath(); p != "" {
		return os.DirFS(p), "disk: " + p, nil
	}

	if fsys, ok := web.DistFS(); ok {
		return fsys, "embedded", nil
	}

	return nil, "", fmt.Errorf("web UI not available: build it with 'cd web && npm run build', or pass --web-dist <path>")
}

// dirHasIndex reports whether dir looks like a built SPA. dir comes from the
// --web-dist flag or $VISIONSTUDIO_WEB_DIST, both set by the operator running
// the CLI locally, not from a remote request.
func dirHasIndex(dir string) bool {
	info, err := os.Stat(dir) // #nosec G703 -- dir is operator-supplied (CLI flag / env var), not request input
	if err != nil || !info.IsDir() {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, "index.html")) // #nosec G703 -- dir is operator-supplied (CLI flag / env var), not request input
	return err == nil
}

// resolveListenAddr computes the host:port to bind from the command's flags and
// warns when the bind host is not loopback (which exposes the DB-backed UI).
func resolveListenAddr(cmd *cobra.Command) (string, error) {
	address, _ := cmd.Flags().GetString("address")
	port, _ := cmd.Flags().GetInt("port")
	addr, nonLoopback, err := parseListenAddr(address, port)
	if err != nil {
		return "", err
	}
	if nonLoopback {
		fmt.Fprintf(os.Stderr,
			"warning: binding %s exposes the VisionStudio UI and its database beyond this machine — ensure this is intended\n",
			addr)
	}
	return addr, nil
}

// parseListenAddr resolves the bind address. --address wins (host:port or
// :port); otherwise it is 127.0.0.1:<port>. nonLoopback is true when an
// explicit non-loopback host was given. Pure function for testability.
func parseListenAddr(address string, port int) (addr string, nonLoopback bool, err error) {
	if address == "" {
		return fmt.Sprintf("127.0.0.1:%d", port), false, nil
	}
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return "", false, fmt.Errorf("invalid --address %q (want host:port, e.g. localhost:9401): %w", address, err)
	}
	if portStr == "" {
		return "", false, fmt.Errorf("invalid --address %q: missing port", address)
	}
	if host == "" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, portStr), !isLoopbackHost(host), nil
}

func isLoopbackHost(h string) bool {
	switch strings.ToLower(h) {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	if ip := net.ParseIP(h); ip != nil {
		return ip.IsLoopback()
	}
	return false
}
