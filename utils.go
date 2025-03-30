package golog

import (
	"path/filepath"
	"runtime"
)

// getCallerInfo returns the file and line number of the caller
// skip is the number of stack frames to skip (1 for direct caller, 2 for caller's caller, etc.)
func getCallerInfo(skip int) (file string, line int) {
	_, file, line, ok := runtime.Caller(skip + 1) // +1 because we want to skip this function
	if !ok {
		return "unknown", 0
	}
	// Get just the filename without the full path
	file = filepath.Base(file)
	return file, line
}
