package main

import (
	"os"

	"github.com/grokify/oscompat/process"
)

// isProcessAlive reports whether a process with the given PID is running. A
// PID whose state cannot be determined counts as not running.
func isProcessAlive(pid int) bool {
	alive, err := process.Exists(pid)
	return err == nil && alive
}

// signalTerminate asks the process to stop: SIGTERM where the platform has it,
// and a kill on Windows, which does not.
func signalTerminate(proc *os.Process) error {
	return process.Signal(proc.Pid)
}

// signalKill stops the process at once. os.Process.Kill works on every
// platform, so it needs no platform-specific code.
func signalKill(proc *os.Process) error {
	return proc.Kill()
}
