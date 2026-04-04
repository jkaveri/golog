package golog

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// testRecordTime is used when the test does not assert on the timestamp.
var testRecordTime = time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)

func TestJSONWriter_Write_nilOut(t *testing.T) {
	jw := &JSONWriter{Out: nil}
	r := newRecordBuilder(testRecordTime, LevelInfo, "x", 0).Build()
	err := jw.Write(context.Background(), r)
	require.Error(t, err)
}

func TestJSONWriter_Write_format(t *testing.T) {
	var buf bytes.Buffer
	jw := NewJSONWriter(&buf)
	ts := time.Date(2025, 3, 22, 12, 30, 0, 0, time.UTC)
	rb := newRecordBuilder(ts, LevelInfo, "hello world", 2)
	rb.AddAttrs(String("k1", "v1"), Int("n", 42))
	r := rb.Build()

	err := jw.Write(context.Background(), r)
	require.NoError(t, err)
	line := buf.String()
	require.True(t, strings.HasSuffix(line, "\n"), "want trailing newline: %q", line)

	var obj map[string]any
	err = json.Unmarshal([]byte(strings.TrimSpace(line)), &obj)
	require.NoError(t, err)
	require.Equal(t, "INFO", obj["level"])
	require.Equal(t, "hello world", obj["msg"])
	require.Equal(t, ts.Format(time.RFC3339Nano), obj["time"])
	require.Equal(t, "v1", obj["k1"])
	require.Equal(t, float64(42), obj["n"])
}

func TestJSONWriter_groupAttrNested(t *testing.T) {
	var buf bytes.Buffer
	jw := NewJSONWriter(&buf)
	rb := newRecordBuilder(testRecordTime, LevelInfo, "m", 1)
	rb.AddAttr(Group("request", String("id", "r1"), Group("user", Int("id", 42))))
	r := rb.Build()

	err := jw.Write(context.Background(), r)
	require.NoError(t, err)
	line := strings.TrimSpace(buf.String())

	var obj map[string]any
	err = json.Unmarshal([]byte(line), &obj)
	require.NoError(t, err)
	request, ok := obj["request"].(map[string]any)
	require.True(t, ok, "request object: %#v", obj["request"])
	require.Equal(t, "r1", request["id"])
	user, ok := request["user"].(map[string]any)
	require.True(t, ok, "request.user: %#v", request["user"])
	require.Equal(t, float64(42), user["id"])
}

func TestJSONWriter_reservedKeyConflict(t *testing.T) {
	var buf bytes.Buffer
	jw := NewJSONWriter(&buf)
	rb := newRecordBuilder(
		time.Date(2025, 3, 22, 12, 30, 0, 0, time.UTC),
		LevelInfo,
		"original",
		3,
	)
	rb.AddAttrs(String("msg", "attr-msg"), String("level", "attr-level"), String("time", "attr-time"))
	r := rb.Build()

	err := jw.Write(context.Background(), r)
	require.NoError(t, err)
	line := strings.TrimSpace(buf.String())

	var obj map[string]any
	err = json.Unmarshal([]byte(line), &obj)
	require.NoError(t, err)
	require.Equal(t, "original", obj["msg"])
	require.Equal(t, "INFO", obj["level"])
	require.Equal(t, r.Time.Format(time.RFC3339Nano), obj["time"])
	attrs, ok := obj["attrs"].(map[string]any)
	require.True(t, ok, "attrs bucket: %#v", obj["attrs"])
	require.Equal(t, "attr-msg", attrs["msg"])
	require.Equal(t, "attr-level", attrs["level"])
	require.Equal(t, "attr-time", attrs["time"])
}

func TestJSONWriter_concurrentWrite(t *testing.T) {
	var buf bytes.Buffer
	jw := NewJSONWriter(&buf)
	r := newRecordBuilder(testRecordTime, LevelInfo, "msg", 0).Build()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = jw.Write(context.Background(), r)
		}()
	}
	wg.Wait()
	require.Equal(t, 8, strings.Count(buf.String(), "\n"))
}

func TestJSONWriter_customTimeLayout(t *testing.T) {
	var buf bytes.Buffer
	jw := NewJSONWriter(&buf)
	jw.TimeLayout = time.RFC3339
	ts := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	rb := newRecordBuilder(ts, LevelInfo, "m", 1)
	rb.AddAttr(Time("at", ts))
	r := rb.Build()

	err := jw.Write(context.Background(), r)
	require.NoError(t, err)
	line := strings.TrimSpace(buf.String())
	var obj map[string]any
	err = json.Unmarshal([]byte(line), &obj)
	require.NoError(t, err)
	require.Equal(t, "m", obj["msg"])
	at, ok := obj["at"].(string)
	require.True(t, ok)
	require.Equal(t, ts.Format(time.RFC3339), at)
}

func TestJSONWriter_scalarKindsAndDurationFormats(t *testing.T) {
	type Args struct {
		durationFormat DurationFormat
		attr           Attr
		checkJSONKey   string
		wantSubstring  string
	}

	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	dur := 3 * time.Second

	testCases := []struct {
		name string
		args Args
	}{
		{
			name: "float64",
			args: Args{
				attr:          Float64("f", 1.25),
				checkJSONKey:  "f",
				wantSubstring: "1.25",
			},
		},
		{
			name: "uint64",
			args: Args{
				attr:          Uint64("u", 99),
				checkJSONKey:  "u",
				wantSubstring: "99",
			},
		},
		{
			name: "bool",
			args: Args{
				attr:          Bool("b", false),
				checkJSONKey:  "b",
				wantSubstring: "false",
			},
		},
		{
			name: "duration-seconds",
			args: Args{
				durationFormat: DurationFormatSeconds,
				attr:           Duration("d", dur),
				checkJSONKey:   "d",
				wantSubstring:  "3",
			},
		},
		{
			name: "duration-nanos",
			args: Args{
				durationFormat: DurationFormatNanos,
				attr:           Duration("d", dur),
				checkJSONKey:   "d",
				wantSubstring:  "3000000000",
			},
		},
		{
			name: "duration-go-string",
			args: Args{
				durationFormat: DurationFormatGo,
				attr:           Duration("d", dur),
				checkJSONKey:   "d",
				wantSubstring:  "3s",
			},
		},
		{
			name: "any-json",
			args: Args{
				attr:          Any("meta", map[string]int{"x": 1}),
				checkJSONKey:  "meta",
				wantSubstring: `"x":1`,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			jw := NewJSONWriter(&buf)
			jw.DurationFormat = tc.args.durationFormat
			rb := newRecordBuilder(testRecordTime, LevelInfo, "m", 1)
			rb.AddAttr(tc.args.attr)
			r := rb.Build()
			err := jw.Write(context.Background(), r)
			require.NoError(t, err)
			line := strings.TrimSpace(buf.String())
			require.Contains(t, line, tc.args.checkJSONKey)
			require.Contains(t, line, tc.args.wantSubstring)
		})
	}

	t.Run("kind-time-field", func(t *testing.T) {
		var buf bytes.Buffer
		jw := NewJSONWriter(&buf)
		rb := newRecordBuilder(testRecordTime, LevelInfo, "m", 1)
		rb.AddAttr(Time("when", ts))
		r := rb.Build()
		err := jw.Write(context.Background(), r)
		require.NoError(t, err)
		var obj map[string]any
		err = json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &obj)
		require.NoError(t, err)
		when, ok := obj["when"].(string)
		require.True(t, ok)
		require.Equal(t, ts.Format(time.RFC3339Nano), when)
	})
}
