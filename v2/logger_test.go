package golog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testWriter struct {
	writeCount   int
	lastAttrKeys []string
	lastAttrs    []Attr
	lastMessage  string
	lastLevel    Level
	lastCtx      context.Context
	captureAttrs bool
}

func (w *testWriter) Write(ctx context.Context, record Record) error {
	w.writeCount++
	w.lastMessage = record.Message
	w.lastLevel = record.Level
	w.lastCtx = ctx
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

func TestLogger_minLevel(t *testing.T) {
	type Args struct {
		minLevel Level
		emit     func(Logger)
	}
	type Expects struct {
		wantWriteCount int
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{
			name: "debug-suppressed-when-min-info",
			args: Args{
				minLevel: LevelInfo,
				emit:     func(log Logger) { log.Debug("hello") },
			},
			expects: Expects{wantWriteCount: 0},
		},
		{
			name: "info-emitted-when-min-info",
			args: Args{
				minLevel: LevelInfo,
				emit:     func(log Logger) { log.Info("x") },
			},
			expects: Expects{wantWriteCount: 1},
		},
		{
			name: "debug-emitted-when-min-debug",
			args: Args{
				minLevel: LevelDebug,
				emit:     func(log Logger) { log.Debug("d") },
			},
			expects: Expects{wantWriteCount: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := &testWriter{}
			log := NewLoggerWriter(w, tc.args.minLevel)
			tc.args.emit(log)
			assert.Equal(t, tc.expects.wantWriteCount, w.writeCount)
		})
	}
}

func TestLogger_SetLevel(t *testing.T) {
	w := &testWriter{}
	log := NewLoggerWriter(w, LevelInfo)
	log.Debug("skip")
	assert.Equal(t, 0, w.writeCount)

	log.SetLevel(LevelDebug)
	log.Debug("x")

	assert.Equal(t, 1, w.writeCount)
	assert.Equal(t, LevelDebug, w.lastLevel)
	assert.Equal(t, "x", w.lastMessage)
}

func TestLogger_nilWriter_noPanic(t *testing.T) {
	log := NewLoggerWriter(nil, LevelDebug)
	require.NotPanics(t, func() { log.Info("m") })
}

func TestLogger_levels_emitSeverity(t *testing.T) {
	type Args struct {
		emit func(Logger)
	}
	type Expects struct {
		want Level
		msg  string
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{
			name: "debug-level",
			args: Args{emit: func(log Logger) { log.Debug("d") }},
			expects: Expects{
				want: LevelDebug,
				msg:  "d",
			},
		},
		{
			name: "info-level",
			args: Args{emit: func(log Logger) { log.Info("i") }},
			expects: Expects{
				want: LevelInfo,
				msg:  "i",
			},
		},
		{
			name: "error-level",
			args: Args{emit: func(log Logger) { log.Error("e") }},
			expects: Expects{
				want: LevelError,
				msg:  "e",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := &testWriter{}
			log := NewLoggerWriter(w, LevelDebug)
			tc.args.emit(log)
			assert.Equal(t, tc.expects, Expects{want: w.lastLevel, msg: w.lastMessage})
		})
	}
}

func TestLogger_enrichers_runInOrder(t *testing.T) {
	var order []string
	e1 := EnricherFunc(func(ctx context.Context, r *RecordBuilder) {
		order = append(order, "e1")
		r.AddAttr(String("a", "1"))
	})
	e2 := EnricherFunc(func(ctx context.Context, r *RecordBuilder) {
		order = append(order, "e2")
		r.AddAttr(String("b", "2"))
	})

	w := &testWriter{captureAttrs: true}
	log := NewLoggerWriter(w, LevelDebug, e1, e2)
	log.Info("msg", String("call", "site"))

	type Expects struct {
		wantOrder  []string
		wantAttrKs []string
	}
	assert.Equal(t, Expects{
		wantOrder:  []string{"e1", "e2"},
		wantAttrKs: []string{"call", "a", "b"},
	}, Expects{wantOrder: order, wantAttrKs: w.lastAttrKeys})
}

func TestLogger_With_attrOrder_scopedCallSiteEnricher(t *testing.T) {
	w := &testWriter{captureAttrs: true}
	en := EnricherFunc(func(ctx context.Context, r *RecordBuilder) {
		r.AddAttr(String("enriched", "x"))
	})
	log := NewLoggerWriter(w, LevelDebug, en).With(String("scoped", "s"))
	log.Info("m", String("call", "c"))
	assert.Equal(t, []string{"scoped", "call", "enriched"}, w.lastAttrKeys)
}

func TestLogger_WithContext_passesContextToWriter(t *testing.T) {
	type ctxKey struct{}

	type Args struct {
		ctx context.Context
	}
	type Expects struct {
		wantWriteCount int
		wantValue      any
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{
			name: "writer-receives-context-value",
			args: Args{
				ctx: context.WithValue(context.Background(), ctxKey{}, 42),
			},
			expects: Expects{
				wantWriteCount: 1,
				wantValue:      42,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := &testWriter{}
			log := NewLoggerWriter(w, LevelDebug).WithContext(tc.args.ctx)
			log.Info("x")
			assert.Equal(t, tc.expects.wantWriteCount, w.writeCount)
			assert.Equal(t, tc.expects.wantValue, w.lastCtx.Value(ctxKey{}))
		})
	}
}

func TestLogger_WithError(t *testing.T) {
	type Args struct {
		err error
	}
	type Expects struct {
		wantAttrKeys []string
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{
			name:    "nil-error-is-no-op",
			args:    Args{err: nil},
			expects: Expects{wantAttrKeys: nil},
		},
		{
			name:    "non-nil-adds-error-attr",
			args:    Args{err: context.Canceled},
			expects: Expects{wantAttrKeys: []string{"error"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := &testWriter{captureAttrs: true}
			base := NewLoggerWriter(w, LevelDebug)
			log := base.WithError(tc.args.err)
			log.Info("x")
			assert.Equal(t, tc.expects.wantAttrKeys, w.lastAttrKeys)
		})
	}
}
