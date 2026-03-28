package golog

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTextWriter_Write_format(t *testing.T) {
	var buf bytes.Buffer
	tw := NewTextWriter(&buf)
	ts := time.Date(2025, 3, 22, 12, 30, 0, 0, time.UTC)
	rb := newRecordBuilder(ts, LevelInfo, "hello world", 2)
	rb.AddAttrs(String("k1", "v1"), Int("n", 42))
	r := rb.Build()

	if err := tw.Write(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	line := buf.String()
	if !strings.HasSuffix(line, "\n") {
		t.Fatalf("want trailing newline: %q", line)
	}
	if !strings.Contains(line, "INFO") {
		t.Fatalf("want INFO in line: %q", line)
	}
	if !strings.Contains(line, `"hello world"`) {
		t.Fatalf("want quoted message: %q", line)
	}
	if !strings.Contains(line, "k1=v1") || !strings.Contains(line, "n=42") {
		t.Fatalf("want attrs: %q", line)
	}
}

func TestTextWriter_Write_nilOut(t *testing.T) {
	tw := &TextWriter{Out: nil}
	r := newRecordBuilder(testRecordTime, LevelInfo, "x", 0).Build()
	err := tw.Write(context.Background(), r)
	if err == nil {
		t.Fatal("want error for nil Out")
	}
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
	lines := strings.Count(buf.String(), "\n")
	if lines != 8 {
		t.Fatalf("want 8 lines, got %d", lines)
	}
}

func TestFormatTextAttr_groupFlat(t *testing.T) {
	attr := Attr{Value: GroupValue(Int("a", 1), String("b", "two"))}
	got := string(formatTextAttr(attr))
	if !strings.Contains(got, "a=1") || !strings.Contains(got, "b=two") {
		t.Fatalf("got %q", got)
	}
}

func TestFormatTextAttr_groupNested(t *testing.T) {
	inner := GroupValue(Int("n", 42))
	a := Attr{Key: "inner", Value: inner}
	attr := Attr{Value: GroupValue(a)}
	got := string(formatTextAttr(attr))
	if got != "inner.n=42" {
		t.Fatalf("want inner.n=42, got %q", got)
	}
}

func TestFormatTextAttr_groupNestedPrefix(t *testing.T) {
	attr := Group("outer", Group("inner", Int("n", 1)))
	got := string(formatTextAttr(attr))
	if got != "outer.inner.n=1" {
		t.Fatalf("want outer.inner.n=1, got %q", got)
	}
}

func TestTextWriter_groupAttrPrefixedKeys(t *testing.T) {
	var buf bytes.Buffer
	tw := NewTextWriter(&buf)
	rb := newRecordBuilder(testRecordTime, LevelInfo, "m", 1)
	rb.AddAttr(Group("outer", Int("n", 42), Group("inner", String("k", "v"))))
	r := rb.Build()
	if err := tw.Write(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	line := buf.String()
	if !strings.Contains(line, "outer.n=42") || !strings.Contains(line, "outer.inner.k=v") {
		t.Fatalf("want prefixed group attrs: %q", line)
	}
}
