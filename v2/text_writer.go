package golog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/jkaveri/golog/v2/internal/buffer"
)

// TextWriter implements [Writer] by writing one human-readable line per log
// record to [TextWriter.Out]: time, level, quoted message, then space-separated
// key=value attributes. For group values, nested keys are prefixed with the
// parent key and '.'. Log level filtering is done by [Config.Level] on the
// logger, not by TextWriter.
//
// The mutex only wraps the final [io.Writer.Write] on [TextWriter.Out] (after
// the
// line is fully formatted into a buffer), so formatting work is not serialized.
// This matches [log/slog] built-in handlers. Use a single TextWriter when [Out]
// is not safe for concurrent writes.
type TextWriter struct {
	Mu sync.Mutex
	// Out is the destination. If nil, Write returns an error.
	Out io.Writer
	// TimeLayout is passed to [time.Time.Format]. If empty, [time.RFC3339Nano]
	// is used.
	TimeLayout string
	// DurationFormat controls [KindDuration] attribute encoding. If empty,
	// [DurationFormatGo] is used.
	DurationFormat DurationFormat
}

// NewTextWriter returns a [TextWriter] that writes to out with default time
// layout [time.RFC3339Nano]. Set [TextWriter.DurationFormat] or
// [TextWriter.TimeLayout] on the returned value to customize encoding.
func NewTextWriter(out io.Writer) *TextWriter {
	return &TextWriter{
		Out:        out,
		TimeLayout: time.RFC3339Nano,
	}
}

var errTextWriterNilOut = errors.New("golog: TextWriter: Out is nil")

func (t *TextWriter) timeLayout() string {
	if t.TimeLayout != "" {
		return t.TimeLayout
	}

	return time.RFC3339Nano
}

func (t *TextWriter) durationFormat() DurationFormat {
	return t.DurationFormat
}

func (t *TextWriter) formatAttr(a Attr) []byte {
	return (textFormat{durationFormat: t.durationFormat()}).attrWithPrefix(
		a,
		"",
	)
}

// Write formats record as one line: timestamp, level, quoted message, then
// attributes. ctx is unused but kept for the [Writer] interface. It must not
// retain record after return.
func (t *TextWriter) Write(ctx context.Context, record Record) error {
	if t.Out == nil {
		return errTextWriterNilOut
	}

	buf := buffer.New()
	defer buf.Free()

	if _, err := buf.WriteString(
		record.Time.Format(t.timeLayout()),
	); err != nil {
		return err
	}

	if err := buf.WriteByte(' '); err != nil {
		return err
	}

	if _, err := buf.WriteString(record.Level.String()); err != nil {
		return err
	}

	if err := buf.WriteByte(' '); err != nil {
		return err
	}

	if _, err := buf.Write(
		strconv.AppendQuote(nil, record.Message),
	); err != nil {
		return err
	}

	var attrErr error

	record.RangeAttrs(func(a Attr) bool {
		if attrErr != nil {
			return false
		}

		formatted := t.formatAttr(a)
		if len(formatted) == 0 {
			return true
		}

		if err := buf.WriteByte(' '); err != nil {
			attrErr = err
			return false
		}

		if _, err := buf.Write(formatted); err != nil {
			attrErr = err
			return false
		}

		return true
	})

	if attrErr != nil {
		return attrErr
	}

	if err := buf.WriteByte('\n'); err != nil {
		return err
	}

	t.Mu.Lock()
	_, err := t.Out.Write(*buf)
	t.Mu.Unlock()

	return err
}

// textFormat implements attribute text formatting (key=value, space-separated
// groups;
// nested keys use dot prefixes). Shared by [TextWriter] and [Value.String].
type textFormat struct {
	durationFormat DurationFormat
}

// formatTextAttr formats an Attr as bytes for text output ([Value.String] path;
// default duration format).
func formatTextAttr(attr Attr) []byte {
	return (textFormat{durationFormat: DurationFormatGo}).attrWithPrefix(
		attr,
		"",
	)
}

func appendDurationText(dst []byte, d time.Duration, f DurationFormat) []byte {
	switch f {
	case DurationFormatSeconds:
		return strconv.AppendFloat(dst, d.Seconds(), 'f', -1, 64)
	case DurationFormatNanos:
		return strconv.AppendInt(dst, d.Nanoseconds(), 10)
	default:
		return append(dst, d.String()...)
	}
}

func (f textFormat) scalar(v Value) []byte {
	var dst []byte

	switch v.Kind() {
	case KindString:
		return append(dst, v.str()...)
	case KindInt64:
		return strconv.AppendInt(dst, v.Int64(), 10)
	case KindUint64:
		return strconv.AppendUint(dst, v.Uint64(), 10)
	case KindFloat64:
		return strconv.AppendFloat(dst, v.float(), 'g', -1, 64)
	case KindBool:
		return strconv.AppendBool(dst, v.bool())
	case KindDuration:
		return appendDurationText(dst, v.duration(), f.durationFormat)
	case KindTime:
		return append(dst, v.time().String()...)
	case KindAny:
		return fmt.Append(dst, v.any)
	default:
		panic(fmt.Sprintf("bad kind: %s", v.Kind()))
	}
}

func (textFormat) joinKey(prefix, key string) string {
	if prefix == "" {
		return key
	}

	if key == "" {
		return prefix
	}

	return prefix + "." + key
}

func (f textFormat) groupWithPrefix(attrs []Attr, prefix string) []byte {
	if len(attrs) == 0 {
		return nil
	}

	var b []byte

	for _, a := range attrs {
		piece := f.attrWithPrefix(a, prefix)
		if len(piece) == 0 {
			continue
		}

		if len(b) > 0 {
			b = append(b, ' ')
		}

		b = append(b, piece...)
	}

	return b
}

func (f textFormat) attrWithPrefix(attr Attr, prefix string) []byte {
	fullKey := f.joinKey(prefix, attr.Key)
	if attr.Value.Kind() == KindGroup {
		return f.groupWithPrefix(attr.Value.group(), fullKey)
	}

	if fullKey == "" {
		return f.scalar(attr.Value)
	}

	var b []byte

	b = append(b, fullKey...)
	b = append(b, '=')
	b = append(b, f.scalar(attr.Value)...)

	return b
}
