# Agent instructions — golog

This repository implements **structured logging for Go**. The active development surface is **`v2/`** (`github.com/jkaveri/golog/v2`): a small `Logger` API, `Attr`/`Value` (slog-style), `Config`, `Writer`, optional `Enricher`s, and `TextWriter` / `JSONWriter`.

Older code may live at the module root; prefer **`v2`** for new work unless the task explicitly targets legacy APIs.

## Build and test

- From `v2/`: `go test ./...`
- From repo root: `go test ./v2/...` if your Go version and workspace include `v2`

Match existing patterns (naming, tests, benchmarks) before introducing new abstractions.

## Library design (do not contradict without cause)

- **Levels:** `Debug`, `Info`, `Error` only. There is **no `Warn`** in the API; operational or recoverable issues belong on **`Info`**, failures on **`Error`**. See package docs in `v2/doc.go`.
- **Structured fields:** Use `Attr` helpers (`String`, `Int`, `Int64`, `Bool`, `Group`, …) passed to `Debug`/`Info`/`Error`, or attach defaults with **`Logger.With(...)`**. Do not document or suggest `WithField` / `WithFields` — those are other libraries; golog uses **`With`** and variadic **`...Attr`**.
- **Errors on the logger:** Use **`WithError(err)`** for the standard `"error"` attribute; combine with **`Error(...)`** at the call site when recording a failure.
- **Context:** When a function takes `context.Context` and logging should see that context (e.g. for `NewContextEnricher`), derive **`log := log.WithContext(ctx)`** once and use that logger for the rest of the function. Avoid mixing context-bound and non-context loggers in the same function without a clear reason.

## Message and field conventions (for examples and docs)

- Messages: human-readable, **English**, prefer **lowercase**; add **`Attr`**s for identifiers and searchable metadata (`request_id`, `trace_id`, `user_id`, `component`, `operation`, `status`, …).
- Do not rely on message text alone for important IDs or dimensions.
- **Security:** do not log secrets, tokens, full PII, or entire request/response bodies in examples or defaults.

## Error logging

- Prefer **one** clear log at a boundary or the point of failure; avoid logging the same returned error again at every layer (aligns with `v2/doc.go`).
- When demonstrating error logging: **`log.WithError(err).Error("...")`** (or equivalent with scoped attrs).

## Files and documentation

- **Do not** expand `README.md` or other markdown unless the user asks for doc changes.
- Prefer **`v2/doc.go`** and existing examples in `*_test.go` for API behavior.
- The file **`v2/logging.mdc`** is Cursor-oriented guidance copied from another project; treat **`AGENTS.md`** and this repo’s **`v2`** API as authoritative when they differ from `.mdc` naming (e.g. `With` + `Attr` vs `WithField`).

## Example (golog v2)

```go
func Handle(ctx context.Context, log golog.Logger) error {
	log = log.WithContext(ctx)

	log.Info("handling request", golog.String("operation", "handle"))

	if err := doWork(); err != nil {
		log.WithError(err).Error("failed to handle request")
		return err
	}
	return nil
}
```
