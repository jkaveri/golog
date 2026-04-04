package golog

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/jkaveri/golog/v2/internal/buffer"
)

// JSONWriter implements [Writer] by writing one JSON object per log record to
// [JSONWriter.Out]. Each line includes "time", "level", and "msg" plus user
// attributes (group values become nested objects). Attributes whose keys
// collide with those reserved names are emitted under a nested "attrs" object.
type JSONWriter struct {
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

// NewJSONWriter returns a [JSONWriter] that writes to out with default time
// layout [time.RFC3339Nano]. Customize [JSONWriter.DurationFormat] or
// [JSONWriter.TimeLayout] on the returned value as needed.
func NewJSONWriter(out io.Writer) *JSONWriter {
	return &JSONWriter{
		Out:        out,
		TimeLayout: time.RFC3339Nano,
	}
}

var errJSONWriterNilOut = errors.New("golog: JSONWriter: Out is nil")

func (j *JSONWriter) timeLayout() string {
	if j.TimeLayout != "" {
		return j.TimeLayout
	}

	return time.RFC3339Nano
}

// Write encodes record as one JSON object and a newline. ctx is unused but
// required by [Writer].
// It must not retain record after return.
func (j *JSONWriter) Write(ctx context.Context, record Record) error {
	if j.Out == nil {
		return errJSONWriterNilOut
	}

	buf := buffer.New()
	defer buf.Free()

	if err := buf.WriteByte('{'); err != nil {
		return err
	}

	first := true
	j.appendKVTime(buf, &first, "time", record.Time, j.timeLayout())
	j.appendKVString(buf, &first, "level", record.Level.String())
	j.appendKVString(buf, &first, "msg", record.Message)

	var encodeErr error

	record.RangeAttrs(func(a Attr) bool {
		if encodeErr != nil {
			return false
		}

		if j.isReservedJSONKey(a.Key) {
			return true
		}

		encodeErr = j.appendAttr(buf, &first, a)

		return encodeErr == nil
	})

	if encodeErr != nil {
		return encodeErr
	}

	hasReserved := false

	record.RangeAttrs(func(a Attr) bool {
		if j.isReservedJSONKey(a.Key) {
			hasReserved = true
			return false
		}

		return true
	})

	if hasReserved {
		j.appendKey(buf, &first, "attrs")

		if err := buf.WriteByte('{'); err != nil {
			return err
		}

		attrsFirst := true

		record.RangeAttrs(func(a Attr) bool {
			if encodeErr != nil {
				return false
			}

			if !j.isReservedJSONKey(a.Key) {
				return true
			}

			encodeErr = j.appendAttr(buf, &attrsFirst, a)

			return encodeErr == nil
		})

		if encodeErr != nil {
			return encodeErr
		}

		if err := buf.WriteByte('}'); err != nil {
			return err
		}
	}

	if err := buf.WriteByte('}'); err != nil {
		return err
	}

	if err := buf.WriteByte('\n'); err != nil {
		return err
	}

	j.Mu.Lock()
	_, err := j.Out.Write(*buf)
	j.Mu.Unlock()

	return err
}

func (*JSONWriter) isReservedJSONKey(key string) bool {
	return key == "time" || key == "level" || key == "msg"
}

func (j *JSONWriter) appendKVString(
	buf *buffer.Buffer,
	first *bool,
	key, value string,
) {
	j.appendKey(buf, first, key)

	*buf = strconv.AppendQuote(*buf, value)
}

func (j *JSONWriter) appendKVTime(
	buf *buffer.Buffer,
	first *bool,
	key string,
	t time.Time,
	layout string,
) {
	j.appendKey(buf, first, key)

	*buf = append(*buf, '"')
	*buf = t.AppendFormat(*buf, layout)
	*buf = append(*buf, '"')
}

func (*JSONWriter) appendKey(buf *buffer.Buffer, first *bool, key string) {
	if !*first {
		*buf = append(*buf, ',')
	}

	*first = false
	*buf = strconv.AppendQuote(*buf, key)
	*buf = append(*buf, ':')
}

// appendAttr and appendJSONValue encode [Attr] / [Value] as JSON (nested
// objects for groups).
func (j *JSONWriter) appendAttr(buf *buffer.Buffer, first *bool, a Attr) error {
	j.appendKey(buf, first, a.Key)
	return j.appendJSONValue(buf, a.Value)
}

func (j *JSONWriter) appendJSONValue(buf *buffer.Buffer, v Value) error {
	switch v.Kind() {
	case KindString:
		*buf = strconv.AppendQuote(*buf, v.Any().(string))
	case KindInt64:
		*buf = strconv.AppendInt(*buf, v.Int64(), 10)
	case KindUint64:
		*buf = strconv.AppendUint(*buf, v.Uint64(), 10)
	case KindFloat64:
		*buf = strconv.AppendFloat(*buf, v.Float64(), 'g', -1, 64)
	case KindBool:
		*buf = strconv.AppendBool(*buf, v.Bool())
	case KindDuration:
		switch j.DurationFormat {
		case DurationFormatSeconds:
			*buf = strconv.AppendFloat(
				*buf,
				v.Duration().Seconds(),
				'f',
				-1,
				64,
			)
		case DurationFormatNanos:
			*buf = strconv.AppendInt(*buf, v.Duration().Nanoseconds(), 10)
		default:
			*buf = strconv.AppendQuote(*buf, v.Duration().String())
		}
	case KindTime:
		*buf = strconv.AppendQuote(*buf, v.Time().Format(j.timeLayout()))
	case KindGroup:
		*buf = append(*buf, '{')

		first := true
		for _, child := range v.Group() {
			if err := j.appendAttr(buf, &first, child); err != nil {
				return err
			}
		}

		*buf = append(*buf, '}')
	case KindAny:
		raw, err := json.Marshal(v.Any())
		if err != nil {
			return err
		}

		*buf = append(*buf, raw...)
	default:
		panic("golog: bad value kind")
	}

	return nil
}
