package golog

import (
	"context"
	"os"
)

const (
	defaultBufferSize = 4 * 1024 // 4KB buffer
)

// Standard JSON log field names used in log output
const (
	// FieldTime is the key for time of log
	FieldTime = "time"
	// FieldLevel is the key for level of log
	FieldLevel = "level"
	// FieldMessage is the key for message of log
	FieldMessage = "msg"

	// FieldCaller is the key for caller of log
	FieldCaller = "caller"
)

var (
	// instance is the global log writer instance
	instance LogWriter = &defaultWriter{
		output: os.Stdout,
	}
	// enrichers contains all registered log enrichers
	enrichers []Enricher
)

// LogWriter defines the interface for log output writers.
// Implementations should handle the actual writing of log entries.
type LogWriter interface {
	// Write writes a log entry with the given level, message, and fields
	Write(level int, msg string, fields map[string]any)
	// Flush ensures all buffered log entries are written
	Flush()
}

// SetWriter sets the global log writer instance.
// This function should be called at application startup to configure logging.
func SetWriter(logger LogWriter) {
	instance = logger
}

// RegisterEnricher adds a new enricher to the global enrichers list.
// Enrichers are called in the order they are registered.
func RegisterEnricher(enricher Enricher) {
	enrichers = append(enrichers, enricher)
}

// With creates a new LogScope with a single key-value field.
// It is a convenience function for creating a scope with a single field.
func With(key string, value any) *LogScope {
	return newScope().With(key, value)
}

// WithFields creates a new LogScope with multiple fields.
// It is a convenience function for creating a scope with multiple fields at once.
func WithFields(fields map[string]any) *LogScope {
	return newScope().WithFields(fields)
}

// WithPairs creates a new LogScope with multiple fields.
// It is a convenience function for creating a scope with multiple fields at once.
func WithPairs(args ...any) *LogScope {
	if len(args)%2 != 0 {
		panic("pairs must have even number of arguments")
	}

	pairs := make(map[string]any)
	for i := 0; i < len(args); i += 2 {
		switch key := args[i].(type) {
		case string:
			pairs[key] = args[i+1]
		default:
			panic("pairs must have alternating key-value arguments")
		}
	}

	return WithFields(pairs)
}

// WithContext creates a new LogScope with the given context.
// It is a convenience function for creating a scope with a context.
func WithContext(ctx context.Context) *LogScope {
	return newScope().WithContext(ctx)
}

// WithError creates a new LogScope with an error field.
// It is a convenience function for creating a scope with an error.
func WithError(err error) *LogScope {
	return newScope().WithError(err)
}

// Debug logs a message at the debug level.
// It creates a new scope and writes the message with the DEBUG level.
func Debug(msg string, args ...any) {
	newScope().Debug(msg, args...)
}

// Info logs a message at the info level.
// It creates a new scope and writes the message with the INFO level.
func Info(msg string, args ...any) {
	newScope().Info(msg, args...)
}

// Error logs a message at the error level.
// It creates a new scope and writes the message with the ERROR level.
func Error(msg string, args ...any) error {
	return newScope().Error(msg, args...)
}

// Flush ensures all buffered log entries are written.
// It calls Flush on the global log writer instance.
func Flush() {
	instance.Flush()
}

var skipFrames = 1

func SetSkipFrames(skip int) {
	skipFrames = skip
}
