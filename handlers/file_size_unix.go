//go:build !windows

package handlers

import (
	"os"
	"syscall"
)

// getActualFileSize returns the actual disk usage and whether the file is sparse (Unix/Linux)
func getActualFileSize(info os.FileInfo, apparentSize int64) (int64, bool) {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		// Blocks * 512 = actual bytes on disk
		actualSize := stat.Blocks * 512
		// File is sparse if apparent size > actual disk usage
		isSparse := apparentSize > actualSize
		return actualSize, isSparse
	}

	// Fallback: use apparent size if syscall fails
	return apparentSize, false
}
