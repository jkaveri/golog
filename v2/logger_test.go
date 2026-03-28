package golog

import (
	"context"
	"slices"
	"testing"
)

type testWriter struct {
	writeCount   int
	lastAttrKeys []string
	lastAttrs    []Attr
	lastMessage  string
	lastLevel    Level
	captureAttrs bool
}

func (w *testWriter) Write(ctx context.Context, record Record) error {
	w.writeCount++
	w.lastMessage = record.Message
	w.lastLevel = record.Level
	if w.captureAttrs {
		w.lastAttrKeys = w.lastAttrKeys[:0]
		w.lastAttrs = w.lastAttrs[:0]
		record.RangeAttrs(func(a Attr) bool {
			w.lastAttrKeys = append(w.lastAttrKeys, a.Key)
			w.lastAttrs = append(w.lastAttrs, a)
			return true
		})
	}
	return nil
}

func TestLogger_belowMinLevel_noWrite(t *testing.T) {
	w := &testWriter{}
	log := NewLoggerWriter(w, LevelInfo)
	log.Debug("hello", String("k", "v"))
	if w.writeCount != 0 {
		t.Fatalf("Write called %d times, want 0", w.writeCount)
	}
}

func TestLogger_nilWriter_noWrite(t *testing.T) {
	log := NewLoggerWriter(nil, LevelDebug)
	log.Info("m") // must not panic
}

func TestLogger_minLevelAllowsInfo(t *testing.T) {
	w := &testWriter{}
	log := NewLoggerWriter(w, LevelInfo)
	log.Debug("skip")
	if w.writeCount != 0 {
		t.Fatal("Debug below MinLevel")
	}
	log.Info("x")
	if w.writeCount != 1 {
		t.Fatalf("writeCount=%d, want 1", w.writeCount)
	}
}

func TestLogger_SetLevel(t *testing.T) {
	w := &testWriter{}
	log := NewLoggerWriter(w, LevelInfo)
	log.Debug("skip")
	if w.writeCount != 0 {
		t.Fatal("Debug should be skipped when MinLevel is Info")
	}
	log.SetLevel(LevelDebug)
	log.Debug("x")
	if w.writeCount != 1 || w.lastLevel != LevelDebug {
		t.Fatalf("after SetLevel: count=%d level=%v msg=%q", w.writeCount, w.lastLevel, w.lastMessage)
	}
}

func TestLogger_enrichersRunInOrder(t *testing.T) {
	w := &testWriter{captureAttrs: true}
	var order []string
	e1 := EnricherFunc(func(ctx context.Context, r *RecordBuilder) {
		order = append(order, "e1")
		r.AddAttr(String("a", "1"))
	})
	e2 := EnricherFunc(func(ctx context.Context, r *RecordBuilder) {
		order = append(order, "e2")
		r.AddAttr(String("b", "2"))
	})
	log := NewLoggerWriter(w, LevelDebug, e1, e2)
	log.Info("msg", String("call", "site"))
	if !slices.Equal(order, []string{"e1", "e2"}) {
		t.Fatalf("enricher order: %v", order)
	}
	wantKeys := []string{"call", "a", "b"}
	if !slices.Equal(w.lastAttrKeys, wantKeys) {
		t.Fatalf("attr keys: %v, want %v", w.lastAttrKeys, wantKeys)
	}
}

func TestLogger_attrOrder_withThenCallThenEnricher(t *testing.T) {
	w := &testWriter{captureAttrs: true}
	en := EnricherFunc(func(ctx context.Context, r *RecordBuilder) {
		r.AddAttr(String("enriched", "x"))
	})
	log := NewLoggerWriter(w, LevelDebug, en).With(String("scoped", "s"))

	log.Info("m", String("call", "c"))
	want := []string{"scoped", "call", "enriched"}
	if !slices.Equal(w.lastAttrKeys, want) {
		t.Fatalf("keys %v, want %v", w.lastAttrKeys, want)
	}
}

func TestLogger_levels(t *testing.T) {
	type Args struct {
		emit func(Logger)
	}
	type Expects struct {
		wantLevel Level
		wantMsg   string
	}
	type Deps struct {
		w *testWriter
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{name: "debug", args: Args{emit: func(log Logger) { log.Debug("d") }}, expects: Expects{wantLevel: LevelDebug, wantMsg: "d"}},
		{name: "info", args: Args{emit: func(log Logger) { log.Info("i") }}, expects: Expects{wantLevel: LevelInfo, wantMsg: "i"}},
		{name: "error", args: Args{emit: func(log Logger) { log.Error("e") }}, expects: Expects{wantLevel: LevelError, wantMsg: "e"}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := &testWriter{}
			deps := Deps{w: w}
			log := NewLoggerWriter(deps.w, LevelDebug)
			tc.args.emit(log)
			if deps.w.lastLevel != tc.expects.wantLevel || deps.w.lastMessage != tc.expects.wantMsg {
				t.Fatalf("level=%v msg=%q want level=%v msg=%q", deps.w.lastLevel, deps.w.lastMessage, tc.expects.wantLevel, tc.expects.wantMsg)
			}
		})
	}
}

func TestLogger_WithContext_passedToWriter(t *testing.T) {
	type ctxKey struct{}
	w := &testWriter{}
	log := NewLoggerWriter(w, LevelDebug).WithContext(context.WithValue(context.Background(), ctxKey{}, 42))
	log.Info("x")
	// testWriter doesn't inspect ctx; smoke only
	if w.writeCount != 1 {
		t.Fatalf("writes=%d", w.writeCount)
	}
}

func TestLogger_WithError_nilNoOp(t *testing.T) {
	w := &testWriter{captureAttrs: true}
	base := NewLoggerWriter(w, LevelDebug)
	_ = base.WithError(nil)
	base.Info("x")
	for _, k := range w.lastAttrKeys {
		if k == "error" {
			t.Fatal("unexpected error attr")
		}
	}
}

func TestLogger_WithError_addsAttr(t *testing.T) {
	w := &testWriter{captureAttrs: true}
	log := NewLoggerWriter(w, LevelDebug).WithError(context.Canceled)
	log.Info("x")
	found := false
	for _, k := range w.lastAttrKeys {
		if k == "error" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("keys %v, want error", w.lastAttrKeys)
	}
}
