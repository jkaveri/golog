// Package golog provides a simple, structured logging library for Go applications.
// It supports JSON output, context-based logging, and field enrichment.
//
// # Architecture
//
// Golog uses three main concepts:
//   - Writers: Implement LogWriter to control where and how logs are formatted (e.g., JSON, human-readable).
//   - Scopes: LogScope holds fields and context; create scopes with With, WithFields, WithContext, WithError, or WithPairs.
//   - Enrichers: Implement Enricher to add fields to log entries; register globally with RegisterEnricher.
//
// # Features
//
//   - Structured JSON or text logging
//   - Context support for request-scoped fields
//   - Field enrichment via Enricher
//   - Multiple log levels (Debug, Info, Error)
//   - Flushable output
//
// # Thread Safety
//
// LogScope is not safe for concurrent use; create a new scope per goroutine or operation.
// LogWriter implementations (NewDefaultWriter, NewJSONWriter) are thread-safe.
//
// # Example
//
//	writer := golog.NewJSONWriter(os.Stdout)
//	golog.SetWriter(writer)
//
//	golog.Info("Hello, world!")
//	golog.With("user_id", 123).Info("User logged in")
//	golog.WithError(err).Error("Operation failed")
//	golog.WithContext(ctx).WithFields(map[string]any{"request_id": "abc"}).Info("Request processed")
//	golog.WithPairs("user_id", 123, "action", "login").Info("User logged in")
//	golog.SetLevel(golog.LevelDebug)
//	golog.SetSkipFrames(2)
package golog
