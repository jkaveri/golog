# Golog

A clean and consistent logging interface for Go applications that provides structured logging capabilities with a simple and intuitive API.

## Features

- Simple and clean interface for logging
- Structured logging with key-value fields
- Support for different log levels (Debug, Info, Error)
- Context-aware logging with scoped fields
- Automatic error tracking
- Extensible logger implementation
- Default FMT logger implementation included
- JSON logger implementation for machine-readable logs
- Automatic file and line information in logs
- Support for log levels configuration
- Thread-safe logging operations
- Performance optimized for high-throughput applications


## Installation

```bash
go get github.com/jkaveri/golog
```

## Quick Start

```go
package main

import "github.com/jkaveri/golog"

func main() {
    // Simple logging
    golog.With().Info("Application started")
    // Output: main.go:42 [INFO][2024-03-30T12:34:56Z] Application started

    // Logging with fields
    golog.With().
        With("user_id", 123).
        With("username", "john_doe").
        Info("User logged in")
    // Output: main.go:45 [INFO][2024-03-30T12:34:56Z] User logged in user_id="123" username="john_doe"

    // Logging with error
    err := someOperation()
    if err != nil {
        golog.WithError(err).Error("Operation failed")
        // Output: main.go:50 [ERROR][2024-03-30T12:34:56Z] Operation failed error="operation failed: invalid input"
    }
}
```

### JSON Logger

The JSON logger is perfect for production environments where logs need to be parsed by log aggregation tools.

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
    golog.With().
        With("version", "1.0.0").
        With("environment", "production").
        Info("Application started")

    // Output will look like:
    // {"time":"2024-03-30T12:34:56Z","level":"INFO","msg":"Application started","version":"1.0.0","environment":"production","caller":"main.go:42"}\n
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
    scope.Debug("Request details",
        golog.With("user_agent", req.UserAgent()),
    )

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

func (w *CustomWriter) Write(level string, msg string, fields map[string]any) {
    // Add custom formatting
    timestamp := time.Now().Format(time.RFC3339)
    logLine := fmt.Sprintf("[%s] %s: %s", timestamp, level, msg)

    // Add fields
    for k, v := range fields {
        logLine += fmt.Sprintf(" %s=%v", k, v)
    }

    // Add file and line information
    if file, line, ok := getCallerInfo(); ok {
        logLine += fmt.Sprintf(" file=%s line=%d", file, line)
    }

    logLine += "\n"

    w.writer.Write([]byte(logLine))
}

func (w *CustomWriter) Flush() {
    // Implement flush logic if needed
}

// Usage
func main() {
    // Create a custom writer that writes to stdout
    writer := NewCustomWriter(os.Stdout)

    // Set it as the default writer
    golog.SetWriter(writer)

    // Now all logs will use your custom format
    golog.Info("Application started",
        golog.With("version", "1.0.0"),
    )
}
```

### Log Enrichment

You can enrich your logs with additional context using the `Enricher` interface.

```go
type RequestEnricher struct {
    requestID string
}

func (e *RequestEnricher) Enrich(fields map[string]any) {
    fields["request_id"] = e.requestID
    fields["timestamp"] = time.Now().Unix()
}

// Usage
func handleRequest(w http.ResponseWriter, r *http.Request) {
    requestID := r.Header.Get("X-Request-ID")
    enricher := &RequestEnricher{requestID: requestID}

    scope := golog.WithEnricher(enricher)
    scope.Info("Processing request")
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
golog.SetLevel(golog.DebugLevel)

// Set minimum log level to Info (default)
golog.SetLevel(golog.InfoLevel)

// Set minimum log level to Error
golog.SetLevel(golog.ErrorLevel)
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

### Unsupported Types Include:

- Complex numbers (`complex64`, `complex128`)
- Channels
- Functions
- Other types not supported by sonic

### Example of Handling Unsupported Types

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

### Functions

- `golog.Debug(msg string, args ...any)` - Log a debug message with optional fields
- `golog.Info(msg string, args ...any)` - Log an info message with optional fields
- `golog.Error(msg string, args ...any)` - Log an error message with optional fields
- `golog.With(key string, value any) *LogScope` - Create a new scope with a field
- `golog.WithFields(fields map[string]any) *LogScope` - Create a new scope with multiple fields
- `golog.WithContext(ctx context.Context) *LogScope` - Create a new scope with context
- `golog.WithError(err error) *LogScope` - Create a new scope with error
- `golog.WithEnricher(enricher Enricher) *LogScope` - Create a new scope with an enricher
- `golog.SetWriter(writer LogWriter)` - Set a custom logger implementation
- `golog.SetLevel(level LogLevel)` - Set the minimum log level
- `golog.Flush()` - Flush any buffered logs

### Types

- `golog.LogWriter` - Interface for logger implementations
- `golog.LogScope` - Represents a logging scope with propagated fields
- `golog.Enricher` - Interface for log enrichment
- `golog.LogLevel` - Enum for log levels

> **Note**: When using fields in logging, certain Go types are not supported and will cause a panic:
> - Complex numbers (`complex64`, `complex128`)
> - Channels
> - Other unsupported types will cause a panic

### Implementations

- `golog.NewDefaulWriter(writer io.Writer) *defaultWriter` - Creates a default text-based logger
- `golog.NewJSONWriter(writer io.Writer) *jsonWriter` - Creates a JSON logger

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
