package golog

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// LogScope represents a logging context with associated fields and enrichers.
// It provides methods for adding fields and writing log entries.
type LogScope struct {
	// writer is the LogWriter instance used to write log entries
	writer LogWriter
	// enrichers contains the list of enrichers to apply to log entries
	enrichers []Enricher
	// fields contains the key-value pairs to include in log entries
	fields map[string]any
	// ctx contains the context associated with this scope
	ctx context.Context
}

// Context returns the context associated with this LogScope.
func (l *LogScope) Context() context.Context {
	return l.ctx
}

// Debug writes a log entry at the debug level.
// The message and any additional arguments are formatted using fmt.Sprintf.
func (l *LogScope) Debug(msg string, args ...any) {
	l.write(LevelDebug, msg, args...)
}

// Info writes a log entry at the info level.
// The message and any additional arguments are formatted using fmt.Sprintf.
func (l *LogScope) Info(msg string, args ...any) {
	l.write(LevelInfo, msg, args...)
}

// Error writes a log entry at the error level.
// The message and any additional arguments are formatted using fmt.Sprintf.
func (l *LogScope) Error(msg string, args ...any) error {
	l.write(LevelError, msg, args...)

	if l.fields["error"] != nil {
		err, ok := l.fields["error"].(error)
		if ok {
			return errors.Wrap(err, fmt.Sprintf(msg, args...))
		}
	}

	return errors.New(fmt.Sprintf(msg, args...))
}

// With adds a key-value field to this LogScope.
// It returns the LogScope for method chaining.
func (l *LogScope) With(key string, value any) *LogScope {
	l.fields[key] = value
	return l
}

// write is an internal method that writes a log entry with the given level and message.
// It applies all registered enrichers before writing.
func (l *LogScope) write(level int, msg string, args ...any) {
	// Check if we should log this level
	if !shouldLog(level) {
		return
	}

	// Apply enrichers
	for _, enricher := range l.enrichers {
		enricher.Enrich(l.ctx, LevelString(level), fmt.Sprintf(msg, args...), l.fields)
	}

	l.writer.Write(level, fmt.Sprintf(msg, args...), l.fields)
}

// WithError adds an error field to this LogScope.
// It returns the LogScope for method chaining.
func (l *LogScope) WithError(err error) *LogScope {
	l.fields["error"] = err.Error()
	return l
}

// WithFields adds multiple key-value fields to this LogScope.
// It returns the LogScope for method chaining.
func (l *LogScope) WithFields(fields map[string]any) *LogScope {
	for k, v := range fields {
		l.fields[k] = v
	}

	return l
}

// WithContext sets the context for this LogScope.
// It returns the LogScope for method chaining.
func (l *LogScope) WithContext(ctx context.Context) *LogScope {
	l.ctx = ctx
	return l
}

// newScope creates a new LogScope with default values.
// It uses the global log writer instance and creates an empty fields map.
func newScope() *LogScope {
	return &LogScope{
		writer: instance,
		fields: make(map[string]any),
		ctx:    context.Background(),
	}
}

// Flush ensures all buffered log entries are written.
// It calls Flush on the underlying log writer.
func (l *LogScope) Flush() {
	l.writer.Flush()
}
