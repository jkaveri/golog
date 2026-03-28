package golog

import (
	"context"
	"strings"
	"testing"
)

func sourceAttrForTest(opts SourceEnricherOptions) (Attr, bool) {
	e, _ := NewSourceEnricher(opts).(sourceEnricher)
	return e.sourceAttr()
}

func TestSourceEnricher_defaultFieldName(t *testing.T) {
	attr, ok := sourceAttrForTest(SourceEnricherOptions{})
	if !ok {
		t.Fatal("want source attr")
	}
	if attr.Key != "source" {
		t.Fatalf("want key source, got %q", attr.Key)
	}
}

func TestSourceEnricher_functionFileLineFormat(t *testing.T) {
	attr, ok := sourceAttrForTest(SourceEnricherOptions{})
	if !ok {
		t.Fatal("want source attr")
	}
	s := attr.Value.String()
	if !strings.Contains(s, " ") {
		t.Fatalf("want function and file separator, got %q", s)
	}
	if !strings.Contains(s, ".go:") {
		t.Fatalf("want file:line in source, got %q", s)
	}
}

func TestSourceEnricher_fileLineOnlyFormat(t *testing.T) {
	attr, ok := sourceAttrForTest(SourceEnricherOptions{Format: SourceFormatFileLine})
	if !ok {
		t.Fatal("want source attr")
	}
	s := attr.Value.String()
	if strings.Contains(s, " ") {
		t.Fatalf("want file:line only, no function prefix, got %q", s)
	}
	if !strings.Contains(s, ".go:") {
		t.Fatalf("want basename:line, got %q", s)
	}
}

func TestSourceEnricher_customFieldName(t *testing.T) {
	attr, ok := sourceAttrForTest(SourceEnricherOptions{FieldName: "caller"})
	if !ok {
		t.Fatal("want source attr")
	}
	if attr.Key != "caller" {
		t.Fatalf("want key caller, got %q", attr.Key)
	}
}

func TestSourceEnricher_skipChangesFrame(t *testing.T) {
	a1, ok1 := sourceAttrForTest(SourceEnricherOptions{})
	a2, ok2 := sourceAttrForTest(SourceEnricherOptions{Skip: 1})
	if !ok1 {
		t.Fatalf("want base attr")
	}
	if !ok2 {
		t.Skip("shifted frame not available on this runtime stack")
	}
	if a1.Value.String() == a2.Value.String() {
		t.Fatalf("want different source values with skip change, both=%q", a1.Value.String())
	}
}

func TestSourceEnricher_deepSkipReturnsNoAttr(t *testing.T) {
	if _, ok := sourceAttrForTest(SourceEnricherOptions{Skip: 1 << 20}); ok {
		t.Fatal("want no attr for deep skip")
	}
}

func TestNewSourceEnricher_returnsEnricher(t *testing.T) {
	e := NewSourceEnricher(SourceEnricherOptions{})
	if e == nil {
		t.Fatal("want non-nil enricher")
	}
}

type captureWriter struct {
	attrs []Attr
}

func (w *captureWriter) Write(ctx context.Context, record Record) error {
	w.attrs = w.attrs[:0]
	record.RangeAttrs(func(a Attr) bool {
		w.attrs = append(w.attrs, a)
		return true
	})
	return nil
}

func TestSourceEnricher_integrationWithLogger(t *testing.T) {
	w := &captureWriter{}
	log := NewLoggerWriter(w, LevelDebug, NewSourceEnricher(SourceEnricherOptions{}))
	log.Info("m")

	found := false
	for _, a := range w.attrs {
		if a.Key == "source" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("source attr not found: %#v", w.attrs)
	}
}
