# golog/v2

`golog/v2` is a small structured logging package with:

- a compact `Logger` interface (`Debug`, `Info`, `Error`, `With`, `WithContext`, `WithError`)
- `Attr`/`Value` types compatible with slog-style logging
- declarative [`Config`] (format, output path, level, time/duration formatting, optional source enricher)
- a single pluggable `Writer` sink (also constructible via [`NewLoggerWriter`] for advanced use)
- optional `Enricher` hooks that append context-derived attributes
- built-in `TextWriter` for readable line output
- built-in `JSONWriter` for JSONL output

## Install

```bash
go get github.com/jkaveri/golog/v2
```

## Quick Start

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
		golog.Group("http",
			golog.String("method", "GET"),
			golog.String("path", "/health"),
		),
	)
}
```

Production preset:

```go
log, err := golog.NewLogger(golog.Config{
	Format: golog.FormatJSON,
	Output: "",
	Level:  golog.LevelDebug,
})
```

Or use config factory helpers:

```go
import gologconfig "github.com/jkaveri/golog/v2/config"

logDev, err := golog.NewLogger(gologconfig.Development())
logProd, err := golog.NewLogger(gologconfig.Production())
```

Example output:

```text
2026-03-27T10:00:00Z INFO "server started" addr=:8080 http.method=GET http.path=/health
```

## Easiest Usage (Static Logger)

For quick usage, call package-level helpers directly. By default they log to `stdout`
using `TextWriter`.

```go
package main

import (
	"context"

	golog "github.com/jkaveri/golog/v2"
)

func main() {
	golog.Info("log msg", golog.String("field_a", "henry"))

	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	golog.WithContext(ctx).Info("request accepted")
}
```

You can override the global logger:

```go
if err := golog.InitDefault(golog.Config{
	Format: golog.FormatJSON,
	Output: "",
	Level:  golog.LevelDebug,
}); err != nil {
	log.Fatal(err)
}
```

## Workflow (Caller -> Output)

```text
+-------------------+
| Caller            |
| log.Info(...)     |
+---------+---------+
          |
          v
+-------------------+
| Logger.log(...)   |
| - level/message   |
| - attrs from With |
+---------+---------+
          |
          v
+-------------------+
| Build Record      |
| Time, Level, Msg  |
| + call-site attrs |
+---------+---------+
          |
          v
+-------------------+
| Enrichers (0..N)  |
| Enrich(ctx,rec)   |
| append attrs      |
+---------+---------+
          |
          v
+-------------------+      no (level < MinLevel)
| Logger.log        +--------------------+
| or Writer == nil  |                    |
+---------+---------+                    |
          |                              |
          | yes                          |
          v                              |
+-------------------+                    |
| writer.Write(...) |                    |
| TextWriter format |                    |
| attrs + newline   |                    |
+---------+---------+                    |
          |                              |
          v                              |
+-------------------+                    |
| Output            |                    |
| io.Writer         |                    |
| stdout/file/etc   |                    |
+-------------------+                    |
```

## Group Attribute Prefixing

For group attributes, text output uses the same rules as [`Value.String`]: the group key
is a prefix for child keys, joined with `.`.

Example:

```go
golog.Group("request",
	golog.String("id", "r-123"),
	golog.Group("user", golog.Int("id", 42)),
)
```

Becomes:

```text
request.id=r-123 request.user.id=42
```

## JSON Writer

`JSONWriter` emits one compact JSON object per line. Standard fields are:

- `time`
- `level`
- `msg`

Attributes are added as JSON fields, and group attrs become nested objects.

```go
log, err := golog.NewLogger(golog.Config{
	Format: golog.FormatJSON,
	Output: "",
	Level:  golog.LevelDebug,
})
if err != nil {
	panic(err)
}

log.Info("request done",
	golog.Group("request",
		golog.String("id", "r-123"),
		golog.Group("user", golog.Int("id", 42)),
	),
)
```

Example JSON line:

```json
{"time":"2026-03-27T10:00:00Z","level":"INFO","msg":"request done","request":{"id":"r-123","user":{"id":42}}}
```

## Context enricher

The context enricher ([`ContextEnricher`], [`NewContextEnricher`], [`FromContext`]) maps typed `context.Context` values to log attributes.

```go
import (
	"context"

	golog "github.com/jkaveri/golog/v2"
)

type requestIDKey struct{}
type userIDKey struct{}

log, err := golog.NewLogger(golog.Config{
	Format: golog.FormatJSON,
	Output: "",
	Level:  golog.LevelDebug,
},
	golog.NewContextEnricher(
		golog.FromContext.String(requestIDKey{}, "request_id"),
		golog.FromContext.Int64(userIDKey{}, "user_id"),
	),
)
if err != nil {
	panic(err)
}

ctx := context.WithValue(context.Background(), requestIDKey{}, "req-123")
ctx = context.WithValue(ctx, userIDKey{}, int64(42))
log.WithContext(ctx).Info("request done")
```

Typed getters provide strict extraction and use specific attribute constructors.

## Source enricher

The source enricher ([`NewSourceEnricher`], [`SourceEnricherOptions`], [`SourceFormat`]) appends caller source (`function file:line`) to each log.

```go
import (
	golog "github.com/jkaveri/golog/v2"
)

log, err := golog.NewLogger(golog.Config{
	Format:            golog.FormatText,
	Output:            "",
	Level:             golog.LevelDebug,
	EnableSource:      true,
	SourceFieldName:   "source",
	SourceFieldFormat: golog.SourceFormatFunctionFileLine,
	SourceSkipFrames:  0,
})
if err != nil {
	panic(err)
}
```

## Migration

Older code used `LoggerConfig` with an explicit `Writer` and `Enrichers` slice. That type is removed. Use declarative [`Config`] with [`NewLogger`], optional variadic [`Enricher`]s (e.g. [`NewContextEnricher`]), and [`NewLoggerWriter`] when you need a custom [`Writer`].

## Core Types

- `Logger`: emits records
- `Record`: one log event (`Time`, `Level`, `Message`, `Attr`s)
- `Enricher`: mutates a `RecordBuilder` before immutable record build
- [`Config`]: declarative setup ([`NewLogger`]); [`Config.Level`] is the minimum level to emit (zero means all levels)
- `Writer`: output backend (`Write` only; level filtering is on the logger)
- `TextWriter`: built-in line-oriented writer
- `JSONWriter`: built-in JSON line writer
- Text attributes use the same rules as [`Value.String`] and [`TextWriter`]

