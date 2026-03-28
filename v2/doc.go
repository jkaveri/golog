// Package golog provides structured logging with records, attributes, and
// writers. It is compatible with slog-style types ([Attr], [Value], [Level])
// while offering a small [Logger] API, declarative [Config] for common setups,
// optional [Enricher] hooks, a [Writer] sink, and built-in writers:
// [TextWriter] for human-readable lines and [JSONWriter] for JSON lines.
//
// # Logging levels and philosophy
//
// Dave Cheney and others have argued that many logging setups are more complex
// than they need to be, and that a large number of named levels is often
// unhelpful in practice. The ideas below inform how to use this package’s
// levels and methods; they are guidance,
// not rules enforced by the library.
//
// **Too many levels are often useless.** Levels like “warning” can be
// ambiguous and are rarely acted on differently from “info” or “error”.
// In day-to-day work, two questions usually matter: what the system is doing
// ([LevelInfo], [Logger.Info]) and what helps
// debugging ([LevelDebug], [Logger.Debug]).
//
// **Avoid terminating the process from logging.** Patterns such as log.Fatal
// skip normal teardown (e.g. deferred functions) and make shutdown unsafe.
// Prefer returning errors and handling them at a top-level main or server
// boundary instead of exiting from a logger.
//
// **Don’t log the same error twice.** Logging an error and then returning it
// often duplicates noise. Prefer to log once at a boundary, or return the error
// for the caller
// to handle—not both for the same failure without a clear reason.
//
// **Logs are for humans.** Prefer clear messages and useful context over
// dumping raw internals. Good logging communicates what a future reader needs
// to debug or operate
// the system—not everything the program could possibly emit.
//
// **Simple default:** use [Logger.Info] (or [Info]) for normal operations,
// [Logger.Debug] (or [Debug]) for troubleshooting, and [Logger.Error] (or
// [Error]) when you deliberately
// record a failure at that layer.
//
// Core idea: effective logging is not about recording everything—it is about
// communicating
// useful information to the people who will debug the system later.
//
// # Examples
//
// Runnable examples are in this package’s example_test.go ([Example],
// [ExampleNewLogger], [ExampleNewLoggerWriter], [ExampleLogger_With],
// [ExampleEnricherFunc], and more). Preset configs are exemplified in the
// config subpackage (import path github.com/jkaveri/golog/v2/config). To
// replace the package-level default logger used by [Info], [Debug], and
// [Error], call [InitDefault] once at startup (see its documentation; no
// example to avoid mutating shared state in tests).
package golog
