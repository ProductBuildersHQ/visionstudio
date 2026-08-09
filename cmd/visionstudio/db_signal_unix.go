//go:build !windows

package main

import (
	"os"
	"syscall"
)

// isProcessAlive checks if a process with the given PID exists and is running.
func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

// signalTerminate sends SIGTERM to the process.
func signalTerminate(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}

// signalKill sends SIGKILL to the process.
func signalKill(proc *os.Process) error {
	return proc.Signal(syscall.SIGKILL)
}
