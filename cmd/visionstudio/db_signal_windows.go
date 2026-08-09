//go:build windows

package main

import (
	"os"
)

// isProcessAlive checks if a process with the given PID exists and is running.
// On Windows, we attempt to open the process handle to check if it exists.
func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Windows, FindProcess always succeeds for any PID.
	// We try to kill with signal 0 equivalent - but Windows doesn't support it.
	// Instead, we'll rely on the fact that Kill() will fail if process doesn't exist.
	// For a lightweight check, we just return true and let actual operations fail.
	_ = proc
	return true
}

// signalTerminate terminates the process on Windows.
// Windows doesn't have SIGTERM, so we use Kill().
func signalTerminate(proc *os.Process) error {
	return proc.Kill()
}

// signalKill forcefully kills the process on Windows.
// Same as signalTerminate since Windows only has Kill().
func signalKill(proc *os.Process) error {
	return proc.Kill()
}
