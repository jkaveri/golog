package golog

import (
	"context"
	"os"
)

var defaultLogger Logger

func init() {
	lg, err := NewLogger(Config{
		Format: FormatText,
		Output: "",
		Level:  LevelDebug,
	})
	if err != nil {
		panic(err)
	}

	defaultLogger = lg
}

// InitDefault replaces the package-level default logger used by [Debug],
// [Info], [Error], [With], [WithContext], [WithError], and [SetLevel]. Call it
// once during process startup when you want JSON output, a log file, or a
// different minimum level than the default
// installed at startup.
//
// Optional enrichers are applied like [NewLogger]; see [Enricher] for details.
//
// It returns an error if [Config] cannot be applied (for example an invalid
// [Config.Format]
// or an unopenable log file path).
func InitDefault(cfg Config, enrichers ...Enricher) error {
	l, err := NewLogger(cfg, enrichers...)
	if err != nil {
		return err
	}

	defaultLogger = l

	return nil
}

// Default returns the package-level default logger, installed before main or
// replaced by [InitDefault]. It is safe to call from any goroutine. Use
// [SetLevel] to change the
// minimum severity of that shared logger.
func Default() Logger {
	if defaultLogger != nil {
		return defaultLogger
	}

	// Defensive fallback. init() stores a default logger.
	return NewLoggerWriter(NewTextWriter(os.Stdout), LevelDebug)
}

// Debug logs with the package-level default logger at [LevelDebug]. It is
// equivalent to
// [Default]().Debug(msg, args...).
func Debug(msg string, args ...Attr) {
	Default().Debug(msg, args...)
}

// Info logs with the package-level default logger at [LevelInfo]. It is
// equivalent to
// [Default]().Info(msg, args...).
func Info(msg string, args ...Attr) {
	Default().Info(msg, args...)
}

// Error logs with the package-level default logger at [LevelError]. It is
// equivalent to
// [Default]().Error(msg, args...).
func Error(msg string, args ...Attr) {
	Default().Error(msg, args...)
}

// With returns [Default]().With(args...) for attaching attributes to all
// subsequent
// logs on the returned child logger.
func With(args ...Attr) Logger {
	return Default().With(args...)
}

// WithContext returns [Default]().WithContext(ctx) so context-backed enrichers
// see ctx.
func WithContext(ctx context.Context) Logger {
	return Default().WithContext(ctx)
}

// WithError returns [Default]().WithError(err), adding an "error" attribute
// when err is non-nil.
func WithError(err error) Logger {
	return Default().WithError(err)
}

// SetLevel sets the minimum severity on the package-level default logger by
// calling [Logger.SetLevel] on [Default]. It affects every top-level [Debug],
// [Info], and [Error] call.
func SetLevel(level Level) {
	Default().SetLevel(level)
}
