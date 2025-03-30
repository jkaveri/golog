// Package golog provides a simple, structured logging library for Go applications.
// It supports JSON output, context-based logging, and field enrichment.
//
// Features:
//   - Structured JSON logging
//   - Context support
//   - Field enrichment
//   - Multiple log levels (Debug, Info, Error)
//   - Flushable output
//
// Example usage:
//
//	writer := golog.NewJSONWriter(os.Stdout)
//	golog.SetWriter(writer)
//
//	golog.Info("Hello, world!")
//	golog.With("user_id", 123).Info("User logged in")
//	golog.WithError(err).Error("Operation failed")
//	golog.WithContext(ctx).WithFields(map[string]any{"request_id": "abc"}).Info("Request processed")
//	golog.WithPairs("user_id", 123, "action", "login").Info("User logged in")
//	golog.SetLevel("debug")
//	golog.SetSkipFrames(2)
package golog
