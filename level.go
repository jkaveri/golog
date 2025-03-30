package golog

import "strings"

// Log levels
const (
	LevelDebug = iota // 0
	LevelInfo         // 1
	LevelError        // 2
)

// levelNames maps level integers to their string representations
var levelNames = map[int]string{
	0: "DEBUG",
	1: "INFO",
	2: "ERROR",
}

// levelValues maps string level names to their integer values
var levelValues = map[string]int{
	"DEBUG": 0,
	"INFO":  1,
	"ERROR": 2,
}

// minLevel is the minimum level that should be logged
var minLevel = LevelInfo

// ParseLevel converts a string level name to its integer value.
// The parsing is case-insensitive.
// Returns -1 if the level name is invalid.
func ParseLevel(level string) int {
	// Convert to uppercase for case-insensitive comparison
	upperLevel := strings.ToUpper(level)
	if value, ok := levelValues[upperLevel]; ok {
		return value
	}
	return -1
}

// LevelString converts an integer level to its string representation.
// Returns "UNKNOWN" if the level is invalid.
func LevelString(level int) string {
	if name, ok := levelNames[level]; ok {
		return name
	}
	return "UNKNOWN"
}

// SetLevel sets the minimum log level that should be logged.
// Only messages with severity >= minLevel will be logged.
// Valid levels are: DEBUG (0), INFO (1), ERROR (2)
func SetLevel(level int) {
	if _, ok := levelNames[level]; ok {
		minLevel = level
	}
}

// shouldLog checks if a message with the given level should be logged
// based on the current minimum level setting
func shouldLog(level int) bool {
	_, ok := levelNames[level]
	if !ok {
		return false
	}

	return level >= minLevel
}
