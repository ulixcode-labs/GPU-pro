//go:build !windows

package main

import (
	"syscall"
)

// killProcess sends SIGTERM to a process (Unix/Linux)
func killProcess(pid int) {
	syscall.Kill(pid, syscall.SIGTERM)
}
