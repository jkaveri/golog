package golog

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// LogFormat selects the built-in line encoder for [Config]. Use [FormatText] for human-readable
// lines or [FormatJSON] for one JSON object per line. Empty string is treated like [FormatText].
type LogFormat string

const (
	// FormatText writes human-readable lines via [TextWriter].
	FormatText LogFormat = "text"
	// FormatJSON writes one JSON object per line via [JSONWriter].
	FormatJSON LogFormat = "json"
)

// DurationFormat controls how [KindDuration] values are rendered by [TextWriter] and [JSONWriter].
// Use [DurationFormatGo] (or empty) for [time.Duration.String], [DurationFormatSeconds] for a float
// of seconds, or [DurationFormatNanos] for integer nanoseconds.
type DurationFormat string

const (
	// DurationFormatGo uses [time.Duration.String] (default).
	DurationFormatGo DurationFormat = ""
	// DurationFormatSeconds writes the duration as floating-point seconds (no unit suffix).
	DurationFormatSeconds DurationFormat = "seconds"
	// DurationFormatNanos writes the duration as integer nanoseconds.
	DurationFormatNanos DurationFormat = "nanos"
)

// Config is the declarative logger configuration used by [NewLogger]. It does not embed
// [Writer] or []Enricher: output is derived from [Config.Format] and [Config.Output], and
// optional [Enricher]s are passed as variadic arguments to [NewLogger] (for example
// [NewContextEnricher]). Source location can be enabled via [Config.EnableSource] or a
// non-empty [Config.SourceFieldName].
type Config struct {
	// Level is the minimum level to emit. Zero is [LevelDebug] (all levels).
	Level Level

	// Format selects [TextWriter] or [JSONWriter].
	Format LogFormat

	// Output is the destination: empty, "-", or "stdout" (case-insensitive) mean [os.Stdout];
	// any other non-empty string is treated as a file path opened append-only (create if missing).
	Output string

	// TimeFormat is the layout for log record timestamps ([time.Time.Format]).
	// If empty, [time.RFC3339Nano] is used.
	TimeFormat string

	// DurationFormat controls duration attribute encoding. If empty, [DurationFormatGo] is used.
	DurationFormat DurationFormat

	// EnableSource turns on the source enricher ([NewSourceEnricher]). It is also enabled if
	// [Config.SourceFieldName] is non-empty.
	EnableSource bool

	// SourceFieldName is the attribute key for caller source (default "source").
	SourceFieldName string

	// SourceSkipFrames is passed to [SourceEnricherOptions.Skip] (extra frames beyond the base skip).
	SourceSkipFrames int

	// SourceFieldFormat is passed to [SourceEnricherOptions.Format].
	SourceFieldFormat SourceFormat
}

func openConfigOutput(output string) (io.Writer, error) {
	s := strings.TrimSpace(output)
	if s == "" || s == "-" || strings.EqualFold(s, "stdout") || strings.EqualFold(s, "stderr") {
		return os.Stdout, nil
	}
	f, err := os.OpenFile(s, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("golog: open log file %q: %w", s, err)
	}
	return f, nil
}

func buildWriterFromConfig(cfg Config) (Writer, error) {
	out, err := openConfigOutput(cfg.Output)
	if err != nil {
		return nil, err
	}
	tl := cfg.TimeFormat
	if tl == "" {
		tl = time.RFC3339Nano
	}
	df := cfg.DurationFormat
	switch cfg.Format {
	case FormatText, "":
		return &TextWriter{Out: out, TimeLayout: tl, DurationFormat: df}, nil
	case FormatJSON:
		return &JSONWriter{Out: out, TimeLayout: tl, DurationFormat: df}, nil
	default:
		return nil, fmt.Errorf("golog: unknown Format %q (want %q or %q)", cfg.Format, FormatText, FormatJSON)
	}
}

func enrichersFromConfig(cfg Config, extra []Enricher) []Enricher {
	var list []Enricher
	if cfg.EnableSource || cfg.SourceFieldName != "" {
		opts := SourceEnricherOptions{
			FieldName: cfg.SourceFieldName,
			Format:    cfg.SourceFieldFormat,
			Skip:      cfg.SourceSkipFrames,
		}
		list = append(list, NewSourceEnricher(opts))
	}
	for _, e := range extra {
		if e != nil {
			list = append(list, e)
		}
	}
	return list
}

// DevelopmentConfig returns a [Config] for local development: [FormatText] on stdout,
// minimum level [LevelDebug], [time.ANSIC] timestamps, and caller source at the "caller"
// attribute ([SourceFormatFileLine], [SourceSkipFrames] 2). Use [NewDevelopmentLogger] or
// merge fields into a custom [Config] with [NewLogger].
func DevelopmentConfig() Config {
	return Config{
		Format:            FormatText,
		Output:            "",
		Level:             LevelDebug,
		TimeFormat:        time.ANSIC,
		EnableSource:      true,
		SourceFieldName:   "caller",
		SourceFieldFormat: SourceFormatFileLine,
		SourceSkipFrames:  2,
	}
}

// ProductionConfig returns a [Config] for production: [FormatJSON] on stdout, minimum level
// [LevelInfo], [time.RFC3339Nano] timestamps, [DurationFormatSeconds] for durations, and caller
// source like [DevelopmentConfig]. Use [NewProductionLogger] or pass to [NewLogger] with overrides.
func ProductionConfig() Config {
	return Config{
		Format:            FormatJSON,
		Output:            "stdout",
		Level:             LevelInfo,
		TimeFormat:        time.RFC3339Nano,
		DurationFormat:    DurationFormatSeconds,
		EnableSource:      true,
		SourceFieldName:   "caller",
		SourceFieldFormat: SourceFormatFileLine,
		SourceSkipFrames:  2,
	}
}

// NewDevelopmentLogger builds a [Logger] from [DevelopmentConfig] and optional extra [Enricher]s.
func NewDevelopmentLogger(enrichers ...Enricher) (Logger, error) {
	return NewLogger(DevelopmentConfig(), enrichers...)
}

// NewProductionLogger builds a [Logger] from [ProductionConfig] and optional extra [Enricher]s.
func NewProductionLogger(enrichers ...Enricher) (Logger, error) {
	return NewLogger(ProductionConfig(), enrichers...)
}
