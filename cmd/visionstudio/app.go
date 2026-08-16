package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/grokify/oscompat/process"
	"github.com/spf13/cobra"

	config "github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
)

// addServeFlags registers the flags shared by the UI-serving commands
// (`ui`, `app start`). The dashboard command defines --port/--data-dir itself.
func addServeFlags(cmd *cobra.Command) {
	cmd.Flags().Int("port", 9400, "Port for the UI + API server (used when --address is unset)")
	cmd.Flags().String("address", "", "Bind address as host:port (e.g. localhost:9401); overrides --port")
	cmd.Flags().String("web-dist", "", "Serve the SPA from this built web/dist directory instead of the embedded copy")
	cmd.Flags().String("data-dir", "", "Path to omnidevx data directory (for report token data)")
}

// uiCmd serves the web UI + API. The SPA is embedded in the binary, so this
// works from any directory. The database must already be running.
func uiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Serve the VisionStudio web UI and API, and open it in the browser",
		Long: `Serve the React SPA and the JSON API on a single port and open it in your
browser. The SPA is embedded in the binary, so this works from any directory.

The database must already be running — use 'visionstudio db start', or use
'visionstudio app start' to bring up the database and UI together.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnifiedServer(cmd)
		},
	}
	addServeFlags(cmd)
	return cmd
}

// appCmd groups the batteries-included lifecycle commands that manage the
// database and UI together.
func appCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Run VisionStudio (database + UI) together",
	}
	cmd.AddCommand(appStartCmd(), appStatusCmd(), appStopCmd(), appRestartCmd())
	return cmd
}

// uiPidFileName tracks the UI+API server process separately from Dolt's
// server.pid, so a stale binary can be found and replaced without touching
// the database. Written by runUnifiedServer on every 'ui'/'app start'
// invocation (foreground or detached); removed on graceful shutdown or by
// whatever stops it (appRestartCmd, a future 'ui stop').
const uiPidFileName = "ui.pid"

func uiPidFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home: %w", err)
	}
	return filepath.Join(home, config.Dir, uiPidFileName), nil
}

func writeUIPIDFile(pid, port int) error {
	path, err := uiPidFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create pid dir: %w", err)
	}
	content := fmt.Sprintf("%d\n%d\n", pid, port)
	return os.WriteFile(path, []byte(content), 0o600)
}

func readUIPIDFile() (pid, port int, err error) {
	path, err := uiPidFilePath()
	if err != nil {
		return 0, 0, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("malformed pid file")
	}
	pid, err = strconv.Atoi(lines[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse pid: %w", err)
	}
	port, err = strconv.Atoi(lines[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse port: %w", err)
	}
	return pid, port, nil
}

func removeUIPIDFile() {
	path, err := uiPidFilePath()
	if err != nil {
		return
	}
	os.Remove(path)
}

// stopUIServer terminates the UI+API server recorded in ui.pid (SIGTERM,
// then SIGKILL after a grace period) and removes the PID file. Reports
// whether a server was actually found running -- not an error if not, since
// restart should work whether or not a prior instance is still up.
func stopUIServer() (wasRunning bool, err error) {
	pid, _, err := readUIPIDFile()
	if err != nil {
		return false, nil // nothing recorded
	}
	if !isProcessAlive(pid) {
		removeUIPIDFile()
		return false, nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return true, fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := signalTerminate(proc); err != nil {
		return true, fmt.Errorf("terminate PID %d: %w", pid, err)
	}
	if waitForProcessExit(pid, 10*time.Second) {
		removeUIPIDFile()
		return true, nil
	}
	_ = signalKill(proc)
	removeUIPIDFile()
	return true, nil
}

func appRestartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the web UI (and start the database if needed)",
		Long: `Stops any UI+API server this machine has recorded as running -- including one
started from a different terminal or detached session -- then starts a fresh
one in the foreground, the same as 'app start'. The most common reason to
need this: a new binary was built (e.g. after 'go build'/'go run' picked up
source changes) and the previously running server is still serving the old
one.

The database is left alone if it's already running (same as 'app start');
this only replaces the UI+API process, since that's what actually changes
when the binary changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			wasRunning, err := stopUIServer()
			if err != nil {
				return err
			}
			if wasRunning {
				cmd.Println("Stopped previous UI+API server.")
			}

			startedByUs, port, err := ensureDoltRunning(cmd)
			if err != nil {
				return err
			}
			if startedByUs {
				cmd.Printf("Started Dolt database on port %d\n", port)
				defer func() {
					if stopErr := stopDolt(); stopErr != nil {
						cmd.PrintErrf("warning: could not stop database: %v\n", stopErr)
					} else {
						cmd.Println("Stopped Dolt database")
					}
				}()
			} else {
				cmd.Printf("Using Dolt database already running on port %d\n", port)
			}
			return runUnifiedServer(cmd)
		},
	}
	addServeFlags(cmd)
	return cmd
}

func appStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the database and web UI together (one command)",
		Long: `Start VisionStudio end to end: ensure the Dolt database is running (starting
it in the background if needed), then serve the web UI and JSON API in the
foreground. Press Ctrl-C to stop.

If this command started the database, it also stops it on exit. A database
that was already running is left untouched. Use 'visionstudio db' and
'visionstudio ui' to run the pieces separately.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			startedByUs, port, err := ensureDoltRunning(cmd)
			if err != nil {
				return err
			}
			if startedByUs {
				cmd.Printf("Started Dolt database on port %d\n", port)
				defer func() {
					if stopErr := stopDolt(); stopErr != nil {
						cmd.PrintErrf("warning: could not stop database: %v\n", stopErr)
					} else {
						cmd.Println("Stopped Dolt database")
					}
				}()
			} else {
				cmd.Printf("Using Dolt database already running on port %d\n", port)
			}
			return runUnifiedServer(cmd)
		},
	}
	addServeFlags(cmd)
	return cmd
}

func appStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show whether the database and UI dependencies are ready",
		RunE: func(cmd *cobra.Command, args []string) error {
			dsn := getDSN(cmd)
			addr := dsnAddr(dsn)
			if pingDSN(dsn) {
				cmd.Printf("Database: reachable at %s\n", addr)
			} else if pid, port, err := readPIDFile(); err == nil && isProcessAlive(pid) {
				cmd.Printf("Database: process running (PID %d, port %d) but not answering yet\n", pid, port)
			} else {
				cmd.Printf("Database: not running (target %s)\n", addr)
				cmd.Println("  Start everything with: visionstudio app start")
			}

			if _, ok := webUIStatus(cmd); ok {
				cmd.Println("Web UI:   available (embedded or on disk)")
			} else {
				cmd.Println("Web UI:   not built — run 'cd web && npm run build' or pass --web-dist")
			}
			return nil
		},
	}
}

func appStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the background Dolt database started by app/db start",
		Long: `Stop the background Dolt database. The web UI runs in the foreground, so it
is stopped with Ctrl-C in its own terminal; this command stops the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, _, err := readPIDFile()
			if err != nil {
				cmd.Println("Database: not running (no PID file).")
				return nil
			}
			if !isProcessAlive(pid) {
				removePIDFile()
				cmd.Println("Database: not running (stale PID file removed).")
				return nil
			}
			if err := stopDolt(); err != nil {
				return err
			}
			cmd.Printf("Stopped Dolt database (PID %d).\n", pid)
			return nil
		},
	}
}

// webUIStatus reports the resolved UI source without starting a server.
func webUIStatus(cmd *cobra.Command) (string, bool) {
	_, source, err := resolveWebUI(cmd)
	if err != nil {
		return "", false
	}
	return source, true
}

// ensureDoltRunning makes sure a Dolt SQL server is reachable for the CLI's
// configured DSN, starting one in the background if needed. It returns whether
// this call started the server (so the caller can stop it on exit) and the
// port in use.
func ensureDoltRunning(cmd *cobra.Command) (startedByUs bool, port int, err error) {
	dsn := getDSN(cmd)
	port = portFromDSN(dsn)
	if port == 0 {
		port = defaultServerPort()
	}

	if isPortListening(port) {
		return false, port, nil
	}

	// Not up — start it in the background, mirroring `db start`.
	if pid, _, perr := readPIDFile(); perr == nil && isProcessAlive(pid) {
		// A process is recorded but not yet listening; give it a moment.
		if waitForPort(port, 10*time.Second) {
			return false, port, nil
		}
	}

	dir := getDataDir(cmd)
	if dir == "" {
		dir = expandHome(defaultDataDir)
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false, port, fmt.Errorf("resolve data dir: %w", err)
	}

	doltArgs := []string{"sql-server", "--host", "127.0.0.1", "--port", fmt.Sprintf("%d", port)}
	proc := exec.Command("dolt", doltArgs...)
	proc.Dir = absDir
	process.SetDetached(proc)
	if err := proc.Start(); err != nil {
		return false, port, fmt.Errorf("start dolt server (is 'dolt' installed and on PATH?): %w", err)
	}
	if err := writePIDFile(proc.Process.Pid, port); err != nil {
		cmd.PrintErrf("warning: could not write pid file: %v\n", err)
	}

	// Persist the DSN so other sessions agree on the port.
	if cfg, cErr := config.Load(); cErr == nil {
		cfg.DSN = fmt.Sprintf("root:@tcp(127.0.0.1:%d)/visionstudio", port)
		if sErr := cfg.Save(); sErr != nil {
			cmd.PrintErrf("warning: could not save config: %v\n", sErr)
		}
	}

	if !waitForPort(port, 10*time.Second) {
		return true, port, fmt.Errorf("dolt server did not become ready on port %d", port)
	}
	return true, port, nil
}

func waitForPort(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if isPortListening(port) {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

// stopDolt terminates the background Dolt server recorded in the PID file
// (SIGTERM, then SIGKILL after a grace period) and removes the PID file.
func stopDolt() error {
	pid, _, err := readPIDFile()
	if err != nil {
		return nil // nothing recorded
	}
	if !isProcessAlive(pid) {
		removePIDFile()
		return nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := signalTerminate(proc); err != nil {
		return fmt.Errorf("terminate PID %d: %w", pid, err)
	}
	if waitForProcessExit(pid, 10*time.Second) {
		removePIDFile()
		return nil
	}
	_ = signalKill(proc)
	removePIDFile()
	return nil
}

func waitForProcessExit(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !isProcessAlive(pid) {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}
