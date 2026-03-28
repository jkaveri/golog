package golog

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

// testRecordTime is used when the test does not assert on the timestamp.
var testRecordTime = time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)

func TestJSONWriter_Write_nilOut(t *testing.T) {
	jw := &JSONWriter{Out: nil}
	r := newRecordBuilder(testRecordTime, LevelInfo, "x", 0).Build()
	err := jw.Write(context.Background(), r)
	if err == nil {
		t.Fatal("want error for nil Out")
	}
}

func TestJSONWriter_Write_format(t *testing.T) {
	var buf bytes.Buffer
	jw := NewJSONWriter(&buf)
	ts := time.Date(2025, 3, 22, 12, 30, 0, 0, time.UTC)
	rb := newRecordBuilder(ts, LevelInfo, "hello world", 2)
	rb.AddAttrs(String("k1", "v1"), Int("n", 42))
	r := rb.Build()

	if err := jw.Write(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	line := buf.String()
	if !strings.HasSuffix(line, "\n") {
		t.Fatalf("want trailing newline: %q", line)
	}

	var obj map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &obj); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got := obj["level"]; got != "INFO" {
		t.Fatalf("want level INFO, got %#v", got)
	}
	if got := obj["msg"]; got != "hello world" {
		t.Fatalf("want msg hello world, got %#v", got)
	}
	if got := obj["time"]; got != ts.Format(time.RFC3339Nano) {
		t.Fatalf("want time %q, got %#v", ts.Format(time.RFC3339Nano), got)
	}
	if got := obj["k1"]; got != "v1" {
		t.Fatalf("want k1 v1, got %#v", got)
	}
	if got := obj["n"]; got != float64(42) {
		t.Fatalf("want n 42, got %#v", got)
	}
}

func TestJSONWriter_groupAttrNested(t *testing.T) {
	var buf bytes.Buffer
	jw := NewJSONWriter(&buf)
	rb := newRecordBuilder(testRecordTime, LevelInfo, "m", 1)
	rb.AddAttr(Group("request", String("id", "r1"), Group("user", Int("id", 42))))
	r := rb.Build()

	if err := jw.Write(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	line := strings.TrimSpace(buf.String())

	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	request, ok := obj["request"].(map[string]any)
	if !ok {
		t.Fatalf("want request object, got %#v", obj["request"])
	}
	if request["id"] != "r1" {
		t.Fatalf("want request.id r1, got %#v", request["id"])
	}
	user, ok := request["user"].(map[string]any)
	if !ok {
		t.Fatalf("want request.user object, got %#v", request["user"])
	}
	if user["id"] != float64(42) {
		t.Fatalf("want request.user.id 42, got %#v", user["id"])
	}
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

	if err := jw.Write(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	line := strings.TrimSpace(buf.String())

	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if obj["msg"] != "original" || obj["level"] != "INFO" || obj["time"] != r.Time.Format(time.RFC3339Nano) {
		t.Fatalf("reserved fields must remain canonical: %#v", obj)
	}
	attrs, ok := obj["attrs"].(map[string]any)
	if !ok {
		t.Fatalf("want attrs bucket, got %#v", obj["attrs"])
	}
	if attrs["msg"] != "attr-msg" || attrs["level"] != "attr-level" || attrs["time"] != "attr-time" {
		t.Fatalf("want conflicted attrs under attrs bucket, got %#v", attrs)
	}
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
	lines := strings.Count(buf.String(), "\n")
	if lines != 8 {
		t.Fatalf("want 8 lines, got %d", lines)
	}
}
