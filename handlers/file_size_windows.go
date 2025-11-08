//go:build windows

package handlers

import (
	"os"
)

// getActualFileSize returns the actual disk usage and whether the file is sparse (Windows)
func getActualFileSize(info os.FileInfo, apparentSize int64) (int64, bool) {
	// On Windows, we use the apparent size as fallback
	// Windows has different APIs for getting actual disk usage (GetCompressedFileSize)
	// but they require more complex syscall handling
	// For now, use apparent size - most files aren't sparse on Windows anyway
	return apparentSize, false
}
