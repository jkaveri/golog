# Golog

Structured logging for Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/jkaveri/golog/v2.svg)](https://pkg.go.dev/github.com/jkaveri/golog/v2)

**Use [v2](v2/README.md).** New projects and documentation should target **`github.com/jkaveri/golog/v2`**. The **module root** import (`github.com/jkaveri/golog` without `/v2`) is the **legacy v1 API**; it remains published **only for backward compatibility** with existing code. Do not choose v1 for new work.

## Table of Contents

- [Table of Contents](#table-of-contents)
- [v2 vs legacy at a glance](#v2-vs-legacy-at-a-glance)
- [v2 (recommended)](#v2-recommended)
- [Legacy API (v1)](#legacy-api-v1)
- [Requirements](#requirements)
- [Features](#features)
- [Installation (legacy v1)](#installation-legacy-v1)
- [Quick Start](#quick-start)
  - [JSON Logger](#json-logger)
- [Interfaces](#interfaces)
  - [LogWriter Interface](#logwriter-interface)
  - [Enricher Interface](#enricher-interface)
  - [Thread Safety](#thread-safety)
- [Advanced Usage](#advanced-usage)
  - [Scoped Logging](#scoped-logging)
  - [Custom Writer Implementation](#custom-writer-implementation)
  - [Log Enrichment](#log-enrichment)
  - [Error Handling](#error-handling)
- [Configuration](#configuration)
  - [Log Levels](#log-levels)
  - [Output Configuration](#output-configuration)
- [Unsupported Types](#unsupported-types)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [License](#license)

## v2 vs legacy at a glance

| Topic | **v2** (`github.com/jkaveri/golog/v2`) | **Legacy v1** (`github.com/jkaveri/golog`) |
|-------|----------------------------------------|--------------------------------------------|
| **Status** | Current API; new features land here | Frozen; bugfixes / compatibility only |
| **Shape** | `Logger` + slog-style `Attr` / `Value`, `Config`, `Writer` | Package-level helpers + `LogScope`, `LogWriter` |
| **Levels** | `Debug`, `Info`, `Error` only (no separate “warn”) | Same three levels |
| **Docs** | [v2/README.md](v2/README.md), [pkg.go.dev](https://pkg.go.dev/github.com/jkaveri/golog/v2) | This file (below), [pkg.go.dev](https://pkg.go.dev/github.com/jkaveri/golog) |

## v2 (recommended)

Install and use the **`/v2` module**:

```bash
go get github.com/jkaveri/golog/v2
```

**Documentation**

- **Guide (presets, JSON, context/source enrichers, diagrams):** [v2/README.md](v2/README.md)
- **API reference:** [pkg.go.dev/github.com/jkaveri/golog/v2](https://pkg.go.dev/github.com/jkaveri/golog/v2)
- **Runnable examples:** [`v2/_examples/`](v2/_examples/)

v2 provides a small `Logger` API (`Debug`, `Info`, `Error`, `With`, `WithContext`, `WithError`), slog-style `Attr` / `Value`, declarative `Config` (text or JSON, level, optional caller/source enrichment), pluggable `Writer` implementations (`TextWriter`, `JSONWriter`), and optional `Enricher` hooks. Package-level helpers (`golog.Info`, etc.) are available after optional [`InitDefault`](https://pkg.go.dev/github.com/jkaveri/golog/v2#InitDefault).

Preset configs are available via [`github.com/jkaveri/golog/v2/config`](https://pkg.go.dev/github.com/jkaveri/golog/v2/config) (`Development`, `Production`, and related helpers).

### v2 quick start

```go
package main

import (
	"github.com/jkaveri/golog/v2"
)

func main() {
	log, err := golog.NewLogger(golog.Config{
		Format: golog.FormatText,
		Output: "", // stdout
		Level:  golog.LevelDebug,
	})
	if err != nil {
		panic(err)
	}

	log.Info("server started",
		golog.String("addr", ":8080"),
	)
}
```

## Legacy API (v1)

Everything from [Requirements](#requirements) onward in this file describes the **legacy** package at the **module root** (`github.com/jkaveri/golog`). That API is **frozen** for **backward compatibility**; **refer to v2** for current behavior, types, and best practices.

## Requirements

- **v2** (`github.com/jkaveri/golog/v2`): Go **1.26+** (see [`v2/go.mod`](v2/go.mod)).
- **Legacy v1** (module root): Go **1.23+** (see [`go.mod`](go.mod)); dependencies include [bytedance/sonic](https://github.com/bytedance/sonic) and [pkg/errors](https://github.com/pkg/errors).

## Features

**v2** (see [v2/README.md](v2/README.md) for detail):

- Compact `Logger` interface with structured `Attr` fields
- Text and JSON line writers, declarative `Config`, optional enrichers (context, source)
- Level filtering on the logger; thread-safe writers

**Legacy v1** (sections below):

- Simple package-level API and scoped fields (`With`, `WithContext`, …)
- Text and JSON `LogWriter` implementations
- Global enrichers and level configuration

## Installation (legacy v1)

For existing code that still uses the **module root** (v1):

```bash
go get github.com/jkaveri/golog
```

For **new** code, use [v2](#v2-recommended) (`github.com/jkaveri/golog/v2`) instead.

## Quick Start

```go
package main

import "github.com/jkaveri/golog"

func main() {
    // Simple logging
    golog.Info("Application started")
    // Output: main.go:N [INFO][2026-03-30T12:34:56Z] Application started

    // Logging with fields
    golog.With("user_id", 123).
        With("username", "john_doe").
        Info("User logged in")
    // Output: main.go:N [INFO][2026-03-30T12:34:56Z] User logged in user_id="123" username="john_doe"

    // Logging with error
    err := someOperation()
    if err != nil {
        golog.WithError(err).Error("Operation failed")
        // Output: main.go:N [ERROR][2026-03-30T12:34:56Z] Operation failed error="operation failed: invalid input"
    }
}
```

### JSON Logger

The JSON logger is suitable for production environments where logs need to be parsed by log aggregation tools.

```go
package main

import (
    "os"
    "github.com/jkaveri/golog"
)

func main() {
    // Create a JSON logger writing to stdout
    jsonLogger := golog.NewJSONWriter(os.Stdout)

    // Set it as the default logger
    golog.SetWriter(jsonLogger)

    // Now all logs will be in JSON format
    golog.With("version", "1.0.0").
        With("environment", "production").
        Info("Application started")

    // Output will look like:
    // {"time":"2026-03-30T12:34:56Z","level":"INFO","msg":"Application started","version":"1.0.0","environment":"production","caller":"main.go:N"}\n
}
```

## Interfaces

The library defines two main interfaces that can be implemented to customize logging behavior:

### LogWriter Interface

```go
type LogWriter interface {
    // Write writes a log entry with the given level, message, and fields
    Write(level int, msg string, fields map[string]any)
    // Flush ensures all buffered log entries are written
    Flush()
}
```

The `LogWriter` interface is responsible for the actual writing of log entries. The library provides two implementations:

- `defaultWriter`: A text-based logger that writes human-readable logs
- `jsonWriter`: A JSON logger that writes machine-readable logs

### Enricher Interface

```go
type Enricher interface {
    // Enrich adds additional fields to a log entry based on the context
    Enrich(ctx context.Context, level string, msg string, fields map[string]any)
}
```

The `Enricher` interface allows you to add additional context to log entries. Enrichers are called before each log entry is written and can modify the fields map to add more information.

### Thread Safety

The `LogScope` type is not thread-safe and should not be shared between goroutines. Each goroutine should create its own scope using the provided factory functions (`With`, `WithFields`, `WithContext`, etc.). The underlying `LogWriter` implementations (`defaultWriter` and `jsonWriter`) are thread-safe and can be safely used from multiple goroutines.

## Advanced Usage

### Scoped Logging

Scoped logging allows you to create a context with predefined fields that will be included in all subsequent log messages.

```go
func handleRequest(ctx context.Context, req *http.Request) {
    // Create a new logging scope with request context
    scope := golog.WithContext(ctx).
        With("request_id", req.Header.Get("X-Request-ID")).
        With("method", req.Method).
        With("path", req.URL.Path)

    // All logs in this scope will include the request fields
    scope.Info("Processing request")

    // Additional fields can be added to specific log lines
    scope.With("user_agent", req.UserAgent()).Debug("Request details")

    // Nested scopes inherit parent fields
    userScope := scope.With("user_id", 123)
    userScope.Info("User-specific operation")
}
```

### Custom Writer Implementation

You can create your own writer implementation by implementing the `LogWriter` interface. This allows you to customize how logs are formatted and where they are written.

```go
type CustomWriter struct {
    writer io.Writer
}

func NewCustomWriter(writer io.Writer) *CustomWriter {
    return &CustomWriter{writer: writer}
}

func (w *CustomWriter) Write(level int, msg string, fields map[string]any) {
    timestamp := time.Now().Format(time.RFC3339)
    levelStr := golog.LevelString(level)
    logLine := fmt.Sprintf("[%s] %s: %s", timestamp, levelStr, msg)

    for k, v := range fields {
        logLine += fmt.Sprintf(" %s=%v", k, v)
    }
    logLine += "\n"

    w.writer.Write([]byte(logLine))
}

func (w *CustomWriter) Flush() {
    // Implement flush logic if needed
}

// Usage
func main() {
    writer := NewCustomWriter(os.Stdout)
    golog.SetWriter(writer)

    golog.With("version", "1.0.0").Info("Application started")
}
```

### Log Enrichment

You can enrich your logs with additional context using the `Enricher` interface and `RegisterEnricher`. Enrichers are applied globally to all log entries.

```go
import (
    "context"
    "github.com/jkaveri/golog"
)

// Register at startup (e.g., in main or init)
func init() {
    golog.RegisterEnricher(golog.EnricherFunc(func(ctx context.Context, level, msg string, fields map[string]any) {
        if id := ctx.Value("request_id"); id != nil {
            fields["request_id"] = id
        }
    }))
}

// Usage
func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := context.WithValue(r.Context(), "request_id", r.Header.Get("X-Request-ID"))
    golog.WithContext(ctx).Info("Processing request")
}
```

### Error Handling

Golog provides convenient methods for error logging.

```go
func processUser(userID string) error {
    user, err := fetchUser(userID)
    if err != nil {
        return golog.WithError(err).
            With("user_id", userID).
            With("operation", "fetch_user").
            Error("Failed to fetch user")
    }

    if err := validateUser(user); err != nil {
        return golog.WithError(err).
            With("user_id", userID).
            With("operation", "validate_user").
            Error("User validation failed")
    }

    return nil
}
```

## Configuration

### Log Levels

You can configure the minimum log level to control which messages are logged.

```go
// Set minimum log level to Debug
golog.SetLevel(golog.LevelDebug)

// Set minimum log level to Info (default)
golog.SetLevel(golog.LevelInfo)

// Set minimum log level to Error
golog.SetLevel(golog.LevelError)

// Parse level from string (e.g., from environment variable)
if level := golog.ParseLevel(os.Getenv("LOG_LEVEL")); level >= 0 {
    golog.SetLevel(level)
}
```

### Output Configuration

```go
// Write to file
file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
logger := golog.NewJSONWriter(file)
golog.SetWriter(logger)

// Write to multiple outputs
multiWriter := io.MultiWriter(os.Stdout, file)
logger := golog.NewJSONWriter(multiWriter)
golog.SetWriter(logger)
```

## Unsupported Types

Golog uses [bytedance/sonic](https://github.com/bytedance/sonic) for JSON encoding. As a result, the library inherits the same type limitations as sonic. When using structs with fields that contain unsupported types, you should use the `json:"-"` tag to skip those fields during logging. Any serialization issues, including unsupported types or invalid JSON structures, will cause a panic.

**Unsupported types include:**

- Complex numbers (`complex64`, `complex128`)
- Channels
- Functions
- Other types not supported by sonic

**Example of handling unsupported types:**

```go
type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Complex   complex64 `json:"-"`  // Skip complex number field to avoid panic
    Channel   chan int  `json:"-"`  // Skip channel field to avoid panic
    Callback  func()    `json:"-"`  // Skip function field to avoid panic
}

func main() {
    user := User{
        ID:      "123",
        Name:    "John",
        Complex: 1 + 2i,
        Channel: make(chan int),
        Callback: func() {},
    }

    // This will work without panicking because unsupported fields are tagged with json:"-"
    golog.With().
        With("user", user).
        Info("User created")

    // This will panic because Complex field is not tagged with json:"-"
    golog.With().
        With("complex", 1+2i).
        Info("This will panic")
}
```

## Best Practices

1. **Use Scoped Logging**: Create scopes for different components or operations to maintain context.
2. **Include Relevant Fields**: Always include fields that help with debugging and monitoring.
3. **Use Appropriate Log Levels**:
   - Debug: Detailed information for debugging
   - Info: General operational information
   - Error: Error conditions that need attention
4. **Structured Data**: Use the field system instead of string interpolation for better parsing.
5. **Context Preservation**: Use `WithContext` when working with request-scoped operations.
6. **Type Limitations**: Be aware that certain Go types are not supported in fields and will cause a panic:
   - Complex numbers (`complex64`, `complex128`)
   - Channels
   - Other unsupported types will cause a panic

## API Reference

The following applies to the **legacy v1** package (`github.com/jkaveri/golog`) only. For v2, see [pkg.go.dev/github.com/jkaveri/golog/v2](https://pkg.go.dev/github.com/jkaveri/golog/v2).

### Functions

- `golog.Debug(msg string, args ...any)` - Log a debug message; args are passed to fmt.Sprintf
- `golog.Info(msg string, args ...any)` - Log an info message; args are passed to fmt.Sprintf
- `golog.Error(msg string, args ...any) error` - Log an error message and return an error for propagation
- `golog.With(key string, value any) *LogScope` - Create a new scope with a field
- `golog.WithFields(fields map[string]any) *LogScope` - Create a new scope with multiple fields
- `golog.WithPairs(args ...any) *LogScope` - Create a scope from alternating key-value pairs (panics if odd length or non-string keys)
- `golog.WithContext(ctx context.Context) *LogScope` - Create a new scope with context
- `golog.WithError(err error) *LogScope` - Create a new scope with error
- `golog.RegisterEnricher(enricher Enricher)` - Register a global enricher
- `golog.SetWriter(writer LogWriter)` - Set a custom logger implementation
- `golog.SetLevel(level int)` - Set the minimum log level (use LevelDebug, LevelInfo, LevelError)
- `golog.ParseLevel(level string) int` - Parse level from string; returns -1 if invalid
- `golog.LevelString(level int) string` - Convert level int to string
- `golog.Flush()` - Flush any buffered logs

### Types

- `golog.LogWriter` - Interface for logger implementations
- `golog.LogScope` - Represents a logging scope with propagated fields (not thread-safe)
- `golog.Enricher` - Interface for log enrichment
- `golog.EnricherFunc` - Function type implementing Enricher

### Level Constants

- `golog.LevelDebug` (0), `golog.LevelInfo` (1), `golog.LevelError` (2)

### Implementations

- `golog.NewDefaultWriter(writer io.Writer) *defaultWriter` - Creates a default text-based logger
- `golog.NewJSONWriter(writer io.Writer) *jsonWriter` - Creates a JSON logger

### Common Pitfalls

- **Unsupported types**: Complex numbers, channels, and functions in fields cause a panic. Use `json:"-"` on struct fields or avoid these types.
- **WithPairs**: Must have an even number of arguments; keys must be strings. Panics otherwise.
- **Thread safety**: LogScope is not thread-safe; create a new scope per goroutine.

## Contributing

Contributions are welcome. Please open an issue to discuss larger changes, and ensure `go test ./...` passes from the repository root (including [`v2/`](v2/)) before submitting a pull request.

## License

This project is licensed under the MIT License; see [LICENSE](LICENSE).
