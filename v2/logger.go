package golog

import (
	"context"
	"time"
)

// Logger is the structured logging interface.
//
// Debug, Info, and Error correspond to [LevelDebug], [LevelInfo], and [LevelError].
// See package documentation (“Logging levels and philosophy”) for when to use each.
//
// Create loggers with [NewLogger] from [Config], or [NewLoggerWriter] for a custom [Writer].
// See also [Example], [ExampleNewLogger], and [ExampleNewLoggerWriter].
type Logger interface {
	// Debug logs at [LevelDebug]. Use for troubleshooting detail; see package docs.
	Debug(msg string, args ...Attr)
	// Info logs at [LevelInfo]. Use for normal operational messages; see package docs.
	Info(msg string, args ...Attr)
	// Error logs at [LevelError]. Use for failures you record at this layer; see package docs.
	Error(msg string, args ...Attr)

	// With returns a derived logger that prepends attrs to every later log line from that
	// logger. It is useful for request-scoped or component-scoped fields without repeating
	// them at each call site. A zero-length args list returns the receiver unchanged.
	// See [ExampleLogger_With].
	With(args ...Attr) Logger

	// WithContext returns a derived logger that associates ctx with each log call so
	// [Enricher] implementations (e.g. [NewContextEnricher]) can read context values.
	// The context is not stored on the emitted [Record] unless an enricher copies it into attributes.
	// See [ExampleLogger_WithContext] and [ExampleNewLogger_contextEnricher].
	WithContext(ctx context.Context) Logger

	// WithError returns a derived logger that adds a string attribute "error" with err.Error()
	// to each subsequent log line. If err is nil, the receiver is returned unchanged.
	// See [ExampleLogger_WithError].
	WithError(err error) Logger
	// SetLevel sets the minimum severity this logger emits. It uses the same rule as
	// [Config.Level] in [NewLogger]: a log is skipped when its event level is strictly
	// less than this threshold (no [Record], no enricher work). The zero value is
	// [LevelDebug], so by default every standard level is emitted; raising the
	// minimum to [LevelInfo] suppresses [LevelDebug] lines only; raising it to
	// [LevelError] leaves only error-severity events. Each [Logger] instance has its own threshold;
	// [Logger.With] returns a child whose threshold is copied at creation time.
	SetLevel(level Level)
}

// LoggerScenario identifies a high-level deployment style for callers that want to
// branch on environment. Use [DevelopmentConfig], [ProductionConfig], [NewDevelopmentLogger],
// or [NewProductionLogger] for ready-made settings.
type LoggerScenario uint8

const (
	// ScenarioCustom uses explicit [Config] with [NewLogger].
	ScenarioCustom LoggerScenario = iota
	// ScenarioDevelopment configures a text logger intended for local development.
	ScenarioDevelopment
	// ScenarioProduction configures a JSON logger intended for production systems.
	ScenarioProduction
)

// NewLogger builds a [Logger] from [Config] and optional extra [Enricher]s.
// It opens the output described by [Config.Output], constructs a [TextWriter] or [JSONWriter],
// applies [Config.Level] as the minimum severity, and optionally prepends a source enricher
// when [Config.EnableSource] or [Config.SourceFieldName] is set.
//
// Returns an error if the format is unknown or the output file cannot be opened.
//
// Example:
//
//	log, err := NewLogger(Config{Format: FormatJSON, Output: "", Level: LevelInfo})
//
// See [ExampleNewLogger] and [ExampleNewLogger_contextEnricher].
func NewLogger(cfg Config, enrichers ...Enricher) (Logger, error) {
	w, err := buildWriterFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &logger{
		ctx:       context.Background(),
		writer:    w,
		minLevel:  cfg.Level,
		enrichers: enrichersFromConfig(cfg, enrichers),
	}, nil
}

// NewLoggerWriter builds a [Logger] with an explicit [Writer] (for tests, benchmarks,
// and custom sinks such as fan-out or filtering). The level argument sets the minimum
// severity ([Config.Level] semantics: zero is [LevelDebug]).
//
// Example:
//
//	log := NewLoggerWriter(NewTextWriter(os.Stdout), LevelInfo)
//
// See [ExampleNewLoggerWriter].
func NewLoggerWriter(w Writer, level Level, enrichers ...Enricher) Logger {
	return &logger{
		ctx:       context.Background(),
		writer:    w,
		minLevel:  level,
		enrichers: append([]Enricher(nil), enrichers...),
	}
}

// logger implements [Logger] using [Writer] and [Enricher] pipelines.
type logger struct {
	ctx       context.Context
	writer    Writer
	minLevel  Level
	enrichers []Enricher
	attrs     []Attr // from [Logger.With] / [Logger.WithError]; prepended to each log
}

func (l *logger) Debug(msg string, args ...Attr) {
	l.log(LevelDebug, msg, args...)
}

func (l *logger) Info(msg string, args ...Attr) {
	l.log(LevelInfo, msg, args...)
}

func (l *logger) Error(msg string, args ...Attr) {
	l.log(LevelError, msg, args...)
}

func (l *logger) With(args ...Attr) Logger {
	if len(args) == 0 {
		return l
	}
	next := *l
	next.attrs = append(append([]Attr(nil), l.attrs...), args...)
	return &next
}

func (l *logger) WithContext(ctx context.Context) Logger {
	next := *l
	next.ctx = ctx
	return &next
}

func (l *logger) WithError(err error) Logger {
	if err == nil {
		return l
	}
	return l.With(String("error", err.Error()))
}

func (l *logger) SetLevel(level Level) {
	l.minLevel = level
}

// log builds a record when level passes [logger.minLevel], runs enrichers via RecordBuilder, then writes.
// Writer errors are ignored to keep the hot path simple.
func (l *logger) log(level Level, msg string, args ...Attr) {
	ctx := l.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	if l.writer == nil {
		return
	}
	if level < l.minLevel {
		return
	}

	record := Record{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
	}
	if len(l.enrichers) == 0 && len(l.attrs) == 0 {
		// Fast path: use call-site attrs directly when no enrichment/scoped attrs are present.
		record.attrs = args
	} else {
		builder := newRecordBuilder(record.Time, level, msg, len(l.attrs)+len(args))
		builder.AddAttrs(l.attrs...)
		builder.AddAttrs(args...)
		for _, e := range l.enrichers {
			if e != nil {
				e.Enrich(ctx, &builder)
			}
		}
		record = builder.Build()
	}

	_ = l.writer.Write(ctx, record)
}
