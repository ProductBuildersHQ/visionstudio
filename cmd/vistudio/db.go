package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/grokify/oscompat/process"
	"github.com/spf13/cobra"

	config "github.com/ProductBuildersHQ/visionstudio/pkg/cliconfig"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
)

const pidFileName = "server.pid"

func dbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management",
	}
	cmd.AddCommand(dbServeCmd(), dbStartCmd(), dbStopCmd(), dbRestartCmd(), dbStatusCmd())
	addDoltDBCommands(cmd)
	return cmd
}

func pidFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home: %w", err)
	}
	return filepath.Join(home, config.Dir, pidFileName), nil
}

func writePIDFile(pid, port int) error {
	path, err := pidFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create pid dir: %w", err)
	}
	content := fmt.Sprintf("%d\n%d\n", pid, port)
	return os.WriteFile(path, []byte(content), 0o600)
}

func readPIDFile() (pid, port int, err error) {
	path, err := pidFilePath()
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

func removePIDFile() {
	path, err := pidFilePath()
	if err != nil {
		return
	}
	os.Remove(path)
}

// isProcessAlive is implemented in db_signal_unix.go and db_signal_windows.go

func isPortListening(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func dbServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve [directory]",
		Short: "Start a Dolt SQL server (foreground)",
		Long: `Start a Dolt SQL server bound to localhost. Press Ctrl-C to stop.

The server uses --data-dir (or VISIONSTUDIO_DATA, or the default data directory)
as the multi-database root. Subdirectories containing .dolt are served as
databases. Other sessions connect via VISIONSTUDIO_DSN.

For background operation, use 'vistudio db start' instead.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var dir string
			if len(args) > 0 {
				dir = expandHome(args[0])
			} else {
				dir = getDataDir(cmd)
				if dir == "" {
					dir = expandHome(defaultDataDir)
				}
			}

			absDir, err := filepath.Abs(dir)
			if err != nil {
				return fmt.Errorf("resolve path: %w", err)
			}

			port, _ := cmd.Flags().GetInt("port")
			dsn := fmt.Sprintf("root:@tcp(127.0.0.1:%d)/visionstudio", port)

			cfg, err := config.Load()
			if err != nil {
				cfg = &config.Config{}
			}
			cfg.DSN = dsn
			if err := cfg.Save(); err != nil {
				cmd.PrintErrf("warning: could not save config: %v\n", err)
			} else {
				cmd.Printf("Saved DSN to config file (all vistudio sessions will use this server)\n")
			}

			cmd.Printf("Starting Dolt SQL server on 127.0.0.1:%d (dir: %s)...\n", port, absDir)
			return service.DBServe(cmd.Context(), absDir, port)
		},
	}
	cmd.Flags().Int("port", 3306, "Port for the SQL server")
	return cmd
}

func dbStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Dolt SQL server in the background",
		Long: `Start a Dolt SQL server as a background process.

Writes a PID file to ~/.productbuildershq/visionstudio/server.pid and saves the
DSN to the config file. Use 'vistudio db stop' to shut it down.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, _, err := readPIDFile()
			if err == nil && isProcessAlive(pid) {
				return fmt.Errorf("server already running (PID %d). Use 'vistudio db restart' to restart", pid)
			}

			dir := getDataDir(cmd)
			if dir == "" {
				dir = expandHome(defaultDataDir)
			}
			absDir, err := filepath.Abs(dir)
			if err != nil {
				return fmt.Errorf("resolve path: %w", err)
			}

			port, _ := cmd.Flags().GetInt("port")

			if isPortListening(port) {
				return fmt.Errorf("port %d is already in use", port)
			}

			doltArgs := []string{"sql-server", "--host", "127.0.0.1", "--port", fmt.Sprintf("%d", port)}
			proc := exec.Command("dolt", doltArgs...)
			proc.Dir = absDir
			proc.Stdout = nil
			proc.Stderr = nil
			process.SetDetached(proc)

			if err := proc.Start(); err != nil {
				return fmt.Errorf("start dolt server: %w", err)
			}

			if err := writePIDFile(proc.Process.Pid, port); err != nil {
				cmd.PrintErrf("warning: could not write pid file: %v\n", err)
			}

			dsn := fmt.Sprintf("root:@tcp(127.0.0.1:%d)/visionstudio", port)
			cfg, err := config.Load()
			if err != nil {
				cfg = &config.Config{}
			}
			cfg.DSN = dsn
			if err := cfg.Save(); err != nil {
				cmd.PrintErrf("warning: could not save config: %v\n", err)
			}

			deadline := time.Now().Add(10 * time.Second)
			for time.Now().Before(deadline) {
				if isPortListening(port) {
					cmd.Printf("Dolt server started (PID %d, port %d)\n", proc.Process.Pid, port)
					return nil
				}
				time.Sleep(200 * time.Millisecond)
			}

			cmd.Printf("Dolt server started (PID %d, port %d) — still waiting for port to become ready\n", proc.Process.Pid, port)
			return nil
		},
	}
	cmd.Flags().Int("port", 13306, "Port for the SQL server")
	return cmd
}

func dbStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the background Dolt SQL server",
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, _, err := readPIDFile()
			if err != nil {
				return fmt.Errorf("no server pid file found — is the server running?")
			}

			if !isProcessAlive(pid) {
				removePIDFile()
				cmd.Println("Server is not running (stale PID file removed).")
				return nil
			}

			proc, err := os.FindProcess(pid)
			if err != nil {
				return fmt.Errorf("find process %d: %w", pid, err)
			}
			if err := signalTerminate(proc); err != nil {
				return fmt.Errorf("terminate PID %d: %w", pid, err)
			}

			deadline := time.Now().Add(10 * time.Second)
			for time.Now().Before(deadline) {
				if !isProcessAlive(pid) {
					removePIDFile()
					cmd.Printf("Dolt server stopped (PID %d).\n", pid)
					return nil
				}
				time.Sleep(200 * time.Millisecond)
			}

			if err := signalKill(proc); err != nil {
				return fmt.Errorf("kill PID %d: %w", pid, err)
			}
			removePIDFile()
			cmd.Printf("Dolt server killed (PID %d).\n", pid)
			return nil
		},
	}
}

func dbRestartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the background Dolt SQL server",
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, oldPort, err := readPIDFile()
			if err == nil && isProcessAlive(pid) {
				proc, findErr := os.FindProcess(pid)
				if findErr == nil {
					_ = signalTerminate(proc)
					deadline := time.Now().Add(10 * time.Second)
					for time.Now().Before(deadline) {
						if !isProcessAlive(pid) {
							break
						}
						time.Sleep(200 * time.Millisecond)
					}
					if isProcessAlive(pid) {
						_ = signalKill(proc)
						time.Sleep(500 * time.Millisecond)
					}
				}
				removePIDFile()
				cmd.Printf("Stopped previous server (PID %d).\n", pid)
			}

			port, _ := cmd.Flags().GetInt("port")
			if port == 0 && oldPort > 0 {
				port = oldPort
			}
			if port == 0 {
				port = 13306
			}

			if err := cmd.Flags().Set("port", fmt.Sprintf("%d", port)); err != nil {
				return fmt.Errorf("set port flag: %w", err)
			}

			startCmd := dbStartCmd()
			startCmd.SetOut(cmd.OutOrStdout())
			startCmd.SetErr(cmd.ErrOrStderr())
			if err := startCmd.Flags().Set("port", fmt.Sprintf("%d", port)); err != nil {
				return fmt.Errorf("set port flag: %w", err)
			}
			return startCmd.RunE(startCmd, nil)
		},
	}
	cmd.Flags().Int("port", 0, "Port for the SQL server (default: previous port or 13306)")
	return cmd
}

func dbStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check the Dolt SQL server status",
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, port, err := readPIDFile()
			if err != nil {
				cmd.Println("Server: not running (no PID file)")
				return nil
			}

			alive := isProcessAlive(pid)
			listening := isPortListening(port)

			if !alive {
				removePIDFile()
				cmd.Printf("Server: not running (stale PID file for %d removed)\n", pid)
				return nil
			}

			status := "running"
			if !listening {
				status = "running but port not responding"
			}

			cmd.Printf("Server: %s\n", status)
			cmd.Printf("  PID:  %d\n", pid)
			cmd.Printf("  Port: %d\n", port)
			cmd.Printf("  DSN:  root:@tcp(127.0.0.1:%d)/visionstudio\n", port)
			return nil
		},
	}
}
