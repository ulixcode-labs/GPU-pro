//go:build windows

package main

import (
	"os/exec"
	"strconv"
)

// killProcess terminates a process on Windows using taskkill
func killProcess(pid int) {
	// Use taskkill command on Windows
	// /F = force termination
	// /PID = process ID
	cmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
	cmd.Run() // Ignore errors - process might already be dead
}
