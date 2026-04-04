package golog

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTextWriter_Write_format(t *testing.T) {
	var buf bytes.Buffer
	tw := NewTextWriter(&buf)
	ts := time.Date(2025, 3, 22, 12, 30, 0, 0, time.UTC)
	rb := newRecordBuilder(ts, LevelInfo, "hello world", 2)
	rb.AddAttrs(String("k1", "v1"), Int("n", 42))
	r := rb.Build()

	err := tw.Write(context.Background(), r)
	require.NoError(t, err)
	line := buf.String()
	require.True(t, strings.HasSuffix(line, "\n"), "want trailing newline: %q", line)
	require.Contains(t, line, "INFO")
	require.Contains(t, line, `"hello world"`)
	require.Contains(t, line, "k1=v1")
	require.Contains(t, line, "n=42")
}

func TestTextWriter_Write_nilOut(t *testing.T) {
	tw := &TextWriter{Out: nil}
	r := newRecordBuilder(testRecordTime, LevelInfo, "x", 0).Build()
	err := tw.Write(context.Background(), r)
	require.Error(t, err)
}

func TestTextWriter_concurrentWrite(t *testing.T) {
	var buf bytes.Buffer
	tw := NewTextWriter(&buf)
	r := newRecordBuilder(testRecordTime, LevelInfo, "msg", 0).Build()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = tw.Write(context.Background(), r)
		}()
	}
	wg.Wait()
	require.Equal(t, 8, strings.Count(buf.String(), "\n"))
}

func TestFormatTextAttr_groupFlat(t *testing.T) {
	attr := Attr{Value: GroupValue(Int("a", 1), String("b", "two"))}
	got := string(formatTextAttr(attr))
	require.Contains(t, got, "a=1")
	require.Contains(t, got, "b=two")
}

func TestFormatTextAttr_groupNested(t *testing.T) {
	inner := GroupValue(Int("n", 42))
	a := Attr{Key: "inner", Value: inner}
	attr := Attr{Value: GroupValue(a)}
	got := string(formatTextAttr(attr))
	require.Equal(t, "inner.n=42", got)
}

func TestFormatTextAttr_groupNestedPrefix(t *testing.T) {
	attr := Group("outer", Group("inner", Int("n", 1)))
	got := string(formatTextAttr(attr))
	require.Equal(t, "outer.inner.n=1", got)
}

func TestTextWriter_groupAttrPrefixedKeys(t *testing.T) {
	var buf bytes.Buffer
	tw := NewTextWriter(&buf)
	rb := newRecordBuilder(testRecordTime, LevelInfo, "m", 1)
	rb.AddAttr(Group("outer", Int("n", 42), Group("inner", String("k", "v"))))
	r := rb.Build()
	err := tw.Write(context.Background(), r)
	require.NoError(t, err)
	line := buf.String()
	require.Contains(t, line, "outer.n=42")
	require.Contains(t, line, "outer.inner.k=v")
}

func TestTextWriter_customTimeLayout(t *testing.T) {
	var buf bytes.Buffer
	tw := NewTextWriter(&buf)
	tw.TimeLayout = time.RFC1123
	ts := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	rb := newRecordBuilder(ts, LevelInfo, "msg", 0)
	r := rb.Build()
	err := tw.Write(context.Background(), r)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(buf.String(), ts.Format(time.RFC1123)), "line: %q", buf.String())
}

func TestTextWriter_durationFormats(t *testing.T) {
	type Args struct {
		format DurationFormat
	}
	type Expects struct {
		wantSubstring string
	}

	d := 1500 * time.Millisecond

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{
			name:    "go-string",
			args:    Args{format: DurationFormatGo},
			expects: Expects{wantSubstring: d.String()},
		},
		{
			name:    "seconds-float",
			args:    Args{format: DurationFormatSeconds},
			expects: Expects{wantSubstring: "1.5"},
		},
		{
			name:    "nanos-int",
			args:    Args{format: DurationFormatNanos},
			expects: Expects{wantSubstring: "1500000000"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			tw := NewTextWriter(&buf)
			tw.DurationFormat = tc.args.format
			rb := newRecordBuilder(testRecordTime, LevelInfo, "m", 1)
			rb.AddAttr(Duration("latency", d))
			r := rb.Build()
			err := tw.Write(context.Background(), r)
			require.NoError(t, err)
			require.Contains(t, buf.String(), tc.expects.wantSubstring)
		})
	}
}

func TestTextWriter_scalarAttrs(t *testing.T) {
	type Args struct {
		attr Attr
	}
	type Expects struct {
		wantSubstring string
	}

	ts := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{name: "float64", args: Args{attr: Float64("f", 0.5)}, expects: Expects{wantSubstring: "f=0.5"}},
		{name: "uint64", args: Args{attr: Uint64("u", 8)}, expects: Expects{wantSubstring: "u=8"}},
		{name: "bool-true", args: Args{attr: Bool("b", true)}, expects: Expects{wantSubstring: "b=true"}},
		{name: "time", args: Args{attr: Time("t", ts)}, expects: Expects{wantSubstring: ts.String()}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			tw := NewTextWriter(&buf)
			rb := newRecordBuilder(testRecordTime, LevelInfo, "m", 1)
			rb.AddAttr(tc.args.attr)
			r := rb.Build()
			err := tw.Write(context.Background(), r)
			require.NoError(t, err)
			require.Contains(t, buf.String(), tc.expects.wantSubstring)
		})
	}
}

func TestFormatTextAttr_joinKeyPrefixOnly(t *testing.T) {
	attr := Group("", String("k", "v"))
	got := string(formatTextAttr(attr))
	require.Equal(t, "k=v", got)
}

func TestFormatTextAttr_groupSkipsEmptyPieces(t *testing.T) {
	attr := Attr{Value: GroupValue(
		Attr{Key: "a", Value: StringValue("1")},
		Attr{Key: "skip", Value: GroupValue()},
		Attr{Key: "b", Value: StringValue("2")},
	)}
	got := string(formatTextAttr(attr))
	require.Contains(t, got, "a=1")
	require.Contains(t, got, "b=2")
}
