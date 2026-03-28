// Source enricher: append caller location ([SourceFormatFunctionFileLine] or
// [SourceFormatFileLine]) to log records.
//
// Exported types here are the source-enricher feature ([NewSourceEnricher],
// [SourceEnricherOptions], [SourceFormat]).
package golog

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
)

// SourceFormat controls how caller source is rendered by the source enricher.
type SourceFormat uint8

const (
	// SourceFormatFunctionFileLine renders "function file:line".
	SourceFormatFunctionFileLine SourceFormat = iota
	// SourceFormatFileLine renders "file:line" (basename only, no function
	// name).
	SourceFormatFileLine
)

// SourceEnricherOptions configures the source enricher.
type SourceEnricherOptions struct {
	// FieldName is the attribute key. Defaults to "source".
	FieldName string
	// Format controls source value rendering: [SourceFormatFunctionFileLine]
	// (default) or
	// [SourceFormatFileLine] for basename:line only.
	Format SourceFormat
	// Skip adds caller frames to skip beyond the internal default.
	Skip int
}

const (
	sourceEnricherDefaultField = "source"
	sourceEnricherBaseSkip     = 2
)

type sourceEnricher struct {
	fieldName string
	format    SourceFormat
	skip      int
}

// NewSourceEnricher returns an [Enricher] that appends a caller location
// attribute using [runtime.Caller]. Combine with [Config.EnableSource] / source
// fields in [Config], or pass manually to [NewLogger] / [NewLoggerWriter]. Tune
// frame skipping with [SourceEnricherOptions.Skip]
// if wrappers sit between the log call and the runtime frame.
func NewSourceEnricher(opts SourceEnricherOptions) Enricher {
	fieldName := opts.FieldName
	if fieldName == "" {
		fieldName = sourceEnricherDefaultField
	}

	skip := sourceEnricherBaseSkip + opts.Skip
	if skip < 0 {
		skip = 0
	}

	return sourceEnricher{
		fieldName: fieldName,
		format:    opts.Format,
		skip:      skip,
	}
}

func (e sourceEnricher) Enrich(ctx context.Context, builder *RecordBuilder) {
	_ = ctx

	if attr, ok := e.sourceAttr(); ok {
		builder.AddAttr(attr)
	}
}

func (e sourceEnricher) sourceAttr() (Attr, bool) {
	pc, file, line, ok := runtime.Caller(e.skip)
	if !ok {
		return Attr{}, false
	}

	switch e.format {
	case SourceFormatFileLine:
		value := fmt.Sprintf("%s:%d", filepath.Base(file), line)
		return String(e.fieldName, value), true
	case SourceFormatFunctionFileLine:
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			return Attr{}, false
		}

		value := fmt.Sprintf(
			"%s %s:%d",
			fn.Name(),
			filepath.Base(file),
			line,
		)

		return String(e.fieldName, value), true
	default:
		return Attr{}, false
	}
}
