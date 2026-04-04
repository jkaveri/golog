package golog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func sourceAttrForTest(opts SourceEnricherOptions) (Attr, bool) {
	e, _ := NewSourceEnricher(opts).(sourceEnricher)
	return e.sourceAttr()
}

func TestSourceEnricher_defaultFieldName(t *testing.T) {
	attr, ok := sourceAttrForTest(SourceEnricherOptions{})
	require.True(t, ok, "want source attr")
	require.Equal(t, "source", attr.Key)
}

func TestSourceEnricher_functionFileLineFormat(t *testing.T) {
	attr, ok := sourceAttrForTest(SourceEnricherOptions{})
	require.True(t, ok, "want source attr")
	s := attr.Value.String()
	require.Contains(t, s, " ")
	require.Contains(t, s, ".go:")
}

func TestSourceEnricher_fileLineOnlyFormat(t *testing.T) {
	attr, ok := sourceAttrForTest(SourceEnricherOptions{Format: SourceFormatFileLine})
	require.True(t, ok, "want source attr")
	s := attr.Value.String()
	require.NotContains(t, s, " ")
	require.Contains(t, s, ".go:")
}

func TestSourceEnricher_customFieldName(t *testing.T) {
	attr, ok := sourceAttrForTest(SourceEnricherOptions{FieldName: "caller"})
	require.True(t, ok, "want source attr")
	require.Equal(t, "caller", attr.Key)
}

func TestSourceEnricher_skipChangesFrame(t *testing.T) {
	a1, ok1 := sourceAttrForTest(SourceEnricherOptions{})
	a2, ok2 := sourceAttrForTest(SourceEnricherOptions{Skip: 1})
	require.True(t, ok1, "want base attr")
	if !ok2 {
		t.Skip("shifted frame not available on this runtime stack")
	}
	require.NotEqual(t, a1.Value.String(), a2.Value.String(), "want different source values with skip change")
}

func TestSourceEnricher_deepSkipReturnsNoAttr(t *testing.T) {
	_, ok := sourceAttrForTest(SourceEnricherOptions{Skip: 1 << 20})
	require.False(t, ok, "want no attr for deep skip")
}

func TestNewSourceEnricher_returnsEnricher(t *testing.T) {
	e := NewSourceEnricher(SourceEnricherOptions{})
	require.NotNil(t, e)
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
	require.True(t, found, "source attr not found: %#v", w.attrs)
}
